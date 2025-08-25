package buoy

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart/timeserieslinechart"
	"github.com/charmbracelet/lipgloss"
)

var buoyTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("45"))
var buoyInfoStyle = lipgloss.NewStyle().Faint(true)
var tideErrStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("196")) // red

// section represents a logically grouped portion of the buoy view.
type section struct {
	title string
	lines []string
	err   error
}

func newSection(title string) section { return section{title: title} }
func (s *section) add(line string) {
	if line != "" {
		s.lines = append(s.lines, line)
	}
}

// renderWaveSection builds the wave summary section.
func renderWaveSection(bd *BuoyData) section {
	sec := newSection("Current Wave Conditions")
	if bd == nil {
		sec.add("No data")
		return sec
	}
	if bd.waveErr != nil {
		sec.err = bd.waveErr
		return sec
	}
	if bd.wave == nil {
		sec.add("Loading...")
		return sec
	}
	ws := bd.wave
	ft := func(m float64) float64 { return m * 3.28084 }
	localTs := ws.time.In(time.Local)
	sec.add(fmt.Sprintf("%.1fft sig (swell %.1fft @ %.0fs %s / wind %.1fft @ %.0fs %s)",
		ft(ws.wvht), ft(ws.swellHeight), ws.swellPeriod, ws.swellDirection,
		ft(ws.windWaveHeight), ws.windWavePeriod, ws.windWaveDirection))
	sec.add(fmt.Sprintf("steep %s | avg %.1fs | mean %d° @ %s",
		strings.ToLower(ws.steepness), ws.averagePeriod, ws.meanWaveDirectionDeg, localTs.Format("15:04")))
	return sec
}

// renderTideSection builds the tide timeseries chart and stats.
func renderTideSection(bd *BuoyData) section {
	sec := newSection("Tide")
	if bd == nil {
		sec.add("No data")
		return sec
	}
	if bd.tideErr != nil {
		sec.err = bd.tideErr
		return sec
	}
	if bd.tide == nil || len(bd.tide.points) == 0 {
		sec.add("No tide data")
		return sec
	}
	if len(bd.tide.points) == 1 {
		sec.add("Insufficient tide points")
		return sec
	}
	// Build chart (adapted from previous implementation)
	layout := "2006-01-02 15:04"
	pts := bd.tide.points
	var minTime, maxTime time.Time
	values := make([]float64, len(pts))
	parsedTimes := make([]time.Time, len(pts))
	for i, p := range pts {
		gmt, err := time.ParseInLocation(layout, p.time, time.UTC)
		if err != nil {
			continue
		}
		localTm := gmt.In(time.Local)
		parsedTimes[i] = localTm
		values[i] = p.value
		if i == 0 || localTm.Before(minTime) {
			minTime = localTm
		}
		if i == 0 || localTm.After(maxTime) {
			maxTime = localTm
		}
	}
	if maxTime.IsZero() {
		sec.add("No parsable tide times")
		return sec
	}
	minV, maxV := values[0], values[0]
	for _, v := range values[1:] {
		if v < minV {
			minV = v
		}
		if v > maxV {
			maxV = v
		}
	}
	if minV == maxV {
		maxV += 0.1
		minV -= 0.1
	}
	width, height := 42, 10
	lc := timeserieslinechart.New(width, height)
	lc.SetTimeRange(minTime, maxTime)
	lc.SetViewTimeAndYRange(minTime, maxTime, minV, maxV)
	hours := int(maxTime.Sub(minTime).Hours())
	if hours <= 0 {
		hours = 1
	}
	xStep := 1
	if hours < lc.GraphWidth() {
		xStep = lc.GraphWidth() / hours
		if xStep < 1 {
			xStep = 1
		}
	}
	lc.SetXStep(xStep)
	lc.Model.XLabelFormatter = func(i int, v float64) string { return time.Unix(int64(v), 0).In(time.Local).Format("15:04") }
	for i, tm := range parsedTimes {
		if tm.IsZero() {
			continue
		}
		lc.Push(timeserieslinechart.TimePoint{Time: tm, Value: values[i]})
	}
	lc.DrawBraille()
	now := time.Now()
	if (now.Equal(minTime) || now.After(minTime)) && (now.Equal(maxTime) || now.Before(maxTime)) {
		viewMin, viewMax := lc.Model.ViewMinX(), lc.Model.ViewMaxX()
		if viewMax > viewMin {
			dx := viewMax - viewMin
			xRel := (float64(now.Unix()) - viewMin) / dx
			if xRel < 0 {
				xRel = 0
			} else if xRel > 1 {
				xRel = 1
			}
			col := int(math.Round(xRel * float64(lc.GraphWidth()-1)))
			col += lc.Model.Origin().X
			if lc.Model.YStep() > 0 {
				col += 1
			}
			if col >= 0 && col < lc.Canvas.Width() {
				lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
				for y := 0; y < lc.Model.Origin().Y; y++ {
					p := canvas.Point{X: col, Y: y}
					cell := lc.Canvas.Cell(p)
					if cell.Rune == '│' && cell.Style.GetForeground() != (lipgloss.Style{}).GetForeground() {
						lc.Canvas.SetCell(p, canvas.NewCellWithStyle('│', lineStyle))
					} else {
						lc.Canvas.SetCell(p, canvas.NewCellWithStyle('│', lineStyle))
					}
				}
			}
		}
	}
	sec.add("(ft) timeseries:")
	sec.add(lc.View())
	legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
	sec.add(legendStyle.Render("─") + " " + buoyInfoStyle.Render("Predicted tide"))
	if now := time.Now(); (now.Equal(minTime) || now.After(minTime)) && (now.Equal(maxTime) || now.Before(maxTime)) {
		sec.add(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("│") + " " + buoyInfoStyle.Render("Current time"))
	}
	tzName, _ := minTime.Zone()
	sec.add(fmt.Sprintf("min %.2f / max %.2f | %s - %s %s", minV, maxV, minTime.Format("15:04"), maxTime.Format("15:04"), tzName))
	return sec
}

// View renders buoy data using section-based layout.
func View(data *BuoyData) string {
	if data == nil {
		return buoyInfoStyle.Render("No buoy configured yet. Configure in $HOME/.surflog.yaml")
	}
	sections := []section{renderWaveSection(data), renderTideSection(data)}
	var b strings.Builder
	b.WriteString("\n")
	first := true
	for _, s := range sections {
		if s.err == nil && len(s.lines) == 0 { // skip empty
			continue
		}
		if !first {
			b.WriteString("\n")
		}
		first = false
		if s.title != "" {
			b.WriteString(buoyTitleStyle.Render(s.title))
			b.WriteString("\n")
		}
		if s.err != nil {
			b.WriteString(tideErrStyle.Render(s.err.Error()))
			continue
		}
		for i, line := range s.lines {
			// Chart block (ASCII) already contains newlines internally; print raw
			if strings.ContainsRune(line, '\n') {
				b.WriteString(line)
			} else {
				b.WriteString(buoyInfoStyle.Render(line))
			}
			if i < len(s.lines)-1 {
				b.WriteString("\n")
			}
		}
	}
	return b.String()
}
