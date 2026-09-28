package domain

// ActivityFeatures is the JSON artifact written beside a .fit object.
type ActivityFeatures struct {
	ActivityID         int              `json:"activityId"`
	SchemaVersion      int              `json:"schemaVersion"`
	ObjectKey          string           `json:"objectKey"`
	MetricsObjectKey   string           `json:"metricsObjectKey"`
	StructureObjectKey string           `json:"structureObjectKey"`
	FeaturesObjectKey  string           `json:"featuresObjectKey,omitempty"`
	Overview           FeaturesOverview `json:"overview"`
	Laps               []FeatureLap     `json:"laps"`
	Signals            FeaturesSignals  `json:"signals"`
}

// FeaturesOverview is session-level context for the coach.
type FeaturesOverview struct {
	DistanceM       int                     `json:"distanceM"`
	DurationSec     int                     `json:"durationSec"`
	AvgPaceSecPerKm *int                    `json:"avgPaceSecPerKm,omitempty"`
	AvgHeartRate    *int                    `json:"avgHeartRate,omitempty"`
	MaxHeartRate    *int                    `json:"maxHeartRate,omitempty"`
	TotalAscentM    *int                    `json:"totalAscentM,omitempty"`
	TotalDescentM   *int                    `json:"totalDescentM,omitempty"`
	TotalCalories   *int                    `json:"totalCalories,omitempty"`
	LapCount        int                     `json:"lapCount"`
	Display         FeaturesOverviewDisplay `json:"display"`
}

// FeaturesOverviewDisplay is human-readable overview text for LLM answers.
type FeaturesOverviewDisplay struct {
	Distance      string `json:"distance,omitempty"`
	Duration      string `json:"duration,omitempty"`
	AvgPace       string `json:"avgPace,omitempty"`
	AvgHeartRate  string `json:"avgHeartRate,omitempty"`
	MaxHeartRate  string `json:"maxHeartRate,omitempty"`
	TotalAscent   string `json:"totalAscent,omitempty"`
	TotalDescent  string `json:"totalDescent,omitempty"`
	TotalCalories string `json:"totalCalories,omitempty"`
	LapCount      string `json:"lapCount,omitempty"`
}

// FeatureLap is a structure lap enriched with relative coaching cues.
type FeatureLap struct {
	Lap
	Role                  string            `json:"role,omitempty"` // active | recovery | warmup | cooldown | other
	IsWork                bool              `json:"isWork"`
	PaceDeltaPctVsOverall *float64          `json:"paceDeltaPctVsOverall,omitempty"`
	PaceDeltaPctVsActive  *float64          `json:"paceDeltaPctVsActive,omitempty"`
	DistanceSharePct      *float64          `json:"distanceSharePct,omitempty"`
	DurationSharePct      *float64          `json:"durationSharePct,omitempty"`
	Display               FeatureLapDisplay `json:"display"`
}

// FeatureLapDisplay is human-readable lap text for LLM answers.
type FeatureLapDisplay struct {
	Role               string `json:"role,omitempty"`
	Distance           string `json:"distance,omitempty"`
	Duration           string `json:"duration,omitempty"`
	AvgPace            string `json:"avgPace,omitempty"`
	AvgHeartRate       string `json:"avgHeartRate,omitempty"`
	MaxHeartRate       string `json:"maxHeartRate,omitempty"`
	MinHeartRate       string `json:"minHeartRate,omitempty"`
	AvgCadence         string `json:"avgCadence,omitempty"`
	TotalAscent        string `json:"totalAscent,omitempty"`
	TotalDescent       string `json:"totalDescent,omitempty"`
	TotalCalories      string `json:"totalCalories,omitempty"`
	PaceDeltaVsOverall string `json:"paceDeltaVsOverall,omitempty"`
	PaceDeltaVsActive  string `json:"paceDeltaVsActive,omitempty"`
	DistanceShare      string `json:"distanceShare,omitempty"`
	DurationShare      string `json:"durationShare,omitempty"`
}

