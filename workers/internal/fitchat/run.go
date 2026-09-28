package fitchat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/racecoach/workers/internal/coach"
	"github.com/racecoach/workers/internal/domain"
	"github.com/racecoach/workers/internal/llm"
	"github.com/racecoach/workers/internal/storage"
)

const maxChatMessageLen = 2000

// chatRequest is the private Symfony → fit-chat contract.
// Artifact keys come from the gateway so we never call Symfony back mid-request.
type chatRequest struct {
	ActivityID        int    `json:"activityId"`
	UserID            int    `json:"userId"`
	Message           string `json:"message"`
	StorageBucket     string `json:"storageBucket"`
	FeaturesObjectKey string `json:"featuresObjectKey"`
	SummaryObjectKey  string `json:"summaryObjectKey,omitempty"`
	Summary           string `json:"summary,omitempty"`
}

type chatResponse struct {
	Reply string `json:"reply"`
}

func handleChat(ctx context.Context, w http.ResponseWriter, r *http.Request, storageClient *storage.Client, llmClient *llm.Client) error {
	var request chatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("failed to decode request: %v", err)
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return nil
	}

	request.Message = strings.TrimSpace(request.Message)
	request.StorageBucket = strings.TrimSpace(request.StorageBucket)
	request.FeaturesObjectKey = strings.TrimSpace(request.FeaturesObjectKey)
	request.Summary = strings.TrimSpace(request.Summary)

	if err := validateChatRequest(request); err != nil {
		log.Printf("invalid chat request: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return nil
	}

	featuresBody, err := storageClient.GetObject(ctx, request.StorageBucket, request.FeaturesObjectKey)
	if err != nil {
		log.Printf("failed to get features object: %v", err)
		return fmt.Errorf("get features object: %w", err)
	}

	var features domain.ActivityFeatures
	if err := json.Unmarshal(featuresBody, &features); err != nil {
		log.Printf("failed to unmarshal features: %v", err)
		writeJSONError(w, http.StatusBadGateway, "invalid features artifact")
		return nil
	}

	coachFeatures := coach.ForCoachPrompt(features)
	featuresJSON, err := json.MarshalIndent(coachFeatures, "", "  ")
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return nil
	}

	summaryBlock := "(none)"
	if request.Summary != "" {
		summaryBlock = request.Summary
	}

	userPrompt := "Answer the athlete using ONLY the context below.\n\n" +
		"## Existing session summary (may be empty)\n" +
		summaryBlock + "\n\n" +
		"## Features JSON (coach-safe view)\n" +
		string(featuresJSON) + "\n\n" +
		"## Athlete question\n" +
		request.Message + "\n\n" +
		"Reply as RaceCoach in plain text."

	log.Printf("generating chat reply for activity=%d", request.ActivityID)

	raw, err := llmClient.CompletionRaw(ctx, llmClient.NewTextCompletionRequest(systemPrompt, userPrompt))
	if err != nil {
		return fmt.Errorf("chat completion: %w", err)
	}

	reply := strings.TrimSpace(raw)
	if reply == "" {
		writeJSONError(w, http.StatusBadGateway, "empty chat reply")
		return nil
	}

	return writeJSON(w, http.StatusOK, chatResponse{
		Reply: reply,
	})
}

func validateChatRequest(request chatRequest) error {
	if request.ActivityID <= 0 {
		return fmt.Errorf("activityId must be a positive integer")
	}
	if request.UserID <= 0 {
		return fmt.Errorf("userId must be a positive integer")
	}
	if request.Message == "" {
		return fmt.Errorf("message is required")
	}
	if utf8.RuneCountInString(request.Message) > maxChatMessageLen {
		return fmt.Errorf("message is too long (max %d characters)", maxChatMessageLen)
	}
	if request.StorageBucket == "" {
		return fmt.Errorf("storageBucket is required")
	}
	if request.FeaturesObjectKey == "" {
		return fmt.Errorf("featuresObjectKey is required")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("failed to encode response: %v", err)
		return fmt.Errorf("encode response: %w", err)
	}
	return nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	_ = writeJSON(w, status, map[string]string{"message": message})
}
