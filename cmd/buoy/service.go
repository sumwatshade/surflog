package buoy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Service interface {
	GetTideData() (TideData, error)
	// GetWaveSummary retrieves the latest detailed wave summary (.spec) entry
	// for a fixed buoy station and distills it into structured data. Currently
	// hard-coded to station 46274 (San Francisco Bar / SF approach) and returns
	// the most recent observation (first non-comment line in the .spec file).
	GetWaveSummary() (WaveSummary, error)
}

var _ Service = (*dataService)(nil)

func NewService() Service {
	return &dataService{}
}

// WaveSummary provides a distilled view of a single line from the NOAA
// detailed wave summary (.spec) file.
// Field descriptions (see https://www.ndbc.noaa.gov/faq/measdes.shtml):
//
//	WVHT: Significant Wave Height (m)
//	SwH / SwP / SwD: Primary Swell Height (m), Period (s), Direction (text)
//	WWH / WWP / WWD: Wind Wave Height (m), Period (s), Direction (text)
//	STEEPNESS: Wave steepness category
//	APD: Average Wave Period (s)
//	MWD: Mean Wave Direction (deg true)
type WaveSummary struct {
	stationId            string
	time                 time.Time
	wvht                 float64
	swellHeight          float64
	swellPeriod          float64
	windWaveHeight       float64
	windWavePeriod       float64
	swellDirection       string
	windWaveDirection    string
	steepness            string
	averagePeriod        float64
	meanWaveDirectionDeg int
}

// waveSummaryDTO is the exported representation used for JSON persistence.
type waveSummaryDTO struct {
	StationID         string    `json:"station_id"`
	Time              time.Time `json:"time"`
	SignificantHeight float64   `json:"significant_height_m"`
	SwellHeight       float64   `json:"swell_height_m"`
	SwellPeriod       float64   `json:"swell_period_s"`
	WindWaveHeight    float64   `json:"wind_wave_height_m"`
	WindWavePeriod    float64   `json:"wind_wave_period_s"`
	SwellDirection    string    `json:"swell_direction"`
	WindWaveDirection string    `json:"wind_wave_direction"`
	Steepness         string    `json:"steepness"`
	AveragePeriod     float64   `json:"average_period_s"`
	MeanWaveDirection int       `json:"mean_wave_direction_deg"`
	Summary           string    `json:"summary"` // human readable string (optional convenience)
}

// MarshalJSON implements custom JSON encoding while keeping internal fields unexported.
func (w WaveSummary) MarshalJSON() ([]byte, error) {
	dto := waveSummaryDTO{
		StationID:         w.stationId,
		Time:              w.time,
		SignificantHeight: w.wvht,
		SwellHeight:       w.swellHeight,
		SwellPeriod:       w.swellPeriod,
		WindWaveHeight:    w.windWaveHeight,
		WindWavePeriod:    w.windWavePeriod,
		SwellDirection:    w.swellDirection,
		WindWaveDirection: w.windWaveDirection,
		Steepness:         w.steepness,
		AveragePeriod:     w.averagePeriod,
		MeanWaveDirection: w.meanWaveDirectionDeg,
		Summary:           w.String(),
	}
	return json.Marshal(dto)
}

// UnmarshalJSON decodes persisted wave summary data back into the internal struct.
func (w *WaveSummary) UnmarshalJSON(b []byte) error {
	// Accept empty or null gracefully.
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	var dto waveSummaryDTO
	if err := json.Unmarshal(b, &dto); err != nil {
		return err
	}
	// Populate internal fields.
	w.stationId = dto.StationID
	w.time = dto.Time
	w.wvht = dto.SignificantHeight
	w.swellHeight = dto.SwellHeight
	w.swellPeriod = dto.SwellPeriod
	w.windWaveHeight = dto.WindWaveHeight
	w.windWavePeriod = dto.WindWavePeriod
	w.swellDirection = dto.SwellDirection
	w.windWaveDirection = dto.WindWaveDirection
	w.steepness = dto.Steepness
	w.averagePeriod = dto.AveragePeriod
	w.meanWaveDirectionDeg = dto.MeanWaveDirection
	return nil
}

func (w *WaveSummary) String() string {
	return fmt.Sprintf("%.1fft sig (swell %.1fft @ %.0fs %s / wind %.1fft @ %.0fs %s) | avg %.1fs | mean %d°",
		w.wvht, w.swellHeight, w.swellPeriod, w.swellDirection, w.windWaveHeight, w.windWavePeriod, w.windWaveDirection, w.averagePeriod, w.meanWaveDirectionDeg)
}

