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
	rightView  string // "journal" or "create"
	buoyData   *buoy.BuoyData
	journal    *journal.Journal
	draftEntry *create.Entry
	width      int
	height     int
	// help / key bindings
	keys keyMap
	help bhelp.Model
}

func initialModel() model {
	return model{rightView: "journal", buoyData: nil, journal: journal.NewJournal(), keys: keys, help: bhelp.New()}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Journal):
			m.rightView = "journal"
		case key.Matches(msg, m.keys.Create):
			m.rightView = "create"
		}
	}

	// buoy update (always run; it internally no-ops when not needed)
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.buoyData, cmd = buoy.HandleUpdate(m.buoyData, msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}

	// propagate updates to active right pane
	if m.rightView == "journal" && m.journal != nil {
		cmd = m.journal.Update(msg, rightPaneWidth(m.width), m.height)
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	// create view currently static
	if len(cmds) == 0 {
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	left := buoy.View(m.buoyData)
	var right string
	switch m.rightView {
	case "journal":
		if m.journal != nil {
			right = m.journal.View()
		} else {
			right = "journal unavailable"
		}
	case "create":
		right = create.View(m.draftEntry)
	default:
		right = "unknown"
	}

	// determine split sizes (30% left min width 24)
	leftW := max(24, int(float64(m.width)*0.3))
	rightW := max(20, m.width-leftW-1)
	leftRendered := lipgloss.NewStyle().Width(leftW).Render(contentStyle.Render(left))
	rightRendered := lipgloss.NewStyle().Width(rightW).Render(contentStyle.Render(right))
	columns := lipgloss.JoinHorizontal(lipgloss.Top, leftRendered, dividerStyle.Render("│"), rightRendered)

	header := headerStyle.Render(appTitle) + " " + tabs(m.rightView, max(0, m.width-10))
	sep := dividerStyle.Render(lipgloss.NewStyle().Width(m.width).Render(strings.Repeat("─", max(0, m.width))))
	foot := m.help.View(m.keys)
	layout := lipgloss.JoinVertical(lipgloss.Left, header, sep, columns, sep, foot)
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

// helper to compute right pane width for updates
func rightPaneWidth(total int) int {
	leftW := max(24, int(float64(total)*0.3))
	return max(20, total-leftW-1)
}
