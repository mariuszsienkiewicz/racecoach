package fitfeatures

import (
	"fmt"
	"math"
	"strings"

	"github.com/racecoach/workers/internal/domain"
)

// Build merges metrics + structure into an LLM-ready features artifact.
// Numeric fields stay authoritative. Display.* strings are for LLM answers.
func Build(metrics domain.ActivityMetrics, structure domain.ActivityStructure, athlete *domain.AthleteContext) domain.ActivityFeatures {
	overview := domain.FeaturesOverview{
		DistanceM:    metrics.DistanceM,
		DurationSec:  metrics.DurationSec,
		AvgHeartRate: metrics.AvgHeartRate,
		MaxHeartRate: metrics.MaxHeartRate,
		LapCount:     len(structure.Laps),
	}
	if overview.MaxHeartRate == nil {
		overview.MaxHeartRate = maxLapHeartRate(structure.Laps)
	}
	if metrics.DistanceM > 0 && metrics.DurationSec > 0 {
		pace := int(math.Round(float64(metrics.DurationSec) / (float64(metrics.DistanceM) / 1000)))
		overview.AvgPaceSecPerKm = &pace
	}

	ascent, descent, calories := sumElevationAndCalories(structure.Laps)
	overview.TotalAscentM = ascent
	overview.TotalDescentM = descent
	overview.TotalCalories = calories
	overview.Display = overviewDisplay(overview)

	byIntensity := buildByIntensity(structure.Laps)
	laps := enrichLaps(structure.Laps, overview, byIntensity, athlete)
	signals := deriveSignals(structure.Laps, overview, byIntensity, laps, athlete)

	return domain.ActivityFeatures{
		ActivityID:       structure.ActivityID,
		SchemaVersion:    5,
		ObjectKey:        structure.ObjectKey,
		MetricsObjectKey: metrics.MetricsObjectKey,
		Athlete:          athlete,
		Overview:         overview,
		Laps:             laps,
		Signals:          signals,
	}
}

func overviewDisplay(o domain.FeaturesOverview) domain.FeaturesOverviewDisplay {
	d := domain.FeaturesOverviewDisplay{
		Distance:      formatDistance(o.DistanceM),
		Duration:      formatDuration(o.DurationSec),
		AvgPace:       formatPacePtr(o.AvgPaceSecPerKm),
		AvgHeartRate:  formatHR(o.AvgHeartRate),
		MaxHeartRate:  formatHR(o.MaxHeartRate),
		TotalAscent:   formatMeters(o.TotalAscentM, "ascent"),
		TotalDescent:  formatMeters(o.TotalDescentM, "descent"),
		TotalCalories: formatCalories(o.TotalCalories),
	}
	if o.LapCount > 0 {
		d.LapCount = fmt.Sprintf("%d laps", o.LapCount)
	}
	return d
}

func enrichLaps(laps []domain.Lap, overview domain.FeaturesOverview, byIntensity domain.FeaturesByIntensity, athlete *domain.AthleteContext) []domain.FeatureLap {
	var activeAvgPace *int
	if byIntensity.Active != nil {
		activeAvgPace = byIntensity.Active.AvgPaceSecPerKm
	}

	out := make([]domain.FeatureLap, 0, len(laps))
	for _, lap := range laps {
		role := normalizeRole(lap.Intensity)
		fl := domain.FeatureLap{
			Lap:    lap,
			Role:   role,
			IsWork: role == "active",
		}
		if lap.AvgHeartRate != nil {
			fl.AvgHrZone = zoneForHR(*lap.AvgHeartRate, athlete)
		}
		if overview.AvgPaceSecPerKm != nil && lap.AvgPaceSecPerKm != nil && *overview.AvgPaceSecPerKm > 0 {
			delta := round1((float64(*lap.AvgPaceSecPerKm) - float64(*overview.AvgPaceSecPerKm)) / float64(*overview.AvgPaceSecPerKm) * 100)
			fl.PaceDeltaPctVsOverall = &delta
		}
		if activeAvgPace != nil && lap.AvgPaceSecPerKm != nil && *activeAvgPace > 0 && role == "active" {
			delta := round1((float64(*lap.AvgPaceSecPerKm) - float64(*activeAvgPace)) / float64(*activeAvgPace) * 100)
			fl.PaceDeltaPctVsActive = &delta
		}
		if overview.DistanceM > 0 && lap.DistanceM > 0 {
			share := round1(float64(lap.DistanceM) / float64(overview.DistanceM) * 100)
			fl.DistanceSharePct = &share
		}
		if overview.DurationSec > 0 && lap.DurationSec > 0 {
			share := round1(float64(lap.DurationSec) / float64(overview.DurationSec) * 100)
			fl.DurationSharePct = &share
		}
		fl.Display = lapDisplay(fl)
		out = append(out, fl)
	}
	return out
}

