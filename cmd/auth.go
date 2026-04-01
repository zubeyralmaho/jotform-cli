package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jotform/jotform-cli/internal/api"
	"github.com/jotform/jotform-cli/internal/auth"
	"github.com/jotform/jotform-cli/internal/ui"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication credentials",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Store your Jotform API key securely",
	RunE:  runLogin,
}

// ── Login TUI Model ────────────────────────────────────────────────────

type loginPhase int

const (
	loginPhaseInput loginPhase = iota
	loginPhaseValidating
	loginPhaseDone
	loginPhaseError
)

type loginModel struct {
	textInput textinput.Model
	spinner   spinner.Model
	phase     loginPhase
	userName  string
	userEmail string
	errMsg    string
	apiKey    string
}

type loginValidatedMsg struct {
	name  string
	email string
}

type loginErrorMsg struct{ err error }

func newLoginModel() loginModel {
	ti := textinput.New()
	ti.Placeholder = "paste your API key here"
	ti.Focus()
	ti.CharLimit = 128
	ti.Width = 50
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(ui.BrandOrange)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(ui.BrandOrange)
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'

	return loginModel{
		textInput: ti,
		spinner:   ui.NewSpinner(),
		phase:     loginPhaseInput,
	}
}

func (m loginModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m loginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			return m, tea.Quit
		case "enter":
			if m.phase == loginPhaseInput {
				key := strings.TrimSpace(m.textInput.Value())
				if key == "" {
					m.errMsg = "API key cannot be empty"
					return m, nil
				}
				m.apiKey = key
				m.phase = loginPhaseValidating
				m.errMsg = ""
				return m, tea.Batch(
					m.spinner.Tick,
					validateAPIKey(key),
				)
			}
			if m.phase == loginPhaseDone || m.phase == loginPhaseError {
				return m, tea.Quit
			}
		}
	case loginValidatedMsg:
		m.phase = loginPhaseDone
		m.userName = msg.name
		m.userEmail = msg.email
		return m, tea.Quit
	case loginErrorMsg:
		m.phase = loginPhaseError
		m.errMsg = msg.err.Error()
		return m, nil
	case spinner.TickMsg:
		if m.phase == loginPhaseValidating {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	if m.phase == loginPhaseInput {
		var cmd tea.Cmd
		m.textInput, cmd = m.textInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m loginModel) View() string {
	var b strings.Builder

	// Logo header
	b.WriteString(ui.Logo())
	b.WriteString("\n\n")
	b.WriteString(ui.Title.Render("  Jotform Login"))
	b.WriteString("\n\n")

	switch m.phase {
	case loginPhaseInput:
		b.WriteString("  " + ui.Label.Render("Enter your API key:") + "\n")
		b.WriteString("  " + m.textInput.View() + "\n")
		if m.errMsg != "" {
			b.WriteString("\n  " + ui.ErrorStyle.Render(m.errMsg) + "\n")
		}
		b.WriteString("\n  " + ui.Muted.Render("press enter to submit  |  esc to cancel") + "\n")

	case loginPhaseValidating:
		b.WriteString("  " + m.spinner.View() + ui.Muted.Render(" Validating API key...") + "\n")

	case loginPhaseDone:
		b.WriteString(ui.SuccessBanner("Logged in successfully") + "\n\n")
		b.WriteString(ui.KeyValuePairs([][2]string{
			{"  User", m.userName},
			{"  Email", m.userEmail},
		}))
		b.WriteString("\n")

	case loginPhaseError:
		b.WriteString(ui.ErrorBanner("Authentication failed") + "\n\n")
		b.WriteString("  " + ui.Muted.Render(m.errMsg) + "\n")
		b.WriteString("  " + ui.Muted.Render("press enter to exit") + "\n")
	}

	return b.String()
}

func validateAPIKey(key string) tea.Cmd {
	return func() tea.Msg {
		client := api.New(key)
		user, err := client.GetUser()
		if err != nil {
			return loginErrorMsg{err: fmt.Errorf("invalid API key: %w", err)}
		}
		if err := auth.SaveAPIKey(key); err != nil {
			return loginErrorMsg{err: err}
		}
		return loginValidatedMsg{name: user.Name, email: user.Email}
	}
}

func runLogin(cmd *cobra.Command, args []string) error {
	p := tea.NewProgram(newLoginModel())
	_, err := p.Run()
	return err
}

// ── Logout ─────────────────────────────────────────────────────────────

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := auth.DeleteAPIKey(); err != nil {
			return err
		}
		fmt.Println(ui.SuccessBanner("Logged out"))
		return nil
	},
}

// ── Whoami ─────────────────────────────────────────────────────────────

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authenticated user",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, err := resolveAPIKey()
		if err != nil {
			return err
		}

		res, err := ui.RunWithSpinner("Fetching user info...", func() (interface{}, error) {
			client := api.New(key)
			return client.GetUser()
		})
		if err != nil {
			return err
		}
		user := res.(*api.User)

		fmt.Println(ui.LogoCompact() + "  " + ui.Title.Render("Jotform"))
		fmt.Println(ui.Separator(40))
		fmt.Println(ui.KeyValuePairs([][2]string{
			{"User", user.Username},
			{"Email", user.Email},
			{"Plan", user.AccountType},
		}))
		return nil
	},
}

func init() {
	authCmd.AddCommand(loginCmd, logoutCmd, whoamiCmd)
	rootCmd.AddCommand(authCmd)
}
