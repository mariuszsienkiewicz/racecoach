package fitfeatures

import (
	"strings"
	"testing"

	"github.com/racecoach/workers/internal/domain"
)

func TestBuildEnrichesLapsAndSignals(t *testing.T) {
	pace := func(v int) *int { return &v }
	hr := func(v int) *int { return &v }

	metrics := domain.ActivityMetrics{
		ActivityID:       1,
		DistanceM:        10000,
		DurationSec:      3000, // 5:00/km
		AvgHeartRate:     hr(150),
		MetricsObjectKey: "x.metrics.json",
	}
	structure := domain.ActivityStructure{
		ActivityID: 1,
		ObjectKey:  "x.fit",
		Laps: []domain.Lap{
			{Index: 0, DistanceM: 5000, DurationSec: 1600, AvgPaceSecPerKm: pace(320), AvgHeartRate: hr(145)},
			{Index: 1, DistanceM: 5000, DurationSec: 1400, AvgPaceSecPerKm: pace(280), AvgHeartRate: hr(155)},
		},
	}

	features := Build(metrics, structure)
	if features.Overview.LapCount != 2 {
		t.Fatalf("lapCount=%d", features.Overview.LapCount)
	}
	if features.Overview.AvgPaceSecPerKm == nil || *features.Overview.AvgPaceSecPerKm != 300 {
		t.Fatalf("avgPace=%v", features.Overview.AvgPaceSecPerKm)
	}
	if features.Signals.SplitBias != "negative" {
		t.Fatalf("splitBias=%s", features.Signals.SplitBias)
	}
	if len(features.Laps) != 2 || features.Laps[0].PaceDeltaPctVsOverall == nil {
		t.Fatalf("expected enriched laps")
	}
	if features.Overview.Display.AvgPace != "5:00 /km" {
		t.Fatalf("overview display avgPace=%q", features.Overview.Display.AvgPace)
	}
	if features.Laps[1].Display.AvgPace != "4:40 /km" {
		t.Fatalf("lap display avgPace=%q", features.Laps[1].Display.AvgPace)
	}
	if features.Signals.Display.FastestLapPace != "4:40 /km" {
		t.Fatalf("signals display fastest=%q", features.Signals.Display.FastestLapPace)
	}
	if features.SchemaVersion != 4 {
		t.Fatalf("schemaVersion=%d", features.SchemaVersion)
	}
}

