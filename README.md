# Chotko

A terminal UI for Zabbix 7.x built with Go and [BubbleTea](https://github.com/charmbracelet/bubbletea).

The name comes from the Russian slang word "чётко" (chotko), meaning "precise" or "on point" - fitting for a monitoring tool.

## Features

- View active Zabbix alerts with severity-based color coding
- Acknowledge problems directly from the terminal
- Host status overview (OK, Problem, Unknown, Maintenance)
- Multiple built-in themes (Nord, Dracula, Gruvbox, Catppuccin, Tokyo Night, Solarized)
- Custom theme support via YAML
- Vim-style keyboard navigation
- Filter alerts by severity or text
- Auto-refresh with configurable interval

## Installation

```bash
go install github.com/harpchad/chotko/cmd/chotko@latest
```

Or build from source:

```bash
git clone https://github.com/harpchad/chotko.git
cd chotko
go build -o chotko ./cmd/chotko
```

## Usage

```bash
# Run with interactive setup wizard (first run)
chotko

# Connect with API token (recommended)
chotko -s https://zabbix.example.com -t YOUR_API_TOKEN

# Connect with username/password
chotko -s https://zabbix.example.com -u Admin -p password

# Use a specific theme
chotko --theme dracula

# Show only high severity alerts
chotko --min-severity 4
```

## Configuration

Configuration is stored in `~/.config/chotko/config.yaml`:

```yaml
server:
  url: "https://zabbix.example.com"

auth:
  # API Token (recommended for Zabbix 5.4+)
  token: "your-api-token-here"
  # Or use username/password
  # username: "Admin"
  # password: "zabbix"

display:
  refresh_interval: 30  # seconds
  min_severity: 0       # 0=all, 1-5=filter
  theme: "nord"
```

## Key Bindings

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `PgDn` / `Ctrl+D` | Page down |
| `PgUp` / `Ctrl+U` | Page up |
| `g` / `Home` | Go to top |
| `G` / `End` | Go to bottom |
| `Tab` | Next pane |
| `Shift+Tab` | Previous pane |
| `a` | Acknowledge selected alert |
| `A` | Acknowledge with message |
| `r` | Refresh data |
| `/` | Filter mode |
| `0-5` | Filter by minimum severity |
| `Ctrl+L` | Clear filter |
| `:` | Command mode |
| `?` | Show help |
| `q` | Quit |

## Themes

Built-in themes:
- `default` - Classic Zabbix-inspired colors
- `nord` - Arctic, cool-toned (default)
- `dracula` - Dark purple/pink aesthetic
- `gruvbox` - Retro warm tones
- `catppuccin` - Soothing pastels
- `tokyonight` - Cool blues and purples
- `solarized` - Precision-balanced

### Custom Themes

Create a custom theme in `~/.config/chotko/themes/mytheme.yaml`:

```yaml
name: "mytheme"
description: "My custom theme"

colors:
  disaster: "#FF0000"
  high: "#FF6600"
  average: "#FFAA00"
  warning: "#FFCC00"
  information: "#6699FF"
  not_classified: "#999999"
  ok: "#00CC00"
  unknown: "#AAAAAA"
  maintenance: "#AA66FF"
  primary: "#6699FF"
  secondary: "#00CC00"
  background: "#1a1a1a"
  foreground: "#EEEEEE"
  muted: "#666666"
  border: "#444444"
  focused_border: "#6699FF"
  highlight: "#333366"
  surface: "#2a2a2a"
```

Then use it with `--theme mytheme` or set in config.

## Requirements

- Zabbix 7.x (API compatibility)
- Terminal with true color support recommended

## License

MIT
