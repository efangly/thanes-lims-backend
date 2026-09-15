// Package partnerapi implements portenvironment.PartnerAPIClient against
// SMtrack's Partner API for third-party device data (docs/partner-api-guide.md).
// Auth is a static X-API-Key header; there is no SDK for this API, so this
// is a plain net/http client - see CONTEXT.md#environment and ADR 0011.
package partnerapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	portenvironment "github.com/efangly/thanes-lims-backend/internal/ports/environment"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// New builds a Client. baseURL is the SMtrack host (e.g.
// "https://smtrack.example.com"), without the "/log/partner" suffix -
// that prefix is added per-request.
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

// APIError is returned when the Partner API responds with a non-2xx
// status. StatusCode lets callers (the poller) distinguish a transient
// failure (429, 5xx, or a network/timeout error) worth a stale-cache
// fallback from a permanent one (401/404) that should surface as a real
// error - see ADR 0011.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("partner api: %d: %s", e.StatusCode, e.Message)
}

// Retryable reports whether the failure is worth a stale-cache fallback
// (rate limited or a server-side hiccup) rather than a hard error.
func (e *APIError) Retryable() bool {
	return e.StatusCode == http.StatusTooManyRequests || e.StatusCode >= 500
}

type envelope struct {
	Success bool            `json:"success"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *Client) do(ctx context.Context, path string) (json.RawMessage, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var env envelope
		_ = json.Unmarshal(body, &env) // best-effort: fall back to empty message on a malformed error body
		return nil, &APIError{StatusCode: resp.StatusCode, Message: env.Message}
	}

	var env envelope
	if err := json.Unmarshal(body, &env); err != nil {
		return nil, fmt.Errorf("partner api: decode response: %w", err)
	}
	return env.Data, nil
}

type metadataDTO struct {
	Serial   string `json:"serial"`
	Name     string `json:"name"`
	Status   bool   `json:"status"`
	Firmware string `json:"firmware"`
	Online   bool   `json:"online"`
}

func (c *Client) FetchMetadata(ctx context.Context, serial string) (portenvironment.PartnerDeviceMetadata, error) {
	data, err := c.do(ctx, "/log/partner/devices/"+url.PathEscape(serial))
	if err != nil {
		return portenvironment.PartnerDeviceMetadata{}, err
	}
	var dto metadataDTO
	if err := json.Unmarshal(data, &dto); err != nil {
		return portenvironment.PartnerDeviceMetadata{}, fmt.Errorf("partner api: decode metadata: %w", err)
	}
	return portenvironment.PartnerDeviceMetadata{
		Serial:   dto.Serial,
		Name:     dto.Name,
		Status:   dto.Status,
		Firmware: dto.Firmware,
		Online:   dto.Online,
	}, nil
}

type readingDTO struct {
	Serial          string    `json:"serial"`
	SendTime        time.Time `json:"sendTime"`
	TempDisplay     float64   `json:"tempDisplay"`
	HumidityDisplay float64   `json:"humidityDisplay"`
}

// FetchLatestReading asks for the last 24h window, newest first, one row -
// the timeseries endpoint has no dedicated "latest" shortcut and requires
// an explicit from/to (max 30 days), per docs/partner-api-guide.md.
func (c *Client) FetchLatestReading(ctx context.Context, serial string) (portenvironment.PartnerDeviceReading, bool, error) {
	now := time.Now().UTC()
	from := now.Add(-24 * time.Hour)
	q := url.Values{
		"from":  {from.Format(time.RFC3339)},
		"to":    {now.Format(time.RFC3339)},
		"limit": {"1"},
		"page":  {"1"},
	}
	path := "/log/partner/devices/" + url.PathEscape(serial) + "/timeseries?" + q.Encode()

	data, err := c.do(ctx, path)
	if err != nil {
		return portenvironment.PartnerDeviceReading{}, false, err
	}

	var rows []readingDTO
	if err := json.Unmarshal(data, &rows); err != nil {
		return portenvironment.PartnerDeviceReading{}, false, fmt.Errorf("partner api: decode timeseries: %w", err)
	}
	if len(rows) == 0 {
		return portenvironment.PartnerDeviceReading{}, false, nil
	}
	r := rows[0]
	return portenvironment.PartnerDeviceReading{
		Serial:          r.Serial,
		SendTime:        r.SendTime,
		TempDisplay:     r.TempDisplay,
		HumidityDisplay: r.HumidityDisplay,
	}, true, nil
}
