package journal

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/sumwatshade/surflog/cmd/create"
)

var journalTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
var journalEmptyStyle = lipgloss.NewStyle().Faint(true)
var journalEntrySpot = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("111"))
var journalEntryMeta = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

func renderEntry(e create.Entry) string {
	lines := []string{
		journalEntrySpot.Render(e.Spot),
		journalEntryMeta.Render(fmt.Sprintf("Location: %s  Waves: %s", e.Location, e.WaveData)),
	}
	if e.Comments != "" {
		lines = append(lines, e.Comments)
	}
	return strings.Join(lines, "\n")
}

func View(journal *Journal) string {
	if journal == nil || len(journal.Entries) == 0 {
		return journalTitleStyle.Render("Journal") + "\n" + journalEmptyStyle.Render("No entries yet. Press 'c' to create one.")
	}
	var rendered []string
	for i := len(journal.Entries) - 1; i >= 0; i-- { // newest first
		rendered = append(rendered, renderEntry(journal.Entries[i]))
	}
	return journalTitleStyle.Render("Journal") + "\n" + strings.Join(rendered, "\n\n")
}
