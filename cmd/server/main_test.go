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
	return connectTestServer(t, newServer(shortFS, "components", detailedFS, "docs"))
}

func connectTestServer(t *testing.T, server *mcp.Server) *mcp.ClientSession {
	t.Helper()
	serverTransport, clientTransport := mcp.NewInMemoryTransports()
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

func TestServerReadsDirectoryRootOverrides(t *testing.T) {
	shortFS := fstest.MapFS{"button.md": {Data: []byte("### Button\nOverride summary.\n")}}
	detailedFS := fstest.MapFS{"button.md": {Data: []byte("# Button\nOverride details.\n")}}
	client := connectTestServer(t, newServer(shortFS, ".", detailedFS, "."))
	for _, toolName := range []string{"get_short_doc", "get_detailed_doc"} {
		result, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: toolName, Arguments: map[string]any{"name": "button"}})
		if err != nil {
			t.Fatal(err)
		}
		if result.IsError {
			t.Fatalf("%s could not read root directory override: %#v", toolName, result.Content)
		}
	}
	result, err := client.ReadResource(t.Context(), &mcp.ReadResourceParams{URI: "daisyui://components/button"})
	if err != nil || len(result.Contents) != 1 || !strings.Contains(result.Contents[0].Text, "Override details.") {
		t.Fatalf("ReadResource(root override) = %#v, %v", result, err)
	}
}

func TestDocumentationResourceURIs(t *testing.T) {
	fsys := fstest.MapFS{
		"docs/button.md": {Data: []byte("Button")},
		"docs/100%.md":   {Data: []byte("Percent")},
	}
	handler := documentationResourceHandler(fsys, "docs", nil)
	for _, uri := range []string{
		"daisyui://components/button?ignored=true",
		"daisyui://components/button?",
		"daisyui://components/button#ignored",
		"daisyui://user@components/button",
		"daisyui://components/%2562utton",
		"daisyui://components/..%2Fbutton",
		"daisyui://components/sub%5Cbutton",
	} {
		t.Run(uri, func(t *testing.T) {
			_, err := handler(t.Context(), &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: uri}})
			if err == nil {
				t.Fatal("invalid or doubly encoded resource URI was accepted")
			}
		})
	}
	result, err := handler(t.Context(), &mcp.ReadResourceRequest{Params: &mcp.ReadResourceParams{URI: "daisyui://components/100%25"}})
	if err != nil || len(result.Contents) != 1 || result.Contents[0].Text != "Percent" {
		t.Fatalf("ReadResource(percent name) = %#v, %v", result, err)
	}
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

func TestSearchRecipesReturnsRankedResultsAndValidatesInput(t *testing.T) {
	client := testServer(t)
	result, err := client.CallTool(t.Context(), &mcp.CallToolParams{
		Name: "search_recipes", Arguments: map[string]any{"query": " login ", "limit": 1},
	})
	if err != nil || result.IsError {
		t.Fatalf("search_recipes = %#v, %v", result, err)
	}
	structured, ok := result.StructuredContent.(map[string]any)
	if !ok || structured["count"] != float64(1) {
		t.Fatalf("unexpected recipe search result: %#v", result.StructuredContent)
	}
	recipes, ok := structured["recipes"].([]any)
	if !ok || len(recipes) != 1 {
		t.Fatalf("unexpected recipe list: %#v", structured["recipes"])
	}
	first, ok := recipes[0].(map[string]any)
	if !ok || first["name"] != "login-form" || first["uri"] != "daisyui://recipes/login-form" {
		t.Fatalf("unexpected first recipe: %#v", recipes[0])
	}
	for _, input := range []map[string]any{{"query": " "}, {"query": "login", "limit": -1}, {"query": "login", "limit": 51}} {
		result, err := client.CallTool(t.Context(), &mcp.CallToolParams{Name: "search_recipes", Arguments: input})
		if err != nil || !result.IsError {
			t.Fatalf("invalid recipe search input %#v: result %#v, error %v", input, result, err)
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

func TestImageSourceValidatesSourceKinds(t *testing.T) {
	for _, input := range []GenerateThemeFromImageInput{
		{ImageURL: "local.png"},
		{ImageURL: "file:///tmp/image.png"},
		{ImageURL: "https:///image.png"},
		{ImageURL: "https://user:password@example.com/image.png"},
		{ImagePath: "https://example.com/image.png"},
	} {
		if source, err := imageSource(input); err == nil {
			t.Errorf("imageSource(%+v) accepted invalid source %q", input, source)
		}
	}
	for _, input := range []GenerateThemeFromImageInput{
		{ImageURL: "https://example.com/image.png"},
		{ImageURL: "HTTPS://example.com/image.png"},
		{ImagePath: `C:\images\palette.png`},
		{ImagePath: "/tmp/palette.png"},
	} {
		if _, err := imageSource(input); err != nil {
			t.Errorf("imageSource(%+v) rejected valid source: %v", input, err)
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
