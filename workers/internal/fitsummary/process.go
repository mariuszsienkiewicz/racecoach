package fitsummary

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	amqpClient "github.com/rabbitmq/amqp091-go"
	amqp "github.com/racecoach/workers/internal/amqp"
	"github.com/racecoach/workers/internal/api"
	"github.com/racecoach/workers/internal/coach"
	"github.com/racecoach/workers/internal/domain"
	"github.com/racecoach/workers/internal/llm"
	"github.com/racecoach/workers/internal/storage"
)

const pipelineHeartbeatInterval = 30 * time.Second

func processDelivery(ctx context.Context, ch *amqpClient.Channel, storageClient *storage.Client, apiClient *api.Client, llmClient *llm.Client, d amqpClient.Delivery) error {
	// Guard against leftover/misconfigured binds which would otherwise republish forever on short-circuit
	if d.RoutingKey != "" && d.RoutingKey != routingFeaturesReady {
		log.Printf("ignoring delivery routingKey=%s (expected %s)", d.RoutingKey, routingFeaturesReady)
		_ = d.Ack(false)
		return nil
	}

	var ev domain.ActivityFeaturesReady
	if err := json.Unmarshal(d.Body, &ev); err != nil {
		amqp.Reject(d)
		return fmt.Errorf("unmarshal: %w", err)
	}

	summaryKey := strings.TrimSuffix(ev.ObjectKey, ".fit") + ".summary.json"
	featuresKey := ev.FeaturesObjectKey
	if featuresKey == "" {
		featuresKey = strings.TrimSuffix(ev.ObjectKey, ".fit") + ".features.json"
	}

	activity, err := apiClient.GetActivity(ctx, ev.ActivityID)
	if err == nil && isSummaryDone(*activity, summaryKey) {
		log.Printf("activity=%d summary already done, short-circuiting", ev.ActivityID)
		if err := publishSummaryReady(ctx, ch, ev, featuresKey, summaryKey); err != nil {
			// Publish failures are transport-level → always delayed retry.
			return scheduleRetry(ctx, ch, d, err)
		}
		if err := d.Ack(false); err != nil {
			return fmt.Errorf("ack: %w", err)
		}
		return nil
	}
	if err != nil {
		err = fmt.Errorf("get activity: %w", err)
		if !api.IsRetryable(err) {
			amqp.Reject(d)
			return err
		}
		return scheduleRetry(ctx, ch, d, err)
	}

	featuresBody, err := storageClient.GetObject(ctx, ev.StorageBucket, featuresKey)
	if err != nil {
		return scheduleRetry(ctx, ch, d, fmt.Errorf("get features object: %w", err))
	}
	var features domain.ActivityFeatures
	if err := json.Unmarshal(featuresBody, &features); err != nil {
		amqp.Reject(d)
		return fmt.Errorf("unmarshal features: %w", err)
	}

	log.Printf("generating summary for activity=%d", ev.ActivityID)

	coachFeatures := coach.ForCoachPrompt(features)
	featuresJSON, err := json.MarshalIndent(coachFeatures, "", "  ")
	if err != nil {
		amqp.Reject(d)
		return fmt.Errorf("marshal features: %w", err)
	}
	userPrompt := "Return ONLY one JSON object matching the schema from the system prompt. No prose before or after it.\n\n" +
		"Prefer signals.display.effort, signals.display.dominantHrZone, signals.intervalPattern, and signals.byIntensity.active.\n" +
		"Use signals.display.athleteZones only to interpret intensity - never paste the full zone table into athlete-facing text.\n" +
		"Never treat recovery/warmup/cooldown pace as slowest work lap.\n\n" +
		"Features JSON:\n" + string(featuresJSON)

	stopHeartbeat := startPipelineHeartbeat(ctx, apiClient, ev.ActivityID)
	defer stopHeartbeat()

	raw, err := llmClient.Completion(ctx, llmClient.NewCompletionRequest(systemPrompt, userPrompt))
	if err != nil {
		err = fmt.Errorf("generate summary: %w", err)
		if !llm.IsRetryable(err) {
			return sendToDLQ(ctx, ch, d, err)
		}
		return scheduleRetry(ctx, ch, d, err)
	}

	var coach domain.CoachSummaryPayload
	if err := json.Unmarshal([]byte(raw), &coach); err != nil {
		preview := raw
		if len(preview) > 280 {
			preview = preview[:280] + "..."
		}
		log.Printf("activity=%d invalid LLM JSON preview=%q", ev.ActivityID, preview)
		return scheduleRetry(ctx, ch, d, fmt.Errorf("unmarshal summary prompt response JSON: %w", err))
	}
	if strings.TrimSpace(coach.Summary) == "" {
		return scheduleRetry(ctx, ch, d, fmt.Errorf("empty summary from LLM"))
	}

	artifact := domain.ActivitySummary{
		ActivityID:        ev.ActivityID,
		SchemaVersion:     1,
		ObjectKey:         ev.ObjectKey,
		FeaturesObjectKey: featuresKey,
		SummaryObjectKey:  summaryKey,
		Headline:          coach.Headline,
		Summary:           coach.Summary,
		Highlights:        coach.Highlights,
		Watchouts:         coach.Watchouts,
		NextFocus:         coach.NextFocus,
	}
	summaryBody, err := json.Marshal(artifact)
	if err != nil {
		amqp.Reject(d)
		return fmt.Errorf("marshal summary: %w", err)
	}

	if err := storageClient.PutObject(ctx, ev.StorageBucket, summaryKey, summaryBody, "application/json"); err != nil {
		return scheduleRetry(ctx, ch, d, fmt.Errorf("put summary object: %w", err))
	}
	log.Printf("uploaded summary to %s (%d bytes)", summaryKey, len(summaryBody))

	uiSummary := coach.Summary
	if strings.TrimSpace(coach.Headline) != "" {
		uiSummary = coach.Headline + "\n\n" + coach.Summary
	}
	if err := apiClient.PatchActivitySummary(ctx, ev.ActivityID, domain.PatchActivitySummaryRequest{
		SummaryObjectKey: summaryKey,
		Summary:          uiSummary,
	}); err != nil {
		err = fmt.Errorf("patch activity summary: %w", err)
		if !api.IsRetryable(err) {
			amqp.Reject(d)
			return err
		}
		return scheduleRetry(ctx, ch, d, err)
	}
	log.Printf("patched activity summary for activity=%d (ready)", ev.ActivityID)

	if err := publishSummaryReady(ctx, ch, ev, featuresKey, summaryKey); err != nil {
		return scheduleRetry(ctx, ch, d, err)
	}

	if err := d.Ack(false); err != nil {
		return fmt.Errorf("ack: %w", err)
	}
	return nil
}

