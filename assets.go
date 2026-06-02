package daisyuimcp

import "embed"

//go:embed components
var ComponentsFS embed.FS

//go:embed docs
var DocsFS embed.FS

//go:embed colors.md
var ColorsData []byte

//go:embed guide/customize.md
var GuideCustomize []byte

//go:embed guide/config.md
var GuideConfig []byte

//go:embed guide/themes.md
var GuideThemes []byte

//go:embed guide/base.md
var GuideBase []byte

//go:embed guide/utilities.md
var GuideUtilities []byte

//go:embed guide/layout-and-typography.md
var GuideLayoutTypography []byte
