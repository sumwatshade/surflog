package buoy

import (
	tea "github.com/charmbracelet/bubbletea"
)

// internal message indicating tide data fetch completed
type tideFetchedMsg struct {
	tide tideData
	err  error
}

// internal message for wave summary fetch completion
type waveFetchedMsg struct {
	wave waveSummary
	err  error
}

// fetchTideCmd performs the HTTP request via the buoy service and returns a tideFetchedMsg
func fetchTideCmd() tea.Cmd {
	return func() tea.Msg {
		svc := NewService()
		td, err := svc.GetTideData()
		return tideFetchedMsg{tide: td, err: err}
	}
}

// fetchWaveCmd retrieves wave summary (latest .spec reading)
func fetchWaveCmd(data *BuoyData) tea.Cmd {
	return func() tea.Msg {
		svc := NewService()
		ws, err := svc.GetWaveSummary()
		return waveFetchedMsg{wave: ws, err: err}
	}
}

// HandleUpdate manages buoy-specific updates. It triggers an initial tide fetch
// the first time we get a window size (a proxy for program start) when no data
// has been loaded yet, and applies fetched tide data when received.
func HandleUpdate(data *BuoyData, msg tea.Msg) (*BuoyData, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if data == nil { // trigger initial load once
			data = &BuoyData{}
			return data, tea.Batch(fetchTideCmd(), fetchWaveCmd(nil))
		}
		_ = m // unused otherwise
	case tideFetchedMsg:
		data.setTide(m.tide, m.err)
		return data, nil
	case waveFetchedMsg:
		data.setWave(m.wave, m.err)
		return data, nil
	}
	return data, nil
}
