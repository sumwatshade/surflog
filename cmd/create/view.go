package create

import "github.com/charmbracelet/lipgloss"

var createTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("219"))
var createNoteStyle = lipgloss.NewStyle().Faint(true)

func View(draftEntry *Entry) string {
	if draftEntry == nil {
		return createTitleStyle.Render("New Entry") + "\n" + createNoteStyle.Render("Entry form not implemented yet. Future: interactive inputs for spot, wave data, comments.")
	}
	return createTitleStyle.Render("New Entry (Draft)")
}
