package fitfeatures

import (
	"testing"

	"github.com/racecoach/workers/internal/domain"
)

func TestClassifyActivityType(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		shape string
		label string
		dist  int
		want  string
	}{
		{name: "intervals", shape: "intervals", label: "easy", dist: 8000, want: "intervals"},
		{name: "near max", shape: "steady", label: "near_max", dist: 5000, want: "race"},
		{name: "hard", shape: "steady", label: "hard", dist: 8000, want: "tempo"},
		{name: "easy long", shape: "steady", label: "easy", dist: 18000, want: "long_run"},
		{name: "easy short", shape: "steady", label: "easy", dist: 6000, want: "easy"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			features := domain.ActivityFeatures{
				Overview: domain.FeaturesOverview{DistanceM: tc.dist},
				Signals: domain.FeaturesSignals{
					SuspectedWorkoutShape: tc.shape,
					Effort:                &domain.EffortSignal{Label: tc.label},
				},
			}
			if got := ClassifyActivityType(features); got != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}
