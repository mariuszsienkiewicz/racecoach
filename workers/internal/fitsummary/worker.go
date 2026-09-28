package fitsummary

import (
	"context"
	"fmt"
	"log"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/racecoach/workers/internal/amqp"
	"github.com/racecoach/workers/internal/api"
	"github.com/racecoach/workers/internal/config"
	"github.com/racecoach/workers/internal/llm"
	"github.com/racecoach/workers/internal/storage"
)

const (
	queueFitSummary           = "fit-summary"
	queueActivitySummaryReady = "activity.summary.ready"
	routingFeaturesReady      = "activity.features.ready"
	routingSummaryReady       = "activity.summary.ready"

	// retry queue and dlq for fit summary
	queueFitSummaryRetry = "fit-summary.retry.30s"
	queueFitSummaryDLQ   = "fit-summary.dlq"

	// message ttl
	messageTTL = int32(30000)
	// maxSummaryRetries is the max value of x-retry-count on a work delivery.
	// First attempt has count=0; after this many delayed redeliveries we DLQ.
	maxSummaryRetries = int32(5)
)

func Run(ctx context.Context, cfg config.Config) error {
	log.Printf("config=%s %s", cfg.S3Bucket, cfg.S3Endpoint)

	conn, err := amqp.Dial(cfg.RabbitURL)
	if err != nil {
		return fmt.Errorf("failed to dial RabbitMQ: %w", err)
	}
	log.Println("connected to RabbitMQ")
	defer conn.Close()

	ch, err := amqp.OpenChannel(conn)
	if err != nil {
		return fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}
	log.Println("opened RabbitMQ channel")
	defer ch.Close()

	if err := setupTopology(ch); err != nil {
		return err
	}

	storageClient := storage.NewClient(cfg.S3Region, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Endpoint, cfg.S3UsePathStyle)
	apiClient := api.NewClient(cfg.APIBaseURL, cfg.APIToken)
	llmClient := llm.NewClient(cfg.LLMBaseURL, cfg.LLMModel, cfg.LLMAPIKey)

	// One delivery so LLM backoff actually pauses the consumer
	if err := amqp.SetQoS(ch, 1); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := amqp.ConsumeQueue(ch, queueFitSummary, "fit-summary", false)
	if err != nil {
		return fmt.Errorf("failed to consume %s: %w", queueFitSummary, err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("queue %s: channel closed", queueFitSummary)
			}
			if err := processDelivery(ctx, ch, storageClient, apiClient, llmClient, d); err != nil {
				log.Printf("process delivery: %v", err)
			}
		}
	}
}

func setupTopology(ch *amqp091.Channel) error {
	if err := amqp.DeclareExchange(ch, amqp.ExchangeMessages, "topic"); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("declared exchange %s (topic)", amqp.ExchangeMessages)

	if err := amqp.DeclareQueue(ch, queueFitSummary); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	if err := amqp.BindQueue(ch, queueFitSummary, amqp.ExchangeMessages, routingFeaturesReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("queue %s bound with %s", queueFitSummary, routingFeaturesReady)

	if err := amqp.DeclareQueue(ch, queueActivitySummaryReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	if err := amqp.BindQueue(ch, queueActivitySummaryReady, amqp.ExchangeMessages, routingSummaryReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("queue %s bound with %s", queueActivitySummaryReady, routingSummaryReady)

	// retry / dlq
	if err := amqp.DeclareQueueWithArgs(ch, queueFitSummaryRetry, amqp091.Table{
		"x-message-ttl":             messageTTL,
		"x-dead-letter-exchange":    amqp.ExchangeMessages,
		"x-dead-letter-routing-key": routingFeaturesReady,
	}); err != nil {
		return fmt.Errorf("topology declare retry: %w", err)
	}

	log.Printf("declared queue %s", queueFitSummaryRetry)
	if err := amqp.DeclareQueue(ch, queueFitSummaryDLQ); err != nil {
		return fmt.Errorf("topology declare dlq: %w", err)
	}
	log.Printf("declared queue %s", queueFitSummaryDLQ)

	return nil
}
