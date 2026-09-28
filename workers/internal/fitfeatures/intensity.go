package fitfeatures

import (
	"fmt"
	"math"
	"strings"

	"github.com/racecoach/workers/internal/domain"
)

func normalizeRole(intensity string) string {
	switch strings.ToLower(strings.TrimSpace(intensity)) {
	case "active":
		return "active"
	case "recovery", "rest":
		return "recovery"
	case "warmup", "warm_up", "warm-up":
		return "warmup"
	case "cooldown", "cool_down", "cool-down":
		return "cooldown"
	case "":
		return "other"
	default:
		return "other"
	}
}

func roleLabel(role string) string {
	switch role {
	case "active":
		return "active work"
	case "recovery":
		return "recovery"
	case "warmup":
		return "warmup"
	case "cooldown":
		return "cooldown"
	default:
		return "unclassified"
	}
}

func buildByIntensity(laps []domain.Lap) domain.FeaturesByIntensity {
	groups := map[string][]domain.Lap{
		"active":   {},
		"recovery": {},
		"warmup":   {},
		"cooldown": {},
		"other":    {},
	}
	for _, lap := range laps {
		role := normalizeRole(lap.Intensity)
		groups[role] = append(groups[role], lap)
	}

	out := domain.FeaturesByIntensity{}
	if b := intensityBucket("active", groups["active"], groups["recovery"]); b != nil {
		out.Active = b
	}
	if b := intensityBucket("recovery", groups["recovery"], nil); b != nil {
		out.Recovery = b
	}
	if b := intensityBucket("warmup", groups["warmup"], nil); b != nil {
		out.Warmup = b
	}
	if b := intensityBucket("cooldown", groups["cooldown"], nil); b != nil {
		out.Cooldown = b
	}
	if b := intensityBucket("other", groups["other"], nil); b != nil {
		out.Other = b
	}
	return out
}

func intensityBucket(role string, laps []domain.Lap, recoveryForRatio []domain.Lap) *domain.IntensityBucket {
	if len(laps) == 0 {
		return nil
	}
	b := &domain.IntensityBucket{
		LapCount:   len(laps),
		LapIndexes: make([]int, 0, len(laps)),
	}
	var dist, dur, hrSum, hrN, cadSum, cadN int
	var maxHR int
	var hasMax bool
	paced := make([]domain.Lap, 0, len(laps))

	for _, lap := range laps {
		b.LapIndexes = append(b.LapIndexes, lap.Index)
		dist += lap.DistanceM
		dur += lap.DurationSec
		if lap.AvgHeartRate != nil && *lap.AvgHeartRate > 0 {
			hrSum += *lap.AvgHeartRate
			hrN++
		}
		if lap.MaxHeartRate != nil && *lap.MaxHeartRate > 0 {
			if !hasMax || *lap.MaxHeartRate > maxHR {
				maxHR = *lap.MaxHeartRate
				hasMax = true
			}
		}
		if lap.AvgCadence != nil && *lap.AvgCadence > 0 {
			cadSum += *lap.AvgCadence
			cadN++
		}
		if lap.AvgPaceSecPerKm != nil && *lap.AvgPaceSecPerKm > 0 && lap.DistanceM > 0 {
			paced = append(paced, lap)
		}
	}
	b.DistanceM = dist
	b.DurationSec = dur
	if dist > 0 && dur > 0 {
		pace := int(math.Round(float64(dur) / (float64(dist) / 1000)))
		b.AvgPaceSecPerKm = &pace
	}
	if hrN > 0 {
		hr := int(math.Round(float64(hrSum) / float64(hrN)))
		b.AvgHeartRate = &hr
	}
	if hasMax {
		b.MaxHeartRate = &maxHR
	}
	if cadN > 0 {
		cad := int(math.Round(float64(cadSum) / float64(cadN)))
		b.AvgCadence = &cad
	}

	if len(paced) > 0 {
		paces := make([]float64, len(paced))
		for i, lap := range paced {
			paces[i] = float64(*lap.AvgPaceSecPerKm)
		}
		if mean := meanFloat(paces); mean > 0 && len(paced) >= 2 {
			cv := round1(stddev(paces) / mean * 100)
			b.PaceVariabilityPct = &cv
		}
		fast, slow := 0, 0
		for i := 1; i < len(paced); i++ {
			if *paced[i].AvgPaceSecPerKm < *paced[fast].AvgPaceSecPerKm {
				fast = i
			}
			if *paced[i].AvgPaceSecPerKm > *paced[slow].AvgPaceSecPerKm {
				slow = i
			}
		}
		fi, si := paced[fast].Index, paced[slow].Index
		fp, sp := *paced[fast].AvgPaceSecPerKm, *paced[slow].AvgPaceSecPerKm
		b.FastestLapIndex = &fi
		b.SlowestLapIndex = &si
		b.FastestLapPaceSecPerKm = &fp
		b.SlowestLapPaceSecPerKm = &sp
	}

	if role == "active" && dur > 0 {
		recDur := 0
		for _, lap := range recoveryForRatio {
			recDur += lap.DurationSec
		}
		if recDur > 0 {
			ratio := round1(float64(dur) / float64(recDur))
			b.WorkRestRatio = &ratio
		}
	}

	b.Display = intensityBucketDisplay(role, b)
	return b
}

