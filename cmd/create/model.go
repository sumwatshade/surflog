package create

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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

func (m *Model) buildForm() {
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Spot").Value(&m.spotStr),
			huh.NewSelect[string]().Title("Perceived Wave Height").Options(selectOptions(HeightOptions)...).Value(&m.heightStr),
			huh.NewText().Title("Comments").Value(&m.commentsStr),
		),
	).WithShowHelp(false)
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
