package theme

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
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

func TestRemoteImagesRejectPrivateNetworks(t *testing.T) {
	_, err := ExtractThemeFromImageContext(t.Context(), "http://127.0.0.1/image.png")
	if err == nil || !strings.Contains(err.Error(), "private or non-public") {
		t.Fatalf("private image URL error = %v, want non-public-address error", err)
	}
}

func TestRemoteImagesRejectCredentialsAndUnsupportedSchemes(t *testing.T) {
	_, err := ExtractThemeFromImageContext(t.Context(), "https://user:secret@example.com/image.png")
	if err == nil || !strings.Contains(err.Error(), "credentials") {
		t.Fatalf("credentialed image URL error = %v, want credentials error", err)
	}
	_, err = ExtractThemeFromImageContext(t.Context(), "ftp://example.com/image.png")
	if err == nil || !strings.Contains(err.Error(), "http or https") {
		t.Fatalf("FTP image URL error = %v, want scheme error", err)
	}
}

func TestPublicImageAddressClassification(t *testing.T) {
	for _, blocked := range []string{"0.1.2.3", "127.0.0.1", "10.0.0.1", "100.64.0.1", "169.254.169.254", "192.0.2.1", "240.0.0.1", "::1", "fc00::1", "2001:db8::1", "::ffff:127.0.0.1", "64:ff9b:1::a00:1", "100::1"} {
		if isPublicImageAddress(net.ParseIP(blocked)) {
			t.Errorf("isPublicImageAddress(%s) = true, want false", blocked)
		}
	}
	for _, public := range []string{"8.8.8.8", "1.1.1.1", "2606:4700:4700::1111"} {
		if !isPublicImageAddress(net.ParseIP(public)) {
			t.Errorf("isPublicImageAddress(%s) = false, want true", public)
		}
	}
}

func TestPalettePreservesMonochromeAndIgnoresTransparentPixels(t *testing.T) {
	for _, foreground := range []color.NRGBA{{R: 255, A: 255}, {R: 128, G: 128, B: 128, A: 255}, {A: 255}, {R: 255, G: 255, B: 255, A: 255}} {
		img := image.NewNRGBA(image.Rect(10, 20, 14, 24))
		for y := 20; y < 24; y++ {
			for x := 10; x < 14; x++ {
				img.SetNRGBA(x, y, color.NRGBA{G: 255, A: 0})
			}
		}
		img.SetNRGBA(11, 21, foreground)
		got, err := extractPalette(t.Context(), img)
		if err != nil {
			t.Fatal(err)
		}
		want, _ := colorful.MakeColor(foreground)
		for _, actual := range []string{got.Primary, got.Secondary, got.Accent, got.Neutral} {
			if actual != want.Hex() {
				t.Errorf("palette color = %s, want visible source color %s", actual, want.Hex())
			}
		}
	}
	_, err := extractPalette(t.Context(), image.NewNRGBA(image.Rect(0, 0, 4, 4)))
	if err == nil || !strings.Contains(err.Error(), "no visible colors") {
		t.Fatalf("transparent image error = %v", err)
	}
}

func TestPaletteIsDeterministic(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := range 32 {
		for x := range 32 {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 8), G: uint8(y * 8), B: uint8((x + y) * 4), A: 255})
		}
	}
	want, err := extractPalette(t.Context(), img)
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		got, err := extractPalette(t.Context(), img)
		if err != nil || got != want {
			t.Fatalf("palette = %+v, %v; want %+v", got, err, want)
		}
	}
}

func TestPaletteFindsSparseVisiblePixelsBetweenSamples(t *testing.T) {
	img := image.NewNRGBA(image.Rect(0, 0, 256, 256))
	// The regular sample grid visits odd coordinates and misses this stroke.
	for y := range 256 {
		img.SetNRGBA(0, y, color.NRGBA{R: 255, A: 255})
	}
	got, err := extractPalette(t.Context(), img)
	if err != nil || got.Primary != "#ff0000" {
		t.Fatalf("sparse image palette = %+v, %v; want red", got, err)
	}
}

type sparseCountingImage struct {
	visible int
	cancel  context.CancelFunc
}

func (i *sparseCountingImage) Bounds() image.Rectangle { return image.Rect(0, 0, 512, 512) }
func (i *sparseCountingImage) ColorModel() color.Model { return color.NRGBAModel }
func (i *sparseCountingImage) At(x, y int) color.Color {
	// All regular samples (x=2 mod 4) are transparent. The fallback can
	// find more than 16,384 visible pixels and should stop at that limit.
	if x%4 == 2 {
		return color.NRGBA{}
	}
	i.visible++
	if i.cancel != nil && i.visible == 10 {
		i.cancel()
	}
	return color.NRGBA{R: 255, A: 255}
}

func TestSparsePaletteFallbackIsBoundedAndCancelable(t *testing.T) {
	img := &sparseCountingImage{}
	if _, err := extractPalette(t.Context(), img); err != nil {
		t.Fatal(err)
	}
	if img.visible != 128*128 {
		t.Fatalf("fallback accumulated %d visible pixels, want 16,384", img.visible)
	}
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	img = &sparseCountingImage{cancel: cancel}
	if _, err := extractPalette(ctx, img); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled fallback error = %v", err)
	}
}

type countingImage struct {
	bounds  image.Rectangle
	samples int
}

func (i *countingImage) Bounds() image.Rectangle { return i.bounds }
func (i *countingImage) ColorModel() color.Model { return color.NRGBAModel }
func (i *countingImage) At(x, y int) color.Color {
	i.samples++
	return color.NRGBA{R: 255, A: 255}
}

func TestPaletteSamplingIsBoundedForAllAspectRatios(t *testing.T) {
	for _, bounds := range []image.Rectangle{image.Rect(0, 0, 16_000_000, 1), image.Rect(0, 0, 1, 16_000_000), image.Rect(0, 0, 4000, 4000)} {
		img := &countingImage{bounds: bounds}
		if _, err := extractPalette(t.Context(), img); err != nil {
			t.Fatal(err)
		}
		if img.samples == 0 || img.samples > 128*128 {
			t.Fatalf("sampled %d pixels for %s", img.samples, bounds)
		}
	}
}

func TestImageExtractionHonorsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := ExtractThemeFromImageContext(ctx, "does-not-exist.png"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled extraction error = %v", err)
	}
}

func TestImageExtractionRejectsDirectory(t *testing.T) {
	if _, err := ExtractThemeFromImage(t.TempDir()); err == nil || !strings.Contains(err.Error(), "regular file") {
		t.Fatalf("directory extraction error = %v", err)
	}
}