func intensityBucketDisplay(role string, b *domain.IntensityBucket) domain.IntensityBucketDisplay {
	d := domain.IntensityBucketDisplay{
		Distance:     formatDistance(b.DistanceM),
		Duration:     formatDuration(b.DurationSec),
		AvgPace:      formatPacePtr(b.AvgPaceSecPerKm),
		AvgHeartRate: formatHR(b.AvgHeartRate),
		MaxHeartRate: formatHR(b.MaxHeartRate),
		FastestLap:   formatLapRef(b.FastestLapIndex, b.FastestLapPaceSecPerKm),
		SlowestLap:   formatLapRef(b.SlowestLapIndex, b.SlowestLapPaceSecPerKm),
	}
	if b.PaceVariabilityPct != nil {
		d.PaceVariability = fmt.Sprintf("%.1f%% pace variability within %s laps", *b.PaceVariabilityPct, role)
	}
	if b.WorkRestRatio != nil {
		d.WorkRestRatio = fmt.Sprintf("work:rest %.1f:1 (by duration)", *b.WorkRestRatio)
	}

	parts := []string{fmt.Sprintf("%d %s laps", b.LapCount, role)}
	if d.Distance != "" {
		parts = append(parts, d.Distance)
	}
	if d.Duration != "" {
		parts = append(parts, d.Duration)
	}
	if d.AvgPace != "" {
		parts = append(parts, "avg "+d.AvgPace)
	}
	if d.AvgHeartRate != "" {
		parts = append(parts, "avg "+d.AvgHeartRate)
	}
	d.Summary = strings.Join(parts, ", ")
	return d
}

func buildIntervalPattern(laps []domain.Lap, byIntensity domain.FeaturesByIntensity) *domain.IntervalPattern {
	if byIntensity.Active == nil || byIntensity.Active.LapCount < 2 {
		return nil
	}
	active := make([]domain.Lap, 0, byIntensity.Active.LapCount)
	for _, lap := range laps {
		if normalizeRole(lap.Intensity) == "active" {
			active = append(active, lap)
		}
	}
	if len(active) < 2 {
		return nil
	}

	// Prefer distance-based reps when distances are consistent
	repDist := medianInt(mapInts(active, func(l domain.Lap) int { return l.DistanceM }))
	distConsistent := true
	for _, lap := range active {
		if repDist <= 0 || math.Abs(float64(lap.DistanceM-repDist)) > float64(repDist)*0.15+20 {
			distConsistent = false
			break
		}
	}
	repDur := medianInt(mapInts(active, func(l domain.Lap) int { return l.DurationSec }))
	durConsistent := true
	for _, lap := range active {
		if repDur <= 0 || math.Abs(float64(lap.DurationSec-repDur)) > float64(repDur)*0.2+5 {
			durConsistent = false
			break
		}
	}
	if !distConsistent && !durConsistent {
		// Still emit a soft pattern from active aggregates
		distConsistent = false
		durConsistent = false
	}

	indexes := make([]int, len(active))
	for i, lap := range active {
		indexes[i] = lap.Index
	}

	p := &domain.IntervalPattern{
		RepCount:         len(active),
		ActiveLapIndexes: indexes,
	}
	if distConsistent && repDist > 0 {
		p.RepDistanceM = &repDist
	}
	if durConsistent && repDur > 0 {
		p.RepDurationSec = &repDur
	}
	if byIntensity.Active.AvgPaceSecPerKm != nil {
		p.AvgActivePaceSecPerKm = byIntensity.Active.AvgPaceSecPerKm
	}
	if byIntensity.Active.AvgHeartRate != nil {
		p.AvgActiveHeartRate = byIntensity.Active.AvgHeartRate
	}
	if byIntensity.Active.PaceVariabilityPct != nil {
		p.ActivePaceVariabilityPct = byIntensity.Active.PaceVariabilityPct
	}
	if byIntensity.Active.WorkRestRatio != nil {
		p.WorkRestRatio = byIntensity.Active.WorkRestRatio
	}

	// Recovery segments between consecutive active laps
	activeSet := map[int]bool{}
	for _, idx := range indexes {
		activeSet[idx] = true
	}
	var recDurs, recDists []int
	for i := 0; i < len(indexes)-1; i++ {
		from, to := indexes[i], indexes[i+1]
		sumDur, sumDist := 0, 0
		found := false
		for _, lap := range laps {
			if lap.Index <= from || lap.Index >= to {
				continue
			}
			if normalizeRole(lap.Intensity) == "recovery" || normalizeRole(lap.Intensity) == "other" {
				sumDur += lap.DurationSec
				sumDist += lap.DistanceM
				found = true
			}
		}
		if found {
			recDurs = append(recDurs, sumDur)
			recDists = append(recDists, sumDist)
		}
	}
	if len(recDurs) > 0 {
		d := medianInt(recDurs)
		p.AvgRecoveryDurationSec = &d
	}
	if len(recDists) > 0 {
		d := medianInt(recDists)
		p.AvgRecoveryDistanceM = &d
	}

	p.Display = intervalPatternDisplay(p)
	return p
}

