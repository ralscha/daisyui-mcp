package main

import (
	"context"
	"fmt"
	"io/fs"
	"net/url"
	"sort"
	"strings"

	daisyuimcp "daisyui-mcp"
	"daisyui-mcp/internal/components"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func guideDocuments() map[string][]byte {
	return map[string][]byte{
		"base":                  daisyuimcp.GuideBase,
		"colors":                daisyuimcp.ColorsData,
		"config":                daisyuimcp.GuideConfig,
		"customize":             daisyuimcp.GuideCustomize,
		"layout-and-typography": daisyuimcp.GuideLayoutTypography,
		"themes":                daisyuimcp.GuideThemes,
		"utilities":             daisyuimcp.GuideUtilities,
	}
}

func registerDocumentationResources(
	server *mcp.Server,
	componentIndex map[string]string,
	docsFS fs.FS,
	docsDir string,
	recipeIndex map[string]string,
) {
	handler := documentationResourceHandler(docsFS, docsDir, recipeIndex)

	templates := []*mcp.ResourceTemplate{
		{
			Name:        "daisyui-component",
			Title:       "daisyUI component documentation",
			Description: "Full documentation for a daisyUI component.",
			MIMEType:    "text/markdown",
			URITemplate: "daisyui://components/{name}",
		},
		{
			Name:        "daisyui-guide",
			Title:       "daisyUI guide",
			Description: "A daisyUI configuration, theme, color, utility, or layout guide.",
			MIMEType:    "text/markdown",
			URITemplate: "daisyui://guides/{name}",
		},
		{
			Name:        "daisyui-recipe",
			Title:       "daisyUI composition recipe",
			Description: "A curated, accessible composition built from daisyUI components.",
			MIMEType:    "text/markdown",
			URITemplate: "daisyui://recipes/{name}",
		},
	}
	for _, template := range templates {
		server.AddResourceTemplate(template, handler)
	}

	for _, component := range componentSummaries(componentIndex, "components") {
		content, err := components.GetContentFS(docsFS, docsDir, component.Name)
		if err != nil {
			continue
		}
		server.AddResource(&mcp.Resource{
			Name:        "component-" + component.Name,
			Title:       component.Name,
			Description: component.Description,
			MIMEType:    "text/markdown",
			Size:        int64(len(content)),
			URI:         component.URI,
		}, handler)
	}

	guides := guideDocuments()
	guideNames := make([]string, 0, len(guides))
	for name := range guides {
		guideNames = append(guideNames, name)
	}
	sort.Strings(guideNames)
	for _, name := range guideNames {
		server.AddResource(&mcp.Resource{
			Name:        "guide-" + name,
			Title:       strings.ReplaceAll(name, "-", " "),
			Description: "daisyUI " + strings.ReplaceAll(name, "-", " ") + " documentation.",
			MIMEType:    "text/markdown",
			Size:        int64(len(guides[name])),
			URI:         "daisyui://guides/" + name,
		}, handler)
	}

	for _, recipe := range componentSummaries(recipeIndex, "recipes") {
		content, err := components.GetContentFS(daisyuimcp.RecipesFS, "recipes", recipe.Name)
		if err != nil {
			continue
		}
		server.AddResource(&mcp.Resource{
			Name:        "recipe-" + recipe.Name,
			Title:       strings.ReplaceAll(recipe.Name, "-", " "),
			Description: recipe.Description,
			MIMEType:    "text/markdown",
			Size:        int64(len(content)),
			URI:         recipe.URI,
		}, handler)
	}
}

func documentationResourceHandler(docsFS fs.FS, docsDir string, recipeIndex map[string]string) mcp.ResourceHandler {
	return func(_ context.Context, request *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		uri := request.Params.URI
		parsed, err := url.Parse(uri)
		if err != nil || parsed.Scheme != "daisyui" {
			return nil, mcp.ResourceNotFoundError(uri)
		}
		name, err := url.PathUnescape(strings.TrimPrefix(parsed.Path, "/"))
		if err != nil || name == "" || strings.Contains(name, "/") {
			return nil, mcp.ResourceNotFoundError(uri)
		}

		var content string
		switch parsed.Host {
		case "components":
			content, err = components.GetContentFS(docsFS, docsDir, name)
		case "guides":
			data, ok := guideDocuments()[name]
			if !ok {
				err = fmt.Errorf("guide not found")
			} else {
				content = string(data)
			}
		case "recipes":
			if _, ok := recipeIndex[name]; !ok {
				err = fmt.Errorf("recipe not found")
			} else {
				content, err = components.GetContentFS(daisyuimcp.RecipesFS, "recipes", name)
			}
		default:
			err = fmt.Errorf("resource type not found")
		}
		if err != nil {
			return nil, mcp.ResourceNotFoundError(uri)
		}

		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{
			URI:      uri,
			MIMEType: "text/markdown",
			Text:     content,
		}}}, nil
	}
}

func registerRecipeTools(server *mcp.Server, recipeIndex map[string]string) {
	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "list_recipes",
			Description: "List curated daisyUI composition recipes for common application patterns.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, _ GetGuideInput) (*mcp.CallToolResult, RecipeListOutput, error) {
			recipes := componentSummaries(recipeIndex, "recipes")
			var text strings.Builder
			fmt.Fprintf(&text, "Available daisyUI recipes (%d total):\n\n", len(recipes))
			for _, recipe := range recipes {
				fmt.Fprintf(&text, "  • %s - %s\n", recipe.Name, recipe.Description)
			}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text.String()}}}, RecipeListOutput{
				Recipes: recipes,
				Count:   len(recipes),
			}, nil
		},
	)

	mcp.AddTool(
		server,
		&mcp.Tool{
			Name:        "get_recipe",
			Description: "Get a curated daisyUI composition recipe with accessible HTML and implementation guidance.",
		},
		func(_ context.Context, _ *mcp.CallToolRequest, input GetComponentInput) (*mcp.CallToolResult, DocumentOutput, error) {
			name := strings.ToLower(strings.TrimSpace(input.Name))
			if _, ok := recipeIndex[name]; !ok {
				return &mcp.CallToolResult{
					IsError: true,
					Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Recipe %q not found. Use list_recipes() to see available recipes.", input.Name)}},
				}, DocumentOutput{}, nil
			}
			content, err := components.GetContentFS(daisyuimcp.RecipesFS, "recipes", name)
			if err != nil {
				return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, DocumentOutput{}, nil
			}
			output := DocumentOutput{Name: name, URI: "daisyui://recipes/" + name, Documentation: content, Warnings: []string{}}
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: content}}}, output, nil
		},
	)
}
