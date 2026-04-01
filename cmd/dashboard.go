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
	1/2/3/4      Switch screens (overview/questions/submissions/actions)
	? or 5       Open help
	b or esc      Go back (or quit from root screen)
	r             Refresh current screen data
	o             Open selected form in browser
	q             Quit`,
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

type dashScreen int

const (
	dashScreenOverview dashScreen = iota
	dashScreenQuestions
	dashScreenSubmissions
	dashScreenActions
	dashScreenHelp
)

type dashboardModel struct {
	client   *api.Client
	phase    dashPhase
	spinner  spinner.Model
	forms    []api.Form
	selected int
	detail   *api.FormProperties
	subs     []api.Submission
	screen   dashScreen
	history  []dashScreen
	errMsg   string
	width    int
	height   int
}

type dashFormsLoadedMsg struct{ forms []api.Form }
type dashDetailLoadedMsg struct{ detail *api.FormProperties }
type dashSubsLoadedMsg struct{ subs []api.Submission }
type dashErrMsg struct{ err error }

func newDashboardModel(client *api.Client) dashboardModel {
	return dashboardModel{
		client:  client,
		phase:   dashPhaseLoading,
		screen:  dashScreenOverview,
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

func loadSubmissions(client *api.Client, formID string) tea.Cmd {
	return func() tea.Msg {
		subs, err := client.GetSubmissions(formID, 0, 20, "created_at", "DESC")
		if err != nil {
			return dashErrMsg{err}
		}
		return dashSubsLoadedMsg{subs: subs}
	}
}

func (m dashboardModel) currentFormID() string {
	if len(m.forms) == 0 || m.selected < 0 || m.selected >= len(m.forms) {
		return ""
	}
	return m.forms[m.selected].ID
}

func (m *dashboardModel) pushScreen(next dashScreen) {
	if m.screen == next {
		return
	}
	m.history = append(m.history, m.screen)
	m.screen = next
}

func (m *dashboardModel) popScreen() {
	if len(m.history) == 0 {
		m.screen = dashScreenOverview
		return
	}
	last := len(m.history) - 1
	m.screen = m.history[last]
	m.history = m.history[:last]
}

func (m dashboardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "esc", "b", "backspace":
			if m.screen != dashScreenOverview || len(m.history) > 0 {
				m.popScreen()
				return m, nil
			}
			return m, tea.Quit

		case "up", "k":
			if m.phase == dashPhaseReady && m.selected > 0 {
				m.selected--
				m.detail = nil
				m.subs = nil
				return m, loadDetail(m.client, m.forms[m.selected].ID)
			}

		case "down", "j":
			if m.phase == dashPhaseReady && m.selected < len(m.forms)-1 {
				m.selected++
				m.detail = nil
				m.subs = nil
				return m, loadDetail(m.client, m.forms[m.selected].ID)
			}

		case "1":
			m.pushScreen(dashScreenOverview)
			return m, nil

		case "2":
			m.pushScreen(dashScreenQuestions)
			return m, nil

		case "3", "s":
			if m.phase == dashPhaseReady && len(m.forms) > 0 {
				m.pushScreen(dashScreenSubmissions)
				if len(m.subs) == 0 {
					return m, loadSubmissions(m.client, m.currentFormID())
				}
			}
			return m, nil

		case "4", "a", "enter":
			m.pushScreen(dashScreenActions)
			return m, nil

		case "?", "5":
			m.pushScreen(dashScreenHelp)
			return m, nil

		case "left", "h":
			if m.screen > dashScreenOverview {
				m.pushScreen(m.screen - 1)
			}
			return m, nil

		case "right", "l", "tab":
			if m.screen < dashScreenHelp {
				m.pushScreen(m.screen + 1)
			}
			return m, nil

		case "r":
			if m.phase == dashPhaseReady {
				if m.currentFormID() == "" {
					return m, loadForms(m.client)
				}
				if m.screen == dashScreenSubmissions {
					return m, loadSubmissions(m.client, m.currentFormID())
				}
				return m, loadDetail(m.client, m.currentFormID())
			}

		case "o":
			if m.phase == dashPhaseReady && len(m.forms) > 0 {
				formURL := fmt.Sprintf("https://form.jotform.com/%s", m.forms[m.selected].ID)
				openBrowser(formURL)
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

	case dashSubsLoadedMsg:
		m.subs = msg.subs

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
			ui.Muted.Render("↑↓ forms  1-5 screens  ? help  b back  r refresh  o open  q quit"),
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
		{"Screen", m.screenTitle()},
	}))
	sb.WriteString("\n\n")

	sb.WriteString(ui.Muted.Render("  [1] Overview  [2] Questions  [3] Submissions  [4] Actions  [5/?] Help") + "\n")
	sb.WriteString(ui.Muted.Render("  [b/esc] Back  [r] Refresh  [o] Open form") + "\n\n")

	switch m.screen {
	case dashScreenOverview:
		sb.WriteString(m.renderOverview(width))
	case dashScreenQuestions:
		sb.WriteString(m.renderQuestions(width))
	case dashScreenSubmissions:
		sb.WriteString(m.renderSubmissions(width))
	case dashScreenActions:
		sb.WriteString(m.renderActions(width))
	case dashScreenHelp:
		sb.WriteString(m.renderHelp(width))
	}

	sb.WriteString("\n\n")
	sb.WriteString(ui.Muted.Render("  Open:  ") + ui.Value.Render(fmt.Sprintf("https://form.jotform.com/%s", f.ID)))

	return sb.String()
}

func (m dashboardModel) screenTitle() string {
	switch m.screen {
	case dashScreenOverview:
		return "Overview"
	case dashScreenQuestions:
		return "Questions"
	case dashScreenSubmissions:
		return "Submissions"
	case dashScreenActions:
		return "Actions"
	case dashScreenHelp:
		return "Help"
	default:
		return "Overview"
	}
}

func (m dashboardModel) renderOverview(width int) string {
	if m.detail == nil {
		return ui.Muted.Render("  Loading details...")
	}
	questions := extractQuestions(m.detail)
	return ui.KeyValue("Questions", fmt.Sprintf("%d fields", len(questions)))
}

func (m dashboardModel) renderQuestions(width int) string {
	if m.detail == nil {
		return ui.Muted.Render("  Loading details...")
	}

	questions := extractQuestions(m.detail)
	questionCount := len(questions)
	var sb strings.Builder

	sb.WriteString(ui.KeyValue("Questions", fmt.Sprintf("%d fields", questionCount)))
	sb.WriteString("\n\n")
	sb.WriteString(ui.Subtitle.Render("Questions") + "\n")
	sb.WriteString(ui.Separator(width) + "\n")

	count := 0
	for _, q := range questions {
		if count >= 12 {
			remaining := questionCount - 12
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

	return sb.String()
}

func (m dashboardModel) renderSubmissions(width int) string {
	var sb strings.Builder

	sb.WriteString(ui.Subtitle.Render("Recent Submissions") + "\n")
	sb.WriteString(ui.Separator(width) + "\n")

	if len(m.subs) == 0 {
		sb.WriteString("\n" + ui.Muted.Render("  No submissions loaded yet. Press r to refresh."))
		return sb.String()
	}

	limit := len(m.subs)
	if limit > 8 {
		limit = 8
	}

	for i := 0; i < limit; i++ {
		s := m.subs[i]
		subID := s.ID
		if len(subID) > 14 {
			subID = subID[:14] + "..."
		}
		sb.WriteString(fmt.Sprintf("  %s  %s\n",
			ui.Value.Render(subID),
			ui.Muted.Render(s.CreatedAt),
		))
	}

	if len(m.subs) > limit {
		sb.WriteString("\n" + ui.Muted.Render(fmt.Sprintf("  ... and %d more", len(m.subs)-limit)))
	}

	return sb.String()
}

func (m dashboardModel) renderActions(width int) string {
	return strings.Join([]string{
		ui.Subtitle.Render("Quick Actions"),
		ui.Separator(width),
		"",
		"  " + ui.Value.Render("o") + "  Open selected form in browser",
		"  " + ui.Value.Render("r") + "  Refresh current screen data",
		"  " + ui.Value.Render("s") + "  Jump to Submissions screen",
		"  " + ui.Value.Render("?") + "  Open Help screen",
		"  " + ui.Value.Render("1") + "  Go to Overview",
		"  " + ui.Value.Render("2") + "  Go to Questions",
		"  " + ui.Value.Render("3") + "  Go to Submissions",
		"  " + ui.Value.Render("4") + "  Go to Actions",
		"  " + ui.Value.Render("5") + "  Go to Help",
		"  " + ui.Value.Render("b/esc") + "  Go back to previous screen",
		"  " + ui.Value.Render("q") + "  Quit dashboard",
	}, "\n")
}

func (m dashboardModel) renderHelp(width int) string {
	return strings.Join([]string{
		ui.Subtitle.Render("Help"),
		ui.Separator(width),
		"",
		ui.Value.Render("Dashboard Navigation"),
		"  ↑/↓ or j/k   Select form",
		"  1            Overview",
		"  2            Questions",
		"  3 or s       Submissions",
		"  4 or a       Actions",
		"  5 or ?       Help",
		"  b, esc       Back",
		"  r            Refresh",
		"  o            Open selected form URL",
		"  q            Quit",
		"",
		ui.Value.Render("Root Commands"),
		"  jotform --help",
		"  jotform dashboard",
		"  jotform status",
		"  jotform diff",
		"  jotform pull",
		"  jotform push",
		"  jotform share",
		"  jotform open",
		"  jotform version",
		"  jotform completion <shell>",
		"",
		ui.Value.Render("CLI Commands"),
		"  jotform auth login",
		"  jotform forms list",
		"  jotform forms get <form-id>",
		"  jotform forms create --file <file>",
		"  jotform forms update <form-id> --file <file>",
		"  jotform forms delete <form-id>",
		"  jotform forms export <form-id>",
		"  jotform forms import --file <file>",
		"  jotform forms status <form-id>",
		"  jotform forms diff <form-id>",
		"  jotform forms apply <form-id>",
		"  jotform submissions list <form-id>",
		"  jotform submissions watch <form-id>",
		"  jotform share <form-id>",
		"  jotform open <form-id>",
		"  jotform init",
		"  jotform clone <form-id>",
		"  jotform ai generate-schema \"prompt\"",
		"  jotform ai analyze <form-id>",
		"  jotform mcp start-server",
		"",
		ui.Muted.Render("Use b/esc to return to your previous screen."),
	}, "\n")
}

func extractQuestions(detail *api.FormProperties) map[string]interface{} {
	if detail == nil {
		return nil
	}

	if len(detail.Questions) > 0 {
		return detail.Questions
	}

	if detail.Properties != nil {
		if rawQuestions, ok := detail.Properties["questions"]; ok {
			if questions, ok := rawQuestions.(map[string]interface{}); ok {
				return questions
			}
		}
	}

	return detail.Questions
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
