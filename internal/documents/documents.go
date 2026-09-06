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
	var fenceCharacter byte
	fenceLength := 0

	for _, lineWithNewline := range lines {
		line := strings.TrimSuffix(lineWithNewline, "\n")
		line = strings.TrimSuffix(line, "\r")
		candidate, validIndent := markdownBlockCandidate(line)

		if fenceCharacter != 0 {
			if validIndent && closesFence(candidate, fenceCharacter, fenceLength) {
				fenceCharacter = 0
				fenceLength = 0
			}
			runeOffset += len([]rune(lineWithNewline))
			continue
		}
		if !validIndent {
			runeOffset += len([]rune(lineWithNewline))
			continue
		}
		if character, length, ok := opensFence(candidate); ok {
			fenceCharacter = character
			fenceLength = length
			runeOffset += len([]rune(lineWithNewline))
			continue
		}

		level := 0
		for level < len(candidate) && level < 6 && candidate[level] == '#' {
			level++
		}
		if level > 0 && (len(candidate) == level || candidate[level] == ' ' || candidate[level] == '\t') {
			title := strings.TrimSpace(candidate[level:])
			title = trimClosingHeadingSequence(title)
			title = strings.TrimSpace(strings.TrimPrefix(title, "~"))
			if title != "" {
				result = append(result, heading{title: title, level: level, startRune: runeOffset})
			}
		}
		runeOffset += len([]rune(lineWithNewline))
	}
	return result
}

func markdownBlockCandidate(line string) (string, bool) {
	indent := 0
	for indent < len(line) && line[indent] == ' ' {
		indent++
	}
	if indent > 3 || (indent < len(line) && line[indent] == '\t') {
		return "", false
	}
	return line[indent:], true
}

func opensFence(line string) (byte, int, bool) {
	if len(line) < 3 || (line[0] != '`' && line[0] != '~') {
		return 0, 0, false
	}
	character := line[0]
	length := 1
	for length < len(line) && line[length] == character {
		length++
	}
	if length < 3 || (character == '`' && strings.ContainsRune(line[length:], '`')) {
		return 0, 0, false
	}
	return character, length, true
}

func closesFence(line string, character byte, minimumLength int) bool {
	length := 0
	for length < len(line) && line[length] == character {
		length++
	}
	return length >= minimumLength && strings.TrimSpace(line[length:]) == ""
}

func trimClosingHeadingSequence(title string) string {
	end := len(title)
	for end > 0 && title[end-1] == '#' {
		end--
	}
	if end == len(title) || end == 0 || (title[end-1] != ' ' && title[end-1] != '\t') {
		return title
	}
	return strings.TrimSpace(title[:end])
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
