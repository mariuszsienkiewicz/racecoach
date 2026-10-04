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

	if shape == "intervals" {
		return "intervals"
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