func intervalPatternDisplay(p *domain.IntervalPattern) domain.IntervalPatternDisplay {
	d := domain.IntervalPatternDisplay{
		AvgActivePace:      formatPacePtr(p.AvgActivePaceSecPerKm),
		AvgActiveHeartRate: formatHR(p.AvgActiveHeartRate),
	}
	rep := fmt.Sprintf("%d ×", p.RepCount)
	switch {
	case p.RepDistanceM != nil && *p.RepDistanceM > 0:
		if *p.RepDistanceM < 1000 {
			rep = fmt.Sprintf("%d × %d m", p.RepCount, *p.RepDistanceM)
		} else {
			rep = fmt.Sprintf("%d × %s", p.RepCount, formatDistance(*p.RepDistanceM))
		}
	case p.RepDurationSec != nil && *p.RepDurationSec > 0:
		rep = fmt.Sprintf("%d × %s", p.RepCount, formatDuration(*p.RepDurationSec))
	default:
		rep = fmt.Sprintf("%d active reps", p.RepCount)
	}
	d.Reps = rep

	if p.AvgRecoveryDurationSec != nil {
		rec := formatDuration(*p.AvgRecoveryDurationSec)
		if p.AvgRecoveryDistanceM != nil && *p.AvgRecoveryDistanceM > 0 {
			d.AvgRecovery = fmt.Sprintf("%s recovery (%s)", rec, formatDistance(*p.AvgRecoveryDistanceM))
		} else {
			d.AvgRecovery = fmt.Sprintf("%s recovery", rec)
		}
	}
	if p.ActivePaceVariabilityPct != nil {
		d.ActivePaceConsistency = fmt.Sprintf("%.1f%% pace variability on active reps", *p.ActivePaceVariabilityPct)
	}
	if p.WorkRestRatio != nil {
		d.WorkRestRatio = fmt.Sprintf("work:rest %.1f:1", *p.WorkRestRatio)
	}

	parts := []string{rep}
	if d.AvgActivePace != "" {
		parts = append(parts, "avg "+d.AvgActivePace)
	}
	if d.AvgRecovery != "" {
		parts = append(parts, "with "+d.AvgRecovery)
	}
	d.Summary = strings.Join(parts, " ")
	return d
}

func formatIntensityBreakdown(counts map[string]int) string {
	if len(counts) == 0 {
		return ""
	}
	order := []string{"warmup", "active", "recovery", "cooldown", "other"}
	parts := make([]string, 0, len(counts))
	seen := map[string]bool{}
	for _, key := range order {
		if n, ok := counts[key]; ok && n > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", n, key))
			seen[key] = true
		}
	}
	for key, n := range counts {
		if seen[key] || n <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d %s", n, key))
	}
	return strings.Join(parts, ", ")
}

func mapInts(laps []domain.Lap, fn func(domain.Lap) int) []int {
	out := make([]int, len(laps))
	for i, lap := range laps {
		out[i] = fn(lap)
	}
	return out
}

func medianInt(vals []int) int {
	if len(vals) == 0 {
		return 0
	}
	cp := append([]int(nil), vals...)
	for i := 0; i < len(cp); i++ {
		for j := i + 1; j < len(cp); j++ {
			if cp[j] < cp[i] {
				cp[i], cp[j] = cp[j], cp[i]
			}
		}
	}
	mid := len(cp) / 2
	if len(cp)%2 == 0 {
		return int(math.Round(float64(cp[mid-1]+cp[mid]) / 2))
	}
	return cp[mid]
}
