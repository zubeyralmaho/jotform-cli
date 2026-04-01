package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	qrcode "github.com/skip2/go-qrcode"
	"github.com/spf13/cobra"
	"github.com/zubeyralmaho/jotform-cli/internal/api"
	"github.com/zubeyralmaho/jotform-cli/internal/config"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
)

var shareCmd = &cobra.Command{
	Use:   "share [form-id]",
	Short: "Display form URL and QR code for sharing",
	Long: `Retrieves the form URL and renders a QR code in the terminal.

Examples:
  jotform share                    # Uses form ID from .jotform.yaml
  jotform share 242753193847060    # Share specific form
  jf share                         # Short alias`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShare,
}

func runShare(cmd *cobra.Command, args []string) error {
	formID, err := config.ResolveFormID(args)
	if err != nil {
		return err
	}

	res, err := ui.RunWithSpinner("Fetching form...", func() (interface{}, error) {
		client, err := newClient()
		if err != nil {
			return nil, err
		}
		return client.GetForm(formID)
	})
	if err != nil {
		return err
	}
	form := res.(*api.FormProperties)

	formURL := fmt.Sprintf("https://form.jotform.com/%s", formID)

	// Render the QR code
	qr := renderQR(formURL)

	// Outer box: navy background, orange header
	header := lipgloss.NewStyle().
		Foreground(ui.White).
		Background(ui.BrandOrange).
		Bold(true).
		Padding(0, 2).
		Render("  " + form.Title + "  ")

	urlLine := "  " + ui.Label.Render("URL:") + " " + ui.Value.Render(formURL)
	idLine := "  " + ui.Label.Render("ID: ") + " " + ui.Subtitle.Render(formID)

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.BrandNavy).
		Background(ui.BrandNavy).
		Padding(1, 2)

	qrBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ui.BrandOrange).
		Padding(1, 2).
		Render(qr)

	content := strings.Join([]string{
		header,
		"",
		urlLine,
		idLine,
		"",
		ui.Muted.Render("  Scan the QR code to open the form:"),
	}, "\n")

	fmt.Println(box.Render(content))
	fmt.Println(qrBox)
	return nil
}

// renderQR generates a real QR code matrix and renders it with block characters.
func renderQR(url string) string {
	code, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		// Fallback: just display the URL in a box if QR generation fails
		return ui.Muted.Render("  (QR code not available)\n  ") + ui.Value.Render(url)
	}

	matrix := code.Bitmap()
	if len(matrix) == 0 {
		return ui.Muted.Render("  (QR code not available)\n  ") + ui.Value.Render(url)
	}

	var sb strings.Builder
	quietZone := 2
	rowWidth := len(matrix[0]) + (quietZone * 2)

	// Top quiet zone.
	for i := 0; i < quietZone; i++ {
		sb.WriteString(strings.Repeat("  ", rowWidth) + "\n")
	}

	for _, row := range matrix {
		sb.WriteString(strings.Repeat("  ", quietZone)) // left quiet zone
		for _, mod := range row {
			if mod {
				sb.WriteString("██") // dark module
			} else {
				sb.WriteString("  ") // light module
			}
		}
		sb.WriteString(strings.Repeat("  ", quietZone) + "\n") // right quiet zone
	}

	// Bottom quiet zone.
	for i := 0; i < quietZone; i++ {
		sb.WriteString(strings.Repeat("  ", rowWidth))
		if i < quietZone-1 {
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

func init() {
	rootCmd.AddCommand(shareCmd)
}