func lapDisplay(fl domain.FeatureLap) domain.FeatureLapDisplay {
	d := domain.FeatureLapDisplay{
		Role:          roleLabel(fl.Role),
		Distance:      formatDistance(fl.DistanceM),
		Duration:      formatDuration(fl.DurationSec),
		AvgPace:       formatPacePtr(fl.AvgPaceSecPerKm),
		AvgHeartRate:  formatHR(fl.AvgHeartRate),
		MaxHeartRate:  formatHR(fl.MaxHeartRate),
		MinHeartRate:  formatHR(fl.MinHeartRate),
		AvgCadence:    formatCadence(fl.AvgCadence),
		TotalAscent:   formatMeters(fl.TotalAscentM, "ascent"),
		TotalDescent:  formatMeters(fl.TotalDescentM, "descent"),
		TotalCalories: formatCalories(fl.TotalCalories),
	}
	if fl.PaceDeltaPctVsOverall != nil {
		d.PaceDeltaVsOverall = formatPaceDeltaVsOverall(*fl.PaceDeltaPctVsOverall)
	}
	if fl.PaceDeltaPctVsActive != nil {
		d.PaceDeltaVsActive = formatPaceDeltaVsActive(*fl.PaceDeltaPctVsActive)
	}
	if fl.DistanceSharePct != nil {
		d.DistanceShare = formatShare(*fl.DistanceSharePct)
	}
	if fl.DurationSharePct != nil {
		d.DurationShare = formatShare(*fl.DurationSharePct)
	}
	d.AvgHrZone = formatZone(fl.AvgHrZone)
	return d
}

func deriveSignals(laps []domain.Lap, overview domain.FeaturesOverview, byIntensity domain.FeaturesByIntensity, featureLaps []domain.FeatureLap, athlete *domain.AthleteContext) domain.FeaturesSignals {
	_ = overview
	signals := domain.FeaturesSignals{
		IntensityCounts:       countBy(laps, func(l domain.Lap) string { return normalizeRole(l.Intensity) }),
		LapTriggerCounts:      countBy(laps, func(l domain.Lap) string { return l.LapTrigger }),
		SplitBias:             "unknown",
		SuspectedWorkoutShape: "unknown",
		ByIntensity:           byIntensity,
		HasAthleteZones:       athlete != nil && len(athlete.Zones) > 0,
	}

	paced := make([]domain.Lap, 0, len(laps))
	for _, lap := range laps {
		if lap.AvgPaceSecPerKm != nil && *lap.AvgPaceSecPerKm > 0 && lap.DistanceM > 0 {
			paced = append(paced, lap)
		}
	}
	if len(paced) > 0 {
		paces := make([]float64, len(paced))
		for i, lap := range paced {
			paces[i] = float64(*lap.AvgPaceSecPerKm)
		}
		mean := meanFloat(paces)
		if mean > 0 {
			cv := round1(stddev(paces) / mean * 100)
			signals.PaceVariabilityPct = &cv
		}

		fastIdx, slowIdx := 0, 0
		for i := 1; i < len(paced); i++ {
			if *paced[i].AvgPaceSecPerKm < *paced[fastIdx].AvgPaceSecPerKm {
				fastIdx = i
			}
			if *paced[i].AvgPaceSecPerKm > *paced[slowIdx].AvgPaceSecPerKm {
				slowIdx = i
			}
		}
		fi, si := paced[fastIdx].Index, paced[slowIdx].Index
		fp, sp := *paced[fastIdx].AvgPaceSecPerKm, *paced[slowIdx].AvgPaceSecPerKm
		signals.FastestLapIndex = &fi
		signals.SlowestLapIndex = &si
		signals.FastestLapPaceSecPerKm = &fp
		signals.SlowestLapPaceSecPerKm = &sp

		if delta, ok := halfSplitDelta(paced); ok {
			signals.SecondHalfPaceDeltaPct = &delta
			switch {
			case delta <= -2:
				signals.SplitBias = "negative"
			case delta >= 2:
				signals.SplitBias = "positive"
			default:
				signals.SplitBias = "even"
			}
		}
	}

	if drift, ok := heartRateDrift(laps); ok {
		signals.HeartRateDriftPct = &drift
	}

	signals.SuspectedWorkoutShape = guessShape(signals, paced, byIntensity)
	// Session-wide split is misleading for interval workouts (warmup/recovery dominate).
	if signals.SuspectedWorkoutShape == "intervals" {
		signals.SplitBias = "unknown"
		signals.SecondHalfPaceDeltaPct = nil
	}

	signals.IntervalPattern = buildIntervalPattern(laps, byIntensity)
	signals.DominantHrZone = dominantZoneFromLaps(featureLaps)
	signals.TimeInZoneSec = timeInZoneSec(featureLaps)
	signals.Effort = deriveEffort(overview, signals, athlete)
	signals.Display = signalsDisplay(signals, athlete)
	return signals
}

