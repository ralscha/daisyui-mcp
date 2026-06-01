package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const llmsURL = "https://daisyui.com/llms.txt"

func detailedDocURL(name string) string {
	return "https://raw.githubusercontent.com/saadeghi/daisyui/refs/heads/master/packages/docs/src/routes/(routes)/components/" + name + "/+page.md"
}

var guideDocs = []struct{ name, url string }{
	{"customize", "https://raw.githubusercontent.com/saadeghi/daisyui/refs/heads/master/packages/docs/src/routes/(routes)/docs/customize/+page.md?plain=1"},
	{"config", "https://raw.githubusercontent.com/saadeghi/daisyui/refs/heads/master/packages/docs/src/routes/(routes)/docs/config/+page.md?plain=1"},
	{"themes", "https://raw.githubusercontent.com/saadeghi/daisyui/refs/heads/master/packages/docs/src/routes/(routes)/docs/themes/+page.md?plain=1"},
	{"base", "https://raw.githubusercontent.com/saadeghi/daisyui/refs/heads/master/packages/docs/src/routes/(routes)/docs/base/+page.md?plain=1"},
	{"utilities", "https://raw.githubusercontent.com/saadeghi/daisyui/refs/heads/master/packages/docs/src/routes/(routes)/docs/utilities/+page.md?plain=1"},
	{"layout-and-typography", "https://raw.githubusercontent.com/saadeghi/daisyui/refs/heads/master/packages/docs/src/routes/(routes)/docs/layout-and-typography/+page.md?plain=1"},
}

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		fatalf("cannot determine working directory: %v", err)
	}

	tempDir, err := os.MkdirTemp(cwd, ".daisyui-update-*")
	if err != nil {
		fatalf("cannot create temporary update directory: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not remove temporary update directory %s: %v\n", tempDir, err)
		}
	}()

	componentsDir := filepath.Join(tempDir, "components")
	if err := os.MkdirAll(componentsDir, 0o755); err != nil {
		fatalf("cannot create components directory: %v", err)
	}

	docsDir := filepath.Join(tempDir, "docs")
	if err := os.MkdirAll(docsDir, 0o755); err != nil {
		fatalf("cannot create docs directory: %v", err)
	}

	guideDir := filepath.Join(tempDir, "guide")
	if err := os.MkdirAll(guideDir, 0o755); err != nil {
		fatalf("cannot create guide directory: %v", err)
	}

	colorsFilePath := filepath.Join(tempDir, "colors.md")

	fmt.Fprintf(os.Stderr, "Fetching %s...\n", llmsURL)

	req, err := http.NewRequest(http.MethodGet, llmsURL, nil)
	if err != nil {
		fatalf("cannot create request: %v", err)
	}
	req.Header.Set("User-Agent", "DaisyUI MCP Updater/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fatalf("error fetching URL: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fatalf("error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fatalf("unexpected HTTP status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fatalf("error reading response body: %v", err)
	}
	content := string(body)

	headerRe := regexp.MustCompile(`(?m)^### (.*)$`)
	loc := headerRe.FindAllStringIndex(content, -1)
	matches := headerRe.FindAllStringSubmatch(content, -1)

	fmt.Fprintf(os.Stderr, "Found %d potential sections.\n", len(matches))

	count := 0
	var componentNames []string
	for i, match := range matches {
		title := strings.TrimSpace(match[1])

		start := loc[i][0]
		end := len(content)
		if i+1 < len(loc) {
			end = loc[i+1][0]
		}

		section := content[start:end]

		if !strings.Contains(section, "#### Class names") && !strings.Contains(section, "#### Syntax") {
			continue
		}

		var sb strings.Builder
		for _, ch := range strings.ToLower(title) {
			if (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-' {
				sb.WriteRune(ch)
			}
		}
		safeName := sb.String()
		if safeName == "" {
			continue
		}

		filePath := filepath.Join(componentsDir, safeName+".md")
		fileContent := strings.TrimSpace(section) + "\n"

		if err := os.WriteFile(filePath, []byte(fileContent), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", filePath, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "Generated %s.md\n", safeName)
		componentNames = append(componentNames, safeName)
		count++
	}

	fmt.Fprintf(os.Stderr, "Successfully generated %d component files.\n", count)

	fmt.Fprintf(os.Stderr, "Downloading detailed documentation for %d components...\n", len(componentNames))
	docsCount := 0
	for _, name := range componentNames {
		url := detailedDocURL(name)
		data, dlErr := downloadFile(client, url)
		if dlErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not download detailed doc for %s: %v\n", name, dlErr)
			continue
		}
		docPath := filepath.Join(docsDir, name+".md")
		if writeErr := os.WriteFile(docPath, data, 0o644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", docPath, writeErr)
			continue
		}
		fmt.Fprintf(os.Stderr, "Downloaded detailed doc: %s.md\n", name)
		docsCount++
	}
	fmt.Fprintf(os.Stderr, "Successfully downloaded %d detailed documentation files.\n", docsCount)

	colorsGenerated := false
	sectionRe := regexp.MustCompile(`(?m)^## daisyUI 5 colors\s*$`)
	sectionLoc := sectionRe.FindStringIndex(content)
	if sectionLoc == nil {
		fmt.Fprintln(os.Stderr, "WARNING: '## daisyUI 5 colors' section not found in llms.txt")
	} else {
		nextSectionRe := regexp.MustCompile(`(?m)^## `)
		nextLoc := nextSectionRe.FindStringIndex(content[sectionLoc[1]:])
		var colorsSection string
		if nextLoc == nil {
			colorsSection = content[sectionLoc[0]:]
		} else {
			colorsSection = content[sectionLoc[0] : sectionLoc[1]+nextLoc[0]]
		}
		colorsSection = strings.TrimSpace(colorsSection) + "\n"
		if err := os.WriteFile(colorsFilePath, []byte(colorsSection), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "Error writing colors.md: %v\n", err)
		} else {
			fmt.Fprintln(os.Stderr, "Generated colors.md")
			colorsGenerated = true
		}
	}

	fmt.Fprintf(os.Stderr, "Downloading %d guide documentation pages...\n", len(guideDocs))
	guideCount := 0
	for _, doc := range guideDocs {
		data, dlErr := downloadFile(client, doc.url)
		if dlErr != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not download guide doc %s: %v\n", doc.name, dlErr)
			continue
		}
		docPath := filepath.Join(guideDir, doc.name+".md")
		if writeErr := os.WriteFile(docPath, data, 0o644); writeErr != nil {
			fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", docPath, writeErr)
			continue
		}
		fmt.Fprintf(os.Stderr, "Downloaded guide doc: %s.md\n", doc.name)
		guideCount++
	}
	fmt.Fprintf(os.Stderr, "Successfully downloaded %d guide documentation pages.\n", guideCount)

	if count == 0 {
		fatalf("no component files were generated")
	}
	if !colorsGenerated {
		fatalf("colors.md was not generated")
	}

	if err := replaceDir(filepath.Join(cwd, "components"), componentsDir); err != nil {
		fatalf("cannot replace components directory: %v", err)
	}
	if err := replaceDir(filepath.Join(cwd, "docs"), docsDir); err != nil {
		fatalf("cannot replace docs directory: %v", err)
	}
	if err := replaceDir(filepath.Join(cwd, "guide"), guideDir); err != nil {
		fatalf("cannot replace guide directory: %v", err)
	}
	if err := replaceFile(filepath.Join(cwd, "colors.md"), colorsFilePath); err != nil {
		fatalf("cannot replace colors.md: %v", err)
	}
	fmt.Fprintln(os.Stderr, "Replaced generated documentation files.")
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
	return io.ReadAll(resp.Body)
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}

