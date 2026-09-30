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
const maxHistoryLen = 12

// chatRequest is the private Symfony to fit-chat contract.
type chatRequest struct {
	ActivityID        int           `json:"activityId"`
	UserID            int           `json:"userId"`
	Message           string        `json:"message"`
	StorageBucket     string        `json:"storageBucket"`
	FeaturesObjectKey string        `json:"featuresObjectKey"`
	SummaryObjectKey  string        `json:"summaryObjectKey,omitempty"`
	Summary           string        `json:"summary,omitempty"`
	History           []historyTurn `json:"history,omitempty"`
}

type historyTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type sseEvent struct {
	Event string
	Data  string
}

type tokenEventData struct {
	Text string `json:"text"`
}

type doneEventData struct {
	Reply string `json:"reply"`
}

type errorEventData struct {
	Message string `json:"message"`
}

func handleChat(ctx context.Context, w http.ResponseWriter, r *http.Request, storageClient *storage.Client, llmClient *llm.Client) error {
	var request chatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		log.Printf("failed to decode request: %v", err)
		writeJSONError(w, http.StatusBadRequest, "invalid JSON body")
		return nil
	}

	normalizeChatRequest(&request)

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

	sessionContext := buildSessionContext(summaryBlock, string(featuresJSON))
	messages := buildChatMessages(systemPrompt, sessionContext, request.History, request.Message)

	log.Printf("generating chat reply for activity=%d", request.ActivityID)

	rc := http.NewResponseController(w)
	startSSE(w)

	// From here the HTTP status is committed as 200 and event-stream.
	// Failures must be SSE event:error (return nil), never a second JSON status from Serve.
	raw, streamErr := llmClient.CompletionStream(
		ctx,
		llmClient.NewChatCompletionRequestStream(messages),
		func(token string) error {
			return writeSSE(rc, w, tokenEvent(token))
		},
	)
	if streamErr != nil {
		log.Printf("chat stream failed activity=%d: %v", request.ActivityID, streamErr)
		_ = writeSSE(rc, w, errorEvent(streamErr.Error()))
		return nil
	}

	reply := strings.TrimSpace(raw)
	if reply == "" {
		log.Printf("chat stream empty reply activity=%d", request.ActivityID)
		_ = writeSSE(rc, w, errorEvent("empty chat reply"))
		return nil
	}

	if err := writeSSE(rc, w, doneEvent(reply)); err != nil {
		log.Printf("failed to write done event activity=%d: %v", request.ActivityID, err)
		return nil
	}
	return nil
}

func startSSE(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
}

func writeSSE(rc *http.ResponseController, w http.ResponseWriter, event sseEvent) error {
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, event.Data); err != nil {
		return fmt.Errorf("write sse: %w", err)
	}
	if err := rc.Flush(); err != nil {
		return fmt.Errorf("flush sse: %w", err)
	}
	return nil
}

func tokenEvent(text string) sseEvent {
	return mustSSEEvent("token", tokenEventData{Text: text})
}

func doneEvent(reply string) sseEvent {
	return mustSSEEvent("done", doneEventData{Reply: reply})
}

func errorEvent(message string) sseEvent {
	return mustSSEEvent("error", errorEventData{Message: message})
}

func mustSSEEvent(event string, payload any) sseEvent {
	data, err := json.Marshal(payload)
	if err != nil {
		panic(fmt.Sprintf("marshal sse %s payload: %v", event, err))
	}
	return sseEvent{Event: event, Data: string(data)}
}

func normalizeChatRequest(request *chatRequest) {
	request.Message = strings.TrimSpace(request.Message)
	request.StorageBucket = strings.TrimSpace(request.StorageBucket)
	request.FeaturesObjectKey = strings.TrimSpace(request.FeaturesObjectKey)
	request.Summary = strings.TrimSpace(request.Summary)
	for i := range request.History {
		request.History[i].Role = strings.ToLower(strings.TrimSpace(request.History[i].Role))
		request.History[i].Content = strings.TrimSpace(request.History[i].Content)
	}
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
	if len(request.History) > maxHistoryLen {
		return fmt.Errorf("history is too long (max %d messages)", maxHistoryLen)
	}
	for _, turn := range request.History {
		if turn.Role == "" {
			return fmt.Errorf("history role is required")
		}
		if turn.Role != "athlete" && turn.Role != "coach" {
			return fmt.Errorf("history role must be either athlete or coach")
		}
		if turn.Content == "" {
			return fmt.Errorf("history content is required")
		}
		if utf8.RuneCountInString(turn.Content) > maxChatMessageLen {
			return fmt.Errorf("history content is too long (max %d characters)", maxChatMessageLen)
		}
	}
	return nil
}

func buildSessionContext(summary string, featuresJSON string) string {
	return "Session context for this activity (facts only; do not invent beyond this).\n\n" +
		"## Existing session summary\n" +
		summary + "\n\n" +
		"## Features JSON (coach-safe view)\n" +
		featuresJSON
}

func buildChatMessages(systemPrompt string, sessionContext string, history []historyTurn, question string) []llm.Message {
	messages := []llm.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: sessionContext},
	}
	for _, turn := range history {
		messages = append(messages, llm.Message{Role: llmRole(turn.Role), Content: turn.Content})
	}
	messages = append(messages, llm.Message{Role: "user", Content: question})
	return messages
}

func llmRole(role string) string {
	if role == "coach" {
		return "assistant"
	}
	return "user"
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