func deriveEffort(overview domain.FeaturesOverview, signals domain.FeaturesSignals, athlete *domain.AthleteContext) *domain.EffortSignal {
	evidence := make([]string, 0, 6)
	if overview.MaxHeartRate != nil {
		evidence = append(evidence, formatHR(overview.MaxHeartRate)+" max HR")
	}
	if overview.AvgHeartRate != nil {
		evidence = append(evidence, formatHR(overview.AvgHeartRate)+" avg HR")
	}
	if overview.Display.AvgPace != "" {
		evidence = append(evidence, overview.Display.AvgPace+" avg pace")
	}
	if signals.SuspectedWorkoutShape == "intervals" {
		evidence = append(evidence, "interval session structure")
	}

	// Personalized path: zones beat absolute bpm heuristics.
	if athlete != nil && len(athlete.Zones) > 0 && signals.DominantHrZone != nil {
		zone := *signals.DominantHrZone
		label := effortFromZone(zone)
		// Only escalate easy/moderate → hard when FIT has real work/rest intensity roles.
		// Km-split "intervals" guesses must not override an aerobic Z1–Z2 long run.
		if signals.SuspectedWorkoutShape == "intervals" &&
			(label == "easy" || label == "moderate") &&
			hasStructuredWorkRest(signals.ByIntensity) {
			label = "hard"
			evidence = append(evidence, "structured work/rest intensity raises effort vs steady zone")
		}
		evidence = append(evidence, formatZone(&zone)+" dominant zone from athlete profile")
		display := formatEffortLabel(label)
		if len(evidence) > 0 {
			display = display + " (" + joinEvidence(evidence) + ")"
		}
		return &domain.EffortSignal{Label: label, Confidence: "high", Evidence: evidence, Display: display}
	}

	// Fallback without profile: conservative, low confidence. Avoid shaming high absolute HR.
	label := "unknown"
	maxHR := overview.MaxHeartRate
	avgHR := overview.AvgHeartRate
	switch {
	case signals.SuspectedWorkoutShape == "intervals" && hasStructuredWorkRest(signals.ByIntensity):
		label = "hard"
	case maxHR != nil && avgHR != nil && *maxHR >= 190 && *avgHR >= 170:
		label = "near_max"
	case maxHR != nil && avgHR != nil && *maxHR >= 185 && *avgHR >= 160:
		label = "hard"
	case avgHR != nil && *avgHR < 140:
		label = "easy"
	default:
		label = "moderate"
		evidence = append(evidence, "no athlete HR zones configured - intensity is approximate")
	}

	display := formatEffortLabel(label)
	if len(evidence) > 0 {
		display = display + " (" + joinEvidence(evidence) + ")"
	}
	return &domain.EffortSignal{Label: label, Confidence: "low", Evidence: evidence, Display: display}
}

