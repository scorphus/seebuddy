// Package openmeteo is a thin client for the Open-Meteo current-weather API.
// It is shared infrastructure used by adapters that want to enrich a lake
// reading with ambient weather based on lat/lon.
package openmeteo

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
const baseURL = "https://api.open-meteo.com/v1/forecast"

var httpClient = &http.Client{Timeout: 10 * time.Second}

// Weather is the parsed current-weather snapshot for a coordinate.
type Weather struct {
	MeasuredAt  time.Time
	TempC       *float64
	HumidityPct *float64
	WindKMH     *float64
	WeatherCode *int32
}

type apiResponse struct {
	Current struct {
		Time               string   `json:"time"`
		Temperature2m      *float64 `json:"temperature_2m"`
		RelativeHumidity2m *float64 `json:"relative_humidity_2m"`
		WindSpeed10m       *float64 `json:"wind_speed_10m"`
		WeatherCode        *int32   `json:"weather_code"`
	} `json:"current"`
}

// FetchWeather returns the current observation for the given coordinates,
// the raw JSON body for storage, or an error.
func FetchWeather(ctx context.Context, lat, lon float64) (Weather, []byte, error) {
	var w Weather

	url := baseURL +
		"?latitude=" + strconv.FormatFloat(lat, 'f', 4, 64) +
		"&longitude=" + strconv.FormatFloat(lon, 'f', 4, 64) +
		"&current=temperature_2m,relative_humidity_2m,wind_speed_10m,weather_code" +
		"&timezone=UTC"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return w, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return w, nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return w, nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return w, raw, fmt.Errorf("upstream status %d", resp.StatusCode)
	}

	w, err = parseResponse(raw)
	return w, raw, err
}

// parseResponse decodes the Open-Meteo current-weather payload.
func parseResponse(raw []byte) (Weather, error) {
	var parsed apiResponse
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return Weather{}, fmt.Errorf("decode: %w", err)
	}
	measuredAt, err := time.Parse("2006-01-02T15:04", parsed.Current.Time)
	if err != nil {
		return Weather{}, fmt.Errorf("parse time %q: %w", parsed.Current.Time, err)
	}
	return Weather{
		MeasuredAt:  measuredAt.UTC(),
		TempC:       parsed.Current.Temperature2m,
		HumidityPct: parsed.Current.RelativeHumidity2m,
		WindKMH:     parsed.Current.WindSpeed10m,
		WeatherCode: parsed.Current.WeatherCode,
	}, nil
}
