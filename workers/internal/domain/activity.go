package domain

type Activity struct {
	ID                 string `json:"id"`
	Title              string `json:"title"`
	Type               string `json:"type"`
	Source             string `json:"source"`
	Status             string `json:"status"`
	StartedAt          string `json:"startedAt"`
	DurationSec        int    `json:"durationSec"`
	DistanceM          int    `json:"distanceM"`
	AvgPaceSecPerKm    int    `json:"avgPaceSecPerKm"`
	AvgHeartRate       int    `json:"avgHeartRate"`
	ElevationGainM     int    `json:"elevationGainM"`
	Summary            string `json:"summary"`
	MetricsObjectKey   string `json:"metricsObjectKey"`
	StructureObjectKey string `json:"structureObjectKey"`
	FeaturesObjectKey  string `json:"featuresObjectKey"`
	SummaryObjectKey   string `json:"summaryObjectKey"`
}
