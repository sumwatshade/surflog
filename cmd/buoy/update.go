package buoy

import (
	tea "github.com/charmbracelet/bubbletea"
)

// internal message indicating tide data fetch completed
type tideFetchedMsg struct {
	data *BuoyData
}

// fetchTideCmd performs the HTTP request via the buoy service and returns a tideFetchedMsg
func fetchTideCmd() tea.Cmd {
	return func() tea.Msg {
		svc := NewService()
		td, err := svc.GetTideData()
		return tideFetchedMsg{data: NewBuoyData(td, err)}
	}
}

// HandleUpdate manages buoy-specific updates. It triggers an initial tide fetch
// the first time we get a window size (a proxy for program start) when no data
// has been loaded yet, and applies fetched tide data when received.
func HandleUpdate(data *BuoyData, msg tea.Msg) (*BuoyData, tea.Cmd) {
	switch m := msg.(type) {
	case tea.WindowSizeMsg:
		if data == nil { // trigger initial load once
			return data, fetchTideCmd()
		}
		_ = m // unused otherwise
	case tideFetchedMsg:
		return m.data, nil
	}
	return data, nil
}
