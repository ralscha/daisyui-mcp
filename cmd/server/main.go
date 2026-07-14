package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"

	daisyuimcp "daisyui-mcp"
	"daisyui-mcp/internal/components"
	"daisyui-mcp/internal/documents"
	"daisyui-mcp/internal/theme"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func componentSource() (fs.FS, string) {
	if dir := os.Getenv("DAISYUI_COMPONENTS_DIR"); dir != "" {
		log.Printf("DaisyUI MCP Server: using components from DAISYUI_COMPONENTS_DIR=%s", dir)
		return os.DirFS(dir), "."
	}
	log.Printf("DaisyUI MCP Server: using embedded components")
	return daisyuimcp.ComponentsFS, "components"
}

func docsSource() (fs.FS, string) {
	if dir := os.Getenv("DAISYUI_DOCS_DIR"); dir != "" {
		log.Printf("DaisyUI MCP Server: using detailed docs from DAISYUI_DOCS_DIR=%s", dir)
		return os.DirFS(dir), "."
	}
	log.Printf("DaisyUI MCP Server: using embedded detailed docs")
	return daisyuimcp.DocsFS, "docs"
}

// Input types
type ListComponentsInput struct{}
type GetComponentInput struct {
	Name string `json:"name" jsonschema:"The name of the DaisyUI component (e.g. 'button', 'card', 'badge'). Case-insensitive."`
}
type GetDetailedDocInput struct {
	Name     string `json:"name" jsonschema:"The name of the DaisyUI component (e.g. 'button', 'card', 'badge'). Case-insensitive."`
	Section  string `json:"section,omitempty" jsonschema:"Optional exact Markdown section heading to return."`
	Page     int    `json:"page,omitempty" jsonschema:"1-based page number. Defaults to 1."`
	PageSize int    `json:"page_size,omitempty" jsonschema:"Maximum characters per page. Defaults to 12000 and cannot exceed 24000."`
}
type GetColorPaletteInput struct{}
type GetGuideInput struct{}
type GenerateThemeInput struct {
	Primary   string `json:"primary" jsonschema:"Primary color in hex or OKLCH format (e.g. '#ff0000' or 'oklch(60% 0.2 30)'). Required."`
	Secondary string `json:"secondary,omitempty" jsonschema:"Secondary color in hex or OKLCH format. Optional."`
	Accent    string `json:"accent,omitempty" jsonschema:"Accent color in hex or OKLCH format. Optional."`
	Neutral   string `json:"neutral,omitempty" jsonschema:"Neutral color in hex or OKLCH format. Optional."`
	Base100   string `json:"base_100,omitempty" jsonschema:"Base 100 color (background) in hex or OKLCH format. Optional."`
	Info      string `json:"info,omitempty" jsonschema:"Info color in hex or OKLCH format. Optional."`
	Success   string `json:"success,omitempty" jsonschema:"Success color in hex or OKLCH format. Optional."`
	Warning   string `json:"warning,omitempty" jsonschema:"Warning color in hex or OKLCH format. Optional."`
	Error     string `json:"error,omitempty" jsonschema:"Error color in hex or OKLCH format. Optional."`

	RadiusSelector string `json:"radius_selector,omitempty" jsonschema:"Radius selector. Optional."`
	RadiusField    string `json:"radius_field,omitempty" jsonschema:"Radius field. Optional."`
	RadiusBox      string `json:"radius_box,omitempty" jsonschema:"Radius box. Optional."`
	SizeSelector   string `json:"size_selector,omitempty" jsonschema:"Size selector. Optional."`
	SizeField      string `json:"size_field,omitempty" jsonschema:"Size field. Optional."`
	Border         string `json:"border,omitempty" jsonschema:"Border. Optional."`
	Depth          string `json:"depth,omitempty" jsonschema:"Depth. Optional."`
	Noise          string `json:"noise,omitempty" jsonschema:"Noise. Optional."`
}
type GenerateThemeFromImageInput struct {
	ImagePath string `json:"image_path,omitempty" jsonschema:"Local file path to the image. Either image_path or image_url must be provided."`
	ImageURL  string `json:"image_url,omitempty" jsonschema:"URL to the image. Either image_path or image_url must be provided."`
}

