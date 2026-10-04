package fit

import "testing"

func TestHumanizeSport(t *testing.T) {
	t.Parallel()

	if got := humanizeSport("running"); got != "Running" {
		t.Fatalf("got %q", got)
	}
	if got := humanizeSport("cross_country_skiing"); got != "Cross Country Skiing" {
		t.Fatalf("got %q", got)
	}
	if got := humanizeSport(""); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestCleanFitString(t *testing.T) {
	t.Parallel()

	if got := cleanFitString("  Tempo Tuesday  "); got != "Tempo Tuesday" {
		t.Fatalf("got %q", got)
	}
	if got := cleanFitString(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
