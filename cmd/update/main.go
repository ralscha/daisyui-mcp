package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"
)

const (
	maxDocBytes     = 5 << 20
	downloadWorkers = 6
)

func sourceURL(version, path string) string {
	return "https://raw.githubusercontent.com/saadeghi/daisyui/refs/tags/v" + version + "/" + path
}

var componentDocLinkRe = regexp.MustCompile(`\[[^\]\r\n]+\]\(https://daisyui\.com/components/([a-z0-9]+(?:-[a-z0-9]+)*)/?(?:#[^)\s]*)?\)`)
var componentIndexLinkRe = regexp.MustCompile(`\[[^\]\r\n]+\]\(\./components/([^\s)]+)\.md\)`)
var componentNameRe = regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
var versionRe = regexp.MustCompile(`^5\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)$`)
var colorsHeadingRe = regexp.MustCompile(`(?m)^## daisyUI 5 colors[\t ]*\r?$`)

type documentDownload struct{ name, url string }

var guideNames = []string{"customize", "config", "themes", "base", "utilities", "layout-and-typography"}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine working directory: %w", err)
	}
	versionData, err := os.ReadFile(filepath.Join(cwd, "daisyui.version"))
	if err != nil {
		return fmt.Errorf("cannot read daisyui.version; run the updater from the project root: %w", err)
	}
	version, err := parseVersion(string(versionData))
	if err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp(cwd, ".daisyui-update-*")
	if err != nil {
		return fmt.Errorf("cannot create temporary update directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove temporary update directory %s: %v\n", tempDir, err)
		}
	}()

	componentsDir := filepath.Join(tempDir, "components")
	if err := os.MkdirAll(componentsDir, 0o755); err != nil {
		return fmt.Errorf("cannot create components directory: %w", err)
	}

	docsDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		return fmt.Errorf("cannot create docs directory: %w", err)
	}

	guideDir := filepath.Join(tempDir, "guide")
	if err := os.MkdirAll(guideDir, 0o755); err != nil {
		return fmt.Errorf("cannot create guide directory: %w", err)
	}

	colorsFilePath := filepath.Join(tempDir, "colors.md")

	fmt.Fprintf(os.Stderr, "Fetching documentation for daisyUI %s...\n", version)
	client := &http.Client{Timeout: 30 * time.Second}
	index, err := downloadFile(client, sourceURL(version, "skills/daisyui/SKILL.md"))
	if err != nil {
		return fmt.Errorf("error fetching component index: %w", err)
	}
	componentNames, err := componentNamesFromIndex(string(index))
	if err != nil {
		return fmt.Errorf("invalid component index: %w", err)
	}
	summaries := make([]documentDownload, 0, len(componentNames))
	componentDocs := make([]documentDownload, 0, len(componentNames))
	for _, name := range componentNames {
		summaries = append(summaries, documentDownload{name, sourceURL(version, "skills/daisyui/components/"+name+".md")})
		componentDocs = append(componentDocs, documentDownload{name, sourceURL(version, "packages/docs/src/routes/(routes)/components/"+name+"/+page.md")})
	}
	fmt.Fprintf(os.Stderr, "Downloading %d component summaries...\n", len(summaries))
	count := downloadDocuments(client, componentsDir, summaries)
	if count != len(componentNames) {
		return fmt.Errorf("downloaded %d of %d component summaries; existing files were left unchanged", count, len(componentNames))
	}
	componentGenerationFailed := false
	for _, name := range componentNames {
		data, err := os.ReadFile(filepath.Join(componentsDir, name+".md"))
		canonicalName, nameErr := componentNameFromSection(string(data))
		if err != nil || nameErr != nil || canonicalName != name {
			fmt.Fprintf(os.Stderr, "Invalid component summary for %s\n", name)
			componentGenerationFailed = true
		}
	}

	fmt.Fprintf(os.Stderr, "Downloading detailed documentation for %d components...\n", len(componentNames))
	docsCount := downloadDocuments(client, docsDir, componentDocs)
	fmt.Fprintf(os.Stderr, "Successfully downloaded %d detailed documentation files.\n", docsCount)

	colorsGenerated := false
	colorsData, err := downloadFile(client, sourceURL(version, "skills/daisyui/colors/SKILL.md"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not download colors: %v\n", err)
	} else {
		colorsData, err = extractColors(colorsData)
		if err == nil {
			err = os.WriteFile(colorsFilePath, colorsData, 0o644)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error writing colors.md: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "Generated colors.md")
			colorsGenerated = true
		}
	}

	guideDocs := make([]documentDownload, 0, len(guideNames))
	for _, name := range guideNames {
		guideDocs = append(guideDocs, documentDownload{name, sourceURL(version, "packages/docs/src/routes/(routes)/docs/"+name+"/+page.md")})
	}
	fmt.Fprintf(os.Stderr, "Downloading %d guide documentation pages...\n", len(guideDocs))
	guideCount := downloadDocuments(client, guideDir, guideDocs)
	fmt.Fprintf(os.Stderr, "Successfully downloaded %d guide documentation pages.\n", guideCount)

	if err := validateGeneratedDocumentation(count, componentGenerationFailed, docsCount, guideCount, len(guideDocs), colorsGenerated); err != nil {
		return fmt.Errorf("generated documentation is incomplete; existing files were left unchanged: %w", err)
	}

	if err := replaceGeneratedDocumentation(cwd, tempDir); err != nil {
		return fmt.Errorf("cannot replace generated documentation: %w", err)
	}
	fmt.Fprintln(os.Stderr, "Replaced generated documentation files.")
	return nil
}

