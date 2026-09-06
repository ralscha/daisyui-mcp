package theme

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
)

type ThemeInput struct {
	Name        string `json:"name,omitempty" jsonschema:"Theme name. Defaults to mytheme."`
	Default     bool   `json:"default,omitempty" jsonschema:"Whether this is the default daisyUI theme."`
	PrefersDark bool   `json:"prefers_dark,omitempty" jsonschema:"Whether this theme should be selected for a dark system preference."`

	Primary   string `json:"primary" jsonschema:"Primary color (hex or OKLCH). Required."`
	Secondary string `json:"secondary,omitempty" jsonschema:"Secondary color (hex or OKLCH). Optional."`
	Accent    string `json:"accent,omitempty" jsonschema:"Accent color (hex or OKLCH). Optional."`
	Neutral   string `json:"neutral,omitempty" jsonschema:"Neutral color (hex or OKLCH). Optional."`
	Base100   string `json:"base_100,omitempty" jsonschema:"Base 100 background color (hex or OKLCH). Optional."`
	Info      string `json:"info,omitempty" jsonschema:"Info color (hex or OKLCH). Optional."`
	Success   string `json:"success,omitempty" jsonschema:"Success color (hex or OKLCH). Optional."`
	Warning   string `json:"warning,omitempty" jsonschema:"Warning color (hex or OKLCH). Optional."`
	Error     string `json:"error,omitempty" jsonschema:"Error color (hex or OKLCH). Optional."`

	RadiusSelector string `json:"radius_selector,omitempty" jsonschema:"Radius selector. Optional."`
	RadiusField    string `json:"radius_field,omitempty" jsonschema:"Radius field. Optional."`
	RadiusBox      string `json:"radius_box,omitempty" jsonschema:"Radius box. Optional."`
	SizeSelector   string `json:"size_selector,omitempty" jsonschema:"Size selector. Optional."`
	SizeField      string `json:"size_field,omitempty" jsonschema:"Size field. Optional."`
	Border         string `json:"border,omitempty" jsonschema:"Border. Optional."`
	Depth          string `json:"depth,omitempty" jsonschema:"Depth. Optional."`
	Noise          string `json:"noise,omitempty" jsonschema:"Noise. Optional."`
}

