package fitfeatures

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
	queueFitFeatures           = "fit-features"
	queueActivityFeaturesReady = "activity.features.ready"
	routingStructureReady      = "activity.structure.ready"
	routingFeaturesReady       = "activity.features.ready"
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

	msgs, err := amqp.ConsumeQueue(ch, queueFitFeatures, "fit-features", false)
	if err != nil {
		return fmt.Errorf("failed to consume %s: %w", queueFitFeatures, err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("shutting down")
			return nil
		case d, ok := <-msgs:
			if !ok {
				return fmt.Errorf("queue %s: channel closed", queueFitFeatures)
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

	if err := amqp.DeclareQueue(ch, queueFitFeatures); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	if err := amqp.BindQueue(ch, queueFitFeatures, amqp.ExchangeMessages, routingStructureReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("queue %s bound with %s", queueFitFeatures, routingStructureReady)

	if err := amqp.DeclareQueue(ch, queueActivityFeaturesReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	if err := amqp.BindQueue(ch, queueActivityFeaturesReady, amqp.ExchangeMessages, routingFeaturesReady); err != nil {
		return fmt.Errorf("topology: %w", err)
	}
	log.Printf("queue %s bound with %s", queueActivityFeaturesReady, routingFeaturesReady)

	return nil
}