func parseVersion(value string) (string, error) {
	version := strings.TrimSpace(value)
	if !versionRe.MatchString(version) {
		return "", fmt.Errorf("daisyui.version must contain a stable daisyUI 5 release such as 5.7.50")
	}
	return version, nil
}

func componentNamesFromIndex(index string) ([]string, error) {
	var names []string
	seen := make(map[string]bool)
	for _, match := range componentIndexLinkRe.FindAllStringSubmatch(index, -1) {
		name := match[1]
		if !componentNameRe.MatchString(name) {
			return nil, fmt.Errorf("unsafe component name %q", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("duplicate component %q", name)
		}
		seen[name] = true
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no component links found")
	}
	return names, nil
}

func extractColors(data []byte) ([]byte, error) {
	loc := colorsHeadingRe.FindIndex(data)
	if loc == nil {
		return nil, fmt.Errorf("daisyUI 5 colors heading not found")
	}
	return append(bytes.TrimSpace(data[loc[0]:]), '\n'), nil
}

func downloadFile(client *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "DaisyUI MCP Updater/1.0")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %s", resp.Status)
	}
	data, err := readLimited(resp.Body, maxDocBytes)
	if err != nil {
		return nil, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}
	return data, nil
}

// Fetch a bounded number of pages in parallel, then report results in input order.
func downloadDocuments(client *http.Client, dir string, documents []documentDownload) int {
	results := make([]error, len(documents))
	jobs := make(chan int)
	var workers sync.WaitGroup
	for range min(downloadWorkers, len(documents)) {
		workers.Go(func() {
			for i := range jobs {
				doc := documents[i]
				data, err := downloadFile(client, doc.url)
				if err == nil {
					err = os.WriteFile(filepath.Join(dir, doc.name+".md"), data, 0o644)
				}
				results[i] = err
			}
		})
	}
	for i := range documents {
		jobs <- i
	}
	close(jobs)
	workers.Wait()

	count := 0
	for i, err := range results {
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not save %s.md: %v\n", documents[i].name, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "Downloaded %s.md\n", documents[i].name)
		count++
	}
	return count
}

func componentNameFromSection(section string) (string, error) {
	match := componentDocLinkRe.FindStringSubmatch(section)
	if len(match) != 2 {
		return "", fmt.Errorf("canonical daisyUI component documentation link not found")
	}
	return match[1], nil
}

func validateGeneratedDocumentation(componentCount int, componentGenerationFailed bool, detailedCount, guideCount, expectedGuideCount int, colorsGenerated bool) error {
	if componentCount == 0 {
		return fmt.Errorf("no component files were generated")
	}
	if componentGenerationFailed {
		return fmt.Errorf("one or more component files could not be generated")
	}
	if detailedCount != componentCount {
		return fmt.Errorf("downloaded %d of %d detailed component documents", detailedCount, componentCount)
	}
	if guideCount != expectedGuideCount {
		return fmt.Errorf("downloaded %d of %d guide documents", guideCount, expectedGuideCount)
	}
	if !colorsGenerated {
		return fmt.Errorf("colors.md was not generated")
	}
	return nil
}

func readLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("response is larger than %d bytes", maxBytes)
	}
	return data, nil
}

// Keep every backup until the entire set is installed, so a failed rename does
// not leave component summaries, detailed docs, and guides from different runs.
func replaceGeneratedDocumentation(dstRoot, srcRoot string) error {
	type replacement struct {
		dst, src, backup    string
		backedUp, installed bool
	}
	var replacements []*replacement
	rollback := func(cause error) error {
		for _, r := range slices.Backward(replacements) {
			if r.installed {
				if err := os.Rename(r.dst, r.src); err != nil {
					cause = errors.Join(cause, fmt.Errorf("cannot undo replacement of %s; original retained at %s: %w", r.dst, r.backup, err))
					continue
				}
			}
			if r.backedUp {
				if err := os.Rename(r.backup, r.dst); err != nil {
					cause = errors.Join(cause, fmt.Errorf("cannot restore %s; original retained at %s: %w", r.dst, r.backup, err))
				}
			}
		}
		return cause
	}
	for _, name := range []string{"components", "docs", "guide", "colors.md"} {
		r := &replacement{dst: filepath.Join(dstRoot, name), src: filepath.Join(srcRoot, name)}
		r.backup = backupPath(r.dst)
		replacements = append(replacements, r)
		if err := os.Rename(r.dst, r.backup); err == nil {
			r.backedUp = true
		} else if !os.IsNotExist(err) {
			return rollback(fmt.Errorf("backup %s: %w", name, err))
		}
		if err := os.Rename(r.src, r.dst); err != nil {
			return rollback(fmt.Errorf("replace %s: %w", name, err))
		}
		r.installed = true
	}
	var cleanupErr error
	for _, r := range replacements {
		if r.backedUp {
			if err := os.RemoveAll(r.backup); err != nil {
				cleanupErr = errors.Join(cleanupErr, fmt.Errorf("documentation replaced, but cannot remove backup %s: %w", r.backup, err))
			}
		}
	}
	return cleanupErr
}

func backupPath(dst string) string {
	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	return filepath.Join(dir, fmt.Sprintf(".%s.bak-%d", base, time.Now().UnixNano()))
}
