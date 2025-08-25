package buoy

// BuoyData holds buoy identifier and associated tide information for the day.
// All fields are unexported to keep the public surface small until stabilized.
type BuoyData struct {
	id      string
	tide    *tideData
	tideErr error
}

type tideData struct {
	stationId string
	points    []struct {
		time  string
		value float64
	}
}

// NewBuoyData constructs a BuoyData from tide data.
func NewBuoyData(td tideData, err error) *BuoyData {
	bd := &BuoyData{tideErr: err}
	if err == nil {
		bd.id = td.stationId
		bd.tide = &td
	}
	return bd
}