func formatEffortLabel(label string) string {
	switch label {
	case "easy":
		return "easy effort"
	case "moderate":
		return "moderate effort"
	case "hard":
		return "hard effort"
	case "near_max":
		return "near-max / race-level effort"
	default:
		return "effort unclear from available HR data"
	}
}

func joinEvidence(parts []string) string {
	out := ""
	for i, p := range parts {
		if i > 0 {
			out += "; "
		}
		out += p
	}
	return out
}

func maxLapHeartRate(laps []domain.Lap) *int {
	var max int
	found := false
	for _, lap := range laps {
		if lap.MaxHeartRate != nil && *lap.MaxHeartRate > 0 {
			if !found || *lap.MaxHeartRate > max {
				max = *lap.MaxHeartRate
				found = true
			}
		}
	}
	if !found {
		return nil
	}
	return &max
}

func signalsDisplay(s domain.FeaturesSignals, athlete *domain.AthleteContext) domain.FeaturesSignalsDisplay {
	d := domain.FeaturesSignalsDisplay{
		SplitBias:             formatSplitBias(s.SplitBias, s.SecondHalfPaceDeltaPct),
		SuspectedWorkoutShape: formatWorkoutShape(s.SuspectedWorkoutShape),
		FastestLap:            formatLapRef(s.FastestLapIndex, s.FastestLapPaceSecPerKm),
		SlowestLap:            formatLapRef(s.SlowestLapIndex, s.SlowestLapPaceSecPerKm),
		FastestLapPace:        formatPacePtr(s.FastestLapPaceSecPerKm),
		SlowestLapPace:        formatPacePtr(s.SlowestLapPaceSecPerKm),
		IntensityBreakdown:    formatIntensityBreakdown(s.IntensityCounts),
		AthleteZones:          formatAthleteZones(athlete),
		DominantHrZone:        formatZone(s.DominantHrZone),
	}
	if s.PaceVariabilityPct != nil {
		d.PaceVariability = fmt.Sprintf("%.1f%% pace variability across all laps", *s.PaceVariabilityPct)
	}
	if s.SecondHalfPaceDeltaPct != nil && s.SplitBias != "unknown" {
		d.SecondHalfPaceDelta = formatPct(*s.SecondHalfPaceDeltaPct, " second-half vs first-half pace")
	}
	if s.HeartRateDriftPct != nil {
		d.HeartRateDrift = formatHRDrift(*s.HeartRateDriftPct)
	}
	if s.ByIntensity.Active != nil && s.ByIntensity.Active.Display.Summary != "" {
		d.ActiveWork = s.ByIntensity.Active.Display.Summary
	}
	if s.IntervalPattern != nil {
		d.IntervalPattern = s.IntervalPattern.Display.Summary
	}
	if s.Effort != nil {
		d.Effort = s.Effort.Display
	}
	return d
}

func halfSplitDelta(paced []domain.Lap) (float64, bool) {
	if len(paced) < 2 {
		return 0, false
	}
	mid := len(paced) / 2
	first := meanPace(paced[:mid])
	second := meanPace(paced[mid:])
	if first <= 0 {
		return 0, false
	}
	return round1((second - first) / first * 100), true
}

func heartRateDrift(laps []domain.Lap) (float64, bool) {
	var hrs []float64
	for _, lap := range laps {
		if lap.AvgHeartRate != nil && *lap.AvgHeartRate > 0 && lap.DurationSec >= 30 {
			hrs = append(hrs, float64(*lap.AvgHeartRate))
		}
	}
	if len(hrs) < 3 {
		return 0, false
	}
	third := len(hrs) / 3
	if third < 1 {
		return 0, false
	}
	first := meanFloat(hrs[:third])
	last := meanFloat(hrs[len(hrs)-third:])
	if first <= 0 {
		return 0, false
	}
	return round1((last - first) / first * 100), true
}

