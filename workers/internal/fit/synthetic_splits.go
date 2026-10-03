package fit

import (
	"math"
	"time"

	"github.com/muktihari/fit/kit/datetime"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/racecoach/workers/internal/domain"
)

const (
	syntheticKmTrigger = "synthetic_km"
	kmSplitDistanceM   = 1000.0
	minPartialSplitM   = 50.0
)

// needsSyntheticKmSplits is true when device laps cannot describe activity pace.
// Typical case: auto-lap off -> one session-end lap covering the whole run.
func needsSyntheticKmSplits(laps []domain.Lap) bool {
	return len(laps) <= 1
}

type distSample struct {
	t    time.Time
	dist float64
	hr   uint8
	okHR bool
}

func buildSyntheticKmLaps(records []*mesgdef.Record) []domain.Lap {
	return syntheticKmLapsFromSamples(recordsToSamples(records))
}

func recordsToSamples(records []*mesgdef.Record) []distSample {
	out := make([]distSample, 0, len(records))
	for _, r := range records {
		if r == nil || r.Distance == basetype.Uint32Invalid {
			continue
		}
		if r.Timestamp.Before(datetime.Epoch()) {
			continue
		}
		d := r.DistanceScaled()
		if math.IsNaN(d) || math.IsInf(d, 0) || d < 0 {
			continue
		}
		if len(out) > 0 && d < out[len(out)-1].dist {
			continue // drop GPS noise
		}
		s := distSample{t: r.Timestamp.UTC(), dist: d}
		if r.HeartRate != basetype.Uint8Invalid {
			s.hr = r.HeartRate
			s.okHR = true
		}
		out = append(out, s)
	}
	return out
}

func syntheticKmLapsFromSamples(samples []distSample) []domain.Lap {
	if len(samples) < 2 {
		return nil
	}

	origin := samples[0].dist
	nextBound := origin + kmSplitDistanceM
	startIdx := 0
	laps := make([]domain.Lap, 0, int((samples[len(samples)-1].dist-origin)/kmSplitDistanceM)+1)

	for i := 1; i < len(samples); i++ {
		for samples[i].dist >= nextBound {
			if i > startIdx {
				laps = append(laps, lapFromSamples(len(laps), samples[startIdx:i+1]))
				startIdx = i
			}
			nextBound += kmSplitDistanceM
		}
	}

	if startIdx < len(samples)-1 {
		leftover := samples[len(samples)-1].dist - samples[startIdx].dist
		if leftover >= minPartialSplitM {
			laps = append(laps, lapFromSamples(len(laps), samples[startIdx:]))
		}
	}

	return laps
}

func lapFromSamples(index int, samples []distSample) domain.Lap {
	start := samples[0]
	end := samples[len(samples)-1]
	distM := int(math.Round(math.Max(0, end.dist-start.dist)))
	durSec := int(math.Round(end.t.Sub(start.t).Seconds()))
	if durSec < 0 {
		durSec = 0
	}

	item := domain.Lap{
		Index:          index,
		StartTimestamp: start.t.Format(time.RFC3339),
		EndTimestamp:   end.t.Format(time.RFC3339),
		DistanceM:      distM,
		DurationSec:    durSec,
		LapTrigger:     syntheticKmTrigger,
	}
	if distM > 0 && durSec > 0 {
		pace := int(math.Round(float64(durSec) / (float64(distM) / 1000)))
		item.AvgPaceSecPerKm = &pace
	}

	var hrSum int
	var hrN int
	var maxHR, minHR int
	haveHR := false
	for _, s := range samples {
		if !s.okHR {
			continue
		}
		hr := int(s.hr)
		hrSum += hr
		hrN++
		if !haveHR {
			maxHR, minHR = hr, hr
			haveHR = true
			continue
		}
		if hr > maxHR {
			maxHR = hr
		}
		if hr < minHR {
			minHR = hr
		}
	}
	if hrN > 0 {
		avg := int(math.Round(float64(hrSum) / float64(hrN)))
		item.AvgHeartRate = &avg
		item.MaxHeartRate = &maxHR
		item.MinHeartRate = &minHR
	}
	return item
}
