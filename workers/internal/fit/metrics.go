package fit

import (
	"bytes"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/muktihari/fit/decoder"
	"github.com/muktihari/fit/kit/datetime"
	"github.com/muktihari/fit/profile/basetype"
	"github.com/muktihari/fit/profile/filedef"
	"github.com/muktihari/fit/profile/mesgdef"
	"github.com/muktihari/fit/profile/typedef"
)

// Metrics is the result of parsing a FIT session (internal).
type Metrics struct {
	DistanceM    int
	DurationSec  int
	AvgHeartRate *int
	MaxHeartRate *int
	Title        string
	StartedAt    *time.Time
}

func ParseMetrics(data []byte) (Metrics, error) {
	lis := filedef.NewListener()
	defer lis.Close()

	dec := decoder.New(bytes.NewReader(data),
		decoder.WithMesgListener(lis),
		decoder.WithBroadcastOnly(),
	)
	if _, err := dec.Decode(); err != nil {
		return Metrics{}, fmt.Errorf("decode: %w", err)
	}

	activity := lis.File().(*filedef.Activity)
	if len(activity.Sessions) == 0 {
		return Metrics{}, fmt.Errorf("no session in fit file")
	}
	session := activity.Sessions[0]

	return Metrics{
		DistanceM:    int(session.TotalDistanceScaled()),
		DurationSec:  int(session.TotalTimerTimeScaled()),
		AvgHeartRate: optionalUint8(session.AvgHeartRate),
		MaxHeartRate: optionalUint8(session.MaxHeartRate),
		Title:        sessionTitle(activity, session),
		StartedAt:    sessionStartTime(session),
	}, nil
}

func sessionTitle(activity *filedef.Activity, session *mesgdef.Session) string {
	for _, workout := range activity.Workouts {
		if workout == nil {
			continue
		}
		if name := cleanFitString(workout.WktName); name != "" {
			return name
		}
	}

	if name := cleanFitString(session.SportProfileName); name != "" {
		return name
	}

	if session.Sport != typedef.SportInvalid {
		return humanizeSport(session.Sport.String())
	}

	return ""
}

func sessionStartTime(session *mesgdef.Session) *time.Time {
	epoch := datetime.Epoch()
	if !session.StartTime.Before(epoch) {
		t := session.StartTime.UTC()
		return &t
	}
	if !session.Timestamp.Before(epoch) {
		t := session.Timestamp.UTC()
		return &t
	}
	return nil
}

func cleanFitString(v string) string {
	if v == basetype.StringInvalid {
		return ""
	}
	trimmed := strings.TrimSpace(v)
	if trimmed == "" || trimmed == "\x00" {
		return ""
	}
	return trimmed
}

func humanizeSport(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "invalid" {
		return ""
	}
	parts := strings.Split(raw, "_")
	for i, part := range parts {
		if part == "" {
			continue
		}
		runes := []rune(part)
		runes[0] = unicode.ToUpper(runes[0])
		parts[i] = string(runes)
	}
	return strings.Join(parts, " ")
}
