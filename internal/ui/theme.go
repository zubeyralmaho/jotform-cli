package ui

import "github.com/charmbracelet/lipgloss"

// ── Brand Color Palette (Jotform 2026) ─────────────────────────────────
// These are the strict, canonical brand hex codes mapped to lipgloss tokens.

var (
	// BrandNavy is the anchor color: structure, borders, panel outlines.
	BrandNavy = lipgloss.Color("#0A1551")
	// BrandOrange is the primary accent: CTAs, active items, headers.
	BrandOrange = lipgloss.Color("#FF6100")
	// BrandBlue is the technical accent: progress, spinners, metrics.
	BrandBlue = lipgloss.Color("#0099FF")
	// BrandYellow is the highlight accent: warnings, search hits, glows.
	BrandYellow = lipgloss.Color("#FFB629")

	// Neutral tones for text on various backgrounds.
	White    = lipgloss.Color("#FFFFFF")
	DimWhite = lipgloss.Color("#B0B0B0")
	DarkGray = lipgloss.Color("#3A3A3A")
)

// ── Reusable Styles ────────────────────────────────────────────────────

var (
	// Title renders a bold orange header line.
	Title = lipgloss.NewStyle().
		Foreground(BrandOrange).
		Bold(true)

	// Subtitle renders a dim blue secondary line.
	Subtitle = lipgloss.NewStyle().
			Foreground(BrandBlue)

	// Label renders a dim white label (left-side of key:value pairs).
	Label = lipgloss.NewStyle().
		Foreground(DimWhite)

	// Value renders a bright white value.
	Value = lipgloss.NewStyle().
		Foreground(White).
		Bold(true)

	// Success renders a confirmation message in blue.
	Success = lipgloss.NewStyle().
		Foreground(BrandBlue).
		Bold(true)

	// Warning renders a warning message in yellow.
	Warning = lipgloss.NewStyle().
		Foreground(BrandYellow)

	// ErrorStyle renders error text in orange (high contrast).
	ErrorStyle = lipgloss.NewStyle().
			Foreground(BrandOrange).
			Bold(true)

	// Muted renders dimmed helper text.
	Muted = lipgloss.NewStyle().
		Foreground(DimWhite)

	// Added renders added items (green-ish via blue).
	Added = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#2ECC71"))

	// Modified renders modified items in yellow.
	Modified = lipgloss.NewStyle().
			Foreground(BrandYellow)

	// Deleted renders deleted items in orange.
	Deleted = lipgloss.NewStyle().
		Foreground(BrandOrange)

	// Panel renders a bordered box with navy border.
	Panel = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BrandNavy).
		Padding(1, 2)

	// ActivePanel renders a box with orange border (focused state).
	ActivePanel = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BrandOrange).
			Padding(1, 2)

	// ListItem renders a normal list row.
	ListItem = lipgloss.NewStyle().
			PaddingLeft(2)

	// ActiveListItem renders the currently selected list row.
	ActiveListItem = lipgloss.NewStyle().
			Foreground(White).
			Background(BrandOrange).
			Bold(true).
			PaddingLeft(2).
			PaddingRight(2)

	// GradientBanner creates the branded success banner box.
	GradientBanner = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(BrandBlue).
			Foreground(White).
			Padding(0, 2).
			Bold(true)
)
