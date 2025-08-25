package create

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var createTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("219"))
var faint = lipgloss.NewStyle().Faint(true)
var errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
var highlight = lipgloss.NewStyle().Foreground(lipgloss.Color("51")).Bold(true)

// View renders the huh form state and supplemental wave info
func View(m *Model) string {
	if m == nil {
		return createTitleStyle.Render("New Entry") + "\n" + faint.Render("(initializing)")
	}
	b := &strings.Builder{}
	fmt.Fprintln(b, createTitleStyle.Render("New Entry"))

	if m.waveErr != nil {
		fmt.Fprintln(b, errStyle.Render("Wave fetch error: "+m.waveErr.Error()))
	}

	if m.waveFetched && m.Entry.WaveSummary.String() != "" {
		fmt.Fprintln(b, faint.Render("\nWave: ")+m.Entry.WaveSummary.String())
	}
	fmt.Fprintln(b, faint.Render("\nDate: ")+m.Entry.SessionAt.Format(time.Kitchen))

	if m.form != nil {
		fmt.Fprintln(b, m.form.View())
	}
	if m.completed && !m.persisted {
		if !m.confirmed {
			fmt.Fprintf(b, "\nReview: %s | %s | %s\n", m.Entry.Spot, m.Entry.SessionAt.Format(time.Kitchen), m.Entry.WaveHeight)
			fmt.Fprintln(b, highlight.Render("Press 'y' to confirm save or 'n' to discard & start over."))
		} else {
			fmt.Fprintf(b, "\nConfirmed. Saving entry...\n")
		}
	}
	return b.String()
}
