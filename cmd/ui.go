package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/sumwatshade/surflog/cmd/buoy"
	"github.com/sumwatshade/surflog/cmd/create"
	"github.com/sumwatshade/surflog/cmd/journal"
)

type model struct {
	view     string
	buoyData *buoy.Data

	journal *journal.Journal

	draftEntry *create.Entry
}

func initialModel() model {
	return model{
		view:     "buoy",
		buoyData: nil,
		journal:  journal.NewJournal(),
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "b":
			m.view = "buoy"
		case "j":
			m.view = "journal"
		case "c":
			m.view = "create"
		}

	}

	return m, nil
}

func (m model) View() string {
	switch m.view {
	case "buoy":
		return buoy.View(m.buoyData)
	case "journal":
		return journal.View(m.journal)
	case "create":
		return create.View(m.draftEntry)
	}

	return ""
}
