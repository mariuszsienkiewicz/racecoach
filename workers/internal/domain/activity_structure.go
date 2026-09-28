package domain

// ActivityStructure is the JSON artifact written beside a .fit object
type ActivityStructure struct {
	ActivityID    int    `json:"activityId"`
	SchemaVersion int    `json:"schemaVersion"`
	ObjectKey     string `json:"objectKey"`
	Laps          []Lap  `json:"laps"`
}

// Lap is a single lap of an activity.
type Lap struct {
	Index           int    `json:"index"`
	StartTimestamp  string `json:"startTimestamp,omitempty"`
	EndTimestamp    string `json:"endTimestamp,omitempty"`
	DistanceM       int    `json:"distanceM"`
	DurationSec     int    `json:"durationSec"`
	AvgPaceSecPerKm *int   `json:"avgPaceSecPerKm,omitempty"`
	AvgHeartRate    *int   `json:"avgHeartRate,omitempty"`
	MaxHeartRate    *int   `json:"maxHeartRate,omitempty"`
	MinHeartRate    *int   `json:"minHeartRate,omitempty"`
	AvgCadence      *int   `json:"avgCadence,omitempty"`
	TotalAscentM    *int   `json:"totalAscentM,omitempty"`
	TotalDescentM   *int   `json:"totalDescentM,omitempty"`
	TotalCalories   *int   `json:"totalCalories,omitempty"`
	LapTrigger      string `json:"lapTrigger,omitempty"`
	Intensity       string `json:"intensity,omitempty"`
}

// PatchActivityStructureRequest is the JSON payload for the PATCH endpoint.
type PatchActivityStructureRequest struct {
	StructureObjectKey string `json:"structureObjectKey"`
}

// ActivityStructureReady is the JSON artifact published to the queue.
type ActivityStructureReady struct {
	ActivityID         int    `json:"activityId"`
	UserID             int    `json:"userId"`
	StorageBucket      string `json:"storageBucket"`
	ObjectKey          string `json:"objectKey"`
	StructureObjectKey string `json:"structureObjectKey"`
}
