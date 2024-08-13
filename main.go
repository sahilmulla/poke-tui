package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type model struct {
	pokemonList list.Model
	loading     bool
	navState    string
	details     PokemonDetail
	stats       []progress.Model
}

func initialModel() model {
	return model{
		pokemonList: initialPokemonListModel(),
		navState:    "list",
		stats:       make([]progress.Model, 6),
	}
}

var (
	accentColor       = lipgloss.Color("11")
	itemStyle         = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.HiddenBorder()).BorderLeft(true)
	selectedItemStyle = itemStyle.Foreground(accentColor).Bold(true).BorderStyle(lipgloss.ThickBorder()).BorderForeground(accentColor)
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := cases.Title(language.English).String(i.Title())

	render := itemStyle.Render
	if index == m.Index() {
		render = func(s ...string) string {
			return selectedItemStyle.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, render(str))
}

func initialPokemonListModel() list.Model {
	l := list.New(getAllPokemonsItem(), itemDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetStatusBarItemName("pokemon", "pokemons")
	l.Paginator.Type = paginator.Arabic

	return l
}

var docStyle = lipgloss.NewStyle().Margin(1, 0)

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case DetailsMsg:
		m.details = msg.data
		for idx := range m.details.Info.Stats {
			newStat := progress.New(progress.WithScaledGradient("#ff0", "#f0f"), progress.WithWidth(15))
			newStat.Full = rune('█')
			newStat.Empty = rune('─')
			newStat.ShowPercentage = false

			m.stats[idx] = newStat
		}
		m.loading = false

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "backspace":
			switch m.navState {
			case "details":
				m.navState = "list"
			}

		case "enter":
			if m.pokemonList.FilterState() != list.Filtering {
				i, ok := m.pokemonList.SelectedItem().(item)
				if ok {
					m.loading = true
					m.navState = "details"
					cmds = append(cmds, getPokemonDetails(i.title))
				}
			}
		}

	case tea.WindowSizeMsg:
		w, h := docStyle.GetFrameSize()
		m.pokemonList.SetSize(msg.Width-w, msg.Height-h)
	}

	var cmd tea.Cmd
	if m.navState == "list" {
		m.pokemonList, cmd = m.pokemonList.Update(msg)
	}
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	s := ""

	switch m.navState {
	case "list":
		s += docStyle.Render(m.pokemonList.View())
	case "details":
		if m.loading {
			s += "loading details..."
		} else {
			s += renderPokemonDetails(m)
		}
	}

	return s
}

var pokemonTypeColors = map[string]string{
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

func renderPokemonDetails(m model) string {
	title := lipgloss.NewStyle().
		BorderStyle(func() lipgloss.Border {
			b := lipgloss.RoundedBorder()
			b.Right = "├"
			return b
		}()).
		Padding(0, 1).
		Bold(true).
		Render(cases.Title(language.English).String(m.details.Info.Name))

	typeInfo := ""
	typeStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("15")).
		Padding(0, 1).
		Bold(true)
	for _, t := range m.details.Info.Types {
		typeInfo += "─" + typeStyle.Background(lipgloss.Color(pokemonTypeColors[t.Type.Name])).Render(strings.ToUpper(t.Type.Name))
	}

	info := " " + lipgloss.JoinHorizontal(
		lipgloss.Top,
		fmt.Sprintf("Height: %d cms", m.details.Info.HeightDecimeter*10),
		strings.Repeat(" ", 3),
		fmt.Sprintf("Weight: %d kgs", m.details.Info.WeightHectogram/10),
	)

	line := strings.Repeat("─", max(0, lipgloss.Width(info)-lipgloss.Width(title)-lipgloss.Width(typeInfo)+2))

	evolution := renderEvolutionTree([]chain{m.details.Evolutions.Chain}, m.details.Info.Species.Name, 1)

	stats := []string{}
	for idx, stat := range m.details.Info.Stats {
		stats = append(stats, lipgloss.JoinVertical(lipgloss.Left, stat.Stat.Name+fmt.Sprintf(" %d", stat.Base), " ", m.stats[idx].ViewAs(float64(stat.Base)/200)), " ")
	}
	statsView := lipgloss.JoinHorizontal(lipgloss.Center, stats...)

	content := lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.JoinHorizontal(lipgloss.Center, title, line, typeInfo+"─"), " ",
		info, " "+lipgloss.NewStyle().Bold(true).Render(func() string {
			gr := m.details.Species.GenderRate
			const UNKNOWN, MALE, FEMALE, BOTH = "?", "♂️", "♀️", "⚥"
			switch {
			case gr < 0:
				return UNKNOWN
			case gr == 0:
				return MALE
			case gr == 8:
				return FEMALE
			default:
				return BOTH
			}
		}()),
		evolution,
		statsView,
	)

	return content
}

