package fitchat

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
