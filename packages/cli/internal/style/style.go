package style

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

var Accent = lipgloss.NewStyle().Foreground(lipgloss.Color("51"))
