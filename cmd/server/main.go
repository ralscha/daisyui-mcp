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

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
