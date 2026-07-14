package documents

import (
	"fmt"
	"strings"
)

const (
	DefaultPageSize = 12_000
	MaxPageSize     = 24_000
)

type Page struct {
	Content           string
	Section           string
	AvailableSections []string
	Page              int
	PageSize          int
	TotalPages        int
	TotalCharacters   int
	HasMore           bool
}

type heading struct {
	title     string
	level     int
	startRune int
}

func Paginate(content, requestedSection string, page, pageSize int) (Page, error) {
	if page < 0 {
		return Page{}, fmt.Errorf("page must be 1 or greater")
	}
	if page == 0 {
		page = 1
	}
	if pageSize < 0 {
		return Page{}, fmt.Errorf("page_size must be greater than 0")
	}
	if pageSize == 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		return Page{}, fmt.Errorf("page_size must not exceed %d characters", MaxPageSize)
	}

	headings := markdownHeadings(content)
	available := make([]string, 0, len(headings))
	for _, h := range headings {
		available = append(available, h.title)
	}

	selected := content
	section := ""
	if strings.TrimSpace(requestedSection) != "" {
		match := findHeading(headings, requestedSection)
		if match < 0 {
			return Page{}, fmt.Errorf("section %q not found; available sections: %s", requestedSection, strings.Join(available, ", "))
		}
		section = headings[match].title
		allRunes := []rune(content)
		end := len(allRunes)
		for i := match + 1; i < len(headings); i++ {
			if headings[i].level <= headings[match].level {
				end = headings[i].startRune
				break
			}
		}
		selected = strings.TrimSpace(string(allRunes[headings[match].startRune:end])) + "\n"
	}

	runes := []rune(selected)
	totalCharacters := len(runes)
	pages := splitPages(runes, pageSize)
	totalPages := len(pages)
	if page > totalPages {
		return Page{}, fmt.Errorf("page %d is out of range; document has %d page(s)", page, totalPages)
	}

	return Page{
		Content:           pages[page-1],
		Section:           section,
		AvailableSections: available,
		Page:              page,
		PageSize:          pageSize,
		TotalPages:        totalPages,
		TotalCharacters:   totalCharacters,
		HasMore:           page < totalPages,
	}, nil
}

func splitPages(content []rune, pageSize int) []string {
	if len(content) == 0 {
		return []string{""}
	}

	pages := make([]string, 0, (len(content)+pageSize-1)/pageSize)
	for start := 0; start < len(content); {
		end := min(start+pageSize, len(content))
		if end < len(content) {
			minimumBreak := start + pageSize/2
			for candidate := end; candidate > minimumBreak; candidate-- {
				if content[candidate-1] == '\n' {
					end = candidate
					break
				}
			}
		}
		pages = append(pages, string(content[start:end]))
		start = end
	}
	return pages
}

func markdownHeadings(content string) []heading {
	lines := strings.SplitAfter(content, "\n")
	result := make([]heading, 0)
	runeOffset := 0
	inFence := false

	for _, lineWithNewline := range lines {
		line := strings.TrimSuffix(lineWithNewline, "\n")
		trimmed := strings.TrimSpace(strings.TrimSuffix(line, "\r"))
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			runeOffset += len([]rune(lineWithNewline))
			continue
		}
		if !inFence {
			level := 0
			for level < len(trimmed) && level < 6 && trimmed[level] == '#' {
				level++
			}
			if level > 0 && len(trimmed) > level && trimmed[level] == ' ' {
				title := strings.TrimSpace(strings.TrimRight(trimmed[level+1:], "#"))
				title = strings.TrimSpace(strings.TrimPrefix(title, "~"))
				if title != "" {
					result = append(result, heading{title: title, level: level, startRune: runeOffset})
				}
			}
		}
		runeOffset += len([]rune(lineWithNewline))
	}
	return result
}

func findHeading(headings []heading, requested string) int {
	normalized := strings.ToLower(strings.TrimSpace(requested))
	for i, h := range headings {
		if strings.ToLower(h.title) == normalized {
			return i
		}
	}
	return -1
}
