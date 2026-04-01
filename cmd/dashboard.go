package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/zubeyralmaho/jotform-cli/internal/api"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:     "dashboard",
	Aliases: []string{"dash", "d"},
	Short:   "Open the interactive TUI dashboard",
	Long: `Opens a full-terminal interactive dashboard with a split-pane layout:
  Left:  Scrollable form list with live selection
  Right: Form details, stats, and quick actions

Controls:
  ↑/↓ or j/k   Navigate forms
  enter         Select form / confirm
  o             Open selected form in browser
  w             Start watching submissions
  q or esc      Quit`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		p := tea.NewProgram(
			newDashboardModel(client),
			tea.WithAltScreen(),
			tea.WithMouseCellMotion(),
		)
		_, err = p.Run()
		return err
	},
}

// ── Dashboard Model ─────────────────────────────────────────────────────

type dashPhase int

const (
	dashPhaseLoading dashPhase = iota
	dashPhaseReady
	dashPhaseError
)

type dashboardModel struct {
	client   *api.Client
	phase    dashPhase
	spinner  spinner.Model
	forms    []api.Form
	selected int
	detail   *api.FormProperties
	errMsg   string
	width    int
	height   int
}

type dashFormsLoadedMsg struct{ forms []api.Form }
type dashDetailLoadedMsg struct{ detail *api.FormProperties }
type dashErrMsg struct{ err error }

func newDashboardModel(client *api.Client) dashboardModel {
	return dashboardModel{
		client:  client,
		phase:   dashPhaseLoading,
		spinner: ui.NewSpinner(),
	}
}

func (m dashboardModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		loadForms(m.client),
	)
}

func loadForms(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		forms, err := client.ListForms(0, 100)
		if err != nil {
			return dashErrMsg{err}
		}
		return dashFormsLoadedMsg{forms}
	}
}

func loadDetail(client *api.Client, formID string) tea.Cmd {
	return func() tea.Msg {
		detail, err := client.GetForm(formID)
		if err != nil {
			return dashErrMsg{err}
		}
		return dashDetailLoadedMsg{detail}
	}
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.phase == dashPhaseReady && m.selected > 0 {
				m.selected--
				m.detail = nil
				return m, loadDetail(m.client, m.forms[m.selected].ID)
			}

		case "down", "j":
			if m.phase == dashPhaseReady && m.selected < len(m.forms)-1 {
				m.selected++
				m.detail = nil
				return m, loadDetail(m.client, m.forms[m.selected].ID)
			}

		case "o":
			if m.phase == dashPhaseReady && len(m.forms) > 0 {
				formURL := fmt.Sprintf("https://form.jotform.com/%s", m.forms[m.selected].ID)
				openBrowser(formURL)
			}

		case "w":
			if m.phase == dashPhaseReady && len(m.forms) > 0 {
				// Return quit first, then start watch in the main process
				return m, tea.Quit
			}
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case dashFormsLoadedMsg:
		m.forms = msg.forms
		m.phase = dashPhaseReady
		if len(m.forms) > 0 {
			return m, loadDetail(m.client, m.forms[0].ID)
		}

	case dashDetailLoadedMsg:
		m.detail = msg.detail

	case dashErrMsg:
		m.phase = dashPhaseError
		m.errMsg = msg.err.Error()
	}

	return m, nil
}

func (m dashboardModel) View() string {
	if m.width == 0 {
		return ""
	}

	switch m.phase {
	case dashPhaseLoading:
		return m.viewLoading()
	case dashPhaseError:
		return m.viewError()
	case dashPhaseReady:
		return m.viewDashboard()
	}
	return ""
}

func (m dashboardModel) viewLoading() string {
	return fmt.Sprintf("\n\n  %s %s\n",
		m.spinner.View(),
		ui.Muted.Render("Loading dashboard..."),
	)
}

func (m dashboardModel) viewError() string {
	return fmt.Sprintf("\n  %s\n  %s\n",
		ui.ErrorBanner("Failed to load forms"),
		ui.Muted.Render(m.errMsg),
	)
}

func (m dashboardModel) viewDashboard() string {
	leftWidth := m.width/3 - 2
	rightWidth := m.width - leftWidth - 6

	left := m.renderFormList(leftWidth)
	right := m.renderDetail(rightWidth)

	// Split pane with navy border separating them
	leftPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.BrandNavy).
		Width(leftWidth).
		Height(m.height - 4).
		Render(left)

	rightPane := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.BrandNavy).
		Width(rightWidth).
		Height(m.height - 4).
		Render(right)

	panes := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane)

	header := ui.LogoCompact() + "  " + ui.Title.Render("Jotform Dashboard") +
		ui.Muted.Render(fmt.Sprintf("  (%d forms)", len(m.forms))) +
		lipgloss.NewStyle().PaddingLeft(4).Render(
			ui.Muted.Render("↑↓ navigate  o open  w watch  q quit"),
		)

	return header + "\n" + panes
}

func (m dashboardModel) renderFormList(width int) string {
	var sb strings.Builder
	sb.WriteString(ui.Subtitle.Render("Forms") + "\n")
	sb.WriteString(ui.Separator(width) + "\n\n")

	for i, f := range m.forms {
		title := f.Title
		if len(title) > width-4 {
			title = title[:width-7] + "..."
		}

		if i == m.selected {
			// Active item: orange background, bold
			row := ui.ActiveListItem.Width(width - 2).Render(title)
			sb.WriteString(row + "\n")
		} else {
			row := ui.ListItem.Render(title)
			sb.WriteString(row + "\n")
		}
	}
	return sb.String()
}

func (m dashboardModel) renderDetail(width int) string {
	if len(m.forms) == 0 {
		return ui.Muted.Render("No forms found.")
	}

	f := m.forms[m.selected]
	var sb strings.Builder

	sb.WriteString(ui.Title.Render(f.Title) + "\n")
	sb.WriteString(ui.Separator(width) + "\n\n")

	sb.WriteString(ui.KeyValuePairs([][2]string{
		{"ID", f.ID},
		{"Status", f.Status},
		{"Submissions", string(f.Count)},
	}))
	sb.WriteString("\n\n")

	if m.detail != nil {
		questionCount := len(m.detail.Questions)
		sb.WriteString(ui.KeyValue("Questions", fmt.Sprintf("%d fields", questionCount)))
		sb.WriteString("\n\n")
		sb.WriteString(ui.Subtitle.Render("Questions") + "\n")
		sb.WriteString(ui.Separator(width) + "\n")

		count := 0
		for _, q := range m.detail.Questions {
			if count >= 8 {
				remaining := questionCount - 8
				if remaining > 0 {
					sb.WriteString(ui.Muted.Render(fmt.Sprintf("  ... and %d more", remaining)))
				}
				break
			}
			if qMap, ok := q.(map[string]interface{}); ok {
				qType := fmt.Sprintf("%v", qMap["type"])
				qText := fmt.Sprintf("%v", qMap["text"])
				if len(qText) > width-20 {
					qText = qText[:width-23] + "..."
				}
				sb.WriteString(fmt.Sprintf("  %s  %s\n",
					ui.Subtitle.Render(qType),
					ui.Muted.Render(qText),
				))
			}
			count++
		}
	} else {
		sb.WriteString(ui.Muted.Render("  Loading details..."))
	}

	sb.WriteString("\n\n")
	sb.WriteString(ui.Muted.Render("  Open:  ") + ui.Value.Render(fmt.Sprintf("https://form.jotform.com/%s", f.ID)))

	return sb.String()
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
