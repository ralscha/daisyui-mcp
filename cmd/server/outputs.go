package main

import (
	"fmt"
	"sort"
	"strings"

	"daisyui-mcp/internal/documents"
	"daisyui-mcp/internal/theme"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type ComponentSummaryOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URI         string `json:"uri"`
}

type ComponentListOutput struct {
	Components []ComponentSummaryOutput `json:"components"`
	Count      int                      `json:"count"`
}

type DocumentOutput struct {
	Name          string   `json:"name"`
	URI           string   `json:"uri"`
	Documentation string   `json:"documentation"`
	Warnings      []string `json:"warnings"`
}

type ComponentDocumentOutput struct {
	Name              string   `json:"name"`
	URI               string   `json:"uri"`
	Documentation     string   `json:"documentation"`
	Section           string   `json:"section,omitempty"`
	AvailableSections []string `json:"available_sections"`
	Page              int      `json:"page"`
	PageSize          int      `json:"page_size"`
	TotalPages        int      `json:"total_pages"`
	TotalCharacters   int      `json:"total_characters"`
	HasMore           bool     `json:"has_more"`
	Warnings          []string `json:"warnings"`
}

type ColorPaletteOutput struct {
	Documentation  string   `json:"documentation"`
	SemanticColors []string `json:"semantic_colors"`
	Modifiers      []string `json:"modifier_patterns"`
	Warnings       []string `json:"warnings"`
}

type ThemeOutput struct {
	CSS      string                 `json:"css"`
	Colors   []theme.GeneratedColor `json:"colors"`
	Warnings []string               `json:"warnings"`
}

type ExtractedColorsOutput struct {
	Primary   string `json:"primary,omitempty"`
	Secondary string `json:"secondary,omitempty"`
	Accent    string `json:"accent,omitempty"`
	Neutral   string `json:"neutral,omitempty"`
}

type ImageThemeOutput struct {
	Source          string                 `json:"source"`
	ExtractedColors ExtractedColorsOutput  `json:"extracted_colors"`
	CSS             string                 `json:"css"`
	Colors          []theme.GeneratedColor `json:"colors"`
	Warnings        []string               `json:"warnings"`
}

type RecipeListOutput struct {
	Recipes []ComponentSummaryOutput `json:"recipes"`
	Count   int                      `json:"count"`
}

func componentSummaries(index map[string]string, resourceType string) []ComponentSummaryOutput {
	names := make([]string, 0, len(index))
	for name := range index {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]ComponentSummaryOutput, 0, len(names))
	for _, name := range names {
		result = append(result, ComponentSummaryOutput{
			Name:        name,
			Description: index[name],
			URI:         fmt.Sprintf("daisyui://%s/%s", resourceType, name),
		})
	}
	return result
}

func detailedDocumentOutput(name string, page documents.Page) ComponentDocumentOutput {
	return ComponentDocumentOutput{
		Name:              name,
		URI:               "daisyui://components/" + name,
		Documentation:     page.Content,
		Section:           page.Section,
		AvailableSections: page.AvailableSections,
		Page:              page.Page,
		PageSize:          page.PageSize,
		TotalPages:        page.TotalPages,
		TotalCharacters:   page.TotalCharacters,
		HasMore:           page.HasMore,
		Warnings:          []string{},
	}
}

func formatDetailedText(output ComponentDocumentOutput) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "Component: %s\n", output.Name)
	if output.Section != "" {
		fmt.Fprintf(&sb, "Section: %s\n", output.Section)
	}
	fmt.Fprintf(&sb, "Page %d of %d (%d characters total)\n\n", output.Page, output.TotalPages, output.TotalCharacters)
	sb.WriteString(output.Documentation)
	if output.HasMore {
		if output.Section != "" {
			fmt.Fprintf(&sb, "\n\n---\nMore content is available. Call get_detailed_doc with name=%q, section=%q, page=%d, and page_size=%d.", output.Name, output.Section, output.Page+1, output.PageSize)
		} else {
			fmt.Fprintf(&sb, "\n\n---\nMore content is available. Call get_detailed_doc with name=%q, page=%d, and page_size=%d.", output.Name, output.Page+1, output.PageSize)
		}
	}
	return sb.String()
}

func themeOutput(generated theme.GeneratedTheme) ThemeOutput {
	return ThemeOutput{CSS: generated.CSS, Colors: generated.Colors, Warnings: generated.Warnings}
}

func guideToolResult(name string, content []byte) (*mcp.CallToolResult, DocumentOutput, error) {
	text := string(content)
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, DocumentOutput{
		Name:          name,
		URI:           "daisyui://guides/" + name,
		Documentation: text,
		Warnings:      []string{},
	}, nil
}
