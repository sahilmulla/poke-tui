package pokemon

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
	"github.com/sahilmulla/poke-tui/pkgs/domain"
	"github.com/sahilmulla/poke-tui/pkgs/styles"
)

func RenderVarieties(d PokemonDetail) string {
	varieties := []string{}
	for _, v := range d.Species.Varieties {
		varieties = append(varieties, styles.TransformTitle(v.Pokemon.Name))
	}

	return styles.SectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		styles.SectionTitleStyle.MarginBottom(1).Render("FORMS"),
		lipgloss.NewStyle().Italic(true).Render(
			wordwrap.String(strings.Join(varieties, ", "), 48))))
}

func heightValue(value float64, unit domain.Unit) string {
	switch unit {
	case domain.US:
		in := value * 3.937
		ft := int(in / 12)
		roundedInches := int(math.Round(in - float64(ft*12)))
		if roundedInches == 12 {
			ft++
			roundedInches = 0
		}

		return fmt.Sprintf(`%d' %d"`, ft, roundedInches)
	default:
		return fmt.Sprintf("%.1f m", value/10)
	}
}
func weightValue(value float64, unit domain.Unit) string {
	switch unit {
	case domain.US:
		return fmt.Sprintf("%.1f lbs", value/4.536)
	default:
		return fmt.Sprintf("%.1f kgs", value/10)
	}
}
func RenderInfo(d PokemonDetail, unit domain.Unit) string {
	height := heightValue(float64(d.Info.HeightDecimeter), unit)
	weight := weightValue(float64(d.Info.WeightHectogram), unit)

	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		styles.SectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, styles.SectionTitleStyle.MarginBottom(1).Render("HEIGHT"), height)),
		styles.SectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, styles.SectionTitleStyle.MarginBottom(1).Render("WEIGHT"), weight)),
		styles.SectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, styles.SectionTitleStyle.MarginBottom(1).Render("GENDER %"), func() string {
			gr := float32(d.Species.GenderRate)
			switch {
			case gr < 0:
				return "Genderless"
			default:
				return fmt.Sprintf("%.1f M %.1f F", (8-gr)/8*100, gr/8*100)
			}
		}())),
		styles.SectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, styles.SectionTitleStyle.MarginBottom(1).Render("Capture Rate"), fmt.Sprintf("%d", d.Species.CaptureRate))),
	)
}

func RenderAbilities(d PokemonDetail) string {
	abilities := []string{}
	for _, a := range d.Info.Abilities {
		slot := fmt.Sprintf("[%d]", a.Slot)
		if a.IsHidden {
			slot = "HID"
		}
		slot = lipgloss.NewStyle().Foreground(styles.AccentColor).Render(slot)
		abilities = append(abilities, slot+" "+styles.TransformTitle(a.Ability.Name))
	}
	return styles.SectionStyle.Render(
		lipgloss.JoinHorizontal(lipgloss.Top,
			styles.SectionTitleStyle.UnsetMarginBottom().MarginRight(3).Render("Abilities"),
			lipgloss.JoinVertical(lipgloss.Left, abilities...)))
}

func RenderEvolutionTree(d PokemonDetail) string {
	babyLegend := lipgloss.NewStyle().Foreground(styles.IsBabyColor).Render("■") + " BABY"
	title := styles.SectionTitleStyle.MarginRight(3).Render("Evolutions")
	return styles.SectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Top, title, babyLegend),
		renderEvolutionTree([]chain{d.Evolutions.Chain}, d.Info.Species.Name, 1)))
}

func renderEvolutionTree(c []chain, infoName string, depth int) string {
	if len(c) == 0 {
		return ""
	}

	speciesStyle := func(val string, isBaby, highlight bool) string {
		s := lipgloss.NewStyle().
			Bold(highlight).
			Underline(highlight)

		if isBaby {
			s = s.Foreground(lipgloss.ANSIColor(styles.IsBabyColor))
		}

		return s.Render(styles.TransformTitle(val))
	}

	rows := []string{}
	for idx, evo := range c {
		connector := "├"
		switch {
		case idx == 0 && idx == len(c)-1:
			connector = "─"
		case idx == 0:
			connector = "┬"
		case idx == len(c)-1:
			// connector = "└"
			connector = "╰"
		}
		connector = lipgloss.NewStyle().Faint(true).Render(" " + connector + "→ ")
		if depth == 1 {
			connector = ""
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, connector+speciesStyle(evo.Species.Name, evo.IsBaby, evo.Species.Name == infoName), renderEvolutionTree(evo.EvolvesTo, infoName, depth+1)))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func RenderStats(d PokemonDetail, statBars *[]progress.Model) string {
	statWidth := 24
	stats := []string{styles.SectionTitleStyle.MarginBottom(1).Render("Base Stats")}
	statVal, statTitle := "", ""
	for idx, stat := range d.Info.Stats {
		statTitle = styles.TransformTitle(stat.Stat.Name)
		statVal = lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", stat.Base))
		statTitle += lipgloss.NewStyle().Faint(true).Render(strings.Repeat(".", statWidth-len(statTitle)-lipgloss.Width(statVal)))
		stats = append(stats, statTitle+statVal+" ["+(*statBars)[idx].View()+"]")
	}

	stats = append(stats, lipgloss.NewStyle().Faint(true).Render(strings.Repeat(".", statWidth)))

	statTitle = "TOTAL"
	statVal = lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%d", d.Info.StatsTotal))
	statTitle += lipgloss.NewStyle().Faint(true).Render(strings.Repeat(".", statWidth-len(statTitle)-lipgloss.Width(statVal)))
	stats = append(stats, lipgloss.NewStyle().Render(statTitle+statVal))

	return styles.SectionStyle.Render(lipgloss.JoinVertical(lipgloss.Left, stats...))
}

func RenderHeader(d PokemonDetail) string {
	title := lipgloss.NewStyle().
		BorderStyle(func() lipgloss.Border {
			b := lipgloss.RoundedBorder()
			b.Right = "├"
			return b
		}()).
		Padding(0, 1).
		Bold(true).
		Render(styles.TransformTitle(d.Info.Name))

	typeInfo := ""
	typeStyle := lipgloss.NewStyle().
		Padding(0, 1).
		Bold(true)
	for _, t := range d.Info.Types {
		typeInfo += "─" + typeStyle.Background(lipgloss.Color(styles.PokemonTypeToColor[t.Type.Name])).Render(strings.ToUpper(t.Type.Name))
	}

	line := strings.Repeat("─", max(0, 56-lipgloss.Width(title)-lipgloss.Width(typeInfo)+2))
	return lipgloss.JoinHorizontal(lipgloss.Center, title, line, typeInfo+"─")
}
