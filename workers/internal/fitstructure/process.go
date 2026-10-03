package fitstructure

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
	"github.com/racecoach/workers/internal/fit"
	"github.com/racecoach/workers/internal/storage"
)

func processDelivery(ctx context.Context, ch *amqpClient.Channel, storageClient *storage.Client, apiClient *api.Client, d amqpClient.Delivery) error {
	var ev domain.ActivityMetricsReady
	if err := json.Unmarshal(d.Body, &ev); err != nil {
		amqp.Reject(d)
		return fmt.Errorf("unmarshal: %w", err)
	}
	log.Printf(
		"structure-pending activity=%d user=%d fit=%s metrics=%s distanceM=%d",
		ev.ActivityID, ev.UserID, ev.ObjectKey, ev.MetricsObjectKey, ev.DistanceM,
	)

	structureKey := strings.TrimSuffix(ev.ObjectKey, ".fit") + ".structure.json"

	activity, err := apiClient.GetActivity(ctx, ev.ActivityID)
	if err == nil && isStructureDone(*activity, structureKey) {
		log.Printf("activity=%d structure already done, short-circuiting", ev.ActivityID)
		if err := publishStructureReady(ctx, ch, ev, structureKey); err != nil {
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

	body, err := storageClient.GetObject(ctx, ev.StorageBucket, ev.ObjectKey)
	if err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("get object: %w", err))
	}

	parsed, err := fit.ParseStructure(body)
	if err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("parse structure: %w", err))
	}

	log.Printf("activity=%d laps=%d", ev.ActivityID, len(parsed.Laps))
	for _, lap := range parsed.Laps {
		log.Printf(
			"  lap[%d] distanceM=%d durationSec=%d pace=%v avgHR=%v maxHR=%v trigger=%s",
			lap.Index, lap.DistanceM, lap.DurationSec, derefInt(lap.AvgPaceSecPerKm), derefInt(lap.AvgHeartRate), derefInt(lap.MaxHeartRate), lap.LapTrigger,
		)
	}
	if len(parsed.Laps) > 0 && parsed.Laps[0].LapTrigger == "synthetic_km" {
		log.Printf("activity=%d using synthetic 1km splits (device laps were empty or single-lap)", ev.ActivityID)
	}

	artifact := domain.ActivityStructure{
		ActivityID:    ev.ActivityID,
		SchemaVersion: 2,
		ObjectKey:     ev.ObjectKey,
		Laps:          parsed.Laps,
	}

	structureBody, err := json.Marshal(artifact)
	if err != nil {
		amqp.Reject(d)
		return fmt.Errorf("marshal structure: %w", err)
	}

	if err := storageClient.PutObject(ctx, ev.StorageBucket, structureKey, structureBody, "application/json"); err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, fmt.Errorf("put structure object: %w", err))
	}
	log.Printf("uploaded structure to %s (%d bytes)", structureKey, len(structureBody))

	if err := apiClient.PatchActivityStructure(ctx, ev.ActivityID, domain.PatchActivityStructureRequest{StructureObjectKey: structureKey}); err != nil {
		return amqp.MaybeRequeue(ctx, d, api.IsRetryable(err), fmt.Errorf("patch activity structure: %w", err))
	}
	log.Printf("patched activity structure for activity=%d", ev.ActivityID)

	if err := publishStructureReady(ctx, ch, ev, structureKey); err != nil {
		return amqp.RequeueAfterBackoff(ctx, d, err)
	}

	if err := d.Ack(false); err != nil {
		return fmt.Errorf("ack: %w", err)
	}

	return nil
}

func publishStructureReady(ctx context.Context, ch *amqpClient.Channel, ev domain.ActivityMetricsReady, structureKey string) error {
	asr := domain.ActivityStructureReady{
		ActivityID:         ev.ActivityID,
		UserID:             ev.UserID,
		StorageBucket:      ev.StorageBucket,
		ObjectKey:          ev.ObjectKey,
		StructureObjectKey: structureKey,
	}
	asrBody, err := json.Marshal(asr)
	if err != nil {
		return fmt.Errorf("marshal activity structure ready: %w", err)
	}
	if err := amqp.PublishMessage(ctx, ch, amqp.ExchangeMessages, routingStructureReady, "application/json", asrBody); err != nil {
		return fmt.Errorf("publish message: %w", err)
	}
	log.Printf("published %s for activity=%d", routingStructureReady, asr.ActivityID)
	return nil
}

func isStructureDone(a domain.Activity, expectedKey string) bool {
	if a.StructureObjectKey == "" {
		return false
	}
	if a.StructureObjectKey != expectedKey {
		return false
	}
	return a.Status == "analyzing" || a.Status == "ready"
}

func derefInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}
