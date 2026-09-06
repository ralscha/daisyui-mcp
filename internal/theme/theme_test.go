package theme

import (
	"strings"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
)

func TestParseColorAcceptsHexWithoutHash(t *testing.T) {
	got := parseColor("ff0000", colorful.Color{})
	if got.Hex() != "#ff0000" {
		t.Fatalf("parseColor() = %s, want #ff0000", got.Hex())
	}
}

func TestParseColorAcceptsOKLCH(t *testing.T) {
	got := parseColor("oklch(50% 0.2 120)", colorful.Color{})
	l, c, h := got.OkLch()
	if l < 0.49 || l > 0.51 || c < 0.19 || c > 0.21 || h < 119 || h > 121 {
		t.Fatalf("parseColor() OKLCH = (%f, %f, %f), want about (0.5, 0.2, 120)", l, c, h)
	}
}

func TestGenerateThemeCSSUsesDefaultShapeValues(t *testing.T) {
	css := GenerateThemeCSS(ThemeInput{Primary: "#ff0000"})
	for _, want := range []string{
		"--radius-selector: 0.25rem;",
		"--radius-field: 0.25rem;",
		"--radius-box: 0.5rem;",
		"--size-selector: 0.25rem;",
		"--size-field: 0.25rem;",
		"--border: 1px;",
		"--depth: 0;",
		"--noise: 0;",
	} {
		if !strings.Contains(css, want) {
			t.Fatalf("generated CSS missing %q:\n%s", want, css)
		}
	}
}

func TestGenerateThemeReturnsColorsContrastAndWarnings(t *testing.T) {
	result := GenerateTheme(ThemeInput{Primary: "not-a-color"})
	if result.CSS == "" {
		t.Fatal("GenerateTheme() returned empty CSS")
	}
	if len(result.Colors) != 9 {
		t.Fatalf("GenerateTheme() returned %d colors, want 9", len(result.Colors))
	}
	for _, color := range result.Colors {
		if color.Value == "" || color.ContentValue == "" || color.ContrastRatio <= 0 {
			t.Fatalf("incomplete generated color: %+v", color)
		}
	}
	foundInvalidWarning := false
	for _, warning := range result.Warnings {
		if strings.Contains(warning, "Invalid primary color") {
			foundInvalidWarning = true
		}
	}
	if !foundInvalidWarning {
		t.Fatalf("warnings = %v, want an invalid primary warning", result.Warnings)
	}
}

func TestGenerateThemeSupportsMetadata(t *testing.T) {
	result := GenerateTheme(ThemeInput{
		Name:        "brand-dark",
		Default:     true,
		PrefersDark: true,
		Primary:     "#2563eb",
	})
	for _, want := range []string{`name: "brand-dark";`, "default: true;", "prefersdark: true;"} {
		if !strings.Contains(result.CSS, want) {
			t.Fatalf("generated CSS missing %q:\n%s", want, result.CSS)
		}
	}
}

func TestGenerateThemeRejectsUnsafeOptions(t *testing.T) {
	result := GenerateTheme(ThemeInput{
		Name:           `bad\"; } body { color: red`,
		Primary:        "#2563eb",
		RadiusSelector: "1rem; color: red",
		Depth:          "url(https://example.com)",
	})
	if strings.Contains(result.CSS, "body") || strings.Contains(result.CSS, "url(") {
		t.Fatalf("unsafe option was emitted into CSS:\n%s", result.CSS)
	}
	if len(result.Warnings) < 3 {
		t.Fatalf("warnings = %v, want invalid-option warnings", result.Warnings)
	}
}

func TestGeneratedContentColorsMeetWCAGAA(t *testing.T) {
	for _, background := range []string{"#000000", "#555555", "#777777", "#999999", "#ffffff", "oklch(55% 0.3 240)"} {
		result := GenerateTheme(ThemeInput{Primary: background})
		primary := result.Colors[0]
		if primary.Name != "base-100" {
			t.Fatalf("first generated color = %q, want base-100", primary.Name)
		}
		for _, color := range result.Colors {
			if color.ContrastRatio < 4.5 {
				t.Fatalf("%s background %s has contrast %.2f, want at least 4.5", color.Name, background, color.ContrastRatio)
			}
		}
	}
}

func TestGeneratedContentColorMeetsWCAGAAAcrossColorSpace(t *testing.T) {
	for lightness := 0.0; lightness <= 1.0; lightness += 0.05 {
		for _, chroma := range []float64{0, 0.1, 0.2, 0.3} {
			for hue := 0.0; hue < 360; hue += 30 {
				background := cssColor(colorful.OkLch(lightness, chroma, hue))
				content := generateContentColor(background)
				if ratio := contrastRatio(background, content); ratio < 4.5 {
					t.Fatalf("background L=%f C=%f H=%f has contrast %.3f, want at least 4.5", lightness, chroma, hue, ratio)
				}
			}
		}
	}
}
