package fitmetrics

import (
	"context"
	"fmt"
	"log"

	amqp091 "github.com/rabbitmq/amqp091-go"
	"github.com/racecoach/workers/internal/amqp"
	"github.com/racecoach/workers/internal/api"
	"github.com/racecoach/workers/internal/config"
	"github.com/racecoach/workers/internal/storage"
)

const (
	queueFitMetrics           = "fit-metrics"
	queueActivityMetricsReady = "activity.metrics.ready"
	routingActivityUploaded   = "activity.uploaded"
	routingMetricsReady       = "activity.metrics.ready"
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

	if err := amqp.SetQoS(ch, 1); err != nil {
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	msgs, err := amqp.ConsumeQueue(ch, queueFitMetrics, "fit-metrics", false)
	if err != nil {
		return fmt.Errorf("failed to consume %s: %w", queueFitMetrics, err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("queue %s: channel closed", queueFitMetrics)
			}
			if err := processDelivery(ctx, ch, storageClient, apiClient, d); err != nil {
				return fmt.Errorf("process delivery: %w", err)
			}
		}
	}
}

func setupTopology(ch *amqp091.Channel) error {
	if err := amqp.DeclareExchange(ch, amqp.ExchangeMessages, "topic"); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("declared exchange %s (topic)", amqp.ExchangeMessages)

	if err := amqp.DeclareQueue(ch, queueFitMetrics); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	if err := amqp.BindQueue(ch, queueFitMetrics, amqp.ExchangeMessages, routingActivityUploaded); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("queue %s bound with %s", queueFitMetrics, routingActivityUploaded)

	if err := amqp.DeclareQueue(ch, queueActivityMetricsReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	if err := amqp.BindQueue(ch, queueActivityMetricsReady, amqp.ExchangeMessages, routingMetricsReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("queue %s bound with %s", queueActivityMetricsReady, routingMetricsReady)

	return nil
}
