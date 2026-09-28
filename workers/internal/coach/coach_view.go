package coach

import "github.com/racecoach/workers/internal/domain"

// forCoachPrompt returns a copy of features safe to send to the LLM.
// Session-wide "slowest lap" on interval workouts is usually recovery walking
// strip those traps so the model cannot cite them as work quality.
func ForCoachPrompt(in domain.ActivityFeatures) domain.ActivityFeatures {
	out := in
	if !shouldPreferActiveSignals(out) {
		return out
	}

	out.Signals.PaceVariabilityPct = nil
	out.Signals.FastestLapIndex = nil
	out.Signals.SlowestLapIndex = nil
	out.Signals.FastestLapPaceSecPerKm = nil
	out.Signals.SlowestLapPaceSecPerKm = nil
	out.Signals.Display.PaceVariability = ""
	out.Signals.Display.FastestLap = ""
	out.Signals.Display.SlowestLap = ""
	out.Signals.Display.FastestLapPace = ""
	out.Signals.Display.SlowestLapPace = ""

	laps := make([]domain.FeatureLap, len(out.Laps))
	copy(laps, out.Laps)
	for i := range laps {
		if laps[i].IsWork || laps[i].Role == "active" {
			continue
		}
		// Keep structure (role/distance/duration) but remove pace so recovery
		// cannot be described as "slowest lap 12:59 /km".
		laps[i].AvgPaceSecPerKm = nil
		laps[i].PaceDeltaPctVsOverall = nil
		laps[i].PaceDeltaPctVsActive = nil
		laps[i].Display.AvgPace = ""
		laps[i].Display.PaceDeltaVsOverall = ""
		laps[i].Display.PaceDeltaVsActive = ""
	}
	out.Laps = laps
	return out
}

func shouldPreferActiveSignals(f domain.ActivityFeatures) bool {
	if f.Signals.IntervalPattern != nil {
		return true
	}
	if f.Signals.SuspectedWorkoutShape == "intervals" {
		return true
	}
	active := f.Signals.ByIntensity.Active
	recovery := f.Signals.ByIntensity.Recovery
	return active != nil && active.LapCount > 0 && recovery != nil && recovery.LapCount > 0
}
