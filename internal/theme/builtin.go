// Package theme provides theming support with built-in and custom themes.
package theme

import "github.com/charmbracelet/lipgloss"

// Built-in theme names
const (
	ThemeDefault    = "default"
	ThemeNord       = "nord"
	ThemeDracula    = "dracula"
	ThemeGruvbox    = "gruvbox"
	ThemeCatppuccin = "catppuccin"
	ThemeTokyoNight = "tokyonight"
	ThemeSolarized  = "solarized"
)

// BuiltinThemes returns all available built-in themes.
func BuiltinThemes() map[string]*Theme {
	return map[string]*Theme{
		ThemeDefault:    DefaultTheme(),
		ThemeNord:       NordTheme(),
		ThemeDracula:    DraculaTheme(),
		ThemeGruvbox:    GruvboxTheme(),
		ThemeCatppuccin: CatppuccinTheme(),
		ThemeTokyoNight: TokyoNightTheme(),
		ThemeSolarized:  SolarizedTheme(),
	}
}

// BuiltinThemeNames returns a list of all built-in theme names.
func BuiltinThemeNames() []string {
	return []string{
		ThemeDefault,
		ThemeNord,
		ThemeDracula,
		ThemeGruvbox,
		ThemeCatppuccin,
		ThemeTokyoNight,
		ThemeSolarized,
	}
}

// DefaultTheme returns the default Zabbix-inspired theme.
func DefaultTheme() *Theme {
	return &Theme{
		Name:        ThemeDefault,
		Description: "Classic Zabbix-inspired colors",
		Colors: ColorPalette{
			// Severity colors (classic Zabbix)
			Disaster:      lipgloss.Color("#FF0000"), // Red
			High:          lipgloss.Color("#FF6600"), // Orange
			Average:       lipgloss.Color("#FFAA00"), // Yellow-Orange
			Warning:       lipgloss.Color("#FFCC00"), // Yellow
			Information:   lipgloss.Color("#6699FF"), // Blue
			NotClassified: lipgloss.Color("#999999"), // Gray

			// Status colors
			OK:          lipgloss.Color("#00CC00"), // Green
			Unknown:     lipgloss.Color("#AAAAAA"), // Light gray
			Maintenance: lipgloss.Color("#AA66FF"), // Purple

			// UI colors
			Primary:       lipgloss.Color("#6699FF"),
			Secondary:     lipgloss.Color("#00CC00"),
			Background:    lipgloss.Color("#1a1a1a"),
			Foreground:    lipgloss.Color("#EEEEEE"),
			Muted:         lipgloss.Color("#666666"),
			Border:        lipgloss.Color("#444444"),
			FocusedBorder: lipgloss.Color("#6699FF"),
			Highlight:     lipgloss.Color("#333366"),
			Surface:       lipgloss.Color("#2a2a2a"),
		},
	}
}

// NordTheme returns the Nord color theme.
// https://www.nordtheme.com/
func NordTheme() *Theme {
	return &Theme{
		Name:        ThemeNord,
		Description: "Arctic, cool-toned Nord palette",
		Colors: ColorPalette{
			// Severity colors (Aurora)
			Disaster:      lipgloss.Color("#BF616A"), // nord11 - red
			High:          lipgloss.Color("#D08770"), // nord12 - orange
			Average:       lipgloss.Color("#EBCB8B"), // nord13 - yellow
			Warning:       lipgloss.Color("#88C0D0"), // nord8 - cyan
			Information:   lipgloss.Color("#81A1C1"), // nord9 - blue
			NotClassified: lipgloss.Color("#4C566A"), // nord3 - gray

			// Status colors
			OK:          lipgloss.Color("#A3BE8C"), // nord14 - green
			Unknown:     lipgloss.Color("#4C566A"), // nord3
			Maintenance: lipgloss.Color("#B48EAD"), // nord15 - purple

			// UI colors (Polar Night + Frost)
			Primary:       lipgloss.Color("#88C0D0"), // nord8 - cyan
			Secondary:     lipgloss.Color("#A3BE8C"), // nord14 - green
			Background:    lipgloss.Color("#2E3440"), // nord0
			Foreground:    lipgloss.Color("#D8DEE9"), // nord4
			Muted:         lipgloss.Color("#4C566A"), // nord3
			Border:        lipgloss.Color("#3B4252"), // nord1
			FocusedBorder: lipgloss.Color("#88C0D0"), // nord8
			Highlight:     lipgloss.Color("#3B4252"), // nord1
			Surface:       lipgloss.Color("#434C5E"), // nord2
		},
	}
}

