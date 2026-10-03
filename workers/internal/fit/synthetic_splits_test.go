package fit

import (
	"testing"
	"time"

	"github.com/racecoach/workers/internal/domain"
)

func TestNeedsSyntheticKmSplits(t *testing.T) {
	t.Parallel()

	if !needsSyntheticKmSplits(nil) {
		t.Fatal("empty should need synthetics")
	}
	if !needsSyntheticKmSplits([]domain.Lap{{Index: 0, DistanceM: 5070}}) {
		t.Fatal("single lap should need synthetics")
	}
	if needsSyntheticKmSplits([]domain.Lap{{Index: 0}, {Index: 1}}) {
		t.Fatal("multi lap should keep device laps")
	}
}

func TestSyntheticKmLapsFromSamplesFadingRace(t *testing.T) {
	t.Parallel()

	start := time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC)
	samples := make([]distSample, 0, 520)
	kmDurations := []int{340, 310, 295, 280, 260}
	elapsed := 0
	samples = append(samples, distSample{t: start, dist: 0, hr: 150, okHR: true})
	for km, dur := range kmDurations {
		for step := 1; step <= 20; step++ {
			elapsed += dur / 20
			dist := float64(km)*1000 + float64(step)*50
			samples = append(samples, distSample{
				t:    start.Add(time.Duration(elapsed) * time.Second),
				dist: dist,
				hr:   uint8(150 + km*5),
				okHR: true,
			})
		}
	}
	// partial 70 m
	elapsed += 18
	samples = append(samples, distSample{
		t:    start.Add(time.Duration(elapsed) * time.Second),
		dist: 5070,
		hr:   175,
		okHR: true,
	})

	laps := syntheticKmLapsFromSamples(samples)
	if len(laps) != 6 {
		t.Fatalf("laps=%d want 6: %+v", len(laps), laps)
	}
	for i, lap := range laps {
		if lap.LapTrigger != syntheticKmTrigger {
			t.Fatalf("lap[%d] trigger=%q", i, lap.LapTrigger)
		}
		if lap.Index != i {
			t.Fatalf("lap[%d] index=%d", i, lap.Index)
		}
		if lap.AvgPaceSecPerKm == nil {
			t.Fatalf("lap[%d] missing pace", i)
		}
	}
	if *laps[0].AvgPaceSecPerKm < 320 {
		t.Fatalf("first km too fast: %d", *laps[0].AvgPaceSecPerKm)
	}
	if *laps[4].AvgPaceSecPerKm > 280 {
		t.Fatalf("5th km too slow: %d", *laps[4].AvgPaceSecPerKm)
	}
	if *laps[0].AvgPaceSecPerKm <= *laps[4].AvgPaceSecPerKm {
		t.Fatalf("expected fading race: first=%d fifth=%d", *laps[0].AvgPaceSecPerKm, *laps[4].AvgPaceSecPerKm)
	}
	if laps[5].DistanceM < 50 || laps[5].DistanceM > 100 {
		t.Fatalf("partial distance=%d", laps[5].DistanceM)
	}
}

func TestSyntheticKmLapsTooShort(t *testing.T) {
	t.Parallel()

	start := time.Date(2024, 5, 1, 10, 0, 0, 0, time.UTC)
	samples := []distSample{
		{t: start, dist: 0},
		{t: start.Add(3 * time.Minute), dist: 800},
	}
	laps := syntheticKmLapsFromSamples(samples)
	if len(laps) != 1 {
		t.Fatalf("laps=%d want 1 partial-only", len(laps))
	}
	// ParseStructure only replaces when len(synth) >= 2
	if len(laps) >= 2 {
		t.Fatal("should not replace device lap for sub-1km activity")
	}
}