var (
	oklchRegex       = regexp.MustCompile(`(?i)^oklch\(\s*([\d.]+)(%)?\s+([\d.]+)\s+([\d.]+)\s*\)$`)
	themeNameRegex   = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,63}$`)
	cssLengthRegex   = regexp.MustCompile(`(?i)^(?:0(?:\.0+)?|(?:\d+(?:\.\d+)?|\.\d+)(?:px|rem|em|ex|ch|cap|ic|lh|rlh|vw|vh|vi|vb|vmin|vmax|svw|svh|lvw|lvh|dvw|dvh|cm|mm|q|in|pc|pt|%))$`)
	themeToggleRegex = regexp.MustCompile(`^(?:0(?:\.0+)?|1(?:\.0+)?)$`)
)

func parseColor(s string, defaultColor colorful.Color) colorful.Color {
	c, _ := parseColorWithStatus(s, defaultColor)
	return c
}

func parseColorWithStatus(s string, defaultColor colorful.Color) (colorful.Color, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return defaultColor, true
	}
	s = strings.TrimSuffix(s, ";")

	if matches := oklchRegex.FindStringSubmatch(s); len(matches) == 5 {
		l, lErr := strconv.ParseFloat(matches[1], 64)
		if matches[2] == "%" {
			l = l / 100.0
		} else if l > 1.0 {
			l = l / 100.0
		}
		c, cErr := strconv.ParseFloat(matches[3], 64)
		h, hErr := strconv.ParseFloat(matches[4], 64)
		if lErr != nil || cErr != nil || hErr != nil || l < 0 || l > 1 || c < 0 || h < 0 || h > 360 {
			return defaultColor, false
		}
		return colorful.OkLch(l, c, h), true
	}

	if !strings.HasPrefix(s, "#") && !strings.HasPrefix(s, "rgb") && !strings.HasPrefix(s, "hsl") && !strings.HasPrefix(s, "oklch") {
		s = "#" + s
	}
	c, err := colorful.Hex(s)
	if err == nil {
		return c, true
	}
	return defaultColor, false
}

func generateContentColor(c colorful.Color) colorful.Color {
	c = cssColor(c)
	_, _, h := c.OkLch()
	if math.IsNaN(h) {
		h = 0
	}

	dark := cssColor(colorful.OkLch(0.2, 0.03, h))
	light := cssColor(colorful.OkLch(0.98, 0.01, h))
	best := dark
	if contrastRatio(c, light) > contrastRatio(c, dark) {
		best = light
	}
	if contrastRatio(c, best) >= 4.5 {
		return best
	}

	// Near the middle of the luminance range, tinted foregrounds can both miss
	// the WCAG AA target. Pure black or white always gives the stronger fallback.
	black := colorful.Color{}
	white := colorful.Color{R: 1, G: 1, B: 1}
	if contrastRatio(c, white) > contrastRatio(c, black) {
		return white
	}
	return black
}

func cssColor(c colorful.Color) colorful.Color {
	l, chroma, h := c.OkLch()
	if chroma < 0.0001 || math.IsNaN(h) {
		h = 0
	}
	return colorful.OkLch(
		math.Round(l*100_000)/100_000,
		math.Round(chroma*1_000)/1_000,
		math.Round(h*1_000)/1_000,
	)
}

func formatOklch(c colorful.Color) string {
	l, chroma, h := cssColor(c).OkLch()
	lStr := strconv.FormatFloat(math.Round(l*100_000)/1_000, 'f', -1, 64)
	cStr := strconv.FormatFloat(chroma, 'f', -1, 64)
	hStr := strconv.FormatFloat(h, 'f', -1, 64)

	return fmt.Sprintf("oklch(%s%% %s %s)", lStr, cStr, hStr)
}

func validatedThemeOption(value, fallback, name string, pattern *regexp.Regexp, warnings *[]string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	if pattern.MatchString(value) {
		return value
	}
	*warnings = append(*warnings, fmt.Sprintf("Invalid %s value %q; the default value %q was used.", name, value, fallback))
	return fallback
}

type GeneratedColor struct {
	Name          string  `json:"name"`
	Value         string  `json:"value"`
	ContentValue  string  `json:"content_value"`
	ContrastRatio float64 `json:"contrast_ratio"`
}

type GeneratedTheme struct {
	Name        string           `json:"name"`
	Default     bool             `json:"default"`
	PrefersDark bool             `json:"prefers_dark"`
	CSS         string           `json:"css"`
	Colors      []GeneratedColor `json:"colors"`
	Warnings    []string         `json:"warnings"`
}

func GenerateTheme(input ThemeInput) GeneratedTheme {
	type colorSpec struct {
		name     string
		input    string
		fallback colorful.Color
	}

	specs := []colorSpec{
		{"primary", input.Primary, colorful.OkLch(0.55, 0.3, 240)},
		{"secondary", input.Secondary, colorful.OkLch(0.70, 0.25, 200)},
		{"accent", input.Accent, colorful.OkLch(0.65, 0.25, 160)},
		{"neutral", input.Neutral, colorful.OkLch(0.50, 0.05, 240)},
		{"base-100", input.Base100, colorful.OkLch(0.98, 0.02, 240)},
		{"info", input.Info, colorful.OkLch(0.70, 0.2, 220)},
		{"success", input.Success, colorful.OkLch(0.65, 0.25, 140)},
		{"warning", input.Warning, colorful.OkLch(0.80, 0.25, 80)},
		{"error", input.Error, colorful.OkLch(0.65, 0.3, 30)},
	}

	parsed := make(map[string]colorful.Color, len(specs))
	warnings := make([]string, 0)
	for _, spec := range specs {
		color, valid := parseColorWithStatus(spec.input, spec.fallback)
		parsed[spec.name] = color
		if spec.input != "" && !valid {
			warnings = append(warnings, fmt.Sprintf("Invalid %s color %q; the default value was used.", spec.name, spec.input))
		}
	}
	themeName := validatedThemeOption(input.Name, "mytheme", "theme name", themeNameRegex, &warnings)

	primary := parsed["primary"]
	secondary := parsed["secondary"]
	accent := parsed["accent"]
	neutral := parsed["neutral"]
	base100 := parsed["base-100"]
	info := parsed["info"]
	success := parsed["success"]
	warning := parsed["warning"]
	errorColor := parsed["error"]

	b1L, b1C, b1H := base100.OkLch()

	var base200, base300 colorful.Color
	if b1L > 0.5 {
		base200 = colorful.OkLch(max(0, b1L-0.05), b1C+0.01, b1H)
		base300 = colorful.OkLch(max(0, b1L-0.10), b1C+0.02, b1H)
	} else {
		base200 = colorful.OkLch(min(1, b1L+0.05), b1C+0.01, b1H)
		base300 = colorful.OkLch(min(1, b1L+0.10), b1C+0.02, b1H)
	}

	var sb strings.Builder
	sb.WriteString("@plugin \"daisyui/theme\" {\n")
	fmt.Fprintf(&sb, "  name: %q;\n", themeName)
	fmt.Fprintf(&sb, "  default: %t;\n", input.Default)
	fmt.Fprintf(&sb, "  prefersdark: %t;\n", input.PrefersDark)
	if b1L > 0.5 {
		sb.WriteString("  color-scheme: \"light\";\n")
	} else {
		sb.WriteString("  color-scheme: \"dark\";\n")
	}

	fmt.Fprintf(&sb, "  --color-base-100: %s;\n", formatOklch(base100))
	fmt.Fprintf(&sb, "  --color-base-200: %s;\n", formatOklch(base200))
	fmt.Fprintf(&sb, "  --color-base-300: %s;\n", formatOklch(base300))
	baseContent := generateContentColor(base100)
	primaryContent := generateContentColor(primary)
	secondaryContent := generateContentColor(secondary)
	accentContent := generateContentColor(accent)
	neutralContent := generateContentColor(neutral)
	infoContent := generateContentColor(info)
	successContent := generateContentColor(success)
	warningContent := generateContentColor(warning)
	errorContent := generateContentColor(errorColor)

	fmt.Fprintf(&sb, "  --color-base-content: %s;\n", formatOklch(baseContent))

	fmt.Fprintf(&sb, "  --color-primary: %s;\n", formatOklch(primary))
	fmt.Fprintf(&sb, "  --color-primary-content: %s;\n", formatOklch(primaryContent))

	fmt.Fprintf(&sb, "  --color-secondary: %s;\n", formatOklch(secondary))
	fmt.Fprintf(&sb, "  --color-secondary-content: %s;\n", formatOklch(secondaryContent))

	fmt.Fprintf(&sb, "  --color-accent: %s;\n", formatOklch(accent))
	fmt.Fprintf(&sb, "  --color-accent-content: %s;\n", formatOklch(accentContent))

	fmt.Fprintf(&sb, "  --color-neutral: %s;\n", formatOklch(neutral))
	fmt.Fprintf(&sb, "  --color-neutral-content: %s;\n", formatOklch(neutralContent))

	fmt.Fprintf(&sb, "  --color-info: %s;\n", formatOklch(info))
	fmt.Fprintf(&sb, "  --color-info-content: %s;\n", formatOklch(infoContent))

	fmt.Fprintf(&sb, "  --color-success: %s;\n", formatOklch(success))
	fmt.Fprintf(&sb, "  --color-success-content: %s;\n", formatOklch(successContent))

	fmt.Fprintf(&sb, "  --color-warning: %s;\n", formatOklch(warning))
	fmt.Fprintf(&sb, "  --color-warning-content: %s;\n", formatOklch(warningContent))

	fmt.Fprintf(&sb, "  --color-error: %s;\n", formatOklch(errorColor))
	fmt.Fprintf(&sb, "  --color-error-content: %s;\n", formatOklch(errorContent))

	radiusSelector := validatedThemeOption(input.RadiusSelector, "0.25rem", "radius_selector", cssLengthRegex, &warnings)
	radiusField := validatedThemeOption(input.RadiusField, "0.25rem", "radius_field", cssLengthRegex, &warnings)
	radiusBox := validatedThemeOption(input.RadiusBox, "0.5rem", "radius_box", cssLengthRegex, &warnings)
	sizeSelector := validatedThemeOption(input.SizeSelector, "0.25rem", "size_selector", cssLengthRegex, &warnings)
	sizeField := validatedThemeOption(input.SizeField, "0.25rem", "size_field", cssLengthRegex, &warnings)
	border := validatedThemeOption(input.Border, "1px", "border", cssLengthRegex, &warnings)
	depth := validatedThemeOption(input.Depth, "0", "depth", themeToggleRegex, &warnings)
	noise := validatedThemeOption(input.Noise, "0", "noise", themeToggleRegex, &warnings)

	fmt.Fprintf(&sb, "  --radius-selector: %s;\n", radiusSelector)
	fmt.Fprintf(&sb, "  --radius-field: %s;\n", radiusField)
	fmt.Fprintf(&sb, "  --radius-box: %s;\n", radiusBox)
	fmt.Fprintf(&sb, "  --size-selector: %s;\n", sizeSelector)
	fmt.Fprintf(&sb, "  --size-field: %s;\n", sizeField)
	fmt.Fprintf(&sb, "  --border: %s;\n", border)
	fmt.Fprintf(&sb, "  --depth: %s;\n", depth)
	fmt.Fprintf(&sb, "  --noise: %s;\n", noise)
	sb.WriteString("}\n")

	pairs := []struct {
		name    string
		color   colorful.Color
		content colorful.Color
	}{
		{"base-100", base100, baseContent},
		{"primary", primary, primaryContent},
		{"secondary", secondary, secondaryContent},
		{"accent", accent, accentContent},
		{"neutral", neutral, neutralContent},
		{"info", info, infoContent},
		{"success", success, successContent},
		{"warning", warning, warningContent},
		{"error", errorColor, errorContent},
	}
	colors := make([]GeneratedColor, 0, len(pairs))
	for _, pair := range pairs {
		color := cssColor(pair.color)
		content := cssColor(pair.content)
		ratio := contrastRatio(color, content)
		colors = append(colors, GeneratedColor{
			Name:          pair.name,
			Value:         formatOklch(color),
			ContentValue:  formatOklch(content),
			ContrastRatio: math.Round(ratio*100) / 100,
		})
		if ratio < 4.5 {
			warnings = append(warnings, fmt.Sprintf("%s and %s-content have a contrast ratio of %.2f:1, below the WCAG AA text target of 4.5:1.", pair.name, pair.name, ratio))
		}
	}

	return GeneratedTheme{
		Name:        themeName,
		Default:     input.Default,
		PrefersDark: input.PrefersDark,
		CSS:         sb.String(),
		Colors:      colors,
		Warnings:    warnings,
	}
}

func GenerateThemeCSS(input ThemeInput) string {
	return GenerateTheme(input).CSS
}

func contrastRatio(a, b colorful.Color) float64 {
	ar, ag, ab := a.Clamped().LinearRgb()
	br, bg, bb := b.Clamped().LinearRgb()
	aLuminance := 0.2126*ar + 0.7152*ag + 0.0722*ab
	bLuminance := 0.2126*br + 0.7152*bg + 0.0722*bb
	if aLuminance < bLuminance {
		aLuminance, bLuminance = bLuminance, aLuminance
	}
	return (aLuminance + 0.05) / (bLuminance + 0.05)
}
