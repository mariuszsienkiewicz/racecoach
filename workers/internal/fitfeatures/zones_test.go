package fitfeatures

import (
	"strings"
	"testing"

	"github.com/racecoach/workers/internal/domain"
)

func TestFormatAthleteZonesIsFactsOnly(t *testing.T) {
	t.Parallel()
	athlete := &domain.AthleteContext{
		Zones: []domain.HrZone{
			{Zone: 1, MinBpm: 127, MaxBpm: 143},
			{Zone: 2, MinBpm: 143, MaxBpm: 159},
		},
	}
	got := formatAthleteZones(athlete)
	if got != "Z1 127–143 bpm; Z2 143–159 bpm" {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(strings.ToLower(got), "judge") {
		t.Fatalf("display must not contain coaching instructions: %q", got)
	}
	if formatAthleteZones(nil) != "" {
		t.Fatalf("nil athlete should yield empty display")
	}
}

func TestZoneForHR(t *testing.T) {
	t.Parallel()
	athlete := &domain.AthleteContext{
		Zones: []domain.HrZone{
			{Zone: 1, MinBpm: 90, MaxBpm: 120},
			{Zone: 2, MinBpm: 120, MaxBpm: 164},
			{Zone: 3, MinBpm: 164, MaxBpm: 175},
			{Zone: 4, MinBpm: 175, MaxBpm: 185},
			{Zone: 5, MinBpm: 185, MaxBpm: 195},
		},
	}
	if got := zoneForHR(160, athlete); got == nil || *got != 2 {
		t.Fatalf("160 bpm should be Z2, got %v", got)
	}
	if got := zoneForHR(190, athlete); got == nil || *got != 5 {
		t.Fatalf("190 bpm should be Z5, got %v", got)
	}
}
