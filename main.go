package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/sahilmulla/poke-tui/pkgs/pokemon"
	"github.com/sahilmulla/poke-tui/pkgs/styles"
)

type NavState int

const (
	POKEMON_LIST NavState = iota
	POKEMON_DETAILS
	LOADING_POKEMON_DETAILS
)

type model struct {
	pokemonList list.Model
	navState    NavState

	details  pokemon.PokemonDetail
	statBars []progress.Model
}

func initialModel() model {
	return model{
		pokemonList: initialPokemonListModel(),
		navState:    POKEMON_LIST,
		statBars:    make([]progress.Model, 6),
	}
}

func initialPokemonListModel() list.Model {
	l := list.New(getAllPokemonItems(), itemDelegate{}, 0, 0)
	l.SetShowTitle(false)
	l.SetStatusBarItemName("pokemon", "pokemons")
	l.Paginator.Type = paginator.Arabic
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{listAdditionalKeys.Enter}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{listAdditionalKeys.Enter, keys.Quit}
	}
	l.DisableQuitKeybindings()

	return l
}

func (m model) Init() tea.Cmd { return nil }

type (
	keyMap struct {
		Quit key.Binding
	}
	listAdditionalKeyMap struct {
		Enter key.Binding
	}
	detailKeyMap struct {
		Back key.Binding
	}
)

var (
	keys = keyMap{
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
	listAdditionalKeys = listAdditionalKeyMap{
		Enter: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "view"),
		),
	}
	detailKeys = detailKeyMap{
		Back: key.NewBinding(
			key.WithKeys("backspace"),
			key.WithHelp("backspace", "list"),
		),
	}
)

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{detailKeys.Back, k.Quit}
}
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp(), {}}
}

type DetailsMsg struct {
	data pokemon.PokemonDetail
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	cmds := []tea.Cmd{}
	switch msg := msg.(type) {
	case DetailsMsg:
		m.details = msg.data
		for idx, stat := range m.details.Info.Stats {
			newStat := progress.New(
				progress.WithScaledGradient("#e24", "#2b8"),
				progress.WithWidth(25))
			newStat.Full = rune('∎')
			newStat.Empty = rune('-')

			newStat.ShowPercentage = false
			newStat.SetSpringOptions(12.0, 1)
			cmds = append(cmds, newStat.SetPercent(float64(stat.Base)/200))
			m.statBars[idx] = newStat
		}
		m.navState = POKEMON_DETAILS

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, detailKeys.Back):
			switch m.navState {
			case POKEMON_DETAILS:
				m.navState = POKEMON_LIST
			}

		case key.Matches(msg, listAdditionalKeys.Enter):
			switch m.navState {
			case POKEMON_LIST:
				if m.pokemonList.FilterState() != list.Filtering {
					i, ok := m.pokemonList.SelectedItem().(item)
					if ok {
						cmds = append(cmds, getPokemonDetails(i.title))
						m.navState = LOADING_POKEMON_DETAILS
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		w, h := styles.DocStyle.GetFrameSize()
		m.pokemonList.SetSize(msg.Width-w, msg.Height-h)

	case progress.FrameMsg:
		for idx, pModel := range m.statBars {
			newModel, newCmd := pModel.Update(msg)
			m.statBars[idx] = newModel.(progress.Model)
			cmds = append(cmds, newCmd)
		}
	}

	var cmd tea.Cmd
	switch m.navState {
	case POKEMON_LIST:
		m.pokemonList, cmd = m.pokemonList.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func renderPokemonDetails(m model) string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		pokemon.RenderHeader(m.details),
		pokemon.RenderInfo(m.details),
		pokemon.RenderAbilities(m.details),
		pokemon.RenderStats(m.details, &m.statBars),
		pokemon.RenderEvolutionTree(m.details))

	return content
}
func (m model) View() string {
	s := ""

	switch m.navState {
	case LOADING_POKEMON_DETAILS:
		s += lipgloss.Place(40, 10,
			lipgloss.Center, lipgloss.Center,
			lipgloss.NewStyle().Render(" LOADING "),
			lipgloss.WithWhitespaceChars("* "),
			lipgloss.WithWhitespaceForeground(lipgloss.ANSIColor(termenv.ANSIBrightBlack)))
	case POKEMON_LIST:
		s += styles.DocStyle.Render(m.pokemonList.View())
	case POKEMON_DETAILS:
		s += renderPokemonDetails(m)
		s += "\n" + lipgloss.NewStyle().Margin(1, 2).Render(help.New().View(keys))
	}

	return s
}

func main() {
	p := tea.NewProgram(
		initialModel(),
		tea.WithAltScreen(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

func getAllPokemonItems() []list.Item {
	result := []list.Item{}
	data := pokemon.GetPokemons()

	for _, el := range data.Results {
		result = append(result, item{title: el.Name})
	}

	return result
}

var pokemonDetailsCache = map[string]pokemon.PokemonDetail{}

func getPokemonDetails(name string) tea.Cmd {
	return func() tea.Msg {
		cached, hit := pokemonDetailsCache[name]

		if hit {
			return DetailsMsg{data: cached}
		}

		fetched := pokemon.GetPokemon(name)
		pokemonDetailsCache[name] = fetched

		return DetailsMsg{data: fetched}
	}
}

type item struct {
	title string
}

func (i item) Title() string       { return i.title }
func (i item) FilterValue() string { return i.title }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}
	str := styles.FormatTitle(i.Title())

	render := styles.ItemStyle.Render
	if index == m.Index() {
		render = func(s ...string) string {
			return styles.SelectedItemStyle.Render(strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, render(str))
}
