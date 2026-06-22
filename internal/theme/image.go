package theme

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/generaltso/vibrant"
	_ "golang.org/x/image/webp"
)

var imageHTTPClient = &http.Client{Timeout: 15 * time.Second}

const (
	maxImageBytes  = 10 << 20
	maxImagePixels = 16_000_000
)

func ExtractThemeFromImage(imagePathOrURL string) (ThemeInput, error) {
	return ExtractThemeFromImageContext(context.Background(), imagePathOrURL)
}

func ExtractThemeFromImageContext(ctx context.Context, imagePathOrURL string) (ThemeInput, error) {
	var data []byte

	if strings.HasPrefix(imagePathOrURL, "http://") || strings.HasPrefix(imagePathOrURL, "https://") {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, imagePathOrURL, nil)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to create image request: %w", err)
		}

		resp, err := imageHTTPClient.Do(req)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to fetch image from URL: %w", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return ThemeInput{}, fmt.Errorf("failed to fetch image, status code: %d", resp.StatusCode)
		}

		data, err = readLimited(resp.Body, maxImageBytes)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to read image from URL: %w", err)
		}
	} else {
		file, err := os.Open(imagePathOrURL)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to open local image file: %w", err)
		}
		defer func() { _ = file.Close() }()

		data, err = readLimited(file, maxImageBytes)
		if err != nil {
			return ThemeInput{}, fmt.Errorf("failed to read local image: %w", err)
		}
	}

	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return ThemeInput{}, fmt.Errorf("failed to decode image metadata: %w", err)
	}
	if config.Width <= 0 || config.Height <= 0 || config.Width > maxImagePixels/config.Height {
		return ThemeInput{}, fmt.Errorf("image dimensions %dx%d exceed the limit of %d pixels", config.Width, config.Height, maxImagePixels)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return ThemeInput{}, fmt.Errorf("failed to decode image: %w", err)
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

func readLimited(r io.Reader, maxBytes int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("image is larger than %d bytes", maxBytes)
	}
	return data, nil
}
