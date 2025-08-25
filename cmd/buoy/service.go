package buoy

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"
)

type Service interface {
	GetTideData() (tideData, error)
}

var _ Service = (*dataService)(nil)

func NewService() Service {
	return &dataService{}
}

// GetTideData retrieves today's tide prediction data for a fixed station.
// Currently hard-coded to station 9410170 (San Francisco, CA) and returns
// times in GMT as provided by the API.
func (s *dataService) GetTideData() (tideData, error) {
	const stationID = "9410170"
	const url = "https://api.tidesandcurrents.noaa.gov/api/prod/datagetter?date=today&station=" + stationID + "&product=predictions&datum=MLLW&time_zone=gmt&units=english&format=json"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return tideData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return tideData{}, errors.New("unexpected status code: " + resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return tideData{}, err
	}

	// Struct matching NOAA response
	var parsed struct {
		Predictions []struct {
			T string `json:"t"`
			V string `json:"v"`
		} `json:"predictions"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return tideData{}, err
	}

	td := tideData{stationId: stationID, points: make([]struct {
		time  string
		value float64
	}, len(parsed.Predictions))}

	for i, p := range parsed.Predictions {
		v, err := strconv.ParseFloat(p.V, 64)
		if err != nil {
			return tideData{}, err
		}
		td.points[i] = struct {
			time  string
			value float64
		}{time: p.T, value: v}
	}

	return td, nil
}

type dataService struct{}