func TestBuildIntervalWorkoutActiveAggregates(t *testing.T) {
	pace := func(v int) *int { return &v }
	hr := func(v int) *int { return &v }

	metrics := domain.ActivityMetrics{
		ActivityID:       78,
		DistanceM:        4259,
		DurationSec:      1588,
		AvgHeartRate:     hr(153),
		MetricsObjectKey: "x.metrics.json",
	}
	structure := domain.ActivityStructure{
		ActivityID: 78,
		ObjectKey:  "x.fit",
		Laps: []domain.Lap{
			{Index: 0, DistanceM: 1717, DurationSec: 600, AvgPaceSecPerKm: pace(349), AvgHeartRate: hr(146), Intensity: "warmup"},
			{Index: 1, DistanceM: 300, DurationSec: 81, AvgPaceSecPerKm: pace(270), AvgHeartRate: hr(178), Intensity: "active"},
			{Index: 2, DistanceM: 154, DurationSec: 120, AvgPaceSecPerKm: pace(779), AvgHeartRate: hr(138), Intensity: "recovery"},
			{Index: 3, DistanceM: 300, DurationSec: 79, AvgPaceSecPerKm: pace(263), AvgHeartRate: hr(162), Intensity: "active"},
			{Index: 4, DistanceM: 162, DurationSec: 119, AvgPaceSecPerKm: pace(735), AvgHeartRate: hr(141), Intensity: "recovery"},
			{Index: 5, DistanceM: 300, DurationSec: 81, AvgPaceSecPerKm: pace(270), AvgHeartRate: hr(162), Intensity: "active"},
			{Index: 6, DistanceM: 156, DurationSec: 114, AvgPaceSecPerKm: pace(731), AvgHeartRate: hr(143), Intensity: "recovery"},
			{Index: 7, DistanceM: 300, DurationSec: 82, AvgPaceSecPerKm: pace(273), AvgHeartRate: hr(162), Intensity: "active"},
			{Index: 8, DistanceM: 869, DurationSec: 312, AvgPaceSecPerKm: pace(359), AvgHeartRate: hr(165), Intensity: "cooldown"},
		},
	}

	features := Build(metrics, structure)
	if features.Signals.SuspectedWorkoutShape != "intervals" {
		t.Fatalf("shape=%s", features.Signals.SuspectedWorkoutShape)
	}
	if features.Signals.ByIntensity.Active == nil || features.Signals.ByIntensity.Active.LapCount != 4 {
		t.Fatalf("active=%v", features.Signals.ByIntensity.Active)
	}
	if features.Signals.IntervalPattern == nil || features.Signals.IntervalPattern.RepCount != 4 {
		t.Fatalf("pattern=%v", features.Signals.IntervalPattern)
	}
	if features.Signals.IntervalPattern.RepDistanceM == nil || *features.Signals.IntervalPattern.RepDistanceM != 300 {
		t.Fatalf("repDistance=%v", features.Signals.IntervalPattern.RepDistanceM)
	}
	if features.Signals.Display.IntervalPattern == "" {
		t.Fatal("expected interval pattern display")
	}
	if !features.Laps[1].IsWork || features.Laps[2].IsWork {
		t.Fatalf("isWork flags wrong: active=%v recovery=%v", features.Laps[1].IsWork, features.Laps[2].IsWork)
	}
	// Overall split should be suppressed for interval sessions.
	if features.Signals.SplitBias != "unknown" {
		t.Fatalf("splitBias should be unknown for intervals, got %s", features.Signals.SplitBias)
	}
	if features.Signals.ByIntensity.Active.AvgPaceSecPerKm == nil {
		t.Fatal("expected active avg pace")
	}
	// 81+79+81+82=323 sec over 1200 m => 269 sec/km
	if *features.Signals.ByIntensity.Active.AvgPaceSecPerKm != 269 {
		t.Fatalf("active avg pace=%d", *features.Signals.ByIntensity.Active.AvgPaceSecPerKm)
	}
	if features.Signals.Effort == nil || features.Signals.Effort.Label != "hard" {
		t.Fatalf("effort=%v", features.Signals.Effort)
	}
}

func TestBuildNearMaxEffortFromMaxHR(t *testing.T) {
	hr := func(v int) *int { return &v }
	metrics := domain.ActivityMetrics{
		ActivityID:   84,
		DistanceM:    5070,
		DurationSec:  1458,
		AvgHeartRate: hr(170),
		MaxHeartRate: hr(205),
	}
	structure := domain.ActivityStructure{
		ActivityID: 84,
		ObjectKey:  "x.fit",
		Laps: []domain.Lap{
			{Index: 0, DistanceM: 5070, DurationSec: 1458, AvgPaceSecPerKm: hr(172), AvgHeartRate: hr(170), MaxHeartRate: hr(205)},
		},
	}
	features := Build(metrics, structure)
	if features.Overview.MaxHeartRate == nil || *features.Overview.MaxHeartRate != 205 {
		t.Fatalf("maxHR=%v", features.Overview.MaxHeartRate)
	}
	if features.Signals.Effort == nil || features.Signals.Effort.Label != "near_max" {
		t.Fatalf("effort=%v", features.Signals.Effort)
	}
	if !strings.Contains(features.Signals.Display.Effort, "near-max") {
		t.Fatalf("effort display=%q", features.Signals.Display.Effort)
	}
}
