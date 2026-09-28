package coach

import (
	"testing"

	"github.com/racecoach/workers/internal/domain"
)

func TestForCoachPromptStripsSessionWideSlowestOnIntervals(t *testing.T) {
	pace := func(v int) *int { return &v }
	idx := 2
	slow := 779
	features := domain.ActivityFeatures{
		Signals: domain.FeaturesSignals{
			SuspectedWorkoutShape:  "intervals",
			SlowestLapIndex:        &idx,
			SlowestLapPaceSecPerKm: &slow,
			Display: domain.FeaturesSignalsDisplay{
				SlowestLap:     "lap 3 (12:59 /km)",
				SlowestLapPace: "12:59 /km",
			},
			ByIntensity: domain.FeaturesByIntensity{
				Active:   &domain.IntensityBucket{LapCount: 4},
				Recovery: &domain.IntensityBucket{LapCount: 3},
			},
			IntervalPattern: &domain.IntervalPattern{RepCount: 4},
		},
		Laps: []domain.FeatureLap{
			{Lap: domain.Lap{Index: 1, AvgPaceSecPerKm: pace(270)}, Role: "active", IsWork: true, Display: domain.FeatureLapDisplay{AvgPace: "4:30 /km"}},
			{Lap: domain.Lap{Index: 2, AvgPaceSecPerKm: pace(779)}, Role: "recovery", IsWork: false, Display: domain.FeatureLapDisplay{AvgPace: "12:59 /km"}},
		},
	}

	out := ForCoachPrompt(features)
	if out.Signals.SlowestLapIndex != nil || out.Signals.Display.SlowestLap != "" {
		t.Fatalf("expected session-wide slowest stripped: %+v", out.Signals)
	}
	if out.Laps[1].Display.AvgPace != "" || out.Laps[1].AvgPaceSecPerKm != nil {
		t.Fatalf("expected recovery pace stripped: %+v", out.Laps[1])
	}
	if out.Laps[0].Display.AvgPace != "4:30 /km" {
		t.Fatalf("active pace should remain: %+v", out.Laps[0])
	}
}
