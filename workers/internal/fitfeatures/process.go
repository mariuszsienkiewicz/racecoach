package fitfeatures

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	amqpClient "github.com/rabbitmq/amqp091-go"
	amqp "github.com/racecoach/workers/internal/amqp"
	"github.com/racecoach/workers/internal/api"
	"github.com/racecoach/workers/internal/domain"
	"github.com/racecoach/workers/internal/storage"
)

func processDelivery(ctx context.Context, ch *amqpClient.Channel, storageClient *storage.Client, apiClient *api.Client, d amqpClient.Delivery) error {
	var ev domain.ActivityStructureReady
	if err := json.Unmarshal(d.Body, &ev); err != nil {
		amqp.Reject(d)
		return fmt.Errorf("unmarshal: %w", err)
	}

	metricsKey := strings.TrimSuffix(ev.ObjectKey, ".fit") + ".metrics.json"
	structureKey := ev.StructureObjectKey
	if structureKey == "" {
		structureKey = strings.TrimSuffix(ev.ObjectKey, ".fit") + ".structure.json"
	}
	featuresKey := strings.TrimSuffix(ev.ObjectKey, ".fit") + ".features.json"

	log.Printf(
		"features-pending activity=%d fit=%s metrics=%s structure=%s",
		ev.ActivityID, ev.ObjectKey, metricsKey, structureKey,
	)

	activity, err := apiClient.GetActivity(ctx, ev.ActivityID)
	if err == nil && isFeaturesDone(*activity, featuresKey) {
		log.Printf("activity=%d features already done, short-circuiting", ev.ActivityID)
		if err := publishFeaturesReady(ctx, ch, ev, metricsKey, structureKey, featuresKey); err != nil {
			return amqp.RequeueAfterBackoff(ctx, d, err)
		}
		if err := d.Ack(false); err != nil {
			return fmt.Errorf("ack: %w", err)
		}
		return nil
	}
	if err != nil {
		return amqp.MaybeRequeue(ctx, d, api.IsRetryable(err), fmt.Errorf("get activity: %w", err))
	}

	metricsBody, err := storageClient.GetObject(ctx, ev.StorageBucket, metricsKey)
	if err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("get metrics object: %w", err))
	}
	var metrics domain.ActivityMetrics
	if err := json.Unmarshal(metricsBody, &metrics); err != nil {
		amqp.Reject(d)
		return fmt.Errorf("unmarshal metrics: %w", err)
	}
	if metrics.MetricsObjectKey == "" {
		metrics.MetricsObjectKey = metricsKey
	}

	structureBody, err := storageClient.GetObject(ctx, ev.StorageBucket, structureKey)
	if err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("get structure object: %w", err))
	}
	var structure domain.ActivityStructure
	if err := json.Unmarshal(structureBody, &structure); err != nil {
		amqp.Reject(d)
		return fmt.Errorf("unmarshal structure: %w", err)
	}

	features := Build(metrics, structure)
	features.StructureObjectKey = structureKey
	features.FeaturesObjectKey = featuresKey

	featuresBody, err := json.Marshal(features)
	if err != nil {
		amqp.Reject(d)
		return fmt.Errorf("marshal features: %w", err)
	}
	if err := storageClient.PutObject(ctx, ev.StorageBucket, featuresKey, featuresBody, "application/json"); err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("put features object: %w", err))
	}
	log.Printf(
		"uploaded features to %s (%d bytes) shape=%s laps=%d",
		featuresKey, len(featuresBody), features.Signals.SuspectedWorkoutShape, features.Overview.LapCount,
	)

	if err := apiClient.PatchActivityFeatures(ctx, ev.ActivityID, domain.PatchActivityFeaturesRequest{FeaturesObjectKey: featuresKey}); err != nil {
		return amqp.MaybeRequeue(ctx, d, api.IsRetryable(err), fmt.Errorf("patch activity features: %w", err))
	}
	log.Printf("patched activity features for activity=%d", ev.ActivityID)

	if err := publishFeaturesReady(ctx, ch, ev, metricsKey, structureKey, featuresKey); err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, err)
	}

	if err := d.Ack(false); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}

func publishFeaturesReady(ctx context.Context, ch *amqpClient.Channel, ev domain.ActivityStructureReady, metricsKey, structureKey, featuresKey string) error {
	afr := domain.ActivityFeaturesReady{
		ActivityID:         ev.ActivityID,
		UserID:             ev.UserID,
		StorageBucket:      ev.StorageBucket,
		ObjectKey:          ev.ObjectKey,
		MetricsObjectKey:   metricsKey,
		StructureObjectKey: structureKey,
		FeaturesObjectKey:  featuresKey,
	}
	body, err := json.Marshal(afr)
	if err != nil {
		return fmt.Errorf("marshal activity features ready: %w", err)
	}
	if err := amqp.PublishMessage(ctx, ch, amqp.ExchangeMessages, routingFeaturesReady, "application/json", body); err != nil {
		return fmt.Errorf("publish message: %w", err)
	}
	log.Printf("published %s for activity=%d", routingFeaturesReady, afr.ActivityID)
	return nil
}

func isFeaturesDone(a domain.Activity, expectedKey string) bool {
	if a.FeaturesObjectKey == "" {
		return false
	}
	if a.FeaturesObjectKey != expectedKey {
		return false
	}
	return a.Status == "analyzing" || a.Status == "ready"
}
