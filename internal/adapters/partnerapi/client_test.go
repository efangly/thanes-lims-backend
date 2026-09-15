package partnerapi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/efangly/thanes-lims-backend/internal/adapters/partnerapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetchMetadata_Success(t *testing.T) {
	var gotPath, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("X-API-Key")
		fmt.Fprint(w, `{
			"success": true,
			"message": "Request Successfully",
			"data": {"serial": "SN-00042", "name": "Fridge — Ward 3", "status": true, "firmware": "1.4.2", "online": true},
			"timestamp": "2026-09-14T07:27:47.943Z",
			"statusCode": 200
		}`)
	}))
	defer srv.Close()

	client := partnerapi.New(srv.URL, "pk_live_test")
	meta, err := client.FetchMetadata(t.Context(), "SN-00042")

	require.NoError(t, err)
	assert.Equal(t, "/log/partner/devices/SN-00042", gotPath)
	assert.Equal(t, "pk_live_test", gotKey)
	assert.Equal(t, "SN-00042", meta.Serial)
	assert.Equal(t, "Fridge — Ward 3", meta.Name)
	assert.True(t, meta.Status)
	assert.Equal(t, "1.4.2", meta.Firmware)
	assert.True(t, meta.Online)
}

func TestFetchMetadata_NotFound_IsPermanent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"success": false, "message": "Device SN-999 not found", "data": null}`)
	}))
	defer srv.Close()

	client := partnerapi.New(srv.URL, "pk_live_test")
	_, err := client.FetchMetadata(t.Context(), "SN-999")

	require.Error(t, err)
	var apiErr *partnerapi.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "Device SN-999 not found", apiErr.Message)
	assert.False(t, apiErr.Retryable())
}

func TestFetchMetadata_RateLimited_IsRetryable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"success": false, "message": "Too Many Requests", "data": null}`)
	}))
	defer srv.Close()

	client := partnerapi.New(srv.URL, "pk_live_test")
	_, err := client.FetchMetadata(t.Context(), "SN-00042")

	require.Error(t, err)
	var apiErr *partnerapi.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusTooManyRequests, apiErr.StatusCode)
	assert.True(t, apiErr.Retryable())
}

func TestFetchMetadata_Unauthorized_IsPermanent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"success": false, "message": "Unauthorized", "data": null}`)
	}))
	defer srv.Close()

	client := partnerapi.New(srv.URL, "bad-key")
	_, err := client.FetchMetadata(t.Context(), "SN-00042")

	var apiErr *partnerapi.APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
	assert.False(t, apiErr.Retryable())
}

func TestFetchLatestReading_ReturnsNewestRow(t *testing.T) {
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		fmt.Fprint(w, `{
			"success": true,
			"message": "Request Successfully",
			"data": [
				{"serial": "SN-00042", "sendTime": "2026-08-15T00:00:00.000Z", "temp": 4.8, "tempDisplay": 4.8, "humidity": 52.1, "humidityDisplay": 52}
			],
			"meta": {"page": 1, "limit": 1, "total": 336, "totalPages": 336},
			"timestamp": "2026-09-14T07:27:48.264Z",
			"statusCode": 200
		}`)
	}))
	defer srv.Close()

	client := partnerapi.New(srv.URL, "pk_live_test")
	reading, found, err := client.FetchLatestReading(t.Context(), "SN-00042")

	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "SN-00042", reading.Serial)
	assert.Equal(t, 4.8, reading.TempDisplay)
	assert.Equal(t, float64(52), reading.HumidityDisplay)
	assert.Contains(t, gotQuery, "limit=1")
	assert.Contains(t, gotQuery, "from=")
	assert.Contains(t, gotQuery, "to=")
}

func TestFetchLatestReading_EmptyWindow_NotFoundNotError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"success": true, "message": "Request Successfully", "data": [], "meta": {"page":1,"limit":1,"total":0,"totalPages":0}, "timestamp": "2026-09-14T07:27:48.264Z", "statusCode": 200}`)
	}))
	defer srv.Close()

	client := partnerapi.New(srv.URL, "pk_live_test")
	_, found, err := client.FetchLatestReading(t.Context(), "SN-00042")

	require.NoError(t, err)
	assert.False(t, found)
}
