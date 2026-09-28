package domain

// ActivityMetrics is the JSON artifact written beside a .fit object
type ActivityMetrics struct {
	ActivityID       int    `json:"activityId"`
	DistanceM        int    `json:"distanceM"`
	DurationSec      int    `json:"durationSec"`
	AvgHeartRate     *int   `json:"avgHeartRate"`
	MaxHeartRate     *int   `json:"maxHeartRate,omitempty"`
	MetricsObjectKey string `json:"metricsObjectKey,omitempty"`
}

// ActivityMetricsReady is the JSON artifact published to the queue.
type ActivityMetricsReady struct {
	ActivityID       int    `json:"activityId"`
	UserID           int    `json:"userId"`
	StorageBucket    string `json:"storageBucket"`
	ObjectKey        string `json:"objectKey"`
	MetricsObjectKey string `json:"metricsObjectKey"`
	DistanceM        int    `json:"distanceM"`
	DurationSec      int    `json:"durationSec"`
	AvgHeartRate     *int   `json:"avgHeartRate"`
	ReadyAt          string `json:"readyAt"`
}
