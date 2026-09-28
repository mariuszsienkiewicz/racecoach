package amqp

import (
	"context"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RequeueBackoff pauses before Nack(requeue=true) so a down dependency
// (API, MinIO, Ollama etc.) cannot hot-loop the queue.
const RequeueBackoff = 30 * time.Second

// Reject drops the message permanently (poison / non-retryable).
func Reject(d amqp.Delivery) {
	_ = d.Nack(false, false)
}

// RequeueAfterBackoff nacks with requeue after a fixed pause.
func RequeueAfterBackoff(ctx context.Context, d amqp.Delivery, cause error) error {
	log.Printf("backing off %s before requeue: %v", RequeueBackoff, cause)
	select {
	case <-ctx.Done():
		_ = d.Nack(false, true)
		return ctx.Err()
	case <-time.After(RequeueBackoff):
	}
	_ = d.Nack(false, true)
	return cause
}

// MaybeRequeue rejects when requeue is false; otherwise backs off then requeues.
func MaybeRequeue(ctx context.Context, d amqp.Delivery, requeue bool, cause error) error {
	if !requeue {
		Reject(d)
		return cause
	}
	return RequeueAfterBackoff(ctx, d, cause)
}
