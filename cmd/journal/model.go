package journal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/viper"
	"github.com/sumwatshade/surflog/cmd/create"
)

// Journal holds underlying entries plus the interactive list model.
type Journal struct {
	Entries []create.Entry `json:"entries"`
	list    list.Model
	ready   bool
	width   int
	height  int
	detail  bool // whether we're showing a single entry
	svc     Service
	// deletion state
	confirmingDelete bool   // user pressed delete, awaiting confirmation
	deleteTargetID   string // id of entry pending deletion
}

var (
	statusBarStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Padding(0, 1)
	filterMatchStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("219")).Bold(true)
	journalTitleBarStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99"))
	detailHeaderStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("81")).Underline(true)
	detailMetaStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	faintStyle           = lipgloss.NewStyle().Faint(true)
)

// NewJournal constructs a journal loading entries via the service rooted in user config dir.
func NewJournal() *Journal {
	j := &Journal{}
	// Assume viper always has journal.dir (set via default in initConfig or user override)
	dir := strings.TrimSpace(viper.GetString("journal.dir"))
	if dir != "" {
		// expand leading ~ or make relative absolute
		if strings.HasPrefix(dir, "~") {
			if home, herr := os.UserHomeDir(); herr == nil {
				dir = filepath.Join(home, strings.TrimPrefix(dir, "~"))
			}
		} else if !filepath.IsAbs(dir) {
			if wd, werr := os.Getwd(); werr == nil {
				dir = filepath.Join(wd, dir)
			}
		}
		if svc, serr := NewFileService(dir); serr == nil {
			if list, lerr := svc.List(); lerr == nil {
				j.Entries = append(j.Entries, list...)
				j.sortEntries()
			}
			j.svc = svc
		}
	}
	return j
}

// AddEntry appends to underlying slice and (if list initialized) inserts item.
func (j *Journal) AddEntry(entry create.Entry) {
	j.Entries = append(j.Entries, entry)
	j.sortEntries()
	if j.ready {
		j.refreshListItems()
	}
}

// Persist creates the entry via the underlying service (if available) and adds it to the list.
func (j *Journal) Persist(entry create.Entry) (create.Entry, error) {
	if j.svc == nil {
		return create.Entry{}, errors.New("journal service unavailable")
	}
	saved, err := j.svc.Create(entry)
	if err != nil {
		return create.Entry{}, err
	}
	j.AddEntry(saved)
	return saved, nil
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
		j.sortEntries()
		items := make([]list.Item, 0, len(j.Entries))
		for i := 0; i < len(j.Entries); i++ { // already sorted desc
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
			if j.detail { // leave detail view
				j.detail = false
				return nil
			}
			if j.confirmingDelete { // cancel deletion
				j.confirmingDelete = false
				j.deleteTargetID = ""
				return nil
			}
			if j.list.FilterState() == list.Filtering {
				j.list.ResetFilter()
				return nil
			}
		case "enter":
			// open detail (even if filtering; keep filter applied so selection context remains)
			j.detail = true
			return nil
		case "x", "delete": // initiate delete (x common; delete key if sent)
			if j.confirmingDelete { // treat as cancel if repeated
				j.confirmingDelete = false
				j.deleteTargetID = ""
				return nil
			}
			if sel, ok := j.list.SelectedItem().(journalItem); ok {
				j.confirmingDelete = true
				j.deleteTargetID = sel.ID
			}
			return nil
		case "y": // confirm deletion if in confirmation state
			if j.confirmingDelete && j.deleteTargetID != "" {
				id := j.deleteTargetID
				j.confirmingDelete = false
				j.deleteTargetID = ""
				return j.deleteEntry(id)
			}
		case "n": // cancel deletion
			if j.confirmingDelete {
				j.confirmingDelete = false
				j.deleteTargetID = ""
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
	// show delete confirmation banner if active
	if j.confirmingDelete {
		var spot string
		if sel, ok := j.list.SelectedItem().(journalItem); ok {
			spot = sel.Spot
		}
		banner := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Bold(true).Render("Delete entry '" + spot + "'? (y/n)")
		return banner + "\n" + j.list.View()
	}
	if j.detail {
		// render selected entry in full page
		sel, ok := j.list.SelectedItem().(journalItem)
		if !ok {
			j.detail = false
			return j.list.View()
		}
		b := &strings.Builder{}
		fmt.Fprintln(b, journalTitleBarStyle.Render("Journal Entry"))
		fmt.Fprintln(b)
		fmt.Fprintln(b, detailHeaderStyle.Render(sel.Spot))
		fmt.Fprintln(b, detailMetaStyle.Render(sel.WaveSummary.String()))
		if sel.Comments != "" {
			fmt.Fprintln(b)
			fmt.Fprintln(b, sel.Comments)
		}
		fmt.Fprintln(b)
		fmt.Fprintln(b, faintStyle.Render("(esc to go back)"))
		return lipgloss.NewStyle().Width(j.width - 4).Render(b.String())
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

// deleteEntry removes an entry by id from service, underlying slice, and list model.
func (j *Journal) deleteEntry(id string) tea.Cmd {
	if id == "" || j.svc == nil { // nothing to do
		return nil
	}
	// delete from service (ignore error for now but could surface)
	_ = j.svc.Delete(id)
	// remove from Entries slice
	for i := range j.Entries {
		if j.Entries[i].ID == id {
			j.Entries = append(j.Entries[:i], j.Entries[i+1:]...)
			break
		}
	}
	if j.ready {
		// rebuild list items (simpler vs removing by index due to filtering)
		items := make([]list.Item, 0, len(j.Entries))
		for i := len(j.Entries) - 1; i >= 0; i-- { // newest first
			items = append(items, journalItem{j.Entries[i]})
		}
		j.list.SetItems(items)
	}
	return nil
}

// sortEntries orders Entries by SessionAt (newest first). Falls back to CreatedAt when SessionAt zero.
func (j *Journal) sortEntries() {
	parse := func(e create.Entry) time.Time {
		if !e.SessionAt.IsZero() {
			return e.SessionAt
		}
		if t, err := time.Parse(time.RFC3339, strings.TrimSpace(e.CreatedAt)); err == nil {
			return t
		}
		return time.Time{}
	}
	sort.SliceStable(j.Entries, func(i, k int) bool {
		return parse(j.Entries[i]).After(parse(j.Entries[k]))
	})
}

// refreshListItems rebuilds list items from sorted Entries.
func (j *Journal) refreshListItems() {
	if !j.ready {
		return
	}
	j.sortEntries()
	items := make([]list.Item, 0, len(j.Entries))
	for _, e := range j.Entries {
		items = append(items, journalItem{e})
	}
	j.list.SetItems(items)
}
