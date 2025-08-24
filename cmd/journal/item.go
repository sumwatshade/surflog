package journal

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sumwatshade/surflog/cmd/create"
)

var (
	itemTitleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("111")).Bold(true)
	itemDescStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	selectedTitleStyle = itemTitleStyle.Copy().Foreground(lipgloss.Color("51"))
	selectedDescStyle  = itemDescStyle.Copy().Foreground(lipgloss.Color("245"))
)

type journalItem struct{ create.Entry }

func (i journalItem) Title() string       { return i.Spot }
func (i journalItem) Description() string { return fmt.Sprintf("%s • %s", i.Location, i.WaveData) }
func (i journalItem) FilterValue() string {
	return strings.ToLower(strings.Join([]string{i.Spot, i.Location, i.WaveData, i.Comments}, " "))
}

type itemDelegate struct{}

func (d itemDelegate) Height() int                               { return 2 }
func (d itemDelegate) Spacing() int                              { return 1 }
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(journalItem)
	if !ok {
		io.WriteString(w, "?")
		return
	}
	title := itemTitleStyle.Render(it.Title())
	desc := itemDescStyle.Render(it.Description())
	if index == m.Index() {
		title = selectedTitleStyle.Render(it.Title())
		desc = selectedDescStyle.Render(it.Description())
	}
	// Highlight filter matches (simple contains highlight for now)
	if f := strings.TrimSpace(m.FilterValue()); f != "" {
		lower := strings.ToLower(title)
		fl := strings.ToLower(f)
		if pos := strings.Index(lower, fl); pos >= 0 {
			// naive highlight
			orig := title[pos : pos+len(f)]
			title = title[:pos] + filterMatchStyle.Render(orig) + title[pos+len(f):]
		}
	}
	io.WriteString(w, lipgloss.JoinVertical(lipgloss.Left, title, desc))
}