// DraculaTheme returns the Dracula color theme.
// https://draculatheme.com/
func DraculaTheme() *Theme {
	return &Theme{
		Name:        ThemeDracula,
		Description: "Dark purple/pink Dracula aesthetic",
		Colors: ColorPalette{
			// Severity colors
			Disaster:      lipgloss.Color("#FF5555"), // Red
			High:          lipgloss.Color("#FFB86C"), // Orange
			Average:       lipgloss.Color("#F1FA8C"), // Yellow
			Warning:       lipgloss.Color("#8BE9FD"), // Cyan
			Information:   lipgloss.Color("#BD93F9"), // Purple
			NotClassified: lipgloss.Color("#6272A4"), // Comment

			// Status colors
			OK:          lipgloss.Color("#50FA7B"), // Green
			Unknown:     lipgloss.Color("#6272A4"), // Comment
			Maintenance: lipgloss.Color("#FF79C6"), // Pink

			// UI colors
			Primary:       lipgloss.Color("#BD93F9"), // Purple
			Secondary:     lipgloss.Color("#50FA7B"), // Green
			Background:    lipgloss.Color("#282A36"), // Background
			Foreground:    lipgloss.Color("#F8F8F2"), // Foreground
			Muted:         lipgloss.Color("#6272A4"), // Comment
			Border:        lipgloss.Color("#44475A"), // Current Line
			FocusedBorder: lipgloss.Color("#FF79C6"), // Pink
			Highlight:     lipgloss.Color("#44475A"), // Selection
			Surface:       lipgloss.Color("#44475A"), // Current Line
		},
	}
}

// GruvboxTheme returns the Gruvbox dark color theme.
// https://github.com/morhetz/gruvbox
func GruvboxTheme() *Theme {
	return &Theme{
		Name:        ThemeGruvbox,
		Description: "Retro warm-toned Gruvbox palette",
		Colors: ColorPalette{
			// Severity colors (bright variants)
			Disaster:      lipgloss.Color("#FB4934"), // red_bright
			High:          lipgloss.Color("#FE8019"), // orange_bright
			Average:       lipgloss.Color("#FABD2F"), // yellow_bright
			Warning:       lipgloss.Color("#83A598"), // blue_bright
			Information:   lipgloss.Color("#D3869B"), // purple_bright
			NotClassified: lipgloss.Color("#A89984"), // gray

			// Status colors
			OK:          lipgloss.Color("#B8BB26"), // green_bright
			Unknown:     lipgloss.Color("#A89984"), // gray
			Maintenance: lipgloss.Color("#D3869B"), // purple_bright

			// UI colors
			Primary:       lipgloss.Color("#8EC07C"), // aqua_bright
			Secondary:     lipgloss.Color("#B8BB26"), // green_bright
			Background:    lipgloss.Color("#282828"), // bg0
			Foreground:    lipgloss.Color("#EBDBB2"), // fg1
			Muted:         lipgloss.Color("#928374"), // gray_245
			Border:        lipgloss.Color("#3C3836"), // bg1
			FocusedBorder: lipgloss.Color("#8EC07C"), // aqua_bright
			Highlight:     lipgloss.Color("#3C3836"), // bg1
			Surface:       lipgloss.Color("#504945"), // bg2
		},
	}
}

