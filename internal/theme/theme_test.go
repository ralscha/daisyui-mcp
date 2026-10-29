package theme

import (
	"encoding/json"
	"math"
	"regexp"
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

func TestParseColorAcceptsCSSAnglesAndChromaPercentages(t *testing.T) {
	for _, input := range []string{
		"oklch(.5 50% -240)",
		"oklch(50% .2 480deg)",
		"OKLCH(5e-1 2e-1 120DEG)",
		"oklch(50% .2 133.333333grad)",
		"oklch(50% .2 2.0943951rad)",
		"oklch(50% .2 .333333333turn)",
	} {
		got, valid := parseColorWithStatus(input, colorful.Color{})
		l, c, h := got.OkLch()
		if !valid || math.Abs(l-.5) > .0001 || math.Abs(c-.2) > .0001 || math.Abs(h-120) > .02 {
			t.Errorf("parseColor(%q) = (%f, %f, %f), valid %t", input, l, c, h, valid)
		}
	}
}

func TestParseColorRejectsOverflowAndInvalidNumbers(t *testing.T) {
	for _, input := range []string{"oklch(50 .2 120)", "oklch(50% -0.2 120)", "oklch(50% 1e300 120)", "oklch(50% .2 1e308turn)", "oklch(50% ..2 120)", "oklch(50% .2 NaN)"} {
		result := GenerateTheme(ThemeInput{Primary: input})
		if len(result.Warnings) == 0 || !strings.Contains(result.Warnings[0], "Invalid primary") {
			t.Errorf("input %q produced no invalid-color warning: %v", input, result.Warnings)
		}
		if strings.Contains(result.CSS, "NaN") || strings.Contains(result.CSS, "Inf") {
			t.Errorf("input %q generated non-finite CSS", input)
		}
		if _, err := json.Marshal(result); err != nil {
			t.Errorf("input %q generated invalid JSON: %v", input, err)
		}
	}
}

func TestFormatOKLCHKeepsDeclaredPrecision(t *testing.T) {
	pattern := regexp.MustCompile(`^oklch\(\d+(?:\.\d{1,3})?% \d+(?:\.\d{1,3})? \d+(?:\.\d{1,3})?\)$`)
	for _, c := range []colorful.Color{colorful.OkLch(.5, .2, 120), colorful.OkLch(.75234, .23333, 53.12345)} {
		if got := formatOklch(c); !pattern.MatchString(got) {
			t.Fatalf("formatted color has excessive precision: %q", got)
		}
	}
	if got := formatOklch(colorful.Color{R: 1, G: 1, B: 1}); got != "oklch(100% 0 0)" {
		t.Fatalf("formatted white = %q", got)
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

func TestGeneratedThemeWarnsAboutDerivedBaseContrast(t *testing.T) {
	result := GenerateTheme(ThemeInput{Primary: "#2563eb", Base100: "#777777"})
	if len(result.Colors) != 9 {
		t.Fatalf("color count = %d, want the nine semantic color pairs", len(result.Colors))
	}
	warnings := strings.Join(result.Warnings, "\n")
	for _, shade := range []string{"base-200", "base-300"} {
		if !strings.Contains(warnings, shade+" and base-content have a contrast ratio") {
			t.Errorf("warnings %q omit the low-contrast %s shade", warnings, shade)
		}
	}
	if strings.Contains(warnings, "base-100-content") {
		t.Error("warning refers to a nonexistent base-100-content color")
	}
	for _, base := range []string{"#ffffff", "#111111"} {
		result := GenerateTheme(ThemeInput{Primary: "#2563eb", Base100: base})
		if len(result.Warnings) != 0 {
			t.Errorf("high-contrast base %s produced warnings: %v", base, result.Warnings)
		}
	}
}
