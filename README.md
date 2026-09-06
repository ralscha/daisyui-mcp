# DaisyUI MCP Server

## Features

- **Single Binary** – One self-contained executable, no Python/Node/runtime required
- **Embedded Docs** – Component Markdown files are baked into the binary at build time — deploy just the `.exe`
- **Theme Helpers** – Generate daisyUI 5 theme CSS from color values or from an image palette
- **Offline by Default** – Runtime component and guide docs are served from embedded files unless explicitly overridden

## MCP Tools

| Tool | Description |
|------|-------------|
| `list_components` | List all available DaisyUI components with names and brief descriptions |
| `search_components` | Search component names and descriptions, ordered by relevance |
| `get_short_doc` | Get a concise summary for a component: CSS classes, HTML syntax, and usage rules |
| `get_detailed_doc` | Get paged detailed documentation, optionally selecting an exact Markdown section |
| `get_color_palette` | List all daisyUI semantic colors, modifier class patterns, and usage rules |
| `get_customize_docs` | Get the daisyUI customization guide (CSS, Tailwind, and daisyUI conventions) |
| `get_config_docs` | Get the daisyUI configuration reference (themes, logs, prefix, and more) |
| `get_themes_docs` | Get the daisyUI themes documentation (built-in themes, applying, customizing, creating) |
| `get_base_style_docs` | Get the daisyUI base/reset styles documentation |
| `get_utilities_docs` | Get the daisyUI utility classes and CSS variables documentation |
| `get_layout_typography_docs` | Get the daisyUI layout and typography documentation |
| `generate_theme` | Generate a complete daisyUI 5 custom theme CSS based on hex or OKLCH colors |
| `generate_theme_from_image` | Extract a color palette from a local or remote image and generate a complete daisyUI 5 custom theme CSS |
| `list_recipes` | List curated recipes for common daisyUI application compositions |
| `get_recipe` | Get an accessible composition recipe with HTML and implementation notes |

Tool responses retain human-readable text and also include structured content. Component listings are returned as arrays; documentation includes page metadata; and generated themes include CSS, semantic color pairs, contrast ratios, and warnings.

### Detailed documentation pages

`get_detailed_doc` returns at most 12,000 characters by default so a large component page does not unexpectedly consume the client's context window. Pass `page` to continue, `page_size` to request up to 24,000 characters, or `section` to select an exact Markdown heading. The structured response reports `available_sections`, `total_pages`, and `has_more`.

For example:

```json
{
  "name": "button",
  "section": "Button sizes",
  "page": 1,
  "page_size": 8000
}
```

## MCP Resources

Clients that support MCP resources can browse or read the complete embedded documentation directly. Tools remain available for compatibility and context-efficient paging.

| URI template | Content |
|-------------|---------|
| `daisyui://components/{name}` | Full component documentation |
| `daisyui://guides/{name}` | Colors, configuration, themes, utilities, and customization guides |
| `daisyui://recipes/{name}` | Curated composition recipes |

The 20 included recipes cover authentication, application navigation, dialogs, dashboards, accessible forms, pricing, landing pages, FAQs, data tables, checkout, uploads, notifications, and common empty or consent states. Every recipe is independently authored for this project, targets daisyUI 5, and records its official documentation references in a provenance section.


## Installation

### Option A — Build from source

Requires [Go](https://go.dev) matching `go.mod` and [Task](https://taskfile.dev).

```bash
git clone https://github.com/ralscha/daisyui-mcp.git
cd daisyui-mcp
task build
```

`task build` refreshes the generated daisyUI documentation before compiling, so it needs network access. The resulting `bin/daisyui-server` (or `bin/daisyui-server.exe` on Windows) is the only file you need to deploy.

#### Other task commands

| Command | Description |
|---------|-------------|
| `task build` | Build all binaries into `bin/` |
| `task update` | Fetch latest DaisyUI docs then rebuild |
| `task test` | Run all tests |
| `task format` | Format all Go source files |
| `task lint` | Run golangci-lint (requires Docker) |

### Option B — Download a pre-built binary

Download the latest release from the [Releases](../../releases) page and place the binary somewhere on your `PATH`.

## Configuration

Add the server to your AI assistant's MCP configuration. Because the binary is self-contained, the configuration is minimal.

```json
{
  "servers": {
    "daisyui": {
      "command": "/path/to/bin/daisyui-server"
    }
  }
}
```

### Environment variables

| Variable | Description |
|----------|-------------|
| `DAISYUI_COMPONENTS_DIR` | Override the embedded component summaries with files from this directory |
| `DAISYUI_DOCS_DIR` | Override the embedded detailed docs with files from this directory |

Override directories should contain Markdown files named after component slugs, for example `button.md` or `card.md`.

## Theme generation

`generate_theme` accepts hex colors such as `#ff0000` or `ff0000`, plus OKLCH values such as `oklch(60% 0.2 30)`. It derives matching WCAG AA `*-content` colors and `base-200`/`base-300` values automatically. Both theme tools accept `name`, `default`, and `prefers_dark` to configure the generated daisyUI theme declaration.

`generate_theme_from_image` accepts exactly one of `image_path` or `image_url`. Image reads are bounded to keep palette extraction predictable. Remote URLs must use HTTP(S), may not contain credentials, and must resolve to public network addresses; redirects receive the same checks.

Theme tools return the generated CSS as text and as structured data. Structured results include theme metadata, each semantic color and its `*-content` partner, their contrast ratio, extracted image colors when applicable, and validation warnings.

## Disclaimer

DaisyUI has an official [Blueprint MCP](https://daisyui.com/blueprint/) with premium features.

This project is **not** that. It's a free, open-source alternative using their publicly available documentation.

If you use daisyUI components in a commercial project, please consider supporting the creators by purchasing their official MCP server.

## License

[MIT](LICENSE)