func renderEvolutionTree(c []chain, infoName string, depth int) string {
	if len(c) == 0 {
		return ""
	}

	speciesStyle := func(val string, highlight bool) string {
		s := lipgloss.NewStyle().Bold(highlight).Underline(highlight)

		return s.Render(cases.Title(language.English).String(val))
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
			connector = "└"
		}
		connector = lipgloss.NewStyle().Foreground(lipgloss.Color("15")).Faint(true).Render(" " + connector + "→ ")
		if depth == 1 {
			connector = " "
		}

		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, connector+speciesStyle(evo.Species.Name, evo.Species.Name == infoName), renderEvolutionTree(evo.EvolvesTo, infoName, depth+1)))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func main() {
	if _, err := tea.NewProgram(initialModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) FilterValue() string { return i.title }

func getAllPokemonsItem() []list.Item {
	result := []list.Item{}
	data := getPokemons()

	for _, el := range data.Results {
		result = append(result, item{title: el.Name})
	}

	return result
}

type (
	PokemonsResponse struct {
		Count   int `json:"count"`
		Results []struct {
			Name string `json:"name"`
		} `json:"results"`
	}

	PokemonResponse struct {
		Id              int    `json:"id"`
		HeightDecimeter int    `json:"height"`
		WeightHectogram int    `json:"weight"`
		Name            string `json:"name"`
		Stats           []struct {
			Base int `json:"base_stat"`
			Stat struct {
				Name string `json:"name"`
			} `json:"stat"`
		} `json:"stats"`
		Species struct {
			Name string `json:"name"`
			Url  string `json:"url"`
		} `json:"species"`
		Types []struct {
			Slot int `json:"slot"`
			Type struct {
				Name string `json:"name"`
			} `json:"type"`
		} `json:"types"`
	}
	PokemonSpeciesResponse struct {
		GenderRate     int `json:"gender_rate"`
		EvolutionChain struct {
			Url string `json:"url"`
		} `json:"evolution_chain"`
	}
	chain struct {
		Species struct {
			Name string `json:"name"`
		} `json:"species"`
		EvolvesTo []chain `json:"evolves_to"`
	}
	EvolutionChainResponse struct {
		Chain    chain `json:"chain"`
		Metadata struct {
			LevelMaxWidth map[int]int
			NodeHeight    map[string]int
		}
	}
	PokemonDetail struct {
		Info       PokemonResponse
		Species    PokemonSpeciesResponse
		Evolutions EvolutionChainResponse
	}
	DetailsMsg struct {
		data PokemonDetail
	}
)

func getPokemon(name string) PokemonDetail {
	data := new(PokemonDetail)
	{
		url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", name)
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		if err := json.Unmarshal(body, &data.Info); err != nil {
			fmt.Println(err)
		}
	}
	{
		url := data.Info.Species.Url
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		if err := json.Unmarshal(body, &data.Species); err != nil {
			fmt.Println(err)
		}
	}
	{
		url := data.Species.EvolutionChain.Url
		resp, err := http.Get(url)

		if err != nil {
			fmt.Println(err)
		}

		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}

		if err := json.Unmarshal(body, &data.Evolutions); err != nil {
			fmt.Println(err)
		}

		data.Evolutions.calcLevelMaxWidth()
		data.Evolutions.calcNodeHeight()
	}

	return *data
}

func (d *EvolutionChainResponse) calcNodeHeight() {
	d.Metadata.NodeHeight = make(map[string]int)
	stack := []chain{d.Chain}
	toProcess := []chain{}

	for len(stack) > 0 {
		curr := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		toProcess = append(toProcess, curr)
		stack = append(stack, curr.EvolvesTo...)
	}

	for i := len(toProcess) - 1; i >= 0; i-- {
		curr := toProcess[i]
		totalHeight := 0
		if len(curr.EvolvesTo) == 0 {
			totalHeight = 1
		} else {
			for _, evo := range curr.EvolvesTo {
				totalHeight += max(d.Metadata.NodeHeight[evo.Species.Name], 1)
			}
		}

		d.Metadata.NodeHeight[curr.Species.Name] = totalHeight
	}
}

func (d *EvolutionChainResponse) calcLevelMaxWidth() {
	d.Metadata.LevelMaxWidth = make(map[int]int)

	level := 1
	startIdx := 0
	endIdx := 1
	queue := make([]chain, 0, 10)
	queue = append(queue, d.Chain)
	for startIdx < endIdx {
		var maxWidth int
		nextEndIdx := endIdx

		for i := startIdx; i < endIdx; i++ {
			curr := queue[i]
			if len(curr.Species.Name) > maxWidth {
				maxWidth = len(curr.Species.Name)
			}
			queue = append(queue, curr.EvolvesTo...)
			nextEndIdx += len(curr.EvolvesTo)
		}

		d.Metadata.LevelMaxWidth[level] = maxWidth
		level++

		startIdx = endIdx
		endIdx = nextEndIdx
	}
}

var pokemonDetailsCache = map[string]PokemonDetail{}

func getPokemonDetails(name string) tea.Cmd {
	return func() tea.Msg {
		cached, hit := pokemonDetailsCache[name]

		if hit {
			return DetailsMsg{data: cached}
		}

		fetched := getPokemon(name)
		pokemonDetailsCache[name] = fetched

		return DetailsMsg{data: fetched}
	}
}

func getPokemons() PokemonsResponse {
	url := "https://pokeapi.co/api/v2/pokemon/?limit=99999"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	data := new(PokemonsResponse)
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println(err)
	}

	return *data
}

func PrettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "  ")
	return string(s)
}
