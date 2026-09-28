package fitsummary

import (
	"testing"

	amqp091 "github.com/rabbitmq/amqp091-go"
)

func TestHeaderInt(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		val  any
		want int
	}{
		{name: "missing", val: nil, want: 0},
		{name: "int32", val: int32(3), want: 3},
		{name: "int64", val: int64(4), want: 4},
		{name: "int", val: 5, want: 5},
		{name: "string ignored", val: "nope", want: 0},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var headers amqp091.Table
			if tc.val != nil {
				headers = amqp091.Table{headerRetryCount: tc.val}
			}
			if got := headerInt(headers, headerRetryCount); got != tc.want {
				t.Fatalf("got %d want %d", got, tc.want)
			}
		})
	}

	if got := headerInt(nil, headerRetryCount); got != 0 {
		t.Fatalf("nil headers: got %d", got)
	}
}

func TestCloneHeadersDoesNotAlias(t *testing.T) {
	t.Parallel()

	src := amqp091.Table{headerRetryCount: int32(1), "x-other": "keep"}
	dst := cloneHeaders(src)
	dst[headerRetryCount] = int32(9)

	if src[headerRetryCount].(int32) != 1 {
		t.Fatalf("clone mutated source")
	}
	if dst["x-other"] != "keep" {
		t.Fatalf("lost unrelated header")
	}
}
