package styles

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var PokemonTypeToColor = map[string]string{
	"normal":   "#A8A77A",
	"fire":     "#EE8130",
	"water":    "#6390F0",
	"electric": "#F7D02C",
	"grass":    "#7AC74C",
	"ice":      "#96D9D6",
	"fighting": "#C22E28",
	"poison":   "#A33EA1",
	"ground":   "#E2BF65",
	"flying":   "#A98FF3",
	"psychic":  "#F95587",
	"bug":      "#A6B91A",
	"rock":     "#B6A136",
	"ghost":    "#735797",
	"dragon":   "#6F35FC",
	"dark":     "#705746",
	"steel":    "#B7B7CE",
	"fairy":    "#D685AD",
}

var (
	AccentColor = lipgloss.ANSIColor(termenv.ANSIYellow)

	ItemStyle         = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.HiddenBorder()).BorderLeft(true)
	SelectedItemStyle = ItemStyle.Foreground(AccentColor).Bold(true).BorderStyle(lipgloss.ThickBorder()).BorderForeground(AccentColor)

	SectionStyle      = lipgloss.NewStyle().BorderStyle(lipgloss.HiddenBorder()).BorderForeground(AccentColor).Padding(0, 1)
	SectionTitleStyle = lipgloss.NewStyle().Padding(0, 1).MarginBottom(1).Background(AccentColor).Foreground(lipgloss.Color("15")).Bold(true).Transform(strings.ToUpper)

	DocStyle = lipgloss.NewStyle().Margin(1).MarginLeft(0)
)

func FormatTitle(s string) string {
	return cases.Title(language.English).String(strings.ReplaceAll(s, "-", " "))
}
