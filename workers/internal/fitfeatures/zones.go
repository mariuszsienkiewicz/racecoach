package fitfeatures

import (
	"fmt"
	"sort"
	"strings"

	"github.com/racecoach/workers/internal/domain"
)

func zoneForHR(hr int, athlete *domain.AthleteContext) *int {
	if athlete == nil || len(athlete.Zones) == 0 || hr <= 0 {
		return nil
	}
	zones := append([]domain.HrZone(nil), athlete.Zones...)
	sort.Slice(zones, func(i, j int) bool { return zones[i].Zone < zones[j].Zone })
	for _, z := range zones {
		if hr >= z.MinBpm && hr <= z.MaxBpm {
			zone := z.Zone
			return &zone
		}
	}
	// Below Z1 → Z1; above Z5 → Z5
	if hr < zones[0].MinBpm {
		z := zones[0].Zone
		return &z
	}
	z := zones[len(zones)-1].Zone
	return &z
}

func formatZone(zone *int) string {
	if zone == nil {
		return ""
	}
	return fmt.Sprintf("Z%d", *zone)
}

func formatAthleteZones(athlete *domain.AthleteContext) string {
	if athlete == nil || len(athlete.Zones) == 0 {
		return ""
	}
	parts := make([]string, 0, len(athlete.Zones))
	for _, z := range athlete.Zones {
		parts = append(parts, fmt.Sprintf("Z%d %d–%d bpm", z.Zone, z.MinBpm, z.MaxBpm))
	}
	return strings.Join(parts, "; ")
}

func effortFromZone(zone int) string {
	switch {
	case zone <= 2:
		return "easy"
	case zone == 3:
		return "moderate"
	case zone == 4:
		return "hard"
	default:
		return "near_max"
	}
}

func dominantZoneFromLaps(laps []domain.FeatureLap) *int {
	weights := map[int]int{}
	for _, lap := range laps {
		if lap.AvgHrZone == nil || lap.DurationSec <= 0 {
			continue
		}
		// Prefer active/work for dominant effort. Count others lightly
		w := lap.DurationSec
		if lap.IsWork || lap.Role == "active" {
			w *= 3
		}
		weights[*lap.AvgHrZone] += w
	}
	if len(weights) == 0 {
		return nil
	}
	bestZone, bestW := 0, -1
	for z, w := range weights {
		if w > bestW || (w == bestW && z > bestZone) {
			bestZone, bestW = z, w
		}
	}
	return &bestZone
}

func timeInZoneSec(laps []domain.FeatureLap) map[string]int {
	out := map[string]int{}
	for _, lap := range laps {
		if lap.AvgHrZone == nil || lap.DurationSec <= 0 {
			continue
		}
		key := formatZone(lap.AvgHrZone)
		out[key] += lap.DurationSec
	}
	if len(out) == 0 {
		return nil
	}
	return out
}
