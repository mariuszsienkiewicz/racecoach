package fitfeatures

import (
	"fmt"
	"math"
)

func formatDistance(meters int) string {
	if meters <= 0 {
		return ""
	}
	km := float64(meters) / 1000
	if meters >= 10_000 {
		return fmt.Sprintf("%.1f km", km)
	}
	return fmt.Sprintf("%.2f km", km)
}

func formatDuration(sec int) string {
	if sec <= 0 {
		return ""
	}
	h := sec / 3600
	m := (sec % 3600) / 60
	s := sec % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

func formatPace(secPerKm int) string {
	if secPerKm <= 0 {
		return ""
	}
	m := secPerKm / 60
	s := secPerKm % 60
	return fmt.Sprintf("%d:%02d /km", m, s)
}

func formatPacePtr(secPerKm *int) string {
	if secPerKm == nil {
		return ""
	}
	return formatPace(*secPerKm)
}

func formatHR(bpm *int) string {
	if bpm == nil || *bpm <= 0 {
		return ""
	}
	return fmt.Sprintf("%d bpm", *bpm)
}

func formatCadence(spm *int) string {
	if spm == nil || *spm <= 0 {
		return ""
	}
	return fmt.Sprintf("%d spm", *spm)
}

func formatMeters(m *int, label string) string {
	if m == nil || *m <= 0 {
		return ""
	}
	return fmt.Sprintf("%d m %s", *m, label)
}

func formatCalories(c *int) string {
	if c == nil || *c <= 0 {
		return ""
	}
	return fmt.Sprintf("%d kcal", *c)
}

func formatPct(v float64, suffix string) string {
	sign := ""
	if v > 0 {
		sign = "+"
	}
	return fmt.Sprintf("%s%.1f%%%s", sign, v, suffix)
}

func formatPaceDeltaVsOverall(deltaPct float64) string {
	// Positive delta = slower than overall
	switch {
	case math.Abs(deltaPct) < 0.05:
		return "even with overall pace"
	case deltaPct > 0:
		return fmt.Sprintf("%.1f%% slower than overall", deltaPct)
	default:
		return fmt.Sprintf("%.1f%% faster than overall", -deltaPct)
	}
}

func formatPaceDeltaVsActive(deltaPct float64) string {
	switch {
	case math.Abs(deltaPct) < 0.05:
		return "even with active-rep average"
	case deltaPct > 0:
		return fmt.Sprintf("%.1f%% slower than active-rep average", deltaPct)
	default:
		return fmt.Sprintf("%.1f%% faster than active-rep average", -deltaPct)
	}
}

func formatShare(pct float64) string {
	return fmt.Sprintf("%.1f%% of session", pct)
}

func formatSplitBias(bias string, deltaPct *float64) string {
	switch bias {
	case "negative":
		if deltaPct != nil {
			return fmt.Sprintf("negative split (second half %.1f%% faster)", -*deltaPct)
		}
		return "negative split (second half faster)"
	case "positive":
		if deltaPct != nil {
			return fmt.Sprintf("positive split (second half %.1f%% slower)", *deltaPct)
		}
		return "positive split (second half slower)"
	case "even":
		return "even split"
	default:
		return ""
	}
}

func formatHRDrift(pct float64) string {
	switch {
	case math.Abs(pct) < 0.05:
		return "stable heart rate across the session"
	case pct > 0:
		return fmt.Sprintf("HR drift +%.1f%% (higher late vs early)", pct)
	default:
		return fmt.Sprintf("HR drift %.1f%% (lower late vs early)", pct)
	}
}

func formatWorkoutShape(shape string) string {
	switch shape {
	case "steady":
		return "steady effort"
	case "intervals":
		return "interval-style pacing"
	case "progression":
		return "progression (faster toward the end)"
	default:
		return ""
	}
}

func formatLapRef(index *int, paceSecPerKm *int) string {
	if index == nil {
		return ""
	}
	if paceSecPerKm != nil {
		return fmt.Sprintf("lap %d (%s)", *index+1, formatPace(*paceSecPerKm))
	}
	return fmt.Sprintf("lap %d", *index+1)
}
