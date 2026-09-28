package fit

import (
	"bytes"
	"fmt"
	"math"
	"time"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/kit/datetime"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
	"github.com/racecoach/workers/internal/domain"
)

// Structure is the parsed workout breakdown (internal).
type Structure struct {
	Laps []domain.Lap `json:"laps"`
}

func ParseStructure(data []byte) (Structure, error) {
	lis := filedef.NewListener()
	defer lis.Close()

	dec := decoder.New(bytes.NewReader(data),
		decoder.WithMesgListener(lis),
		decoder.WithBroadcastOnly(),
	)
	if _, err := dec.Decode(); err != nil {
		return Structure{}, fmt.Errorf("decode: %w", err)
	}

	activity, ok := lis.File().(*filedef.Activity)
	if !ok || activity == nil {
		return Structure{}, fmt.Errorf("not an activity file")
	}

	out := Structure{Laps: make([]domain.Lap, 0, len(activity.Laps))}
	for i, raw := range activity.Laps {
		if raw == nil {
			continue
		}
		out.Laps = append(out.Laps, mapLap(i, raw))
	}

	// Zero laps is valid (easy runs / activities without auto-lap)
	return out, nil
}

func mapLap(index int, lap *mesgdef.Lap) domain.Lap {
	item := domain.Lap{Index: index}

	if lap.LapTrigger != typedef.LapTriggerInvalid {
		item.LapTrigger = lap.LapTrigger.String()
	}
	if lap.Intensity != typedef.IntensityInvalid {
		item.Intensity = lap.Intensity.String()
	}

	if !lap.StartTime.Before(datetime.Epoch()) {
		item.StartTimestamp = lap.StartTime.UTC().Format(time.RFC3339)
	}

	if !lap.Timestamp.Before(datetime.Epoch()) {
		item.EndTimestamp = lap.Timestamp.UTC().Format(time.RFC3339)
	}

	if dist, ok := scaledOrInt(lap.TotalDistanceScaled(), lap.TotalDistance != basetype.Uint32Invalid); ok {
		item.DistanceM = dist
	}
	if dur, ok := scaledOrInt(lap.TotalTimerTimeScaled(), lap.TotalTimerTime != basetype.Uint32Invalid); ok {
		item.DurationSec = dur
	}
	if item.DistanceM > 0 && item.DurationSec > 0 {
		pace := int(math.Round(float64(item.DurationSec) / (float64(item.DistanceM) / 1000)))
		item.AvgPaceSecPerKm = &pace
	}

	item.AvgHeartRate = optionalUint8(lap.AvgHeartRate)
	item.MaxHeartRate = optionalUint8(lap.MaxHeartRate)
	item.MinHeartRate = optionalUint8(lap.MinHeartRate)
	item.AvgCadence = optionalUint8(lap.AvgCadence)
	item.TotalAscentM = optionalUint16(lap.TotalAscent)
	item.TotalDescentM = optionalUint16(lap.TotalDescent)
	item.TotalCalories = optionalUint16(lap.TotalCalories)

	return item
}

func scaledOrInt(scaled float64, rawOK bool) (int, bool) {
	if !rawOK || math.Float64bits(scaled) == basetype.Float64Invalid {
		return 0, false
	}
	if math.IsNaN(scaled) || math.IsInf(scaled, 0) || scaled < 0 {
		return 0, false
	}
	return int(math.Round(scaled)), true
}

func optionalUint8(v uint8) *int {
	if v == basetype.Uint8Invalid {
		return nil
	}
	x := int(v)
	return &x
}

func optionalUint16(v uint16) *int {
	if v == basetype.Uint16Invalid {
		return nil
	}
	x := int(v)
	return &x
}
