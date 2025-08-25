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

func (w *WaveSummary) String() string {
	return fmt.Sprintf("%.1fft sig (swell %.1fft @ %.0fs %s / wind %.1fft @ %.0fs %s) | steep %s | avg %.1fs | mean %d°",
		w.wvht, w.swellHeight, w.swellPeriod, w.swellDirection, w.windWaveHeight, w.windWavePeriod, w.windWaveDirection, w.steepness, w.averagePeriod, w.meanWaveDirectionDeg)
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

	// Scan lines, find first non-comment, non-empty line (latest reading)
	var latestLine string
	for _, line := range splitLines(string(body)) {
		if len(line) == 0 || line[0] == '#' {
			continue
		}
		latestLine = line
		break
	}
	if latestLine == "" {
		return WaveSummary{}, errors.New("no data lines in spec file")
	}

	fields := fieldsCondense(latestLine)
	if len(fields) < 15 { // require all expected columns
		return WaveSummary{}, errors.New("unexpected column count in spec line: " + latestLine)
	}

	// Parse date/time components
	year, err := strconv.Atoi(fields[0])
	if err != nil {
		return WaveSummary{}, err
	}
	month, err := strconv.Atoi(fields[1])
	if err != nil {
		return WaveSummary{}, err
	}
	day, err := strconv.Atoi(fields[2])
	if err != nil {
		return WaveSummary{}, err
	}
	hour, err := strconv.Atoi(fields[3])
	if err != nil {
		return WaveSummary{}, err
	}
	minute, err := strconv.Atoi(fields[4])
	if err != nil {
		return WaveSummary{}, err
	}
	ts := time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)

	parseF := func(s string) (float64, error) {
		return strconv.ParseFloat(s, 64)
	}
	wvht, err := parseF(fields[5])
	if err != nil {
		return WaveSummary{}, err
	}
	swellH, err := parseF(fields[6])
	if err != nil {
		return WaveSummary{}, err
	}
	swellP, err := parseF(fields[7])
	if err != nil {
		return WaveSummary{}, err
	}
	windH, err := parseF(fields[8])
	if err != nil {
		return WaveSummary{}, err
	}
	windP, err := parseF(fields[9])
	if err != nil {
		return WaveSummary{}, err
	}
	swellDir := fields[10]
	windDir := fields[11]
	steep := fields[12]
	apd, err := parseF(fields[13])
	if err != nil {
		return WaveSummary{}, err
	}
	mwd, err := strconv.Atoi(fields[14])
	if err != nil {
		return WaveSummary{}, err
	}

	return WaveSummary{
		stationId:            stationID,
		time:                 ts,
		wvht:                 wvht,
		swellHeight:          swellH,
		swellPeriod:          swellP,
		windWaveHeight:       windH,
		windWavePeriod:       windP,
		swellDirection:       swellDir,
		windWaveDirection:    windDir,
		steepness:            steep,
		averagePeriod:        apd,
		meanWaveDirectionDeg: mwd,
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