// GetTideData retrieves today's tide prediction data for a fixed station.
// Currently hard-coded to station 9410170 (San Francisco, CA) and returns
// times in GMT as provided by the API.
func (s *dataService) GetTideData() (TideData, error) {
	const stationID = "9410170"
	const url = "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?date=today&station=" + stationID + "&product=predictions&datum=MLLW&time_zone=gmt&units=english&format=json"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return TideData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TideData{}, errors.New("unexpected status code: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return TideData{}, err
	}

	// Struct matching NOAA response
	var parsed struct {
		Predictions []struct {
			T string `json:"t"`
			V string `json:"v"`
		} `json:"predictions"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return TideData{}, err
	}

	td := TideData{stationId: stationID, points: make([]struct {
		time  string
		value float64
	}, len(parsed.Predictions))}

	for i, p := range parsed.Predictions {
		v, err := strconv.ParseFloat(p.V, 64)
		if err != nil {
			return TideData{}, err
		}
		td.points[i] = struct {
			time  string
			value float64
		}{time: p.T, value: v}
	}

	return td, nil
}

// GetWaveSummary fetches the latest detailed wave summary (.spec) file for a
// fixed buoy station and returns the most recent observation parsed into a
// WaveSummary struct.
func (s *dataService) GetWaveSummary() (WaveSummary, error) {
	const stationID = "46274"
	const url = "https://www.ndbc.noaa.gov/data/realtime2/" + stationID + ".spec"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return WaveSummary{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return WaveSummary{}, errors.New("unexpected status code: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return WaveSummary{}, err
	}

	lines := splitLines(string(body))
	// collect up to 5 most recent data lines
	var dataLines []string
	for _, line := range lines {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		dataLines = append(dataLines, line)
		if len(dataLines) == 5 { // we only need first 5 (already newest first in file)
			break
		}
	}
	if len(dataLines) == 0 {
		return WaveSummary{}, errors.New("no data lines in spec file")
	}

	type parsed struct {
		ts       time.Time
		wvht     float64
		swellH   float64
		swellP   float64
		windH    float64
		windP    float64
		swellDir string
		windDir  string
		steep    string
		apd      float64
		mwd      int
	}

	var parsedRows []parsed
	for _, ln := range dataLines {
		fields := fieldsCondense(ln)
		if len(fields) < 15 {
			continue // skip malformed
		}
		// Parse timestamp
		year, err1 := strconv.Atoi(fields[0])
		mon, err2 := strconv.Atoi(fields[1])
		day, err3 := strconv.Atoi(fields[2])
		hour, err4 := strconv.Atoi(fields[3])
		minute, err5 := strconv.Atoi(fields[4])
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil || err5 != nil {
			continue
		}
		ts := time.Date(year, time.Month(mon), day, hour, minute, 0, 0, time.UTC)
		// helper parse float with graceful skip
		parseF := func(v string) (float64, bool) {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return 0, false
			}
			return f, true
		}
		wvht, ok1 := parseF(fields[5])
		swellH, ok2 := parseF(fields[6])
		swellP, ok3 := parseF(fields[7])
		windH, ok4 := parseF(fields[8])
		windP, ok5 := parseF(fields[9])
		apd, ok6 := parseF(fields[13])
		mwd, err := strconv.Atoi(fields[14])
		if err != nil { // skip direction if invalid
			mwd = 0
		}
		if !(ok1 && ok2 && ok3 && ok4 && ok5 && ok6) {
			// If any numeric field failed parsing, skip this row for averaging to avoid bias.
			continue
		}
		parsedRows = append(parsedRows, parsed{
			ts:       ts,
			wvht:     wvht,
			swellH:   swellH,
			swellP:   swellP,
			windH:    windH,
			windP:    windP,
			swellDir: fields[10],
			windDir:  fields[11],
			steep:    fields[12],
			apd:      apd,
			mwd:      mwd,
		})
	}
	if len(parsedRows) == 0 {
		return WaveSummary{}, errors.New("no parsable data rows")
	}

	// Average numeric fields
	var sumWvht, sumSwellH, sumSwellP, sumWindH, sumWindP, sumApd float64
	var sumMwd float64
	for _, r := range parsedRows {
		sumWvht += r.wvht
		sumSwellH += r.swellH
		sumSwellP += r.swellP
		sumWindH += r.windH
		sumWindP += r.windP
		sumApd += r.apd
		sumMwd += float64(r.mwd)
	}
	n := float64(len(parsedRows))
	latest := parsedRows[0] // first row is most recent

	return WaveSummary{
		stationId:            stationID,
		time:                 latest.ts,
		wvht:                 sumWvht / n,
		swellHeight:          sumSwellH / n,
		swellPeriod:          sumSwellP / n,
		windWaveHeight:       sumWindH / n,
		windWavePeriod:       sumWindP / n,
		swellDirection:       latest.swellDir,
		windWaveDirection:    latest.windDir,
		steepness:            latest.steep,
		averagePeriod:        sumApd / n,
		meanWaveDirectionDeg: int(sumMwd/n + 0.5), // simple rounded average
	}, nil
}

// splitLines splits on both \r and \n while keeping things simple.
func splitLines(s string) []string {
	var out []string
	start := 0
	for i, ch := range s {
		if ch == '\n' { // line end
			line := s[start:i]
			// trim trailing CR
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			out = append(out, line)
			start = i + 1
		}
	}
	if start < len(s) { // last line
		line := s[start:]
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		out = append(out, line)
	}
	return out
}

// fieldsCondense splits a line on any run of whitespace.
func fieldsCondense(line string) []string {
	var f []string
	fieldStart := -1
	for i, ch := range line {
		if ch == ' ' || ch == '\t' || ch == '\r' {
			if fieldStart >= 0 {
				f = append(f, line[fieldStart:i])
				fieldStart = -1
			}
		} else {
			if fieldStart < 0 {
				fieldStart = i
			}
		}
	}
	if fieldStart >= 0 {
		f = append(f, line[fieldStart:])
	}
	return f
}

type dataService struct{}
