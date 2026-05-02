package openmeteo

import (
	_ "embed"
	"testing"
	"time"
)

//go:embed testdata/sample.json
var sampleFixture []byte

func TestParseResponse(t *testing.T) {
	t.Parallel()

	w, err := parseResponse(sampleFixture)
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}

	wantTime := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)
	if !w.MeasuredAt.Equal(wantTime) {
		t.Errorf("MeasuredAt = %v, want %v", w.MeasuredAt, wantTime)
	}
	if w.TempC == nil || *w.TempC != 14.5 {
		t.Errorf("TempC = %v, want 14.5", w.TempC)
	}
	if w.HumidityPct == nil || *w.HumidityPct != 65 {
		t.Errorf("HumidityPct = %v, want 65", w.HumidityPct)
	}
	if w.WindKMH == nil || *w.WindKMH != 12.3 {
		t.Errorf("WindKMH = %v, want 12.3", w.WindKMH)
	}
	if w.WeatherCode == nil || *w.WeatherCode != 3 {
		t.Errorf("WeatherCode = %v, want 3", w.WeatherCode)
	}
	if w.IsDay == nil || !*w.IsDay {
		t.Errorf("IsDay = %v, want true", w.IsDay)
	}
	if w.WindDirectionDeg == nil || *w.WindDirectionDeg != 270 {
		t.Errorf("WindDirectionDeg = %v, want 270", w.WindDirectionDeg)
	}
}

func TestParseResponse_NightTime(t *testing.T) {
	t.Parallel()

	w, err := parseResponse([]byte(`{"current":{"time":"2026-05-01T03:00","is_day":0,"weather_code":0}}`))
	if err != nil {
		t.Fatalf("parseResponse: %v", err)
	}
	if w.IsDay == nil || *w.IsDay {
		t.Errorf("IsDay = %v, want false", w.IsDay)
	}
}

func TestParseResponse_BadTime(t *testing.T) {
	t.Parallel()

	_, err := parseResponse([]byte(`{"current":{"time":"not-a-time"}}`))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
