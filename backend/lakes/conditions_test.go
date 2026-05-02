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

	lakes := []adapters.Lake{
		{Slug: "shared", Name: "Shared Lake"},
		{Slug: "shared", Name: "Shared Lake"},
	}
	views := buildView(lakes, nil, time.Now())

	if len(views) != 1 {
		t.Fatalf("got %d views, want 1 (deduped)", len(views))
	}
}

func TestBuildView_StaleWaterFreshWeather(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 2, 22, 0, 0, 0, time.UTC)
	water := 10.8
	wind := 5.0
	code := int32(3)
	day := true

	lakes := []adapters.Lake{{Slug: "tegernsee", Name: "Tegernsee"}}
	rows := []LatestReadingPerLakePerAdapterRow{
		// Sensor: 12 hours old → stale.
		{LakeSlug: "tegernsee", Adapter: "gkd", MeasuredAt: now.Add(-12 * time.Hour), WaterTempC: &water},
		// Weather: 5 minutes old → fresh.
		{LakeSlug: "tegernsee", Adapter: "generic", MeasuredAt: now.Add(-5 * time.Minute), WindSpeedKmh: &wind, WeatherCode: &code, IsDay: &day},
	}
	views := buildView(lakes, rows, now)

	v := views[0].Latest
	if v == nil {
		t.Fatal("Latest = nil")
	}

	// Water: present, water-temp populated, stale.
	if v.Water == nil {
		t.Fatal("Water = nil")
	}
	if v.Water.TempC == nil || *v.Water.TempC != water {
		t.Errorf("Water.TempC = %v, want %v", v.Water.TempC, water)
	}
	if !v.Water.Stale {
		t.Error("Water.Stale = false, want true (12h old)")
	}
	if v.Water.Adapter != "gkd" {
		t.Errorf("Water.Adapter = %q, want gkd", v.Water.Adapter)
	}

	// Weather: present and fresh, with all weather fields.
	if v.Weather == nil {
		t.Fatal("Weather = nil")
	}
	if v.Weather.WindSpeedKMH == nil || *v.Weather.WindSpeedKMH != wind {
		t.Errorf("Weather.WindSpeedKMH = %v, want %v", v.Weather.WindSpeedKMH, wind)
	}
	if v.Weather.Stale {
		t.Error("Weather.Stale = true, want false (5min old)")
	}
	if v.Weather.Adapter != "generic" {
		t.Errorf("Weather.Adapter = %q, want generic", v.Weather.Adapter)
	}
}

func TestBuildView_FreshSensorAndWeather(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 2, 22, 0, 0, 0, time.UTC)
	water := 14.3
	air := 12.5
	humidity := 80.0
	wind := 8.6

	lakes := []adapters.Lake{{Slug: "langwieder", Name: "Langwieder See"}}
	rows := []LatestReadingPerLakePerAdapterRow{
		// Sensor (wachplan) provides air + humidity in addition to water.
		{LakeSlug: "langwieder", Adapter: "wachplan", MeasuredAt: now.Add(-3 * time.Minute),
			WaterTempC: &water, AirTempC: &air, HumidityPct: &humidity},
		// Weather row also has air/humidity from openmeteo, plus wind.
		{LakeSlug: "langwieder", Adapter: "generic", MeasuredAt: now.Add(-1 * time.Minute),
			WindSpeedKmh: &wind},
	}
	views := buildView(lakes, rows, now)

	v := views[0].Latest
	if v == nil || v.Water == nil || v.Weather == nil {
		t.Fatal("expected both Water and Weather populated")
	}
	if v.Water.AirTempC == nil || *v.Water.AirTempC != air {
		t.Errorf("Water.AirTempC = %v, want %v", v.Water.AirTempC, air)
	}
	if v.Weather.WindSpeedKMH == nil || *v.Weather.WindSpeedKMH != wind {
		t.Errorf("Weather.WindSpeedKMH = %v, want %v", v.Weather.WindSpeedKMH, wind)
	}
	if v.Water.Stale || v.Weather.Stale {
		t.Errorf("expected both fresh, got Water.Stale=%v Weather.Stale=%v", v.Water.Stale, v.Weather.Stale)
	}
}

func TestBuildView_WeatherOnlyLake(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 2, 22, 0, 0, 0, time.UTC)
	wind := 8.6
	code := int32(2)

	lakes := []adapters.Lake{{Slug: "ammersee", Name: "Ammersee"}}
	rows := []LatestReadingPerLakePerAdapterRow{
		{LakeSlug: "ammersee", Adapter: "generic", MeasuredAt: now.Add(-2 * time.Minute),
			WindSpeedKmh: &wind, WeatherCode: &code},
	}
	views := buildView(lakes, rows, now)

	v := views[0].Latest
	if v == nil {
		t.Fatal("Latest = nil")
	}
	if v.Water != nil {
		t.Errorf("Water = %+v, want nil (no sensor)", v.Water)
	}
	if v.Weather == nil {
		t.Fatal("Weather = nil")
	}
	if v.Weather.WindSpeedKMH == nil || *v.Weather.WindSpeedKMH != wind {
		t.Errorf("Weather.WindSpeedKMH = %v, want %v", v.Weather.WindSpeedKMH, wind)
	}
}

func TestBuildView_SensorOnlyLake(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 2, 22, 0, 0, 0, time.UTC)
	water := 14.5

	lakes := []adapters.Lake{{Slug: "isolated", Name: "Isolated Lake"}}
	rows := []LatestReadingPerLakePerAdapterRow{
		{LakeSlug: "isolated", Adapter: "wachplan", MeasuredAt: now.Add(-1 * time.Minute), WaterTempC: &water},
	}
	views := buildView(lakes, rows, now)

	v := views[0].Latest
	if v == nil || v.Water == nil {
		t.Fatal("Water = nil")
	}
	if v.Weather != nil {
		t.Errorf("Weather = %+v, want nil (no weather row)", v.Weather)
	}
}
