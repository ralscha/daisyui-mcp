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
