package fit

import (
	"bytes"
	"fmt"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/profile/filedef"
)

// Metrics is the result of parsing a FIT session (internal).
type Metrics struct {
	DistanceM    int
	DurationSec  int
	AvgHeartRate *int
	MaxHeartRate *int
}

func ParseMetrics(data []byte) (Metrics, error) {
	lis := filedef.NewListener()
	defer lis.Close()

	dec := decoder.New(bytes.NewReader(data),
		decoder.WithMesgListener(lis),
		decoder.WithBroadcastOnly(),
	)
	if _, err := dec.Decode(); err != nil {
		return Metrics{}, fmt.Errorf("decode: %w", err)
	}

	activity := lis.File().(*filedef.Activity)
	session := activity.Sessions[0]

	return Metrics{
		DistanceM:    int(session.TotalDistanceScaled()),
		DurationSec:  int(session.TotalTimerTimeScaled()),
		AvgHeartRate: optionalUint8(session.AvgHeartRate),
		MaxHeartRate: optionalUint8(session.MaxHeartRate),
	}, nil
}
