package lakes

import (
	"testing"
	"time"

	"github.com/scorphus/muenchner-see-buddy/backend/adapters"
)

func TestBuildView_NoReadings(t *testing.T) {
	t.Parallel()

	lakes := []adapters.Lake{
		{Slug: "a", Name: "Lake A"},
		{Slug: "b", Name: "Lake B"},
	}
	views := buildView(lakes, nil, time.Now())

	if len(views) != 2 {
		t.Fatalf("got %d views, want 2", len(views))
	}
	for _, v := range views {
		if v.Latest != nil {
			t.Errorf("lake %q: Latest = %+v, want nil", v.Slug, v.Latest)
		}
	}
}

func TestBuildView_DedupesAcrossAdapters(t *testing.T) {
	t.Parallel()

	// Same lake declared by two adapters; should appear once.
	lakes := []adapters.Lake{
		{Slug: "shared", Name: "Shared Lake"},
		{Slug: "shared", Name: "Shared Lake"},
	}
	views := buildView(lakes, nil, time.Now())

	if len(views) != 1 {
		t.Fatalf("got %d views, want 1 (deduped)", len(views))
	}
}

func TestBuildView_FreshAndStale(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)

	lakes := []adapters.Lake{
		{Slug: "fresh", Name: "Fresh Lake"},
		{Slug: "stale", Name: "Stale Lake"},
		{Slug: "empty", Name: "No Reading Lake"},
	}
	temp := 14.56
	rows := []LatestReadingPerLakePerAdapterRow{
		{LakeSlug: "fresh", Adapter: "wachplan", MeasuredAt: now.Add(-30 * time.Minute), WaterTempC: &temp},
		{LakeSlug: "stale", Adapter: "wachplan", MeasuredAt: now.Add(-3 * time.Hour), WaterTempC: &temp},
	}
	views := buildView(lakes, rows, now)

	if len(views) != 3 {
		t.Fatalf("got %d views, want 3", len(views))
	}

	bySlug := map[string]LakeView{}
	for _, v := range views {
		bySlug[v.Slug] = v
	}

	if v := bySlug["fresh"]; v.Latest == nil || v.Latest.Stale {
		t.Errorf("fresh: Latest = %+v, want stale=false", v.Latest)
	}
	if v := bySlug["stale"]; v.Latest == nil || !v.Latest.Stale {
		t.Errorf("stale: Latest = %+v, want stale=true", v.Latest)
	}
	if v := bySlug["empty"]; v.Latest != nil {
		t.Errorf("empty: Latest = %+v, want nil", v.Latest)
	}
}

func TestBuildView_MergesWeatherWhenSensorIsFresh(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)
	water := 14.3
	wind := 8.6
	code := int32(0)
	day := true

	lakes := []adapters.Lake{
		{Slug: "langwieder", Name: "Langwieder See"},
	}
	rows := []LatestReadingPerLakePerAdapterRow{
		// Sensor: 5 minutes ago, fresh.
		{LakeSlug: "langwieder", Adapter: "wachplan", MeasuredAt: now.Add(-5 * time.Minute), WaterTempC: &water},
		// Weather: 1 minute ago, fresh.
		{LakeSlug: "langwieder", Adapter: "generic", MeasuredAt: now.Add(-1 * time.Minute), WindSpeedKmh: &wind, WeatherCode: &code, IsDay: &day},
	}
	views := buildView(lakes, rows, now)

	if len(views) != 1 || views[0].Latest == nil {
		t.Fatalf("expected one view with Latest set, got %+v", views)
	}
	v := views[0].Latest

	if v.WaterTempC == nil || *v.WaterTempC != water {
		t.Errorf("WaterTempC = %v, want %v", v.WaterTempC, water)
	}
	// Weather should be merged in because sensor is fresh.
	if v.WindSpeedKMH == nil || *v.WindSpeedKMH != wind {
		t.Errorf("WindSpeedKMH = %v, want %v", v.WindSpeedKMH, wind)
	}
	if v.WeatherCode == nil || *v.WeatherCode != code {
		t.Errorf("WeatherCode = %v, want %v", v.WeatherCode, code)
	}
	if v.IsDay == nil || !*v.IsDay {
		t.Errorf("IsDay = %v, want true", v.IsDay)
	}
	if v.Stale {
		t.Error("Stale = true, want false")
	}
}

func TestBuildView_DropsWeatherWhenSensorIsStale(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)
	water := 14.56
	wind := 8.6
	code := int32(0)

	lakes := []adapters.Lake{
		{Slug: "lusssee", Name: "Lußsee"},
	}
	rows := []LatestReadingPerLakePerAdapterRow{
		// Sensor: 9 hours ago, stale.
		{LakeSlug: "lusssee", Adapter: "wachplan", MeasuredAt: now.Add(-9 * time.Hour), WaterTempC: &water},
		// Weather: 1 minute ago, fresh — but won't be blended because sensor is stale.
		{LakeSlug: "lusssee", Adapter: "generic", MeasuredAt: now.Add(-1 * time.Minute), WindSpeedKmh: &wind, WeatherCode: &code},
	}
	views := buildView(lakes, rows, now)

	v := views[0].Latest
	if v == nil {
		t.Fatal("Latest = nil")
	}

	if v.WaterTempC == nil || *v.WaterTempC != water {
		t.Errorf("WaterTempC = %v, want %v", v.WaterTempC, water)
	}
	if !v.Stale {
		t.Error("expected Stale=true")
	}
	// Weather must NOT be blended in.
	if v.WindSpeedKMH != nil {
		t.Errorf("WindSpeedKMH = %v, want nil (sensor stale)", v.WindSpeedKMH)
	}
	if v.WeatherCode != nil {
		t.Errorf("WeatherCode = %v, want nil (sensor stale)", v.WeatherCode)
	}
}

func TestBuildView_WeatherOnlyLake(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)
	wind := 8.6
	code := int32(2)

	lakes := []adapters.Lake{
		{Slug: "ammersee", Name: "Ammersee"},
	}
	// Only a generic/openmeteo row — no sensor.
	rows := []LatestReadingPerLakePerAdapterRow{
		{LakeSlug: "ammersee", Adapter: "generic", MeasuredAt: now.Add(-2 * time.Minute), WindSpeedKmh: &wind, WeatherCode: &code},
	}
	views := buildView(lakes, rows, now)

	v := views[0].Latest
	if v == nil {
		t.Fatal("Latest = nil")
	}
	if v.WaterTempC != nil {
		t.Errorf("WaterTempC = %v, want nil (no sensor)", v.WaterTempC)
	}
	if v.WindSpeedKMH == nil || *v.WindSpeedKMH != wind {
		t.Errorf("WindSpeedKMH = %v, want %v", v.WindSpeedKMH, wind)
	}
	if v.Stale {
		t.Error("expected Stale=false")
	}
}
