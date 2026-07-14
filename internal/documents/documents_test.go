package documents

import (
	"strings"
	"testing"
)

const sampleDocument = `# Button

Introduction.

## Sizes

Small and large buttons.

### Small

Use btn-sm.

## Colors

Primary and secondary buttons.
`

func TestPaginateUsesDefaultsAndReportsContinuation(t *testing.T) {
	result, err := Paginate(strings.Repeat("x", DefaultPageSize+10), "", 0, 0)
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if result.Page != 1 || result.TotalPages != 2 || !result.HasMore {
		t.Fatalf("unexpected pagination metadata: %+v", result)
	}
	if len([]rune(result.Content)) != DefaultPageSize {
		t.Fatalf("content length = %d, want %d", len([]rune(result.Content)), DefaultPageSize)
	}
}

func TestPaginateSelectsMarkdownSection(t *testing.T) {
	result, err := Paginate(sampleDocument, "Sizes", 1, 1_000)
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if !strings.Contains(result.Content, "## Sizes") || !strings.Contains(result.Content, "### Small") {
		t.Fatalf("selected section is incomplete:\n%s", result.Content)
	}
	if strings.Contains(result.Content, "## Colors") {
		t.Fatalf("selected section includes the next peer section:\n%s", result.Content)
	}
	if result.Section != "Sizes" {
		t.Fatalf("section = %q, want Sizes", result.Section)
	}
}

func TestPaginateRejectsUnknownSectionAndOutOfRangePage(t *testing.T) {
	if _, err := Paginate(sampleDocument, "Missing", 1, 1_000); err == nil {
		t.Fatal("expected an unknown section error")
	}
	if _, err := Paginate("short", "", 2, 1_000); err == nil {
		t.Fatal("expected an out-of-range page error")
	}
}

func TestPaginateDoesNotSplitLinesWhenThereIsANearbyBoundary(t *testing.T) {
	content := strings.Repeat("short line\n", 20)
	first, err := Paginate(content, "", 1, 60)
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if !strings.HasSuffix(first.Content, "\n") {
		t.Fatalf("page split a Markdown line: %q", first.Content)
	}
	second, err := Paginate(content, "", 2, 60)
	if err != nil {
		t.Fatalf("Paginate(page 2) error = %v", err)
	}
	if first.Content+second.Content == "" {
		t.Fatal("pages unexpectedly empty")
	}
}

func TestMarkdownHeadingsIgnoresCodeFences(t *testing.T) {
	content := "# Real\n\n```md\n## Not a section\n```\n\n## Also real\n"
	result, err := Paginate(content, "", 1, 1_000)
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if len(result.AvailableSections) != 2 {
		t.Fatalf("sections = %v, want two real headings", result.AvailableSections)
	}
}

func TestMarkdownHeadingsNormalizesDaisyUIDisplayMarker(t *testing.T) {
	result, err := Paginate("### ~Button sizes\n\nDetails.\n", "Button sizes", 1, 1_000)
	if err != nil {
		t.Fatalf("Paginate() error = %v", err)
	}
	if result.Section != "Button sizes" {
		t.Fatalf("section = %q, want normalized daisyUI heading", result.Section)
	}
}
