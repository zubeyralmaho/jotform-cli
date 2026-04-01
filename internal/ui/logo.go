package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Logo returns the Jotform geometric logo rendered with ANSI half-blocks.
// The four SVG paths are approximated as colored block art:
//   - Blue diagonal (top-left)
//   - Orange diagonal (center)
//   - Yellow square (bottom-right)
//   - Navy anchor (bottom-left corner)
func Logo() string {
	blue := lipgloss.NewStyle().Foreground(BrandBlue)
	orange := lipgloss.NewStyle().Foreground(BrandOrange)
	yellow := lipgloss.NewStyle().Foreground(BrandYellow)
	navy := lipgloss.NewStyle().Foreground(BrandNavy)

	b := blue.Render
	o := orange.Render
	y := yellow.Render
	n := navy.Render

	// Half-block characters for sub-cell resolution
	const (
		full  = "█"
		upper = "▀"
		lower = "▄"
	)

	_ = upper // used in specific lines below

	lines := []string{
		"  " + b(full+full) + "              ",
		" " + b(full+full+full+full) + "             ",
		"  " + b(full+full+full+full) + "            ",
		"   " + b(full+full) + " " + o(full+full) + "           ",
		"        " + o(full+full+full+full) + "        ",
		"         " + o(full+full+full+full) + "       ",
		"          " + o(full+full) + " " + y(full+full) + "    ",
		"              " + y(full+full+full+full) + " ",
		" " + n(lower+lower) + "          " + y(full+full+full+full) + " ",
		" " + n(full+full) + "           " + y(full+full) + "  ",
	}

	return strings.Join(lines, "\n")
}

// LogoCompact returns a smaller single-line stylized logo mark.
func LogoCompact() string {
	blue := lipgloss.NewStyle().Foreground(BrandBlue).Render
	orange := lipgloss.NewStyle().Foreground(BrandOrange).Render
	yellow := lipgloss.NewStyle().Foreground(BrandYellow).Render
	navy := lipgloss.NewStyle().Foreground(BrandNavy).Render

	return navy("▄") + blue("◆") + orange("◆") + yellow("◆")
}

// LogoWithText returns the logo next to the wordmark.
func LogoWithText() string {
	logo := Logo()
	wordmark := lipgloss.NewStyle().
		Foreground(BrandOrange).
		Bold(true).
		Render("jotform")

	tagline := lipgloss.NewStyle().
		Foreground(DimWhite).
		Render("Form-as-Code CLI")

	text := lipgloss.JoinVertical(lipgloss.Left,
		"",
		"",
		"  "+wordmark,
		"  "+tagline,
		"",
	)

	return lipgloss.JoinHorizontal(lipgloss.Top, logo, "  ", text)
}