func guessShape(signals domain.FeaturesSignals, paced []domain.Lap, byIntensity domain.FeaturesByIntensity) string {
	// Strong evidence: device-tagged active work + recovery.
	if hasStructuredWorkRest(byIntensity) {
		return "intervals"
	}

	if len(paced) < 2 {
		return "unknown"
	}

	if mostlyKmSplits(paced) {
		return shapeFromOverallPace(signals, paced)
	}

	cv := 0.0
	if signals.PaceVariabilityPct != nil {
		cv = *signals.PaceVariabilityPct
	}

	if cv >= 12 && len(paced) >= 6 {
		flips := 0
		for i := 2; i < len(paced); i++ {
			d1 := *paced[i-1].AvgPaceSecPerKm - *paced[i-2].AvgPaceSecPerKm
			d2 := *paced[i].AvgPaceSecPerKm - *paced[i-1].AvgPaceSecPerKm
			if d1*d2 < 0 {
				flips++
			}
		}
		// Require a clear alternating pattern, not mild long-run noise.
		if flips >= int(math.Ceil(float64(len(paced)-2)*0.66)) {
			return "intervals"
		}
	}

	return shapeFromOverallPace(signals, paced)
}

func shapeFromOverallPace(signals domain.FeaturesSignals, paced []domain.Lap) string {
	if len(paced) >= 3 {
		n := len(paced) / 3
		if n < 1 {
			n = 1
		}
		first := meanPace(paced[:n])
		last := meanPace(paced[len(paced)-n:])
		if first > 0 && (first-last)/first*100 >= 3 {
			return "progression"
		}
	}

	cv := 0.0
	if signals.PaceVariabilityPct != nil {
		cv = *signals.PaceVariabilityPct
	}
	if cv < 8 {
		return "steady"
	}
	return "unknown"
}

func hasStructuredWorkRest(byIntensity domain.FeaturesByIntensity) bool {
	return byIntensity.Active != nil && byIntensity.Active.LapCount >= 2 &&
		byIntensity.Recovery != nil && byIntensity.Recovery.LapCount >= 1
}

func mostlyKmSplits(laps []domain.Lap) bool {
	if len(laps) == 0 {
		return false
	}
	km := 0
	for _, lap := range laps {
		if isKmSplitTrigger(lap.LapTrigger) {
			km++
		}
	}
	return float64(km)/float64(len(laps)) >= 0.7
}

func isKmSplitTrigger(trigger string) bool {
	switch strings.ToLower(strings.TrimSpace(trigger)) {
	case "synthetic_km", "distance", "distance_manual":
		return true
	default:
		return false
	}
}

func sumElevationAndCalories(laps []domain.Lap) (ascent, descent, calories *int) {
	var a, d, c int
	var hasA, hasD, hasC bool
	for _, lap := range laps {
		if lap.TotalAscentM != nil {
			a += *lap.TotalAscentM
			hasA = true
		}
		if lap.TotalDescentM != nil {
			d += *lap.TotalDescentM
			hasD = true
		}
		if lap.TotalCalories != nil {
			c += *lap.TotalCalories
			hasC = true
		}
	}
	if hasA {
		ascent = &a
	}
	if hasD {
		descent = &d
	}
	if hasC {
		calories = &c
	}
	return
}

func countBy(laps []domain.Lap, key func(domain.Lap) string) map[string]int {
	out := map[string]int{}
	for _, lap := range laps {
		k := key(lap)
		if k == "" {
			continue
		}
		out[k]++
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func meanPace(laps []domain.Lap) float64 {
	paces := make([]float64, 0, len(laps))
	for _, lap := range laps {
		if lap.AvgPaceSecPerKm != nil {
			paces = append(paces, float64(*lap.AvgPaceSecPerKm))
		}
	}
	return meanFloat(paces)
}

func meanFloat(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	var sum float64
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}

func stddev(vals []float64) float64 {
	if len(vals) < 2 {
		return 0
	}
	m := meanFloat(vals)
	var sum float64
	for _, v := range vals {
		d := v - m
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(vals)))
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}
