package theme

import "github.com/charmbracelet/lipgloss"

// Theme contains all styling information for the application.
type Theme struct {
	Name        string
	Description string
	Colors      ColorPalette
}

// Styles contains pre-built lipgloss styles derived from a theme.
// These are computed once when a theme is loaded for performance.
type Styles struct {
	// Base styles
	App    lipgloss.Style
	Title  lipgloss.Style
	Subtle lipgloss.Style

	// Pane styles
	PaneFocused  lipgloss.Style
	PaneBlurred  lipgloss.Style
	PaneTitle    lipgloss.Style
	PaneSubtitle lipgloss.Style

	// Status bar styles
	StatusBar     lipgloss.Style
	StatusOK      lipgloss.Style
	StatusProblem lipgloss.Style
	StatusUnknown lipgloss.Style
	StatusMaint   lipgloss.Style

	// Alert list styles
	AlertSelected lipgloss.Style
	AlertNormal   lipgloss.Style
	AlertSeverity [6]lipgloss.Style // Index 0-5 for severity levels
	AlertHost     lipgloss.Style
	AlertName     lipgloss.Style
	AlertDuration lipgloss.Style
	AlertAcked    lipgloss.Style

	// Detail pane styles
	DetailLabel lipgloss.Style
	DetailValue lipgloss.Style
	DetailTag   lipgloss.Style

	// Tab styles
	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabBar      lipgloss.Style

	// Command input styles
	CommandPrompt lipgloss.Style
	CommandInput  lipgloss.Style
	CommandHint   lipgloss.Style

	// Modal styles
	ModalBackground lipgloss.Style
	ModalBox        lipgloss.Style
	ModalTitle      lipgloss.Style
	ModalText       lipgloss.Style
	ModalButton     lipgloss.Style
	ModalButtonAlt  lipgloss.Style

	// Help styles
	HelpKey  lipgloss.Style
	HelpDesc lipgloss.Style
}

// NewStyles creates a Styles struct from a Theme.
func NewStyles(t *Theme) *Styles {
	c := t.Colors

	return &Styles{
		// Base styles
		App:    lipgloss.NewStyle().Background(c.Background),
		Title:  lipgloss.NewStyle().Foreground(c.Primary).Bold(true),
		Subtle: lipgloss.NewStyle().Foreground(c.Muted),

		// Pane styles
		PaneFocused: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.FocusedBorder),
		PaneBlurred: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Border),
		PaneTitle: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true).
			Padding(0, 1),
		PaneSubtitle: lipgloss.NewStyle().
			Foreground(c.Muted).
			Padding(0, 1),

		// Status bar styles
		StatusBar: lipgloss.NewStyle().
			Padding(0, 1),
		StatusOK: lipgloss.NewStyle().
			Foreground(c.OK).
			Bold(true),
		StatusProblem: lipgloss.NewStyle().
			Foreground(c.Disaster).
			Bold(true),
		StatusUnknown: lipgloss.NewStyle().
			Foreground(c.Unknown),
		StatusMaint: lipgloss.NewStyle().
			Foreground(c.Maintenance),

		// Alert list styles
		AlertSelected: lipgloss.NewStyle().
			Background(c.Highlight).
			Foreground(c.Foreground).
			Bold(true),
		AlertNormal: lipgloss.NewStyle().
			Foreground(c.Foreground),
		AlertSeverity: [6]lipgloss.Style{
			lipgloss.NewStyle().Foreground(c.NotClassified),       // 0
			lipgloss.NewStyle().Foreground(c.Information),         // 1
			lipgloss.NewStyle().Foreground(c.Warning),             // 2
			lipgloss.NewStyle().Foreground(c.Average),             // 3
			lipgloss.NewStyle().Foreground(c.High),                // 4
			lipgloss.NewStyle().Foreground(c.Disaster).Bold(true), // 5
		},
		AlertHost: lipgloss.NewStyle().
			Foreground(c.Secondary),
		AlertName: lipgloss.NewStyle().
			Foreground(c.Foreground),
		AlertDuration: lipgloss.NewStyle().
			Foreground(c.Muted),
		AlertAcked: lipgloss.NewStyle().
			Foreground(c.OK),

		// Detail pane styles
		DetailLabel: lipgloss.NewStyle().
			Foreground(c.Muted).
			Width(12),
		DetailValue: lipgloss.NewStyle().
			Foreground(c.Foreground),
		DetailTag: lipgloss.NewStyle().
			Foreground(c.Secondary).
			Background(c.Surface).
			Padding(0, 1),

		// Tab styles
		TabActive: lipgloss.NewStyle().
			Foreground(c.Primary).
			Background(c.Surface).
			Bold(true).
			Padding(0, 2),
		TabInactive: lipgloss.NewStyle().
			Foreground(c.Muted).
			Padding(0, 2),
		TabBar: lipgloss.NewStyle().
			BorderBottom(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(c.Border),

		// Command input styles
		CommandPrompt: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true),
		CommandInput: lipgloss.NewStyle().
			Foreground(c.Foreground),
		CommandHint: lipgloss.NewStyle().
			Foreground(c.Muted).
			Italic(true),

		// Modal styles
		ModalBackground: lipgloss.NewStyle(),
		ModalBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(c.Primary).
			Padding(1, 2),
		ModalTitle: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true).
			MarginBottom(1),
		ModalText: lipgloss.NewStyle().
			Foreground(c.Foreground),
		ModalButton: lipgloss.NewStyle().
			Foreground(c.Background).
			Background(c.Primary).
			Padding(0, 2),
		ModalButtonAlt: lipgloss.NewStyle().
			Foreground(c.Foreground).
			Background(c.Surface).
			Padding(0, 2),

		// Help styles
		HelpKey: lipgloss.NewStyle().
			Foreground(c.Primary).
			Bold(true),
		HelpDesc: lipgloss.NewStyle().
			Foreground(c.Muted),
	}
}
