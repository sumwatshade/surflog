package create

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/sumwatshade/surflog/cmd/buoy"
)

// Entry represents a single surf journal entry.
// ID is assigned by the journal service when creating a new entry.
type Entry struct {
	ID          string           `json:"id"`
	Spot        string           `json:"spot"`
	WaveHeight  string           `json:"wave_height"`
	WaveSummary buoy.WaveSummary `json:"wave_summary"`
	SessionAt   time.Time        `json:"session_at"`
	Comments    string           `json:"comments"`
	CreatedAt   string           `json:"created_at"`
}

// Height options for perceived wave height.
var HeightOptions = []string{"Ankle", "Knee", "Waist", "Chest", "Shoulder", "Head", "Overhead"}

// Model using huh form
type Model struct {
	Entry          Entry
	form           *huh.Form
	spotInput      *huh.Input // keep reference to first input to force focus
	waveService    buoy.Service
	waveErr        error
	waveFetched    bool
	timeStr        string
	spotStr        string
	heightStr      string
	commentsStr    string
	persisted      bool
	completed      bool // form has been completed
	confirmed      bool // user confirmed save
	lastTimeParsed string
}

func NewModel() *Model {
	m := &Model{waveService: buoy.NewService()}
	now := time.Now()
	def := time.Date(now.Year(), now.Month(), now.Day(), 7, 30, 0, 0, now.Location())
	m.timeStr = def.Format("2006-01-02 15:04")
	m.heightStr = HeightOptions[0]
	m.buildForm()
	return m
}

// Focus first input (spot) for convenience.
func (m *Model) Focus() {
	if m == nil || m.form == nil {
		return
	}
	if m.spotInput != nil {
		// Attempt to focus; ignore if library signature differs.
		_ = m.spotInput.Focus
		m.spotInput.Focus()
	}
}

func (m *Model) buildForm() {
	spot := huh.NewInput().Title("Spot").Value(&m.spotStr)
	m.spotInput = spot
	m.form = huh.NewForm(
		huh.NewGroup(
			spot,
			huh.NewSelect[string]().Title("Perceived Wave Height").Options(selectOptions(HeightOptions)...).Value(&m.heightStr),
			huh.NewText().Title("Comments").Value(&m.commentsStr),
		),
	).WithShowHelp(false).WithTheme(oceanTheme())
	// Explicit first-field focus.
	m.Focus()
}

func selectOptions(vals []string) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(vals))
	for _, v := range vals {
		opts = append(opts, huh.NewOption(v, v))
	}
	return opts
}

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	if m == nil {
		return nil
	}
	if m.form == nil {
		m.buildForm()
	}
	var cmd tea.Cmd
	if updated, ucmd := m.form.Update(msg); ucmd != nil {
		cmd = ucmd
		if f, ok := updated.(*huh.Form); ok {
			m.form = f
		}
	}
	if m.form.State == huh.StateCompleted && !m.completed {
		m.completed = true
		m.Entry.Spot = m.spotStr
		m.Entry.WaveHeight = m.heightStr
		m.Entry.Comments = m.commentsStr
		m.Entry.SessionAt = parseTimeOrDefault(m.timeStr)
		return cmd
	}
	if !m.waveFetched && m.timeStr != m.lastTimeParsed {
		if _, err := time.Parse("2006-01-02 15:04", m.timeStr); err == nil {
			m.lastTimeParsed = m.timeStr
			return tea.Batch(cmd, m.fetchWaveSummaryCmd())
		}
	}
	return cmd
}

func parseTimeOrDefault(v string) time.Time {
	if t, err := time.Parse("2006-01-02 15:04", v); err == nil {
		return t
	}
	if t2, err := time.Parse("15:04", v); err == nil {
		now := time.Now()
		return time.Date(now.Year(), now.Month(), now.Day(), t2.Hour(), t2.Minute(), 0, 0, now.Location())
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 7, 30, 0, 0, now.Location())
}

func (m *Model) fetchWaveSummaryCmd() tea.Cmd {
	return func() tea.Msg {
		ws, err := m.waveService.GetWaveSummary()
		return waveSummaryMsg{Summary: ws, Err: err}
	}
}

// IsDraft indicates form not yet completed.
func (m *Model) IsDraft() bool { return m != nil && !m.completed }

// IsDoneAndUnpersisted returns true only after user confirmed save.
func (m *Model) IsDoneAndUnpersisted() bool {
	return m != nil && m.completed && m.confirmed && !m.persisted
}
func (m *Model) MarkPersisted() {
	if m != nil {
		m.persisted = true
	}
}

type waveSummaryMsg struct {
	Summary buoy.WaveSummary
	Err     error
}

// oceanTheme builds a custom ocean-colored theme matching application palette.
func oceanTheme() *huh.Theme {
	t := huh.ThemeBase()
	deep := lipgloss.Color("24")    // deep blue background accent
	cyan := lipgloss.Color("44")    // cyan titles
	accent := lipgloss.Color("159") // seafoam accent
	grey := lipgloss.Color("246")   // text
	faint := lipgloss.Color("245")  // faint text
	errCol := lipgloss.Color("203") // error

	t.FieldSeparator = lipgloss.NewStyle().SetString("\n\n")

	// Focused field styles
	t.Focused.Base = lipgloss.NewStyle().PaddingLeft(1).BorderStyle(lipgloss.ThickBorder()).BorderLeft(true).BorderForeground(cyan)
	t.Focused.Title = t.Focused.Title.Foreground(cyan).Bold(true)
	t.Focused.Description = t.Focused.Description.Foreground(faint)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(errCol)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(errCol)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(accent)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(accent)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(accent)
	t.Focused.Option = t.Focused.Option.Foreground(grey)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(accent)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(accent)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(grey)
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(faint)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color("15")).Background(cyan)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(grey).Background(deep)
	t.Focused.Next = t.Focused.FocusedButton
	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(accent)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(faint)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(cyan)
	t.Focused.TextInput.Text = t.Focused.TextInput.Text.Foreground(grey)

	// Blurred copies focused, hides border, neutral indicators
	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.SelectSelector = lipgloss.NewStyle().SetString("  ")
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()
	t.Blurred.FocusedButton = t.Focused.FocusedButton
	t.Blurred.TextInput.Cursor = t.Focused.TextInput.Cursor
	t.Blurred.TextInput.Placeholder = t.Focused.TextInput.Placeholder
	t.Blurred.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(faint)
	t.Blurred.TextInput.Text = t.Focused.TextInput.Text

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
	return t
}
