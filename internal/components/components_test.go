package components_test

import (
	"slices"
	"testing"
	"testing/fstest"

	"daisyui-mcp/internal/components"
)

func makeFS(files map[string]string) fstest.MapFS {
	fsys := fstest.MapFS{}
	for name, content := range files {
		fsys["components/"+name+".md"] = &fstest.MapFile{Data: []byte(content)}
	}
	return fsys
}

const buttonMD = `### button
Button is used to trigger an action.

[button docs](https://daisyui.com/components/button/)

#### Class names
- component: ` + "`btn`" + `
- modifier: ` + "`btn-primary`" + `, ` + "`btn-secondary`" + `
- size: ` + "`btn-sm`" + `, ` + "`btn-lg`" + `

#### Syntax
` + "```" + `html
<button class="btn btn-primary">Click me</button>
` + "```"

const cardMD = `### card
Card is used to group and display content.

[card docs](https://daisyui.com/components/card/)

#### Class names
- component: ` + "`card`" + `

#### Syntax
` + "```" + `html
<div class="card">...</div>
` + "```"

func TestLoadIndexFS_ReturnsAllComponents(t *testing.T) {
	fsys := makeFS(map[string]string{"button": buttonMD, "card": cardMD})
	index := components.LoadIndexFS(fsys, "components")

	if len(index) != 2 {
		t.Fatalf("expected 2 components, got %d", len(index))
	}
	if index["button"] == "" {
		t.Error("button should have a description")
	}
	if index["card"] == "" {
		t.Error("card should have a description")
	}
}

func TestLoadIndexFS_DescriptionContent(t *testing.T) {
	fsys := makeFS(map[string]string{"button": buttonMD})
	index := components.LoadIndexFS(fsys, "components")

	want := "Button is used to trigger an action."
	if index["button"] != want {
		t.Errorf("description = %q, want %q", index["button"], want)
	}
}

func TestLoadIndexFS_EmptyDir(t *testing.T) {
	fsys := fstest.MapFS{}
	index := components.LoadIndexFS(fsys, "components")
	if len(index) != 0 {
		t.Errorf("expected empty index for empty FS, got %d entries", len(index))
	}
}

func TestLoadIndexFS_MalformedFile(t *testing.T) {
	fsys := fstest.MapFS{
		"components/broken.md": {Data: []byte("no header here\njust text")},
	}
	index := components.LoadIndexFS(fsys, "components")
	if len(index) != 0 {
		t.Error("malformed file should be silently skipped")
	}
}

func TestGetContentFS_ReturnsContent(t *testing.T) {
	fsys := makeFS(map[string]string{"button": buttonMD})
	content, err := components.GetContentFS(fsys, "components", "button")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content != buttonMD {
		t.Errorf("content mismatch")
	}
}

func TestGetContentFS_CaseInsensitive(t *testing.T) {
	fsys := makeFS(map[string]string{"button": buttonMD})
	content, err := components.GetContentFS(fsys, "components", "BUTTON")
	if err != nil {
		t.Fatalf("unexpected error for uppercase name: %v", err)
	}
	if content == "" {
		t.Error("expected content for uppercase name")
	}
}

func TestGetContentFS_NotFound(t *testing.T) {
	fsys := makeFS(map[string]string{"button": buttonMD})
	_, err := components.GetContentFS(fsys, "components", "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent component")
	}
}

func TestGetContentFS_PathTraversalRejected(t *testing.T) {
	fsys := makeFS(map[string]string{"button": buttonMD})
	_, err := components.GetContentFS(fsys, "components", "../button")
	if err == nil {
		t.Error("expected error for path traversal attempt with ../")
	}
	_, err = components.GetContentFS(fsys, "components", "sub/button")
	if err == nil {
		t.Error("expected error for path traversal attempt with /")
	}
}

func TestFormatList_ContainsComponents(t *testing.T) {
	index := map[string]string{
		"button": "Button is used to trigger an action.",
		"card":   "Card is used to group content.",
	}
	out := components.FormatList(index)

	for _, want := range []string{"button", "card", "2 total"} {
		if !containsStr(out, want) {
			t.Errorf("FormatList output missing %q", want)
		}
	}
}

func TestFormatList_Empty(t *testing.T) {
	out := components.FormatList(map[string]string{})
	if !containsStr(out, "No components") {
		t.Errorf("expected 'No components' message, got: %s", out)
	}
}

func TestSuggestions_SubstringMatch(t *testing.T) {
	index := map[string]string{"button": "", "submit-button": "", "card": ""}
	got := components.Suggestions(index, "butt")
	assertContains(t, got, "button")
	assertContains(t, got, "submit-button")
}

func TestSuggestions_PrefixFallback(t *testing.T) {
	index := map[string]string{"carousel": "", "card": "", "button": ""}
	got := components.Suggestions(index, "ca")
	assertContains(t, got, "carousel")
	assertContains(t, got, "card")
}

func TestSuggestions_MaxFive(t *testing.T) {
	index := map[string]string{
		"ab1": "", "ab2": "", "ab3": "", "ab4": "", "ab5": "", "ab6": "",
	}
	got := components.Suggestions(index, "ab")
	if len(got) > 5 {
		t.Errorf("expected at most 5 suggestions, got %d", len(got))
	}
}

func TestSuggestions_NoMatch(t *testing.T) {
	index := map[string]string{"button": "", "card": ""}
	got := components.Suggestions(index, "zzz")
	if len(got) != 0 {
		t.Errorf("expected no suggestions for completely unrelated name, got %v", got)
	}
}

func containsStr(s, sub string) bool {
	return len(s) > 0 && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func assertContains(t *testing.T, slice []string, want string) {
	t.Helper()
	if slices.Contains(slice, want) {
		return
	}
	t.Errorf("expected %q in suggestions %v", want, slice)
}
