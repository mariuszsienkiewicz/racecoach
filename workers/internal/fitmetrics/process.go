package fitmetrics

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	amqpClient "github.com/rabbitmq/amqp091-go"
	amqp "github.com/racecoach/workers/internal/amqp"
	"github.com/racecoach/workers/internal/api"
	"github.com/racecoach/workers/internal/domain"
	"github.com/racecoach/workers/internal/fit"
	"github.com/racecoach/workers/internal/storage"
)

func processDelivery(ctx context.Context, ch *amqpClient.Channel, storageClient *storage.Client, apiClient *api.Client, d amqpClient.Delivery) error {
	var ap domain.ActivityUploaded
	if err := json.Unmarshal(d.Body, &ap); err != nil {
		amqp.Reject(d)
		return fmt.Errorf("unmarshal: %w", err)
	}
	log.Printf("activity=%d bucket=%s key=%s", ap.ActivityID, ap.StorageBucket, ap.ObjectKey)

	activity, err := apiClient.GetActivity(ctx, ap.ActivityID)

	// check if metrics are already done
	expectedKey := strings.TrimSuffix(ap.ObjectKey, ".fit") + ".metrics.json"
	if err == nil && isMetricsDone(*activity, expectedKey) {
		log.Printf("activity=%d metrics already done, short-circuiting", ap.ActivityID)
		amr := domain.ActivityMetricsReady{
			ActivityID:       ap.ActivityID,
			UserID:           ap.UserID,
			StorageBucket:    ap.StorageBucket,
			ObjectKey:        ap.ObjectKey,
			MetricsObjectKey: expectedKey,
			DistanceM:        activity.DistanceM,
			DurationSec:      activity.DurationSec,
			AvgHeartRate:     &activity.AvgHeartRate,
			ReadyAt:          time.Now().UTC().Format(time.RFC3339),
		}
		amrBody, err := json.Marshal(amr)
		if err != nil {
			amqp.Reject(d)
			return fmt.Errorf("marshal activity metrics ready: %w", err)
		}
		err = amqp.PublishMessage(ctx, ch, amqp.ExchangeMessages, routingMetricsReady, "application/json", amrBody)
		if err != nil {
			return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("publish message: %w", err))
		}
		log.Printf("published %s for activity=%d", routingMetricsReady, ap.ActivityID)

		if err := d.Ack(false); err != nil {
			return fmt.Errorf("ack: %w", err)
		}

		return nil
	}

	if err != nil {
		return amqp.MaybeRequeue(ctx, d, api.IsRetryable(err), fmt.Errorf("get activity: %w", err))
	}

	body, err := storageClient.GetObject(ctx, ap.StorageBucket, ap.ObjectKey)
	if err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("get object: %w", err))
	}
	log.Printf("downloaded %d bytes (event says %d)", len(body), ap.FileSizeBytes)

	checksum := sha256.Sum256(body)
	if hex.EncodeToString(checksum[:]) != ap.ChecksumSHA256 {
		amqp.Reject(d)
		return fmt.Errorf("checksum mismatch: event says %s, downloaded %s", ap.ChecksumSHA256, hex.EncodeToString(checksum[:]))
	}
	log.Println("checksum matches")

	parsed, err := fit.ParseMetrics(body)
	if err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("parse metrics: %w", err))
	}

	artifact := domain.ActivityMetrics{
		ActivityID:   ap.ActivityID,
		DistanceM:    parsed.DistanceM,
		DurationSec:  parsed.DurationSec,
		AvgHeartRate: parsed.AvgHeartRate,
		MaxHeartRate: parsed.MaxHeartRate,
	}
	log.Printf("metrics: activity=%d distanceM=%d durationSec=%d avgHR=%v maxHR=%v",
		artifact.ActivityID, artifact.DistanceM, artifact.DurationSec, artifact.AvgHeartRate, artifact.MaxHeartRate)

	metricsKey := strings.TrimSuffix(ap.ObjectKey, ".fit") + ".metrics.json"
	artifact.MetricsObjectKey = metricsKey

	metricsBody, err := json.Marshal(artifact)
	if err != nil {
		amqp.Reject(d)
		return fmt.Errorf("marshal metrics: %w", err)
	}
	if err := storageClient.PutObject(ctx, ap.StorageBucket, metricsKey, metricsBody, "application/json"); err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("put metrics object: %w", err))
	}
	log.Printf("uploaded metrics to %s (%d bytes)", metricsKey, len(metricsBody))

	if err := apiClient.PatchActivityMetrics(ctx, ap.ActivityID, artifact); err != nil {
		return amqp.MaybeRequeue(ctx, d, api.IsRetryable(err), fmt.Errorf("patch activity metrics: %w", err))
	}
	log.Printf("patched activity metrics for activity=%d", ap.ActivityID)

	amr := domain.ActivityMetricsReady{
		ActivityID:       ap.ActivityID,
		UserID:           ap.UserID,
		StorageBucket:    ap.StorageBucket,
		ObjectKey:        ap.ObjectKey,
		MetricsObjectKey: metricsKey,
		DistanceM:        artifact.DistanceM,
		DurationSec:      artifact.DurationSec,
		AvgHeartRate:     artifact.AvgHeartRate,
		ReadyAt:          time.Now().UTC().Format(time.RFC3339),
	}
	amrBody, err := json.Marshal(amr)
	if err != nil {
		amqp.Reject(d)
		return fmt.Errorf("marshal activity metrics ready: %w", err)
	}
	err = amqp.PublishMessage(ctx, ch, amqp.ExchangeMessages, routingMetricsReady, "application/json", amrBody)
	if err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("publish message: %w", err))
	}
	log.Printf("published %s for activity=%d", routingMetricsReady, ap.ActivityID)

	if err := d.Ack(false); err != nil {
		return fmt.Errorf("ack: %w", err)
	}

	return nil
}

func isMetricsDone(a domain.Activity, expectedKey string) bool {
	if a.MetricsObjectKey == "" {
		return false
	}
	if a.MetricsObjectKey != expectedKey {
		return false
	}

	return a.Status == "analyzing" || a.Status == "ready"
}
