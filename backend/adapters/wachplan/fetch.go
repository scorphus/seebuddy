package wachplan

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const userAgent = "muenchner-see-buddy/0.1 (+https://github.com/scorphus/muenchner-see-buddy)"
const upstreamURL = "https://sensors.mein-wachplan.de/json/"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// rawResponse mirrors the upstream JSON. Numeric values arrive as strings.
type rawResponse struct {
	Error        string `json:"error,omitempty"`
	ID           string `json:"id"`
	AppID        string `json:"app_id"`
	DevID        string `json:"dev_id"`
	TTNTimestamp string `json:"ttn_timestamp"`
	GtwID        string `json:"gtw_id"`
	GtwRSSI      string `json:"gtw_rssi"`
	GtwSNR       string `json:"gtw_snr"`
	DevValue1    string `json:"dev_value_1"` // air temp °C
	DevValue2    string `json:"dev_value_2"` // water temp °C
	DevValue3    string `json:"dev_value_3"` // humidity %
	DevValue4    string `json:"dev_value_4"` // battery V
}

func fetch(ctx context.Context, sensorID string) (rawResponse, []byte, error) {
	var parsed rawResponse

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, upstreamURL+"?sensor_id="+sensorID, nil)
	if err != nil {
		return parsed, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return parsed, nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return parsed, nil, fmt.Errorf("read body: %w", err)
	}
	parsed, err = parseResponse(raw)
	return parsed, raw, err
}

// parseResponse decodes the upstream JSON and surfaces upstream-reported errors.
func parseResponse(raw []byte) (rawResponse, error) {
	var parsed rawResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return parsed, fmt.Errorf("decode: %w", err)
	}
	if parsed.Error != "" {
		return parsed, fmt.Errorf("upstream: %s", parsed.Error)
	}
	return parsed, nil
}

func parseFloatPtr(s string) *float64 {
	if s == "" {
		return nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func parseIntPtr32(s string) *int32 {
	if s == "" {
		return nil
	}
	n, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return nil
	}
	v := int32(n)
	return &v
}
