package main

import (
	"context"
	"io/fs"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	daisyuimcp "daisyui-mcp"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var recipeReferencePattern = regexp.MustCompile(`https://daisyui\.com/components/([a-z0-9-]+)/`)

func testServer(t *testing.T) *mcp.ClientSession {
	t.Helper()
	shortFS := fstest.MapFS{
		"components/button.md": {Data: []byte("### button\nA clickable button.\n\n#### Syntax\n`<button class=\"btn\">Save</button>`\n")},
	}
	detailedFS := fstest.MapFS{
		"docs/button.md": {Data: []byte("# Button\n\nIntroduction.\n\n## Sizes\n\n" + strings.Repeat("Size details. ", 20) + "\n\n## Colors\n\nColor details.\n")},
	}

	serverTransport, clientTransport := mcp.NewInMemoryTransports()
	server := newServer(shortFS, "components", detailedFS, "docs")
	serverSession, err := server.Connect(t.Context(), serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect() error = %v", err)
	}
	t.Cleanup(func() { _ = serverSession.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	clientSession, err := client.Connect(t.Context(), clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect() error = %v", err)
	}
	t.Cleanup(func() { _ = clientSession.Close() })
	return clientSession
}

func TestDetailedDocToolReturnsPagedStructuredContent(t *testing.T) {
	client := testServer(t)
	result, err := client.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "get_detailed_doc",
		Arguments: map[string]any{
			"name":      "button",
			"section":   "Sizes",
			"page":      1,
			"page_size": 60,
		},
	})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	if result.IsError {
		t.Fatalf("CallTool() returned a tool error: %v", result.Content)
	}
	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content type = %T, want map[string]any", result.StructuredContent)
	}
	if structured["section"] != "Sizes" || structured["has_more"] != true {
		t.Fatalf("unexpected structured content: %#v", structured)
	}
	if _, ok := structured["documentation"]; !ok {
		t.Fatalf("structured content is missing documentation: %#v", structured)
	}
}

func TestDocumentationResourcesAndTemplates(t *testing.T) {
	client := testServer(t)

	templates, err := client.ListResourceTemplates(t.Context(), nil)
	if err != nil {
		t.Fatalf("ListResourceTemplates() error = %v", err)
	}
	wantTemplates := map[string]bool{
		"daisyui://components/{name}": false,
		"daisyui://guides/{name}":     false,
		"daisyui://recipes/{name}":    false,
	}
	for _, template := range templates.ResourceTemplates {
		if _, ok := wantTemplates[template.URITemplate]; ok {
			wantTemplates[template.URITemplate] = true
		}
	}
	for uri, found := range wantTemplates {
		if !found {
			t.Errorf("resource template %q was not advertised", uri)
		}
	}

	resource, err := client.ReadResource(context.Background(), &mcp.ReadResourceParams{URI: "daisyui://components/button"})
	if err != nil {
		t.Fatalf("ReadResource() error = %v", err)
	}
	if len(resource.Contents) != 1 || !strings.Contains(resource.Contents[0].Text, "## Sizes") {
		t.Fatalf("unexpected component resource: %#v", resource.Contents)
	}

	recipe, err := client.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: "daisyui://recipes/login-form"})
	if err != nil {
		t.Fatalf("ReadResource(recipe) error = %v", err)
	}
	if len(recipe.Contents) != 1 || !strings.Contains(recipe.Contents[0].Text, "autocomplete=\"email\"") {
		t.Fatalf("unexpected recipe resource: %#v", recipe.Contents)
	}
}

func TestGuideToolsReturnTheirOwnDocuments(t *testing.T) {
	client := testServer(t)
	guides := map[string]string{
		"get_customize_docs":         "customize",
		"get_config_docs":            "config",
		"get_themes_docs":            "themes",
		"get_base_style_docs":        "base",
		"get_utilities_docs":         "utilities",
		"get_layout_typography_docs": "layout-and-typography",
	}
	for toolName, documentName := range guides {
		result, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: toolName, Arguments: map[string]any{}})
		if err != nil {
			t.Fatalf("CallTool(%s) error = %v", toolName, err)
		}
		structured, ok := result.StructuredContent.(map[string]any)
		if !ok || structured["name"] != documentName || structured["documentation"] == "" {
			t.Errorf("%s returned unexpected document: %#v", toolName, result.StructuredContent)
		}
	}
}

func TestListComponentsReturnsStructuredArray(t *testing.T) {
	client := testServer(t)
	result, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: "list_components", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("CallTool() error = %v", err)
	}
	structured, ok := result.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("structured content type = %T, want map[string]any", result.StructuredContent)
	}
	components, ok := structured["components"].([]any)
	if !ok || len(components) != 1 {
		t.Fatalf("components = %#v, want one structured component", structured["components"])
	}
}

