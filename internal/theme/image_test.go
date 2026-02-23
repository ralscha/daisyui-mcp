package theme

import (
	"testing"
)

func TestExtractThemeFromImage(t *testing.T) {
	url := "https://picsum.photos/id/237/200/300"

	themeInput, err := ExtractThemeFromImage(url)
	if err != nil {
		t.Fatalf("Failed to extract theme: %v", err)
	}

	t.Logf("Extracted Theme: %+v", themeInput)

	if themeInput.Primary == "" && themeInput.Secondary == "" && themeInput.Accent == "" && themeInput.Neutral == "" {
		t.Errorf("Expected at least one color to be extracted, got all empty")
	}
}
