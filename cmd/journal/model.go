package journal

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sumwatshade/surflog/cmd/create"
)

// Journal holds underlying entries plus the interactive list model.
type Journal struct {
	Entries []create.Entry `json:"entries"`
	list    list.Model
	ready   bool
	width   int
	height  int
}

var (
	statusBarStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)
	filterMatchStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("219")).Bold(true)
	journalTitleBarStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
)

// NewJournal constructs a journal with some sample entries.
func NewJournal() *Journal {
	j := &Journal{}
	// Seed with sample entries for now; in future load from disk.
	samples := []create.Entry{
		{Spot: "Lower Trestles", Location: "San Clemente, CA", WaveData: "3-4ft SW", Comments: "Fun morning session, light offshore."},
		{Spot: "Mavericks", Location: "Half Moon Bay, CA", WaveData: "18ft NW", Comments: "Just watching, not paddling out."},
		{Spot: "Pipeline", Location: "Oahu, HI", WaveData: "8-10ft N", Comments: "Heavy barrels, crowded."},
		{Spot: "Rincon", Location: "Santa Barbara, CA", WaveData: "5ft W", Comments: "Long peeling rights, glassy."},
		{Spot: "Malibu", Location: "Malibu, CA", WaveData: "2-3ft S", Comments: "Cruisy logs everywhere."},
	}
	for _, e := range samples {
		j.AddEntry(e)
	}
	return j
}

// AddEntry appends to underlying slice and (if list initialized) inserts item.
func (j *Journal) AddEntry(entry create.Entry) {
	j.Entries = append(j.Entries, entry)
	if j.ready {
		j.list.InsertItem(0, journalItem{entry}) // newest first
	}
}

// ensureList creates or resizes the list model based on dimensions.
func (j *Journal) ensureList(width, height int) {
	if width == 0 || height == 0 {
		return
	}
	j.width = width
	j.height = height
	listHeight := max(5, height-6) // leave space for header/footer around view
	if !j.ready {
		items := make([]list.Item, 0, len(j.Entries))
		// newest first
		for i := len(j.Entries) - 1; i >= 0; i-- {
			items = append(items, journalItem{j.Entries[i]})
		}
		l := list.New(items, itemDelegate{}, width-4, listHeight) // -4 for padding
		l.Title = "Journal"
		l.SetShowStatusBar(true)
		l.SetShowPagination(true)
		l.SetFilteringEnabled(true)
		l.Styles.Title = journalTitleBarStyle
		l.Styles.StatusBar = statusBarStyle
		l.Styles.PaginationStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		l.Styles.HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
		j.list = l
		j.ready = true
		return
	}
	// resize
	j.list.SetSize(width-4, listHeight)
}

// Update handles messages specific to the journal list.
func (j *Journal) Update(msg tea.Msg, width, height int) tea.Cmd {
	j.ensureList(width, height)
	if !j.ready {
		return nil
	}
	switch m := msg.(type) {
	case tea.KeyMsg:
		switch m.String() {
		case "esc":
			if j.list.FilterState() == list.Filtering {
				j.list.ResetFilter()
				return nil
			}
		}
	}
	var cmd tea.Cmd
	j.list, cmd = j.list.Update(msg)
	return cmd
}

// View renders the journal list.
func (j *Journal) View() string {
	if !j.ready {
		return journalTitleBarStyle.Render("Journal") + "\n" + "Loading..."
	}
	if len(j.Entries) == 0 {
		return journalTitleBarStyle.Render("Journal") + "\n" + lipgloss.NewStyle().Faint(true).Render("No entries yet. Press 'c' to create one.")
	}
	return j.list.View()
}

// helper until Go generics version or shared util
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
