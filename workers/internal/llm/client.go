package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	model      string
	apiToken   string
}

type CompletionRequest struct {
	Model          string          `json:"model"`
	Messages       []Message       `json:"messages"`
	Stream         bool            `json:"stream"`
	Temperature    float64         `json:"temperature"`
	ResponseFormat *ResponseFormat `json:"response_format,omitempty"`
}

type ResponseFormat struct {
	Type string `json:"type"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CompletionResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

func NewClient(baseURL string, model string, apiToken string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 180 * time.Second},
		baseURL:    strings.TrimRight(baseURL, "/"),
		model:      model,
		apiToken:   apiToken,
	}
}

func (c *Client) NewCompletionRequest(systemPrompt string, userPrompt string) CompletionRequest {
	return CompletionRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream:         false,
		Temperature:    0.2,
		ResponseFormat: &ResponseFormat{Type: "json_object"},
	}
}

func (c *Client) NewTextCompletionRequest(systemPrompt string, userPrompt string) CompletionRequest {
	return CompletionRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Stream:      false,
		Temperature: 0.4,
	}
}

func (c *Client) NewChatCompletionRequest(messages []Message) CompletionRequest {
	return CompletionRequest{
		Model:       c.model,
		Messages:    messages,
		Stream:      false,
		Temperature: 0.4,
	}
}

func (c *Client) Completion(ctx context.Context, request CompletionRequest) (string, error) {
	content, err := c.CompletionRaw(ctx, request)
	if err != nil {
		return "", err
	}
	return ExtractJSONObject(content), nil
}

func (c *Client) CompletionRaw(ctx context.Context, request CompletionRequest) (string, error) {
	body, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("marshal: %w", err)
	}

	url := c.baseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var completionResponse CompletionResponse
	if err := json.Unmarshal(respBody, &completionResponse); err != nil {
		return "", fmt.Errorf("unmarshal: %w", err)
	}
	if len(completionResponse.Choices) == 0 {
		return "", fmt.Errorf("no choices in completion response")
	}
	content := completionResponse.Choices[0].Message.Content
	return content, nil
}

// ExtractJSONObject strips markdown fences and keeps the first JSON object if the model wrapped it in prose (e.g. "Here is the JSON:\n{...}")
func ExtractJSONObject(s string) string {
	s = stripMarkdownFence(s)
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return strings.TrimSpace(s[start : end+1])
	}
	return strings.TrimSpace(s)
}

func stripMarkdownFence(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "```") {
		return s
	}
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(s), "json") {
		s = strings.TrimSpace(s[4:])
	}
	if i := strings.LastIndex(s, "```"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// HTTPError is returned for non-2xx API responses.
type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("api status %d", e.StatusCode)
	}
	return fmt.Sprintf("api status %d: %s", e.StatusCode, e.Body)
}

// IsRetryable reports whether the caller should Nack(requeue=true)
func IsRetryable(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode >= 500
	}
	return true
}

// IsUnavailable reports transport / upstream-down failures (LLM offline, DNS, timeouts, 5xx).
// Callers should back off before requeuing
func IsUnavailable(err error) bool {
	if err == nil {
		return false
	}
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode >= 500
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		return IsUnavailable(urlErr.Err)
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no such host") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "connection reset")
}
