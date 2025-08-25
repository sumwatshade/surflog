package buoy

// BuoyData holds buoy identifier and associated tide information for the day.
// All fields are unexported to keep the public surface small until stabilized.
type BuoyData struct {
	tide    *TideData
	tideErr error
	wave    *WaveSummary
	waveErr error
}

type TideData struct {
	stationId string
	points    []struct {
		time  string
		value float64
	}
}

// setWave populates wave summary fields (internal helper used after fetching).
func (b *BuoyData) setWave(ws WaveSummary, err error) {
	b.waveErr = err
	if err == nil {
		b.wave = &ws
	}
}

func (b *BuoyData) setTide(td TideData, err error) {
	b.tideErr = err
	if err == nil {
		b.tide = &td
	}
}
