package buoy

// BuoyData holds buoy identifier and associated tide information for the day.
// All fields are unexported to keep the public surface small until stabilized.
type BuoyData struct {
	tide    *tideData
	tideErr error
	wave    *waveSummary
	waveErr error
}

type tideData struct {
	stationId string
	points    []struct {
		time  string
		value float64
	}
}

// setWave populates wave summary fields (internal helper used after fetching).
func (b *BuoyData) setWave(ws waveSummary, err error) {
	b.waveErr = err
	if err == nil {
		b.wave = &ws
	}
}

func (b *BuoyData) setTide(td tideData, err error) {
	b.tideErr = err
	if err == nil {
		b.tide = &td
	}
}
