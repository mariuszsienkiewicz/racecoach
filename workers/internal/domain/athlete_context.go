package domain

// AthleteContext is the physiology snapshot workers use for zone-aware analysis.
type AthleteContext struct {
	HrMax       int      `json:"hrMax"`
	HrRest      *int     `json:"hrRest,omitempty"`
	ZoneMethod  string   `json:"zoneMethod"`
	PrimaryGoal string   `json:"primaryGoal,omitempty"`
	Zones       []HrZone `json:"zones"`
}

type HrZone struct {
	Zone   int `json:"zone"`
	MinBpm int `json:"minBpm"`
	MaxBpm int `json:"maxBpm"`
}

type AthleteContextResponse struct {
	ActivityID string          `json:"activityId"`
	HasProfile bool            `json:"hasProfile"`
	Athlete    *AthleteContext `json:"athlete"`
}
