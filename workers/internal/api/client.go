package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/racecoach/workers/internal/domain"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiToken   string
}

func NewClient(baseURL, apiToken string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 15 * time.Second},
		baseURL:    baseURL,
		apiToken:   apiToken,
	}
}

func (c *Client) PatchActivityMetrics(ctx context.Context, activityID int, metrics domain.ActivityMetrics) error {
	body, err := json.Marshal(metrics)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/activities/%d/metrics", c.baseURL, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("X-Worker-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return nil
}

func (c *Client) PatchActivityStructure(ctx context.Context, activityID int, request domain.PatchActivityStructureRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/activities/%d/structure", c.baseURL, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("X-Worker-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return nil
}

func (c *Client) PatchActivityFeatures(ctx context.Context, activityID int, request domain.PatchActivityFeaturesRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/activities/%d/features", c.baseURL, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("X-Worker-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return nil
}

func (c *Client) PatchActivitySummary(ctx context.Context, activityID int, request domain.PatchActivitySummaryRequest) error {
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	url := fmt.Sprintf("%s/api/activities/%d/summary", c.baseURL, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("X-Worker-Token", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return nil
}

func (c *Client) PatchPipelineHeartbeat(ctx context.Context, activityID int) error {
	url := fmt.Sprintf("%s/api/activities/%d/pipeline-heartbeat", c.baseURL, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, url, http.NoBody)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("X-Worker-Token", c.apiToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	return nil
}

func (c *Client) GetActivity(ctx context.Context, activityID int) (*domain.Activity, error) {
	url := fmt.Sprintf("%s/api/activities/%d", c.baseURL, activityID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("X-Worker-Token", c.apiToken)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var activity domain.Activity
	if err := json.Unmarshal(respBody, &activity); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &activity, nil
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

// IsRetryable reports whether the caller should Nack(requeue=true).
// Network errors and 5xx are retryable.
func IsRetryable(err error) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode >= 500
	}
	return true
}
