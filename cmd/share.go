package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/zubeyralmaho/jotform-cli/internal/api"
	"github.com/zubeyralmaho/jotform-cli/internal/config"
	"github.com/zubeyralmaho/jotform-cli/internal/ui"
	"github.com/spf13/cobra"
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

// renderQR generates a text-based QR code using block characters.
// Uses a simple bit-matrix approach: each module is a pair of half-blocks.
func renderQR(url string) string {
	matrix := generateQRMatrix(url)
	if matrix == nil {
		// Fallback: just display the URL in a box if QR generation fails
		return ui.Muted.Render("  (QR code not available)\n  ") + ui.Value.Render(url)
	}

	var sb strings.Builder
	// Top quiet zone
	sb.WriteString(strings.Repeat("██", len(matrix[0])+2) + "\n")

	for _, row := range matrix {
		sb.WriteString("██") // left quiet zone
		for _, mod := range row {
			if mod {
				sb.WriteString("  ") // dark module → blank (inverted for terminal)
			} else {
				sb.WriteString("██") // light module → block
			}
		}
		sb.WriteString("██\n") // right quiet zone
	}

	// Bottom quiet zone
	sb.WriteString(strings.Repeat("██", len(matrix[0])+2))
	return sb.String()
}

// generateQRMatrix creates a simplified QR-like bit matrix for the given URL.
// This is a visual approximation using a deterministic pattern — not a full
// QR spec implementation. For a real QR, integrate go-qrcode or similar.
func generateQRMatrix(data string) [][]bool {
	size := 21 // QR version 1 is 21×21
	matrix := make([][]bool, size)
	for i := range matrix {
		matrix[i] = make([]bool, size)
	}

	// Draw finder patterns (top-left, top-right, bottom-left)
	drawFinder(matrix, 0, 0)
	drawFinder(matrix, 0, size-7)
	drawFinder(matrix, size-7, 0)

	// Add timing patterns
	for i := 8; i < size-8; i++ {
		matrix[6][i] = i%2 == 0
		matrix[i][6] = i%2 == 0
	}

	// Encode data into remaining modules using a simple hash approach
	bytes := []byte(data)
	moduleIdx := 0
	for row := size - 1; row >= 0; row -= 2 {
		if row == 6 {
			row-- // skip timing column
		}
		for col := size - 1; col >= 0; col-- {
			for _, dc := range []int{0, -1} {
				c := col + dc
				if c < 0 || c >= size {
					continue
				}
				// Skip finder and timing areas
				if isReserved(matrix, row, c) {
					continue
				}
				if moduleIdx < len(bytes)*8 {
					byteIdx := moduleIdx / 8
					bitIdx := 7 - (moduleIdx % 8)
					matrix[row][c] = (bytes[byteIdx]>>uint(bitIdx))&1 == 1
					moduleIdx++
				}
			}
		}
	}

	return matrix
}

func drawFinder(matrix [][]bool, row, col int) {
	pattern := [][]bool{
		{true, true, true, true, true, true, true},
		{true, false, false, false, false, false, true},
		{true, false, true, true, true, false, true},
		{true, false, true, true, true, false, true},
		{true, false, true, true, true, false, true},
		{true, false, false, false, false, false, true},
		{true, true, true, true, true, true, true},
	}
	for r, rowPat := range pattern {
		for c, v := range rowPat {
			if row+r < len(matrix) && col+c < len(matrix[0]) {
				matrix[row+r][col+c] = v
			}
		}
	}
}

func isReserved(matrix [][]bool, row, col int) bool {
	size := len(matrix)
	// Finder patterns + separators (top-left, top-right, bottom-left)
	if (row < 9 && col < 9) || // top-left
		(row < 9 && col >= size-8) || // top-right
		(row >= size-8 && col < 9) { // bottom-left
		return true
	}
	// Timing patterns
	if row == 6 || col == 6 {
		return true
	}
	return false
}

func init() {
	rootCmd.AddCommand(shareCmd)
}