func replaceDir(dst, src string) error {
	backup := backupPath(dst)
	if err := os.Rename(dst, backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("backup existing directory: %w", err)
	}
	if err := os.Rename(src, dst); err != nil {
		if restoreErr := os.Rename(backup, dst); restoreErr != nil && !os.IsNotExist(restoreErr) {
			return fmt.Errorf("replace directory: %w; restore failed: %v", err, restoreErr)
		}
		return fmt.Errorf("replace directory: %w", err)
	}
	if err := os.RemoveAll(backup); err != nil {
		return fmt.Errorf("remove backup: %w", err)
	}
	return nil
}

func replaceFile(dst, src string) error {
	backup := backupPath(dst)
	if err := os.Rename(dst, backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("backup existing file: %w", err)
	}
	if err := os.Rename(src, dst); err != nil {
		if restoreErr := os.Rename(backup, dst); restoreErr != nil && !os.IsNotExist(restoreErr) {
			return fmt.Errorf("replace file: %w; restore failed: %v", err, restoreErr)
		}
		return fmt.Errorf("replace file: %w", err)
	}
	if err := os.Remove(backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove backup: %w", err)
	}
	return nil
}

func backupPath(dst string) string {
	dir := filepath.Dir(dst)
	base := filepath.Base(dst)
	return filepath.Join(dir, fmt.Sprintf(".%s.bak-%d", base, time.Now().UnixNano()))
}
