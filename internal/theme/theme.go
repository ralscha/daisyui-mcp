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

var oklchRegex = regexp.MustCompile(`(?i)oklch\(\s*([\d.]+)%?\s+([\d.]+)\s+([\d.]+)\s*\)`)

func parseColor(s string, defaultColor colorful.Color) colorful.Color {
	if s == "" {
		return defaultColor
	}
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(s, ";")

	if matches := oklchRegex.FindStringSubmatch(s); len(matches) == 4 {
		l, _ := strconv.ParseFloat(matches[1], 64)
		if strings.Contains(matches[0], "%") {
			l = l / 100.0
		} else if l > 1.0 {
			l = l / 100.0
		}
		c, _ := strconv.ParseFloat(matches[2], 64)
		h, _ := strconv.ParseFloat(matches[3], 64)
		return colorful.OkLch(l, c, h)
	}

	if !strings.HasPrefix(s, "#") && !strings.HasPrefix(s, "rgb") && !strings.HasPrefix(s, "hsl") && !strings.HasPrefix(s, "oklch") {
		s = "#" + s
	}
	c, err := colorful.Hex(s)
	if err == nil {
		return c
	}
	return defaultColor
}

func generateContentColor(c colorful.Color) colorful.Color {
	l, _, h := c.OkLch()
	if l > 0.6 {
		return colorful.OkLch(0.2, 0.05, h)
	}
	return colorful.OkLch(0.98, 0.01, h)
}

func formatOklch(c colorful.Color) string {
	l, chroma, h := c.OkLch()
	if chroma < 0.0001 || math.IsNaN(h) {
		h = 0
	}
	lStr := strconv.FormatFloat(math.Round(l*100), 'f', -1, 64)
	cStr := strconv.FormatFloat(math.Round(chroma*1000)/1000, 'f', -1, 64)
	hStr := strconv.FormatFloat(math.Round(h*1000)/1000, 'f', -1, 64)

	return fmt.Sprintf("oklch(%s%% %s %s)", lStr, cStr, hStr)
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func GenerateThemeCSS(input ThemeInput) string {
	primary := parseColor(input.Primary, colorful.OkLch(0.55, 0.3, 240))
	secondary := parseColor(input.Secondary, colorful.OkLch(0.70, 0.25, 200))
	accent := parseColor(input.Accent, colorful.OkLch(0.65, 0.25, 160))
	neutral := parseColor(input.Neutral, colorful.OkLch(0.50, 0.05, 240))
	base100 := parseColor(input.Base100, colorful.OkLch(0.98, 0.02, 240))
	info := parseColor(input.Info, colorful.OkLch(0.70, 0.2, 220))
	success := parseColor(input.Success, colorful.OkLch(0.65, 0.25, 140))
	warning := parseColor(input.Warning, colorful.OkLch(0.80, 0.25, 80))
	errorColor := parseColor(input.Error, colorful.OkLch(0.65, 0.3, 30))

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
	sb.WriteString("  name: \"mytheme\";\n")
	sb.WriteString("  default: false;\n")
	sb.WriteString("  prefersdark: false;\n")
	if b1L > 0.5 {
		sb.WriteString("  color-scheme: \"light\";\n")
	} else {
		sb.WriteString("  color-scheme: \"dark\";\n")
	}

	fmt.Fprintf(&sb, "  --color-base-100: %s;\n", formatOklch(base100))
	fmt.Fprintf(&sb, "  --color-base-200: %s;\n", formatOklch(base200))
	fmt.Fprintf(&sb, "  --color-base-300: %s;\n", formatOklch(base300))
	fmt.Fprintf(&sb, "  --color-base-content: %s;\n", formatOklch(generateContentColor(base100)))

	fmt.Fprintf(&sb, "  --color-primary: %s;\n", formatOklch(primary))
	fmt.Fprintf(&sb, "  --color-primary-content: %s;\n", formatOklch(generateContentColor(primary)))

	fmt.Fprintf(&sb, "  --color-secondary: %s;\n", formatOklch(secondary))
	fmt.Fprintf(&sb, "  --color-secondary-content: %s;\n", formatOklch(generateContentColor(secondary)))

	fmt.Fprintf(&sb, "  --color-accent: %s;\n", formatOklch(accent))
	fmt.Fprintf(&sb, "  --color-accent-content: %s;\n", formatOklch(generateContentColor(accent)))

	fmt.Fprintf(&sb, "  --color-neutral: %s;\n", formatOklch(neutral))
	fmt.Fprintf(&sb, "  --color-neutral-content: %s;\n", formatOklch(generateContentColor(neutral)))

	fmt.Fprintf(&sb, "  --color-info: %s;\n", formatOklch(info))
	fmt.Fprintf(&sb, "  --color-info-content: %s;\n", formatOklch(generateContentColor(info)))

	fmt.Fprintf(&sb, "  --color-success: %s;\n", formatOklch(success))
	fmt.Fprintf(&sb, "  --color-success-content: %s;\n", formatOklch(generateContentColor(success)))

	fmt.Fprintf(&sb, "  --color-warning: %s;\n", formatOklch(warning))
	fmt.Fprintf(&sb, "  --color-warning-content: %s;\n", formatOklch(generateContentColor(warning)))

	fmt.Fprintf(&sb, "  --color-error: %s;\n", formatOklch(errorColor))
	fmt.Fprintf(&sb, "  --color-error-content: %s;\n", formatOklch(generateContentColor(errorColor)))

	radiusSelector := defaultString(input.RadiusSelector, "0.25rem")
	radiusField := defaultString(input.RadiusField, "0.25rem")
	radiusBox := defaultString(input.RadiusBox, "0.5rem")
	sizeSelector := defaultString(input.SizeSelector, "0.25rem")
	sizeField := defaultString(input.SizeField, "0.25rem")
	border := defaultString(input.Border, "1px")
	depth := defaultString(input.Depth, "0")
	noise := defaultString(input.Noise, "0")

	fmt.Fprintf(&sb, "  --radius-selector: %s;\n", radiusSelector)
	fmt.Fprintf(&sb, "  --radius-field: %s;\n", radiusField)
	fmt.Fprintf(&sb, "  --radius-box: %s;\n", radiusBox)
	fmt.Fprintf(&sb, "  --size-selector: %s;\n", sizeSelector)
	fmt.Fprintf(&sb, "  --size-field: %s;\n", sizeField)
	fmt.Fprintf(&sb, "  --border: %s;\n", border)
	fmt.Fprintf(&sb, "  --depth: %s;\n", depth)
	fmt.Fprintf(&sb, "  --noise: %s;\n", noise)
	sb.WriteString("}\n")

	return sb.String()
}
