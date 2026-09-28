package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/url"
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