// FeaturesSignals are compact derived hints so the LLM need not recompute basics.
type FeaturesSignals struct {
	PaceVariabilityPct     *float64               `json:"paceVariabilityPct,omitempty"`
	FastestLapIndex        *int                   `json:"fastestLapIndex,omitempty"`
	SlowestLapIndex        *int                   `json:"slowestLapIndex,omitempty"`
	FastestLapPaceSecPerKm *int                   `json:"fastestLapPaceSecPerKm,omitempty"`
	SlowestLapPaceSecPerKm *int                   `json:"slowestLapPaceSecPerKm,omitempty"`
	SplitBias              string                 `json:"splitBias,omitempty"` // negative | positive | even | unknown
	SecondHalfPaceDeltaPct *float64               `json:"secondHalfPaceDeltaPct,omitempty"`
	HeartRateDriftPct      *float64               `json:"heartRateDriftPct,omitempty"`
	IntensityCounts        map[string]int         `json:"intensityCounts,omitempty"`
	LapTriggerCounts       map[string]int         `json:"lapTriggerCounts,omitempty"`
	SuspectedWorkoutShape  string                 `json:"suspectedWorkoutShape,omitempty"` // steady | intervals | progression | unknown
	Effort                 *EffortSignal          `json:"effort,omitempty"`
	ByIntensity            FeaturesByIntensity    `json:"byIntensity"`
	IntervalPattern        *IntervalPattern       `json:"intervalPattern,omitempty"`
	Display                FeaturesSignalsDisplay `json:"display"`
}

// EffortSignal is a coarse intensity label derived from HR (and shape), for the LLM.
type EffortSignal struct {
	Label    string   `json:"label"` // easy | moderate | hard | near_max | unknown
	Evidence []string `json:"evidence,omitempty"`
	Display  string   `json:"display,omitempty"`
}

// FeaturesByIntensity aggregates laps by FIT intensity role.
type FeaturesByIntensity struct {
	Active   *IntensityBucket `json:"active,omitempty"`
	Recovery *IntensityBucket `json:"recovery,omitempty"`
	Warmup   *IntensityBucket `json:"warmup,omitempty"`
	Cooldown *IntensityBucket `json:"cooldown,omitempty"`
	Other    *IntensityBucket `json:"other,omitempty"`
}

// IntensityBucket is totals/averages for one intensity role.
type IntensityBucket struct {
	LapCount               int                    `json:"lapCount"`
	LapIndexes             []int                  `json:"lapIndexes"`
	DistanceM              int                    `json:"distanceM"`
	DurationSec            int                    `json:"durationSec"`
	AvgPaceSecPerKm        *int                   `json:"avgPaceSecPerKm,omitempty"`
	AvgHeartRate           *int                   `json:"avgHeartRate,omitempty"`
	MaxHeartRate           *int                   `json:"maxHeartRate,omitempty"`
	AvgCadence             *int                   `json:"avgCadence,omitempty"`
	PaceVariabilityPct     *float64               `json:"paceVariabilityPct,omitempty"`
	FastestLapIndex        *int                   `json:"fastestLapIndex,omitempty"`
	SlowestLapIndex        *int                   `json:"slowestLapIndex,omitempty"`
	FastestLapPaceSecPerKm *int                   `json:"fastestLapPaceSecPerKm,omitempty"`
	SlowestLapPaceSecPerKm *int                   `json:"slowestLapPaceSecPerKm,omitempty"`
	WorkRestRatio          *float64               `json:"workRestRatio,omitempty"` // only on active: activeDur / recoveryDur
	Display                IntensityBucketDisplay `json:"display"`
}

