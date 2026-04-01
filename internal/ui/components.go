package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── Spinner ────────────────────────────────────────────────────────────

// NewSpinner returns a brand-blue dot spinner.
func NewSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(BrandBlue)
	return s
}

// ── Spinner TUI (blocking) ─────────────────────────────────────────────

// SpinnerResult holds the outcome of an async operation run under a spinner.
type SpinnerResult struct {
	Result interface{}
	Err    error
}

type spinnerDoneMsg SpinnerResult

type spinnerModel struct {
	spinner spinner.Model
	message string
	done    bool
	result  SpinnerResult
	fn      func() (interface{}, error)
}

func (m spinnerModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			res, err := m.fn()
			return spinnerDoneMsg{Result: res, Err: err}
		},
	)
}

func (m spinnerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case spinnerDoneMsg:
		m.done = true
		m.result = SpinnerResult(msg)
		return m, tea.Quit
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m spinnerModel) View() string {
	if m.done {
		return ""
	}
	return m.spinner.View() + " " + Muted.Render(m.message) + "\n"
}

// RunWithSpinner executes fn while showing a branded spinner with the given message.
// Returns the result and error from fn.
func RunWithSpinner(message string, fn func() (interface{}, error)) (interface{}, error) {
	m := spinnerModel{
		spinner: NewSpinner(),
		message: message,
		fn:      fn,
	}
	p := tea.NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("spinner error: %w", err)
	}
	result := finalModel.(spinnerModel).result
	return result.Result, result.Err
}

// ── Staggered List ─────────────────────────────────────────────────────

type staggerTickMsg struct{}

type staggerModel struct {
	rows      []string
	visible   int
	done      bool
	interval  time.Duration
}

func (m staggerModel) Init() tea.Cmd {
	return tea.Tick(m.interval, func(time.Time) tea.Msg {
		return staggerTickMsg{}
	})
}

func (m staggerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case staggerTickMsg:
		m.visible++
		if m.visible >= len(m.rows) {
			m.done = true
			return m, tea.Quit
		}
		return m, tea.Tick(m.interval, func(time.Time) tea.Msg {
			return staggerTickMsg{}
		})
	case tea.KeyMsg:
		// Any keypress skips animation
		m.visible = len(m.rows)
		m.done = true
		return m, tea.Quit
	}
	return m, nil
}

func (m staggerModel) View() string {
	if m.visible > len(m.rows) {
		m.visible = len(m.rows)
	}
	return strings.Join(m.rows[:m.visible], "\n") + "\n"
}

// RenderStaggered animates rows cascading into view with a per-row delay.
func RenderStaggered(rows []string, interval time.Duration) {
	if len(rows) == 0 {
		return
	}
	m := staggerModel{
		rows:     rows,
		visible:  1, // show first row immediately
		interval: interval,
	}
	p := tea.NewProgram(m)
	p.Run()
}

// ── Banner ─────────────────────────────────────────────────────────────

// SuccessBanner renders a branded success message in a box.
func SuccessBanner(message string) string {
	return GradientBanner.Render("  " + message + "  ")
}

// ErrorBanner renders an error message in a branded box.
func ErrorBanner(message string) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(BrandOrange).
		Foreground(BrandOrange).
		Padding(0, 2).
		Bold(true)
	return box.Render("  " + message + "  ")
}

// ── Key-Value Formatting ───────────────────────────────────────────────

// KeyValue renders a styled "label: value" pair.
func KeyValue(label, value string) string {
	return Label.Render(label+":") + " " + Value.Render(value)
}

// KeyValuePairs renders multiple key-value pairs aligned.
func KeyValuePairs(pairs [][2]string) string {
	maxLen := 0
	for _, p := range pairs {
		if len(p[0]) > maxLen {
			maxLen = len(p[0])
		}
	}

	var lines []string
	for _, p := range pairs {
		padded := fmt.Sprintf("%-*s", maxLen, p[0])
		lines = append(lines, Label.Render(padded)+" "+Value.Render(p[1]))
	}
	return strings.Join(lines, "\n")
}

// ── Helpers ────────────────────────────────────────────────────────────

// Separator renders a dim horizontal line.
func Separator(width int) string {
	return Muted.Render(strings.Repeat("─", width))
}