func TestSearchComponentsReturnsRankedStructuredResults(t *testing.T) {
	client := testServer(t)
	result, err := client.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "search_components",
		Arguments: map[string]any{"query": "clickable"},
	})
	if err != nil {
		t.Fatalf("CallTool(search_components) error = %v", err)
	}
	if result.IsError {
		t.Fatalf("search_components returned a tool error: %#v", result.Content)
	}
	structured, ok := result.StructuredContent.(map[string]any)
	if !ok || structured["count"] != float64(1) {
		t.Fatalf("unexpected search structured content: %#v", result.StructuredContent)
	}
	components, ok := structured["components"].([]any)
	if !ok || len(components) != 1 {
		t.Fatalf("unexpected search components: %#v", structured["components"])
	}
}

func TestSearchComponentsRejectsInvalidInput(t *testing.T) {
	client := testServer(t)
	for _, arguments := range []map[string]any{{"query": " "}, {"query": "button", "limit": 51}} {
		result, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: "search_components", Arguments: arguments})
		if err != nil {
			t.Fatalf("CallTool(search_components) error = %v", err)
		}
		if !result.IsError {
			t.Fatalf("search_components accepted invalid input %#v", arguments)
		}
	}
}

func TestThemeAndRecipeToolsReturnStructuredContent(t *testing.T) {
	client := testServer(t)

	themeResult, err := client.CallTool(t.Context(), &mcp.CallToolParams{
		Name:      "generate_theme",
		Arguments: map[string]any{"name": "brand", "primary": "#2563eb", "default": true, "prefers_dark": true},
	})
	if err != nil {
		t.Fatalf("CallTool(generate_theme) error = %v", err)
	}
	themeOutput, ok := themeResult.StructuredContent.(map[string]any)
	if !ok {
		t.Fatalf("theme structured content type = %T", themeResult.StructuredContent)
	}
	colors, ok := themeOutput["colors"].([]any)
	if !ok || len(colors) != 9 || themeOutput["css"] == "" {
		t.Fatalf("unexpected theme structured content: %#v", themeOutput)
	}
	if _, ok := themeOutput["warnings"].([]any); !ok {
		t.Fatalf("theme warnings are not a structured array: %#v", themeOutput["warnings"])
	}
	if themeOutput["name"] != "brand" || themeOutput["default"] != true || themeOutput["prefers_dark"] != true {
		t.Fatalf("theme metadata is missing from structured content: %#v", themeOutput)
	}
	css, _ := themeOutput["css"].(string)
	if !strings.Contains(css, `name: "brand";`) || !strings.Contains(css, "default: true;") || !strings.Contains(css, "prefersdark: true;") {
		t.Fatalf("generated theme is missing requested metadata: %s", css)
	}

	recipeResult, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: "list_recipes", Arguments: map[string]any{}})
	if err != nil {
		t.Fatalf("CallTool(list_recipes) error = %v", err)
	}
	recipeOutput, ok := recipeResult.StructuredContent.(map[string]any)
	if !ok || recipeOutput["count"] != float64(20) {
		t.Fatalf("unexpected recipe structured content: %#v", recipeResult.StructuredContent)
	}
}

func TestImageThemeToolRequiresExactlyOneSource(t *testing.T) {
	client := testServer(t)
	for _, arguments := range []map[string]any{
		{},
		{"image_path": "local.png", "image_url": "https://example.com/image.png"},
	} {
		result, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: "generate_theme_from_image", Arguments: arguments})
		if err != nil {
			t.Fatalf("CallTool(generate_theme_from_image) error = %v", err)
		}
		if !result.IsError {
			t.Fatalf("generate_theme_from_image accepted sources %#v", arguments)
		}
	}
}

func TestEmbeddedRecipesHaveRequiredSectionsAndValidReferences(t *testing.T) {
	paths, err := fs.Glob(daisyuimcp.RecipesFS, "recipes/*.md")
	if err != nil {
		t.Fatalf("fs.Glob(recipes) error = %v", err)
	}
	if len(paths) != 20 {
		t.Fatalf("embedded recipe count = %d, want 20", len(paths))
	}

	for _, path := range paths {
		data, err := fs.ReadFile(daisyuimcp.RecipesFS, path)
		if err != nil {
			t.Errorf("ReadFile(%q) error = %v", path, err)
			continue
		}
		content := string(data)
		for _, heading := range []string{"## Accessibility notes", "## Provenance"} {
			if !strings.Contains(content, heading) {
				t.Errorf("%s is missing %q", path, heading)
			}
		}
		if !strings.Contains(content, "- Adapted or copied code: No") {
			t.Errorf("%s does not declare its authorship provenance", path)
		}

		matches := recipeReferencePattern.FindAllStringSubmatch(content, -1)
		if len(matches) == 0 {
			t.Errorf("%s has no official daisyUI component references", path)
		}
		for _, match := range matches {
			componentPath := "components/" + match[1] + ".md"
			if _, err := fs.Stat(daisyuimcp.ComponentsFS, componentPath); err != nil {
				t.Errorf("%s references missing embedded documentation %s", path, componentPath)
			}
		}
	}
}
