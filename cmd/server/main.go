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
type GetColorPaletteInput struct{}
type GetGuideInput struct{}
type GenerateThemeInput struct {
	Primary   string `json:"primary" jsonschema:"Primary color in Hex format (e.g. '#ff0000'). Required."`
	Secondary string `json:"secondary,omitempty" jsonschema:"Secondary color in Hex format. Optional."`
	Accent    string `json:"accent,omitempty" jsonschema:"Accent color in Hex format. Optional."`
	Neutral   string `json:"neutral,omitempty" jsonschema:"Neutral color in Hex format. Optional."`
	Base100   string `json:"base_100,omitempty" jsonschema:"Base 100 color (background) in Hex format. Optional."`
	Info      string `json:"info,omitempty" jsonschema:"Info color in Hex format. Optional."`
	Success   string `json:"success,omitempty" jsonschema:"Success color in Hex format. Optional."`
	Warning   string `json:"warning,omitempty" jsonschema:"Warning color in Hex format. Optional."`
	Error     string `json:"error,omitempty" jsonschema:"Error color in Hex format. Optional."`

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

func main() {
	fsys, fsDir := componentSource()
	docsFS, docsDir := docsSource()

	index := components.LoadIndexFS(fsys, fsDir)
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
		) (*mcp.CallToolResult, any, error) {
			text := components.FormatList(index)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: text},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
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
					Content: []mcp.Content{
						&mcp.TextContent{Text: text},
					},
				}, nil, nil
			}

			content, err := components.GetContentFS(fsys, fsDir, name)
			if err != nil {
				log.Printf("error loading component %q: %v", name, err)
				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{
							Text: fmt.Sprintf("Error: Unable to load documentation for '%s'. Please try again.", input.Name),
						},
					},
				}, nil, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: content},
				},
			}, nil, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name: "get_detailed_doc",
			Description: "Get the full documentation page for a specific DaisyUI component. " +
				"Including all variants, examples, and advanced usage. " +
				"Use get_short_doc() for a quick summary.",
		},
		func(
			_ context.Context,
			_ *mcp.CallToolRequest,
			input GetComponentInput,
		) (*mcp.CallToolResult, any, error) {
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
					Content: []mcp.Content{
						&mcp.TextContent{Text: text},
					},
				}, nil, nil
			}

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: content},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.ColorsData)},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.GuideCustomize)},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.GuideConfig)},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.GuideThemes)},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.GuideBase)},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.GuideUtilities)},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: string(daisyuimcp.GuideLayoutTypography)},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
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
			css := theme.GenerateThemeCSS(themeInput)
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: css},
				},
			}, nil, nil
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
		) (*mcp.CallToolResult, any, error) {
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
				}, nil, nil
			}

			themeInput, err := theme.ExtractThemeFromImageContext(ctx, source)
			if err != nil {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{
						&mcp.TextContent{Text: fmt.Sprintf("Error extracting theme from image: %v", err)},
					},
				}, nil, nil
			}

			css := theme.GenerateThemeCSS(themeInput)

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

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: extractedColors + css},
				},
			}, nil, nil
		},
	)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
