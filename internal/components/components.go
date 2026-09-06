package components

import (
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
)

func FormatList(index map[string]string) string {
	if len(index) == 0 {
		return "No components found."
	}

	type entry struct {
		name string
		desc string
	}

	entries := make([]entry, 0, len(index))
	for name, desc := range index {
		entries = append(entries, entry{name, desc})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	var sb strings.Builder
	fmt.Fprintf(&sb, "Available DaisyUI Components (%d total):\n\n", len(entries))
	for _, e := range entries {
		fmt.Fprintf(&sb, "  • %s - %s\n", e.name, e.desc)
	}
	sb.WriteString("\nUse get_short_doc(name) for a concise summary, or get_detailed_doc(name) for the full documentation page.")

	return sb.String()
}

func LoadIndexFS(fsys fs.FS, dir string) map[string]string {
	index := make(map[string]string)

	entries, err := fs.Glob(fsys, dir+"/*.md")
	if err != nil || len(entries) == 0 {
		return index
	}

	for _, filePath := range entries {
		data, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			continue
		}

		lines := strings.Split(strings.TrimSpace(string(data)), "\n")
		name := strings.ToLower(strings.TrimSuffix(path.Base(filePath), ".md"))
		description := ""
		foundHeading := false

		for i, line := range lines {
			if strings.HasPrefix(line, "### ") {
				foundHeading = true
				for _, candidate := range lines[i+1:] {
					description = strings.TrimSpace(candidate)
					if description != "" {
						break
					}
				}
				break
			}
		}

		if foundHeading && name != "" {
			index[name] = description
		}
	}

	return index
}

func GetContentFS(fsys fs.FS, dir, name string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if strings.ContainsAny(normalized, "/\\") {
		return "", fmt.Errorf("invalid component name: %q", name)
	}
	path := dir + "/" + normalized + ".md"

	data, err := fs.ReadFile(fsys, path)
	if err != nil {
		return "", fmt.Errorf("component not found in embedded FS: %s", path)
	}
	return string(data), nil
}

func Suggestions(index map[string]string, name string) []string {
	type candidate struct {
		key   string
		score int
	}

	var results []candidate
	for k := range index {
		if s := suggestionScore(k, name); s > 0 {
			results = append(results, candidate{k, s})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].key < results[j].key
	})

	if len(results) > 5 {
		results = results[:5]
	}

	out := make([]string, len(results))
	for i, c := range results {
		out[i] = c.key
	}
	return out
}

// Search returns component names ordered by relevance to the query. Every query
// term must match either the component name or its short description.
func Search(index map[string]string, query string, limit int) []string {
	terms := strings.Fields(strings.ToLower(strings.TrimSpace(query)))
	if len(terms) == 0 || limit <= 0 {
		return []string{}
	}

	type match struct {
		name  string
		score int
	}
	matches := make([]match, 0)
	for name, description := range index {
		searchableDescription := strings.ToLower(description)
		score := 0
		matched := true
		for _, term := range terms {
			switch {
			case name == term:
				score += 100
			case strings.HasPrefix(name, term):
				score += 60
			case strings.Contains(name, term):
				score += 40
			case strings.Contains(searchableDescription, term):
				score += 20
			default:
				fuzzyScore := suggestionScore(name, term)
				if fuzzyScore == 0 {
					matched = false
					break
				}
				score += fuzzyScore / 2
			}
		}
		if matched {
			matches = append(matches, match{name: name, score: score})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].score != matches[j].score {
			return matches[i].score > matches[j].score
		}
		return matches[i].name < matches[j].name
	})
	if len(matches) > limit {
		matches = matches[:limit]
	}

	result := make([]string, len(matches))
	for i, match := range matches {
		result[i] = match.name
	}
	return result
}

func suggestionScore(key, name string) int {
	if key == name {
		return 100
	}
	if strings.HasPrefix(key, name) || strings.HasPrefix(name, key) {
		return 80
	}
	if strings.Contains(key, name) || strings.Contains(name, key) {
		return 60
	}
	switch levenshteinDistance(key, name) {
	case 1:
		return 50
	case 2:
		return 30
	}
	if ts := trigramJaccard(key, name); ts >= 0.4 {
		return int(ts * 20)
	}
	return 0
}

func levenshteinDistance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	la, lb := len(ra), len(rb)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			if ra[i-1] == rb[j-1] {
				curr[j] = prev[j-1]
			} else {
				curr[j] = 1 + min(prev[j], min(curr[j-1], prev[j-1]))
			}
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func trigramJaccard(a, b string) float64 {
	ta := buildTrigrams(a)
	tb := buildTrigrams(b)
	if len(ta) == 0 || len(tb) == 0 {
		return 0
	}
	intersection := 0
	for t := range ta {
		if tb[t] {
			intersection++
		}
	}
	union := len(ta) + len(tb) - intersection
	return float64(intersection) / float64(union)
}

func buildTrigrams(s string) map[string]bool {
	r := []rune(s)
	m := make(map[string]bool, max(0, len(r)-2))
	for i := 0; i+2 < len(r); i++ {
		m[string(r[i:i+3])] = true
	}
	return m
}
