# DaisyUI MCP Server

## Features

- **Single Binary** – One self-contained executable, no Python/Node/runtime required
- **Embedded Docs** – Component Markdown files are baked into the binary at build time — deploy just the `.exe`

## MCP Tools

| Tool | Description |
|------|-------------|
| `list_components` | List all available DaisyUI components with names and brief descriptions |
| `get_short_doc` | Get a concise summary for a component: CSS classes, HTML syntax, and usage rules |
| `get_detailed_doc` | Get the full documentation page for a component, including all variants and advanced usage |
| `get_color_palette` | List all daisyUI semantic colors, modifier class patterns, and usage rules |
| `get_customize_docs` | Get the daisyUI customization guide (CSS, Tailwind, and daisyUI conventions) |
| `get_config_docs` | Get the daisyUI configuration reference (themes, logs, prefix, and more) |
| `get_themes_docs` | Get the daisyUI themes documentation (built-in themes, applying, customizing, creating) |
| `get_base_style_docs` | Get the daisyUI base/reset styles documentation |
| `get_utilities_docs` | Get the daisyUI utility classes and CSS variables documentation |
| `get_layout_typography_docs` | Get the daisyUI layout and typography documentation |


## Installation

### Option A — Build from source

Requires [Go](https://go.dev) and [Task](https://taskfile.dev).

```bash
git clone https://github.com/your-org/daisyui-mcp.git
cd daisyui-mcp

# Build both binaries into bin/
task build
```

The resulting `bin/daisyui-server` (or `bin/daisyui-server.exe` on Windows) is the only file you need to deploy.

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

## Disclaimer

DaisyUI has an official [Blueprint MCP](https://daisyui.com/blueprint/) with premium features.

This project is **not** that. It's a free, open-source alternative using their publicly available documentation.

If you use daisyUI components in a commercial project, please consider supporting the creators by purchasing their official MCP server.

## License

[MIT](LICENSE)
