package theme

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"os"
	"strings"

	"github.com/generaltso/vibrant"
	_ "golang.org/x/image/webp"
)

func ExtractThemeFromImage(imagePathOrURL string) (ThemeInput, error) {
	var img image.Image
	var err error

	if strings.HasPrefix(imagePathOrURL, "http://") || strings.HasPrefix(imagePathOrURL, "https://") {
		resp, err := http.Get(imagePathOrURL)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to fetch image from URL: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return ThemeInput{}, fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
		}

		img, _, err = image.Decode(resp.Body)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to decode image from URL: %w", err)
		}
	} else {
		file, err := os.Open(imagePathOrURL)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to open local image file: %w", err)
		}
		defer file.Close()

		img, _, err = image.Decode(file)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to decode local image: %w", err)
		}
	}

	palette, err := vibrant.NewPaletteFromImage(img)
	if err != nil {
		return ThemeInput{}, fmt.Errorf("failed to generate color palette: %w", err)
	}

	swatches := palette.ExtractAwesome()
	themeInput := ThemeInput{}

	if swatch, ok := swatches["Vibrant"]; ok && swatch != nil {
		themeInput.Primary = swatch.Color.RGBHex()
	}
	if swatch, ok := swatches["LightVibrant"]; ok && swatch != nil {
		themeInput.Secondary = swatch.Color.RGBHex()
	} else if swatch, ok := swatches["Muted"]; ok && swatch != nil {
		themeInput.Secondary = swatch.Color.RGBHex()
	}
	if swatch, ok := swatches["DarkVibrant"]; ok && swatch != nil {
		themeInput.Accent = swatch.Color.RGBHex()
	}
	if swatch, ok := swatches["DarkMuted"]; ok && swatch != nil {
		themeInput.Neutral = swatch.Color.RGBHex()
	}

	return themeInput, nil
}
