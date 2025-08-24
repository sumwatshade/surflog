package cmd

import (
	"strings"

	bhelp "github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sumwatshade/surflog/cmd/buoy"
	"github.com/sumwatshade/surflog/cmd/create"
	"github.com/sumwatshade/surflog/cmd/journal"
)

type model struct {
	view     string
	buoyData *buoy.Data

	journal *journal.Journal

	draftEntry *create.Entry

	width  int
	height int

	// help / key bindings
	keys keyMap
	help bhelp.Model
}

func initialModel() model {
	m := model{
		view:     "buoy",
		buoyData: nil,
		journal:  journal.NewJournal(),
		keys:     keys,
		help:     bhelp.New(),
	}
	return m
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Buoy):
			m.view = "buoy"
		case key.Matches(msg, m.keys.Journal):
			m.view = "journal"
		case key.Matches(msg, m.keys.Create):
			m.view = "create"
		}
	}

	switch m.view {
	case "buoy":
		return buoy.HandleUpdate(m, msg)
	case "journal":
		return journal.HandleUpdate(m, msg)
	case "create":
		return create.HandleUpdate(m, msg)
	}

	return m, nil
}

func (m model) View() string {
	// Base content per view
	var content string
	switch m.view {
	case "buoy":
		content = buoy.View(m.buoyData)
	case "journal":
		content = journal.View(m.journal)
	case "create":
		content = create.View(m.draftEntry)
	default:
		content = "unknown view"
	}

	// Wrap content
	body := contentStyle.Render(content)

	header := headerStyle.Render(appTitle) + " " + tabs(m.view, max(0, m.width-10))
	sep := dividerStyle.Render(lipgloss.NewStyle().Width(m.width).Render(strings.Repeat("─", max(0, m.width))))
	foot := m.help.View(m.keys)

	layout := lipgloss.JoinVertical(lipgloss.Left, header, sep, body, sep, foot)
	if m.width > 0 {
		layout = lipgloss.NewStyle().Width(m.width).Render(layout)
	}
	return layout
}

// small helper until Go 1.21+ min/max generics maybe
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
