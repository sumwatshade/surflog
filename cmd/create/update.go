package create

import (
	tea "github.com/charmbracelet/bubbletea"
)

// UpdateModel updates the creation form model and returns potential command.
func UpdateModel(m *Model, msg tea.Msg) (*Model, tea.Cmd) {
	if m == nil {
		m = NewModel()
	}
	switch msg := msg.(type) {
	case waveSummaryMsg:
		if msg.Err != nil {
			m.waveErr = msg.Err
		} else {
			m.Entry.WaveSummary = msg.Summary
			m.waveFetched = true
		}
		return m, nil
	}

	// If form completed but not confirmed/persisted, watch for confirmation keys.
	if m.completed && !m.confirmed && !m.persisted {
		if km, ok := msg.(tea.KeyMsg); ok {
			s := km.String()
			if s == "y" || s == "enter" { // confirm save
				m.confirmed = true
				return m, nil
			}
			if s == "n" || s == "esc" { // discard and reset
				return NewModel(), nil
			}
		}
	}
	cmd := m.Update(msg)
	return m, cmd
}
