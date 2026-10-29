package documents

import (
	"fmt"
	"strings"
	"unicode/utf8"
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
	startByte int
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
		end := len(content)
		for i := match + 1; i < len(headings); i++ {
			if headings[i].level <= headings[match].level {
				end = headings[i].startByte
				break
			}
		}
		selected = strings.TrimSpace(content[headings[match].startByte:end]) + "\n"
	}

	totalCharacters := utf8.RuneCountInString(selected)
	pageContent, totalPages := selectPage(selected, page, pageSize)
	if page > totalPages {
		return Page{}, fmt.Errorf("page %d is out of range; document has %d page(s)", page, totalPages)
	}

	return Page{
		Content:           pageContent,
		Section:           section,
		AvailableSections: available,
		Page:              page,
		PageSize:          pageSize,
		TotalPages:        totalPages,
		TotalCharacters:   totalCharacters,
		HasMore:           page < totalPages,
	}, nil
}

// selectPage counts pages without allocating copies of the complete document or
// every page. Byte offsets always follow complete runes, including for Unicode.
func selectPage(content string, requestedPage, pageSize int) (string, int) {
	if len(content) == 0 {
		return "", 1
	}

	var selected string
	totalPages := 0
	for start := 0; start < len(content); {
		end, lastBreak := start, 0
		for count := 1; count <= pageSize && end < len(content); count++ {
			r, width := utf8.DecodeRuneInString(content[end:])
			end += width
			if r == '\n' && count > pageSize/2 {
				lastBreak = end
			}
		}
		if end < len(content) && lastBreak > start {
			end = lastBreak
		}
		totalPages++
		if totalPages == requestedPage {
			selected = content[start:end]
		}
		start = end
	}
	return selected, totalPages
}

func markdownHeadings(content string) []heading {
	result := make([]heading, 0)
	byteOffset := 0
	var fenceCharacter byte
	fenceLength := 0

	for lineWithNewline := range strings.SplitAfterSeq(content, "\n") {
		line := strings.TrimSuffix(lineWithNewline, "\n")
		line = strings.TrimSuffix(line, "\r")
		candidate, validIndent := markdownBlockCandidate(line)

		if fenceCharacter != 0 {
			if validIndent && closesFence(candidate, fenceCharacter, fenceLength) {
				fenceCharacter = 0
				fenceLength = 0
			}
			byteOffset += len(lineWithNewline)
			continue
		}
		if !validIndent {
			byteOffset += len(lineWithNewline)
			continue
		}
		if character, length, ok := opensFence(candidate); ok {
			fenceCharacter = character
			fenceLength = length
			byteOffset += len(lineWithNewline)
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
				result = append(result, heading{title: title, level: level, startByte: byteOffset})
			}
		}
		byteOffset += len(lineWithNewline)
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
