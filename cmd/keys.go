package cmd

import "github.com/charmbracelet/bubbles/key"

// keyMap defines all key bindings for the application. It satisfies key.Map so
// it can be passed directly to bubbles/help.Model for automatic rendering.
type keyMap struct {
	Buoy    key.Binding
	Journal key.Binding
	Create  key.Binding
	Help    key.Binding
	Quit    key.Binding
}

// ShortHelp returns keybindings shown in the mini help view.
func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Buoy, k.Journal, k.Create, k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view (columns).
func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Buoy, k.Journal, k.Create}, // first column
		{k.Help, k.Quit},              // second column
	}
}

// keys is the exported set of key bindings used across the app.
var keys = keyMap{
	Buoy: key.NewBinding(
		key.WithKeys("b"),
		key.WithHelp("b", "buoy view"),
	),
	Journal: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "journal view"),
	),
	Create: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create entry"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