// CatppuccinTheme returns the Catppuccin Mocha color theme.
// https://github.com/catppuccin/catppuccin
func CatppuccinTheme() *Theme {
	return &Theme{
		Name:        ThemeCatppuccin,
		Description: "Soothing pastel Catppuccin Mocha palette",
		Colors: ColorPalette{
			// Severity colors
			Disaster:      lipgloss.Color("#F38BA8"), // Red
			High:          lipgloss.Color("#FAB387"), // Peach
			Average:       lipgloss.Color("#F9E2AF"), // Yellow
			Warning:       lipgloss.Color("#89DCEB"), // Sky
			Information:   lipgloss.Color("#89B4FA"), // Blue
			NotClassified: lipgloss.Color("#6C7086"), // Overlay0

			// Status colors
			OK:          lipgloss.Color("#A6E3A1"), // Green
			Unknown:     lipgloss.Color("#6C7086"), // Overlay0
			Maintenance: lipgloss.Color("#CBA6F7"), // Mauve

			// UI colors
			Primary:       lipgloss.Color("#B4BEFE"), // Lavender
			Secondary:     lipgloss.Color("#A6E3A1"), // Green
			Background:    lipgloss.Color("#1E1E2E"), // Base
			Foreground:    lipgloss.Color("#CDD6F4"), // Text
			Muted:         lipgloss.Color("#6C7086"), // Overlay0
			Border:        lipgloss.Color("#313244"), // Surface0
			FocusedBorder: lipgloss.Color("#B4BEFE"), // Lavender
			Highlight:     lipgloss.Color("#45475A"), // Surface1
			Surface:       lipgloss.Color("#313244"), // Surface0
		},
	}
}

// TokyoNightTheme returns the Tokyo Night color theme.
// https://github.com/folke/tokyonight.nvim
func TokyoNightTheme() *Theme {
	return &Theme{
		Name:        ThemeTokyoNight,
		Description: "Cool blues and purples Tokyo Night palette",
		Colors: ColorPalette{
			// Severity colors
			Disaster:      lipgloss.Color("#F7768E"), // red
			High:          lipgloss.Color("#FF9E64"), // orange
			Average:       lipgloss.Color("#E0AF68"), // yellow
			Warning:       lipgloss.Color("#7DCFFF"), // cyan
			Information:   lipgloss.Color("#7AA2F7"), // blue
			NotClassified: lipgloss.Color("#565F89"), // comment

			// Status colors
			OK:          lipgloss.Color("#9ECE6A"), // green
			Unknown:     lipgloss.Color("#565F89"), // comment
			Maintenance: lipgloss.Color("#BB9AF7"), // magenta

			// UI colors
			Primary:       lipgloss.Color("#7AA2F7"), // blue
			Secondary:     lipgloss.Color("#9ECE6A"), // green
			Background:    lipgloss.Color("#1A1B26"), // bg_dark
			Foreground:    lipgloss.Color("#C0CAF5"), // fg
			Muted:         lipgloss.Color("#565F89"), // comment
			Border:        lipgloss.Color("#292E42"), // bg_highlight
			FocusedBorder: lipgloss.Color("#BB9AF7"), // magenta
			Highlight:     lipgloss.Color("#292E42"), // bg_highlight
			Surface:       lipgloss.Color("#24283B"), // bg
		},
	}
}

// SolarizedTheme returns the Solarized dark color theme.
// https://ethanschoonover.com/solarized/
func SolarizedTheme() *Theme {
	return &Theme{
		Name:        ThemeSolarized,
		Description: "Precision-balanced Solarized dark palette",
		Colors: ColorPalette{
			// Severity colors
			Disaster:      lipgloss.Color("#DC322F"), // red
			High:          lipgloss.Color("#CB4B16"), // orange
			Average:       lipgloss.Color("#B58900"), // yellow
			Warning:       lipgloss.Color("#268BD2"), // blue
			Information:   lipgloss.Color("#6C71C4"), // violet
			NotClassified: lipgloss.Color("#586E75"), // base01

			// Status colors
			OK:          lipgloss.Color("#859900"), // green
			Unknown:     lipgloss.Color("#586E75"), // base01
			Maintenance: lipgloss.Color("#D33682"), // magenta

			// UI colors
			Primary:       lipgloss.Color("#268BD2"), // blue
			Secondary:     lipgloss.Color("#859900"), // green
			Background:    lipgloss.Color("#002B36"), // base03
			Foreground:    lipgloss.Color("#839496"), // base0
			Muted:         lipgloss.Color("#586E75"), // base01
			Border:        lipgloss.Color("#073642"), // base02
			FocusedBorder: lipgloss.Color("#2AA198"), // cyan
			Highlight:     lipgloss.Color("#073642"), // base02
			Surface:       lipgloss.Color("#073642"), // base02
		},
	}
}
