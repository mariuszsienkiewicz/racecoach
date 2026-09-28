package domain

// CoachSummaryPayload is the JSON object returned by the LLM.
type CoachSummaryPayload struct {
	Headline   string   `json:"headline"`
	Summary    string   `json:"summary"`
	Highlights []string `json:"highlights"`
	Watchouts  []string `json:"watchouts"`
	NextFocus  string   `json:"nextFocus"`
}

// ActivitySummary is the JSON artifact written beside a .fit object.
type ActivitySummary struct {
	ActivityID        int      `json:"activityId"`
	SchemaVersion     int      `json:"schemaVersion"`
	ObjectKey         string   `json:"objectKey"`
	FeaturesObjectKey string   `json:"featuresObjectKey,omitempty"`
	SummaryObjectKey  string   `json:"summaryObjectKey,omitempty"`
	Headline          string   `json:"headline"`
	Summary           string   `json:"summary"`
	Highlights        []string `json:"highlights"`
	Watchouts         []string `json:"watchouts"`
	NextFocus         string   `json:"nextFocus"`
}

// PatchActivitySummaryRequest is the JSON payload for the PATCH endpoint.
type PatchActivitySummaryRequest struct {
	SummaryObjectKey string `json:"summaryObjectKey"`
	Summary          string `json:"summary"`
}

// ActivitySummaryReady is published after summary is stored and activity is ready.
type ActivitySummaryReady struct {
	ActivityID        int    `json:"activityId"`
	UserID            int    `json:"userId"`
	StorageBucket     string `json:"storageBucket"`
	ObjectKey         string `json:"objectKey"`
	FeaturesObjectKey string `json:"featuresObjectKey"`
	SummaryObjectKey  string `json:"summaryObjectKey"`
}
