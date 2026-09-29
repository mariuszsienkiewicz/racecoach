package fitchat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/racecoach/workers/internal/llm"
)

func TestHandleChatValidation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name       string
		body       string
		wantStatus int
		wantSubstr string
	}{
		{
			name:       "missing message",
			body:       `{"activityId":1,"userId":2,"message":"   ","storageBucket":"b","featuresObjectKey":"k"}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "message is required",
		},
		{
			name:       "bad activity id",
			body:       `{"activityId":0,"userId":2,"message":"hi","storageBucket":"b","featuresObjectKey":"k"}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "activityId",
		},
		{
			name:       "missing features key",
			body:       `{"activityId":1,"userId":2,"message":"hi","storageBucket":"b"}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "featuresObjectKey",
		},
		{
			name:       "invalid json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "invalid JSON",
		},
		{
			name:       "history invalid role",
			body:       `{"activityId":1,"userId":2,"message":"hi","storageBucket":"b","featuresObjectKey":"k","history":[{"role":"system","content":"x"}]}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "history role must be either athlete or coach",
		},
		{
			name:       "history empty content",
			body:       `{"activityId":1,"userId":2,"message":"hi","storageBucket":"b","featuresObjectKey":"k","history":[{"role":"athlete","content":"   "}]}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "history content is required",
		},
		{
			name:       "history too many turns",
			body:       historyBodyWithTurns(maxHistoryLen + 1),
			wantStatus: http.StatusBadRequest,
			wantSubstr: "history is too long",
		},
		{
			name:       "history content too long",
			body:       `{"activityId":1,"userId":2,"message":"hi","storageBucket":"b","featuresObjectKey":"k","history":[{"role":"coach","content":"` + strings.Repeat("a", maxChatMessageLen+1) + `"}]}`,
			wantStatus: http.StatusBadRequest,
			wantSubstr: "history content is too long",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req := httptest.NewRequest(http.MethodPost, "/v1/chat", strings.NewReader(tc.body))
			rec := httptest.NewRecorder()
			if err := handleChat(req.Context(), rec, req, nil, nil); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if rec.Code != tc.wantStatus {
				t.Fatalf("status=%d want %d body=%q", rec.Code, tc.wantStatus, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), tc.wantSubstr) {
				t.Fatalf("body=%q want substr %q", rec.Body.String(), tc.wantSubstr)
			}
		})
	}
}

func TestNormalizeChatRequestTrimsHistoryInPlace(t *testing.T) {
	t.Parallel()

	req := chatRequest{
		Message:           "  hi  ",
		StorageBucket:     "  bucket  ",
		FeaturesObjectKey: "  key  ",
		Summary:           "  note  ",
		History: []historyTurn{
			{Role: " Athlete ", Content: "  earlier  "},
			{Role: "COACH", Content: "  reply  "},
		},
	}

	normalizeChatRequest(&req)

	if req.Message != "hi" || req.StorageBucket != "bucket" || req.FeaturesObjectKey != "key" || req.Summary != "note" {
		t.Fatalf("scalar trim failed: %+v", req)
	}
	if req.History[0].Role != "athlete" || req.History[0].Content != "earlier" {
		t.Fatalf("history[0]=%+v", req.History[0])
	}
	if req.History[1].Role != "coach" || req.History[1].Content != "reply" {
		t.Fatalf("history[1]=%+v", req.History[1])
	}
}

func TestValidateChatRequestAcceptsNormalizedHistory(t *testing.T) {
	t.Parallel()

	req := chatRequest{
		ActivityID:        1,
		UserID:            2,
		Message:           "follow up",
		StorageBucket:     "b",
		FeaturesObjectKey: "k",
		History: []historyTurn{
			{Role: "athlete", Content: "first"},
			{Role: "coach", Content: "answer"},
		},
	}
	if err := validateChatRequest(req); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBuildSessionContextExcludesQuestion(t *testing.T) {
	t.Parallel()

	ctx := buildSessionContext("Solid long run.", `{"signals":{}}`)
	if !strings.Contains(ctx, "Solid long run.") {
		t.Fatalf("missing summary: %q", ctx)
	}
	if !strings.Contains(ctx, `{"signals":{}}`) {
		t.Fatalf("missing features: %q", ctx)
	}
	if strings.Contains(ctx, "Athlete question") || strings.Contains(strings.ToLower(ctx), "how was my pace") {
		t.Fatalf("session context must not include the question: %q", ctx)
	}
}

func TestBuildChatMessages(t *testing.T) {
	t.Parallel()

	t.Run("no history", func(t *testing.T) {
		t.Parallel()
		msgs := buildChatMessages("SYS", "CTX", nil, "What about recovery?")
		if len(msgs) != 3 {
			t.Fatalf("len=%d want 3: %+v", len(msgs), msgs)
		}
		assertMsg(t, msgs[0], "system", "SYS")
		assertMsg(t, msgs[1], "user", "CTX")
		assertMsg(t, msgs[2], "user", "What about recovery?")
	})

	t.Run("with history", func(t *testing.T) {
		t.Parallel()
		history := []historyTurn{
			{Role: "athlete", Content: "How was the pace?"},
			{Role: "coach", Content: "Mostly steady."},
		}
		msgs := buildChatMessages("SYS", "CTX", history, "And the last km?")
		if len(msgs) != 5 {
			t.Fatalf("len=%d want 5: %+v", len(msgs), msgs)
		}
		assertMsg(t, msgs[0], "system", "SYS")
		assertMsg(t, msgs[1], "user", "CTX")
		assertMsg(t, msgs[2], "user", "How was the pace?")
		assertMsg(t, msgs[3], "assistant", "Mostly steady.")
		assertMsg(t, msgs[4], "user", "And the last km?")
	})

	t.Run("question appears only once at the end", func(t *testing.T) {
		t.Parallel()
		question := "Should I push tomorrow?"
		ctx := buildSessionContext("(none)", `{"ok":true}`)
		msgs := buildChatMessages(systemPrompt, ctx, nil, question)

		count := 0
		for _, m := range msgs {
			if m.Content == question {
				count++
			}
			if strings.Contains(m.Content, question) && m.Content != question {
				t.Fatalf("question leaked into another message: %+v", m)
			}
		}
		if count != 1 {
			t.Fatalf("question occurrences=%d want 1", count)
		}
		if msgs[len(msgs)-1].Content != question {
			t.Fatalf("last message want question, got %+v", msgs[len(msgs)-1])
		}
	})
}

func assertMsg(t *testing.T, got llm.Message, role, content string) {
	t.Helper()
	if got.Role != role || got.Content != content {
		t.Fatalf("got role=%q content=%q want role=%q content=%q", got.Role, got.Content, role, content)
	}
}

func historyBodyWithTurns(n int) string {
	var b strings.Builder
	b.WriteString(`{"activityId":1,"userId":2,"message":"hi","storageBucket":"b","featuresObjectKey":"k","history":[`)
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		role := "athlete"
		if i%2 == 1 {
			role = "coach"
		}
		b.WriteString(`{"role":"` + role + `","content":"turn"}`)
	}
	b.WriteString(`]}`)
	return b.String()
}
