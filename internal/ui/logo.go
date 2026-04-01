package ui

import (
    "strings"

    "github.com/charmbracelet/lipgloss"
)

// Renkler zaten theme.go dosyasında tanımlı olduğu için 
// burada tekrar var (...) bloğu açmıyoruz.
var(
	DarkGrey    = lipgloss.Color("#444444")
)

// buildLogo, logo bloğunu ANSI reset kaçaklarını önleyerek güvenli bir şekilde inşa eder.
func buildLogo() string {
	// Temel beyaz arkaplan stili
	bg := lipgloss.NewStyle().Background(White)

	b := bg.Foreground(BrandBlue).Render
	o := bg.Foreground(BrandOrange).Render
	y := bg.Foreground(BrandYellow).Render
	n := bg.Foreground(BrandNavy).Render

	// KRİTİK NOKTA: İç boşlukların siyah çıkmasını engellemek için 
	// boşlukları da özel olarak beyaz arka planla render eden yardımcı fonksiyon.
	sp := func(width int) string {
		return bg.Render(strings.Repeat(" ", width))
	}

	lines := []string{
		sp(10) + b("▄██▄") + sp(4) + o("▄██▄"),
		sp(8) + b("▄████▀") + sp(2) + o("▄████▀"),
		sp(6) + b("▄████▀") + sp(2) + o("▄████▀") + sp(2),
		sp(6) + b("▀██▀") + sp(2) + o("▄████▀") + sp(2) ,
		sp(10) + o("▄████▀") + sp(3) + y("▄██▄"),
		sp(10) + o("▀██▀") + sp(3) + y("▄████▀"),
		sp(6) + n("▄▄") + sp(9) + y("▀██▀"),
		sp(6) + n("███"),
		sp(6), // Navy'nin altına eklediğimiz beyaz boşluk satırı
	}

	return strings.Join(lines, "\n")
}

// Logo returns the Jotform geometric logo.
func Logo() string {
	logo := buildLogo()

	return lipgloss.NewStyle().
		Background(White).
		Padding(1, 2).
		Render(logo)
}

// LogoCompact returns a smaller single-line stylized logo mark.
func LogoCompact() string {
	b := lipgloss.NewStyle().Foreground(BrandBlue).Render
	o := lipgloss.NewStyle().Foreground(BrandOrange).Render
	y := lipgloss.NewStyle().Foreground(BrandYellow).Render
	n := lipgloss.NewStyle().Foreground(BrandNavy).Render

	return n("◤") + b("▰") + o("▰") + y("▰")
}

// LogoWithText returns the logo next to the wordmark flawlessly aligned 
// inside a seamless white brand island.
func LogoWithText() string {
	// Logoyu alıyoruz ve sağ tarafında tırtık olmaması için Width(28) ile sabitliyoruz
	logoMark := lipgloss.NewStyle().
		Background(White).
		Width(28).
		Render(buildLogo())

	bg := lipgloss.NewStyle().Background(White)

	wordmark := bg.Foreground(BrandOrange).Bold(true).Render("jotform")
	tagline := bg.Foreground(DarkGrey).Render("Form-as-Code CLI")

	textLines := lipgloss.JoinVertical(lipgloss.Left,
		"",
		"",
		"         " + wordmark,
		tagline,
		"",
		"",
		"",
		"",
		"", // Logonun yüksekliğiyle eşleşmesi için boşluklar eklendi
	)

	// Metin bloğunu da sabitliyoruz
	textBlock := lipgloss.NewStyle().
		Background(White).
		Width(22).
		Render(textLines)

	// İki bloğu yan yana getiriyoruz
	combined := lipgloss.JoinHorizontal(lipgloss.Top, logoMark, textBlock)

	// Tüm yapıyı son kez padding ile sarıp mükemmel beyaz kutuyu oluşturuyoruz
	return lipgloss.NewStyle().
		Background(White).
		Padding(1, 2).
		Render(combined)
}