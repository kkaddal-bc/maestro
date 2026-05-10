package style

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func init() {
	lipgloss.SetColorProfile(termenv.TrueColor)
}

var Success = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
var Secondary = lipgloss.NewStyle().Faint(true)
