package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
	"testing"
)

func TestIsUnavailable(t *testing.T) {
	t.Parallel()

	refused := &net.OpError{Op: "dial", Err: errors.New("connection refused")}
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil", err: nil, want: false},
		{name: "connection refused wrap", err: fmt.Errorf("do: %w", refused), want: true},
		{name: "url error", err: &url.Error{Op: "Post", URL: "http://x", Err: refused}, want: true},
		{name: "deadline", err: context.DeadlineExceeded, want: true},
		{name: "http 503", err: &HTTPError{StatusCode: 503, Body: "down"}, want: true},
		{name: "http 400", err: &HTTPError{StatusCode: 400, Body: "bad"}, want: false},
		{name: "plain", err: errors.New("unmarshal: unexpected EOF"), want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsUnavailable(tc.err); got != tc.want {
				t.Fatalf("IsUnavailable(%v)=%v want %v", tc.err, got, tc.want)
			}
		})
	}
}

func TestExtractJSONObjectFromProse(t *testing.T) {
	in := "Here is the summary JSON:\n{\n  \"headline\": \"Hi\",\n  \"summary\": \"Body\"\n}\nThanks!"
	got := ExtractJSONObject(in)
	want := "{\n  \"headline\": \"Hi\",\n  \"summary\": \"Body\"\n}"
	if got != want {
		t.Fatalf("got=%q want=%q", got, want)
	}
}

func TestExtractJSONObjectFromFence(t *testing.T) {
	in := "```json\n{\"headline\":\"A\",\"summary\":\"B\"}\n```"
	got := ExtractJSONObject(in)
	if got != `{"headline":"A","summary":"B"}` {
		t.Fatalf("got=%q", got)
	}
}

func TestConsumeCompletionStream(t *testing.T) {
	t.Parallel()

	body := strings.Join([]string{
		`data: {"choices":[{"delta":{"role":"assistant"}}]}`,
		``,
		`data: {"choices":[{"delta":{"content":"Hel"}}]}`,
		`data: {"choices":[{"delta":{"content":"lo"}}]}`,
		`data:{"choices":[{"delta":{"content":"!"}}]}`,
		`data: [DONE]`,
		``,
	}, "\n")

	var deltas []string
	got, err := consumeCompletionStream(strings.NewReader(body), func(delta string) error {
		deltas = append(deltas, delta)
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello!" {
		t.Fatalf("full=%q want %q", got, "Hello!")
	}
	if strings.Join(deltas, "") != "Hello!" || len(deltas) != 3 {
		t.Fatalf("deltas=%v want [Hel lo !]", deltas)
	}
}

func TestConsumeCompletionStreamStopsOnDeltaError(t *testing.T) {
	t.Parallel()

	body := strings.Join([]string{
		`data: {"choices":[{"delta":{"content":"A"}}]}`,
		`data: {"choices":[{"delta":{"content":"B"}}]}`,
		`data: [DONE]`,
	}, "\n")

	stop := errors.New("stop")
	got, err := consumeCompletionStream(strings.NewReader(body), func(delta string) error {
		if delta == "B" {
			return stop
		}
		return nil
	})
	if !errors.Is(err, stop) {
		t.Fatalf("err=%v want stop", err)
	}
	if got != "AB" {
		t.Fatalf("full=%q want AB (partial before callback error)", got)
	}
}

func TestConsumeCompletionStreamInvalidJSON(t *testing.T) {
	t.Parallel()

	_, err := consumeCompletionStream(strings.NewReader("data: {not-json}\n"), nil)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestIsRetryable(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{name: "429", err: &HTTPError{StatusCode: 429}, want: true},
		{name: "503", err: &HTTPError{StatusCode: 503}, want: true},
		{name: "408", err: &HTTPError{StatusCode: 408}, want: true},
		{name: "404", err: &HTTPError{StatusCode: 404}, want: false},
		{name: "401", err: &HTTPError{StatusCode: 401}, want: false},
		{name: "400", err: &HTTPError{StatusCode: 400}, want: false},
		{name: "network", err: context.DeadlineExceeded, want: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := IsRetryable(tc.err); got != tc.want {
				t.Fatalf("IsRetryable(%v)=%v want %v", tc.err, got, tc.want)
			}
		})
	}
}