func publishSummaryReady(ctx context.Context, ch *amqpClient.Channel, ev domain.ActivityFeaturesReady, featuresKey, summaryKey string) error {
	asr := domain.ActivitySummaryReady{
		ActivityID:        ev.ActivityID,
		UserID:            ev.UserID,
		StorageBucket:     ev.StorageBucket,
		ObjectKey:         ev.ObjectKey,
		FeaturesObjectKey: featuresKey,
		SummaryObjectKey:  summaryKey,
	}
	body, err := json.Marshal(asr)
	if err != nil {
		return fmt.Errorf("marshal activity summary ready: %w", err)
	}
	if err := amqp.PublishMessage(ctx, ch, amqp.ExchangeMessages, routingSummaryReady, "application/json", body); err != nil {
		return fmt.Errorf("publish message: %w", err)
	}
	log.Printf("published %s for activity=%d", routingSummaryReady, asr.ActivityID)
	return nil
}

func isSummaryDone(a domain.Activity, expectedKey string) bool {
	if a.SummaryObjectKey == "" {
		return false
	}
	if a.SummaryObjectKey != expectedKey {
		return false
	}
	return a.Status == "ready"
}

func startPipelineHeartbeat(ctx context.Context, apiClient *api.Client, activityID int) func() {
	hbCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		tick := func() {
			beatCtx, beatCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer beatCancel()
			if err := apiClient.PatchPipelineHeartbeat(beatCtx, activityID); err != nil {
				log.Printf("pipeline heartbeat failed activity=%d: %v", activityID, err)
			}
		}
		tick()
		ticker := time.NewTicker(pipelineHeartbeatInterval)
		defer ticker.Stop()
		for {
			select {
			case <-hbCtx.Done():
				return
			case <-ticker.C:
				tick()
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}