// IntensityBucketDisplay is human-readable intensity-bucket text.
type IntensityBucketDisplay struct {
	Summary         string `json:"summary,omitempty"`
	Distance        string `json:"distance,omitempty"`
	Duration        string `json:"duration,omitempty"`
	AvgPace         string `json:"avgPace,omitempty"`
	AvgHeartRate    string `json:"avgHeartRate,omitempty"`
	MaxHeartRate    string `json:"maxHeartRate,omitempty"`
	PaceVariability string `json:"paceVariability,omitempty"`
	FastestLap      string `json:"fastestLap,omitempty"`
	SlowestLap      string `json:"slowestLap,omitempty"`
	WorkRestRatio   string `json:"workRestRatio,omitempty"`
}

// IntervalPattern describes repeating active + recovery structure.
type IntervalPattern struct {
	RepCount                 int                    `json:"repCount"`
	ActiveLapIndexes         []int                  `json:"activeLapIndexes"`
	RepDistanceM             *int                   `json:"repDistanceM,omitempty"`
	RepDurationSec           *int                   `json:"repDurationSec,omitempty"`
	AvgActivePaceSecPerKm    *int                   `json:"avgActivePaceSecPerKm,omitempty"`
	AvgActiveHeartRate       *int                   `json:"avgActiveHeartRate,omitempty"`
	AvgRecoveryDurationSec   *int                   `json:"avgRecoveryDurationSec,omitempty"`
	AvgRecoveryDistanceM     *int                   `json:"avgRecoveryDistanceM,omitempty"`
	ActivePaceVariabilityPct *float64               `json:"activePaceVariabilityPct,omitempty"`
	WorkRestRatio            *float64               `json:"workRestRatio,omitempty"`
	Display                  IntervalPatternDisplay `json:"display"`
}

// IntervalPatternDisplay is a single coach-ready pattern sentence plus parts.
type IntervalPatternDisplay struct {
	Summary               string `json:"summary,omitempty"`
	Reps                  string `json:"reps,omitempty"`
	AvgActivePace         string `json:"avgActivePace,omitempty"`
	AvgActiveHeartRate    string `json:"avgActiveHeartRate,omitempty"`
	AvgRecovery           string `json:"avgRecovery,omitempty"`
	ActivePaceConsistency string `json:"activePaceConsistency,omitempty"`
	WorkRestRatio         string `json:"workRestRatio,omitempty"`
}

// FeaturesSignalsDisplay is human-readable signal text for LLM answers.
type FeaturesSignalsDisplay struct {
	PaceVariability       string `json:"paceVariability,omitempty"`
	FastestLap            string `json:"fastestLap,omitempty"`
	SlowestLap            string `json:"slowestLap,omitempty"`
	FastestLapPace        string `json:"fastestLapPace,omitempty"`
	SlowestLapPace        string `json:"slowestLapPace,omitempty"`
	SplitBias             string `json:"splitBias,omitempty"`
	SecondHalfPaceDelta   string `json:"secondHalfPaceDelta,omitempty"`
	HeartRateDrift        string `json:"heartRateDrift,omitempty"`
	SuspectedWorkoutShape string `json:"suspectedWorkoutShape,omitempty"`
	IntensityBreakdown    string `json:"intensityBreakdown,omitempty"`
	ActiveWork            string `json:"activeWork,omitempty"`
	IntervalPattern       string `json:"intervalPattern,omitempty"`
	Effort                string `json:"effort,omitempty"`
}

// PatchActivityFeaturesRequest is the JSON payload for the PATCH endpoint.
type PatchActivityFeaturesRequest struct {
	FeaturesObjectKey string `json:"featuresObjectKey"`
}

// ActivityFeaturesReady is published after features are stored.
type ActivityFeaturesReady struct {
	ActivityID         int    `json:"activityId"`
	UserID             int    `json:"userId"`
	StorageBucket      string `json:"storageBucket"`
	ObjectKey          string `json:"objectKey"`
	MetricsObjectKey   string `json:"metricsObjectKey"`
	StructureObjectKey string `json:"structureObjectKey"`
	FeaturesObjectKey  string `json:"featuresObjectKey"`
}
