package fitsummary

import (
	"context"
	"fmt"
	"log"
	"maps"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/racecoach/workers/internal/amqp"
)

const headerRetryCount = "x-retry-count"

// scheduleRetry parks a transient failure on the TTL retry queue.
// After maxSummaryRetries delayed attempts the message goes to the DLQ.
func scheduleRetry(ctx context.Context, ch *amqp091.Channel, d amqp091.Delivery, cause error) error {
	count := headerInt(d.Headers, headerRetryCount)
	if count >= int(maxSummaryRetries) {
		return sendToDLQ(ctx, ch, d, cause)
	}

	next := count + 1
	headers := cloneHeaders(d.Headers)
	headers[headerRetryCount] = int32(next)

	// Default exchange: routing key == queue name.
	if err := amqp.PublishMessageWithHeaders(ctx, ch, "", queueFitSummaryRetry, "application/json", d.Body, headers); err != nil {
		return fmt.Errorf("schedule retry: %w", err)
	}
	if err := d.Ack(false); err != nil {
		return fmt.Errorf("ack after schedule retry: %w", err)
	}

	log.Printf("scheduled retry %d/%d delay=%dms: %v", next, maxSummaryRetries, messageTTL, cause)
	return nil
}

func sendToDLQ(ctx context.Context, ch *amqp091.Channel, d amqp091.Delivery, cause error) error {
	count := headerInt(d.Headers, headerRetryCount)
	headers := cloneHeaders(d.Headers)
	headers[headerRetryCount] = int32(count)

	if err := amqp.PublishMessageWithHeaders(ctx, ch, "", queueFitSummaryDLQ, "application/json", d.Body, headers); err != nil {
		return fmt.Errorf("send to DLQ: %w", err)
	}
	if err := d.Ack(false); err != nil {
		return fmt.Errorf("ack after DLQ: %w", err)
	}

	log.Printf("moved to DLQ after %d/%d retries: %v", count, maxSummaryRetries, cause)
	return nil
}

func cloneHeaders(src amqp091.Table) amqp091.Table {
	dst := make(amqp091.Table, len(src))
	maps.Copy(dst, src)
	return dst
}

// headerInt reads a numeric AMQP header without panicking on type variants.
func headerInt(headers amqp091.Table, key string) int {
	if headers == nil {
		return 0
	}
	raw, ok := headers[key]
	if !ok || raw == nil {
		return 0
	}
	switch n := raw.(type) {
	case int:
		return n
	case int8:
		return int(n)
	case int16:
		return int(n)
	case int32:
		return int(n)
	case int64:
		return int(n)
	case uint8:
		return int(n)
	case uint16:
		return int(n)
	case uint32:
		return int(n)
	case uint64:
		return int(n)
	default:
		return 0
	}
}