func newServer(fsys fs.FS, fsDir string, docsFS fs.FS, docsDir string) *mcp.Server {
	index := components.LoadIndexFS(fsys, fsDir)
	recipeIndex := components.LoadIndexFS(daisyuimcp.RecipesFS, "recipes")
	log.Printf("DaisyUI MCP Server: loaded %d components", len(index))

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "DaisyUI MCP Server",
			Version: "1.0.0",
		},
		nil,
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "list_components",
			Description: "List all available DaisyUI components. " +
				"Returns a formatted list with names and brief descriptions. " +
				"Use this to discover what components are available before calling get_short_doc() " +
				"or get_detailed_doc() for detailed documentation.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ ListComponentsInput,
		) (*mcp.CallToolResult, ComponentListOutput, error) {
			text := components.FormatList(index)
			summaries := componentSummaries(index, "components")
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: text},
				},
			}, ComponentListOutput{Components: summaries, Count: len(summaries)}, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_short_doc",
			Description: "Get a concise summary for a specific DaisyUI component. " +
				"Returns the component's CSS classes, HTML syntax examples, " +
				"and usage rules. " +
				"Use get_detailed_doc() for the full documentation page.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			input GetComponentInput,
		) (*mcp.CallToolResult, DocumentOutput, error) {
			name := strings.ToLower(strings.TrimSpace(input.Name))

			if _, ok := index[name]; !ok {
				suggestions := components.Suggestions(index, name)

				var text string
				if len(suggestions) > 0 {
					text = fmt.Sprintf(
						"Component '%s' not found.\n\nDid you mean one of these?\n  • %s\n\nUse list_components() to see all available components.",
						input.Name,
						strings.Join(suggestions, "\n  • "),
					)
				} else {
					text = fmt.Sprintf(
						"Component '%s' not found.\n\nUse list_components() to see all available components.",
						input.Name,
					)
				}

				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: text},
					},
				}, DocumentOutput{}, nil
			}

			content, err := components.GetContentFS(fsys, fsDir, name)
			if err != nil {
				log.Printf("error loading component %q: %v", name, err)
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Error: Unable to load documentation for '%s'. Please try again.", input.Name),
						},
					},
				}, DocumentOutput{}, nil
			}

			output := DocumentOutput{Name: name, URI: "daisyui://components/" + name, Documentation: content, Warnings: []string{}}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: content},
				},
			}, output, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_detailed_doc",
			Description: "Get detailed documentation for a specific DaisyUI component. " +
				"Returns a 12000-character page by default; use page/page_size to continue or section to select an exact Markdown heading. " +
				"Use the daisyui://components/{name} resource when the complete document is explicitly needed.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			input GetDetailedDocInput,
		) (*mcp.CallToolResult, ComponentDocumentOutput, error) {
			name := strings.ToLower(strings.TrimSpace(input.Name))

			content, err := components.GetContentFS(docsFS, docsDir, name)
			if err != nil {
				suggestions := components.Suggestions(index, name)
				var text string
				if len(suggestions) > 0 {
					text = fmt.Sprintf(
						"Detailed documentation for '%s' not found.\n\nDid you mean one of these?\n  • %s\n\nUse list_components() to see all available components.",
						input.Name,
						strings.Join(suggestions, "\n  • "),
					)
				} else {
					text = fmt.Sprintf(
						"Detailed documentation for '%s' not found.\n\nUse list_components() to see all available components.",
						input.Name,
					)
				}
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: text},
					},
				}, ComponentDocumentOutput{}, nil
			}

			page, err := documents.Paginate(content, input.Section, input.Page, input.PageSize)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: "Unable to select detailed documentation: " + err.Error()}},
				}, ComponentDocumentOutput{}, nil
			}
			output := detailedDocumentOutput(name, page)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: formatDetailedText(output)},
				},
			}, output, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_color_palette",
			Description: "List all daisyUI semantic colors (primary, secondary, accent, neutral, " +
				"base-100/200/300, info, success, warning, error and their *-content variants) " +
				"together with their modifier class patterns and usage rules. " +
				"Use this to understand how to apply colors in daisyUI components and custom themes.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ GetColorPaletteInput,
		) (*mcp.CallToolResult, ColorPaletteOutput, error) {
			output := ColorPaletteOutput{
				Documentation:  string(daisyuimcp.ColorsData),
				SemanticColors: []string{"base-100", "base-200", "base-300", "base-content", "primary", "primary-content", "secondary", "secondary-content", "accent", "accent-content", "neutral", "neutral-content", "info", "info-content", "success", "success-content", "warning", "warning-content", "error", "error-content"},
				Modifiers:      []string{"bg-{color}", "text-{color}", "border-{color}", "from-{color}", "via-{color}", "to-{color}"},
				Warnings:       []string{},
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.ColorsData)},
				},
			}, output, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_customize_docs",
			Description: "Get the DaisyUI documentation for customizing components. " +
				"Explains how to customize daisyUI component styles using CSS, Tailwind, and daisyUI conventions.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ GetGuideInput,
		) (*mcp.CallToolResult, DocumentOutput, error) {
			return guideToolResult("customize", daisyuimcp.GuideCustomize)
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_config_docs",
			Description: "Get the DaisyUI configuration reference. " +
				"Covers all daisyUI plugin config options such as themes, logs, prefix, and more.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ GetGuideInput,
		) (*mcp.CallToolResult, DocumentOutput, error) {
			return guideToolResult("config", daisyuimcp.GuideConfig)
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_themes_docs",
			Description: "Get the DaisyUI themes documentation. " +
				"Lists all built-in themes, explains how to apply, customize, and create new themes.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ GetGuideInput,
		) (*mcp.CallToolResult, DocumentOutput, error) {
			return guideToolResult("themes", daisyuimcp.GuideThemes)
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_base_style_docs",
			Description: "Get the DaisyUI base style documentation. " +
				"Describes the default base/reset styles applied by daisyUI and how to control them.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ GetGuideInput,
		) (*mcp.CallToolResult, DocumentOutput, error) {
			return guideToolResult("base", daisyuimcp.GuideBase)
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_utilities_docs",
			Description: "Get the DaisyUI utility classes and CSS variables documentation. " +
				"Covers all daisyUI utility classes and CSS custom properties available for styling.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ GetGuideInput,
		) (*mcp.CallToolResult, DocumentOutput, error) {
			return guideToolResult("utilities", daisyuimcp.GuideUtilities)
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_layout_typography_docs",
			Description: "Get the DaisyUI layout and typography documentation. " +
				"Covers layout helpers and typography styles included or recommended with daisyUI.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			_ GetGuideInput,
		) (*mcp.CallToolResult, DocumentOutput, error) {
			return guideToolResult("layout-and-typography", daisyuimcp.GuideLayoutTypography)
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "generate_theme",
			Description: "Generate a complete DaisyUI 5 custom theme CSS based on provided colors. " +
				"Takes a primary color (required) and optional secondary, accent, neutral, base-100, etc. " +
				"Automatically calculates appropriate -content colors for contrast and base-200/300 shades. " +
				"Returns the CSS code for the custom theme.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			input GenerateThemeInput,
		) (*mcp.CallToolResult, ThemeOutput, error) {
			themeInput := theme.ThemeInput{
				Primary:        input.Primary,
				Secondary:      input.Secondary,
				Accent:         input.Accent,
				Neutral:        input.Neutral,
				Base100:        input.Base100,
				Info:           input.Info,
				Success:        input.Success,
				Warning:        input.Warning,
				Error:          input.Error,
				RadiusSelector: input.RadiusSelector,
				RadiusField:    input.RadiusField,
				RadiusBox:      input.RadiusBox,
				SizeSelector:   input.SizeSelector,
				SizeField:      input.SizeField,
				Border:         input.Border,
				Depth:          input.Depth,
				Noise:          input.Noise,
			}
			generated := theme.GenerateTheme(themeInput)
			output := themeOutput(generated)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: generated.CSS},
				},
			}, output, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "generate_theme_from_image",
			Description: "Extracts a color palette from a local or remote image and generates a complete DaisyUI 5 custom theme CSS. " +
				"Provide either an image_path (local file) or image_url (remote image).",
		},
		func(
			ctx context.Context,
			_ *mcp.CallToolRequest,
			input GenerateThemeFromImageInput,
		) (*mcp.CallToolResult, ImageThemeOutput, error) {
			var source string
			if input.ImageURL != "" {
				source = input.ImageURL
			} else if input.ImagePath != "" {
				source = input.ImagePath
			} else {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: "Error: Either image_path or image_url must be provided."},
					},
				}, ImageThemeOutput{}, nil
			}

			themeInput, err := theme.ExtractThemeFromImageContext(ctx, source)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error extracting theme from image: %v", err)},
					},
				}, ImageThemeOutput{}, nil
			}

			generated := theme.GenerateTheme(themeInput)
			warnings := append([]string{}, generated.Warnings...)
			if themeInput.Primary == "" {
				warnings = append(warnings, "The image did not contain a vibrant primary swatch; the default primary color was used.")
			}

			extractedColors := fmt.Sprintf("/* Extracted Colors from %s:\n", source)
			if themeInput.Primary != "" {
				extractedColors += fmt.Sprintf("   Primary: %s\n", themeInput.Primary)
			}
			if themeInput.Secondary != "" {
				extractedColors += fmt.Sprintf("   Secondary: %s\n", themeInput.Secondary)
			}
			if themeInput.Accent != "" {
				extractedColors += fmt.Sprintf("   Accent: %s\n", themeInput.Accent)
			}
			if themeInput.Neutral != "" {
				extractedColors += fmt.Sprintf("   Neutral: %s\n", themeInput.Neutral)
			}
			extractedColors += "*/\n\n"

			output := ImageThemeOutput{
				Source: source,
				ExtractedColors: ExtractedColorsOutput{
					Primary: themeInput.Primary, Secondary: themeInput.Secondary, Accent: themeInput.Accent, Neutral: themeInput.Neutral,
				},
				CSS:      generated.CSS,
				Colors:   generated.Colors,
				Warnings: warnings,
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: extractedColors + generated.CSS},
				},
			}, output, nil
		},
	)

	registerRecipeTools(server, recipeIndex)
	registerDocumentationResources(server, index, docsFS, docsDir, recipeIndex)
	return server
}

func main() {
	fsys, fsDir := componentSource()
	docsFS, docsDir := docsSource()
	server := newServer(fsys, fsDir, docsFS, docsDir)
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
