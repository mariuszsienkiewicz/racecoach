package fitfeatures

import "github.com/racecoach/workers/internal/domain"

const longRunDistanceM = 10_000

// ClassifyActivityType maps pipeline signals onto the athlete-facing activity type tag.
func ClassifyActivityType(features domain.ActivityFeatures) string {
	shape := features.Signals.SuspectedWorkoutShape
	effort := "unknown"
	if features.Signals.Effort != nil && features.Signals.Effort.Label != "" {
		effort = features.Signals.Effort.Label
	}
	distanceM := features.Overview.DistanceM
	dominant := features.Signals.DominantHrZone
	structuredIntervals := shape == "intervals" && hasStructuredWorkRest(features.Signals.ByIntensity)

	// True interval sessions (FIT work/rest roles). Aerobic dominant zone still wins the tag
	// only when we lack structured intensity, otherwise keep intervals.
	if structuredIntervals {
		return "intervals"
	}

	if features.Signals.HasAthleteZones && dominant != nil {
		switch {
		case *dominant >= 5 && effort == "near_max":
			return "race"
		case *dominant >= 4:
			return "tempo"
		case *dominant == 3 && distanceM >= longRunDistanceM:
			return "long_run"
		case *dominant == 3:
			return "tempo"
		case distanceM >= longRunDistanceM:
			return "long_run"
		default:
			return "easy"
		}
	}

	switch effort {
	case "near_max":
		return "race"
	case "hard":
		return "tempo"
	case "moderate":
		if distanceM >= longRunDistanceM {
			return "long_run"
		}
		return "tempo"
	case "easy":
		if distanceM >= longRunDistanceM {
			return "long_run"
		}
		return "easy"
	default:
		if distanceM >= longRunDistanceM {
			return "long_run"
		}
		return "easy"
	}
}
