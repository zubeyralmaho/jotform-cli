package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBrandColors_Defined(t *testing.T) {
	assert.NotEmpty(t, string(BrandNavy))
	assert.NotEmpty(t, string(BrandOrange))
	assert.NotEmpty(t, string(BrandBlue))
	assert.NotEmpty(t, string(BrandYellow))
	assert.NotEmpty(t, string(White))
	assert.NotEmpty(t, string(DimWhite))
	assert.NotEmpty(t, string(DarkGray))
}

func TestStyles_NotNil(t *testing.T) {
	styles := []struct {
		name  string
		style interface{}
	}{
		{"Title", Title},
		{"Subtitle", Subtitle},
		{"Label", Label},
		{"Value", Value},
		{"Success", Success},
		{"Warning", Warning},
		{"ErrorStyle", ErrorStyle},
		{"Muted", Muted},
		{"Added", Added},
		{"Modified", Modified},
		{"Deleted", Deleted},
		{"Panel", Panel},
		{"ActivePanel", ActivePanel},
		{"ListItem", ListItem},
		{"ActiveListItem", ActiveListItem},
		{"GradientBanner", GradientBanner},
	}

	for _, s := range styles {
		t.Run(s.name, func(t *testing.T) {
			assert.NotNil(t, s.style)
		})
	}
}

func TestStyles_RenderNonEmpty(t *testing.T) {
	assert.NotEmpty(t, Title.Render("test"))
	assert.NotEmpty(t, Subtitle.Render("test"))
	assert.NotEmpty(t, Label.Render("test"))
	assert.NotEmpty(t, Value.Render("test"))
	assert.NotEmpty(t, Success.Render("test"))
	assert.NotEmpty(t, Warning.Render("test"))
	assert.NotEmpty(t, ErrorStyle.Render("test"))
	assert.NotEmpty(t, Muted.Render("test"))
}
