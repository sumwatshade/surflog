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

// View renders buoy and tide information including a sparkline of today's tide predictions.
func View(data *BuoyData) string {
	b := &strings.Builder{}
	b.WriteString(buoyTitleStyle.Render("Buoy Data"))
	b.WriteString("\n")
	if data == nil {
		b.WriteString(buoyInfoStyle.Render("No buoy configured yet. Configure in $HOME/.surflog.yaml"))
		return b.String()
	}
	if data.id != "" {
		b.WriteString("Buoy ID: ")
		b.WriteString(data.id)
		b.WriteString("\n")
	}
	if data.tideErr != nil {
		b.WriteString(tideErrStyle.Render("tide error: "))
		b.WriteString(tideErrStyle.Render(data.tideErr.Error()))
		return b.String()
	}
	if data.tide != nil && len(data.tide.points) > 1 {
		// Parse times & collect values
		layout := "2006-01-02 15:04"
		pts := data.tide.points
		var minTime, maxTime time.Time
		values := make([]float64, len(pts))
		parsedTimes := make([]time.Time, len(pts))
		for i, p := range pts {
			// Parse NOAA GMT timestamp then convert to local time for display
			gmt, err := time.ParseInLocation(layout, p.time, time.UTC)
			if err != nil {
				continue // skip malformed
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
		if maxTime.IsZero() { // fallback
			b.WriteString(buoyInfoStyle.Render("No parsable tide times"))
			return b.String()
		}
		// compute min/max values
		minV, maxV := values[0], values[0]
		for _, v := range values[1:] {
			if v < minV {
				minV = v
			}
			if v > maxV {
				maxV = v
			}
		}
		if minV == maxV { // add small padding
			maxV += 0.1
			minV -= 0.1
		}
		// chart dimensions
		width := 42
		height := 10
		lc := timeserieslinechart.New(width, height)
		lc.SetTimeRange(minTime, maxTime)
		lc.SetViewTimeAndYRange(minTime, maxTime, minV, maxV)
		// Attempt hourly ticks: compute total hours and set X step so labels align roughly each hour.
		hours := int(maxTime.Sub(minTime).Hours())
		if hours <= 0 {
			hours = 1
		}
		// we want about one label per hour; ntcharts linechart draws labels every XStep columns.
		// So derive xStep as max(1, graphWidth / hours)
		xStep := 1
		if hours < lc.GraphWidth() {
			// ensure at most ~hours labels
			xStep = lc.GraphWidth() / hours
			if xStep < 1 {
				xStep = 1
			}
		}
		lc.SetXStep(xStep)
		// set formatter to show HH:MM
		lc.Model.XLabelFormatter = func(i int, v float64) string {
			return time.Unix(int64(v), 0).In(time.Local).Format("15:04")
		}
		for i, tm := range parsedTimes {
			if tm.IsZero() {
				continue
			}
			lc.Push(timeserieslinechart.TimePoint{Time: tm, Value: values[i]})
		}
		lc.DrawBraille()
		// mark current time with a vertical line if within range (draw after chart so it overlays)
		now := time.Now() // local time
		if (now.Equal(minTime) || now.After(minTime)) && (now.Equal(maxTime) || now.Before(maxTime)) {
			viewMin := lc.Model.ViewMinX()
			viewMax := lc.Model.ViewMaxX()
			if viewMax > viewMin {
				dx := viewMax - viewMin
				xRel := (float64(now.Unix()) - viewMin) / dx
				if xRel < 0 {
					xRel = 0
				} else if xRel > 1 {
					xRel = 1
				}
				col := int(math.Round(xRel * float64(lc.GraphWidth()-1))) // scale to graph width minus 1 (0-indexed)
				// convert to canvas column
				col += lc.Model.Origin().X
				if lc.Model.YStep() > 0 {
					col += 1
				}
				if col >= 0 && col < lc.Canvas.Width() {
					lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("226"))
					for y := 0; y < lc.Model.Origin().Y; y++ {
						p := canvas.Point{X: col, Y: y}
						// Avoid overwriting Y axis line (if any)
						cell := lc.Canvas.Cell(p)
						if cell.Rune == '│' && cell.Style.GetForeground() != (lipgloss.Style{}).GetForeground() {
							// already an axis; brighten it
							lc.Canvas.SetCell(p, canvas.NewCellWithStyle('│', lineStyle))
						} else {
							lc.Canvas.SetCell(p, canvas.NewCellWithStyle('│', lineStyle))
						}
					}
				}
			}
		}
		b.WriteString("Tide (ft) timeseries:\n")
		b.WriteString(lc.View())
		// legend & stats
		b.WriteString("\n")
		legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("45"))
		b.WriteString(legendStyle.Render("─"))
		b.WriteString(" ")
		b.WriteString(buoyInfoStyle.Render("Predicted tide"))
		b.WriteString("\n")
		// current time legend (only if inside range)
		if now := time.Now(); (now.Equal(minTime) || now.After(minTime)) && (now.Equal(maxTime) || now.Before(maxTime)) {
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("226")).Render("│"))
			b.WriteString(" ")
			b.WriteString(buoyInfoStyle.Render("Current time"))
			b.WriteString("\n")
		}
		// Display local timezone abbreviation
		tzName, _ := minTime.Zone()
		b.WriteString(buoyInfoStyle.Render(fmt.Sprintf("min %.2f ft / max %.2f ft | %s - %s %s", minV, maxV, minTime.Format("15:04"), maxTime.Format("15:04"), tzName)))
	} else if data.tide != nil {
		b.WriteString(buoyInfoStyle.Render("Insufficient tide points"))
	} else {
		b.WriteString(buoyInfoStyle.Render("No tide data"))
	}
	return b.String()
}
