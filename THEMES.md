# Themes

Chotko includes several built-in themes to match your terminal aesthetic. You can also create custom themes.

## Built-in Themes

### Default

Classic Zabbix-inspired colors with familiar severity color coding.

```bash
chotko --theme default
```

### Nord

Arctic, cool-toned palette based on the [Nord](https://www.nordtheme.com/) color scheme.

```bash
chotko --theme nord
```

### Dracula

Dark purple/pink aesthetic based on the [Dracula](https://draculatheme.com/) color scheme.

```bash
chotko --theme dracula
```

### Catppuccin

Soothing pastel colors based on [Catppuccin Mocha](https://github.com/catppuccin/catppuccin).

```bash
chotko --theme catppuccin
```

### Tokyo Night

Cool blues and purples based on [Tokyo Night](https://github.com/folke/tokyonight.nvim).

```bash
chotko --theme tokyonight
```

### Gruvbox

Retro warm tones based on [Gruvbox](https://github.com/morhetz/gruvbox).

```bash
chotko --theme gruvbox
```

### Solarized

Precision-balanced colors based on [Solarized Dark](https://ethanschoonover.com/solarized/).

```bash
chotko --theme solarized
```

## Custom Themes

Create custom themes by adding YAML files to `~/.config/chotko/themes/`.

### Example Custom Theme

Create `~/.config/chotko/themes/mytheme.yaml`:

```yaml
name: "mytheme"
description: "My custom theme"

colors:
  # Severity colors
  disaster: "#FF0000"
  high: "#FF6600"
  average: "#FFAA00"
  warning: "#FFCC00"
  information: "#6699FF"
  not_classified: "#999999"

  # Status colors
  ok: "#00CC00"
  unknown: "#AAAAAA"
  maintenance: "#AA66FF"

  # UI colors
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

Then use it:

```bash
chotko --theme mytheme
```

Or set it in your config file (`~/.config/chotko/config.yaml`):

```yaml
display:
  theme: "mytheme"
```

## Color Reference

### Severity Colors

| Severity | Description | Usage |
|----------|-------------|-------|
| `disaster` | Critical alerts | Severity 5 |
| `high` | High severity | Severity 4 |
| `average` | Average severity | Severity 3 |
| `warning` | Warning level | Severity 2 |
| `information` | Informational | Severity 1 |
| `not_classified` | Unclassified | Severity 0 |

### Status Colors

| Status | Description |
|--------|-------------|
| `ok` | Host/service is healthy |
| `unknown` | Status cannot be determined |
| `maintenance` | In maintenance mode |

### UI Colors

| Color | Description |
|-------|-------------|
| `primary` | Primary accent color (active tabs, links) |
| `secondary` | Secondary accent (success states) |
| `background` | Main background color |
| `foreground` | Primary text color |
| `muted` | Subdued text (labels, hints) |
| `border` | Inactive borders |
| `focused_border` | Active/focused pane borders |
| `highlight` | Selected row background |
| `surface` | Elevated surface backgrounds |
