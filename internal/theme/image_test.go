package theme

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestExtractThemeFromImage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "palette.png")
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for y := range 64 {
		for x := range 64 {
			switch {
			case x < 32 && y < 32:
				img.Set(x, y, color.RGBA{R: 220, G: 40, B: 90, A: 255})
			case x >= 32 && y < 32:
				img.Set(x, y, color.RGBA{R: 30, G: 170, B: 210, A: 255})
			case x < 32:
				img.Set(x, y, color.RGBA{R: 40, G: 60, B: 80, A: 255})
			default:
				img.Set(x, y, color.RGBA{R: 240, G: 220, B: 120, A: 255})
			}
		}
	}
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		t.Fatalf("Failed to encode test image: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Failed to close test image: %v", err)
	}

	themeInput, err := ExtractThemeFromImage(path)
	if err != nil {
		t.Fatalf("Failed to extract theme: %v", err)
	}

	t.Logf("Extracted Theme: %+v", themeInput)

	if themeInput.Primary == "" && themeInput.Secondary == "" && themeInput.Accent == "" && themeInput.Neutral == "" {
		t.Errorf("Expected at least one color to be extracted, got all empty")
	}
}

func TestReadLimitedRejectsOversizedImage(t *testing.T) {
	_, err := readLimited(bytes.NewReader([]byte("123456")), 5)
	if err == nil {
		t.Fatal("expected oversized image to be rejected")
	}
}
