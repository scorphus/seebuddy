package lakes

import (
	"context"
	"fmt"
	"time"

	"github.com/scorphus/muenchner-see-buddy/backend/adapters"
)

const stalenessThreshold = 90 * time.Minute

type LakeView struct {
	Slug   string         `json:"slug"`
	Name   string         `json:"name"`
	Region string         `json:"region"`
	Lat    float64        `json:"lat"`
	Lon    float64        `json:"lon"`
	Latest *LatestReading `json:"latest"`
}

type LatestReading struct {
	Adapter      string    `json:"adapter"`
	MeasuredAt   time.Time `json:"measured_at"`
	AgeSeconds   int64     `json:"age_seconds"`
	Stale        bool      `json:"stale"`
	WaterTempC   *float64  `json:"water_temp_c"`
	AirTempC     *float64  `json:"air_temp_c"`
	HumidityPct  *float64  `json:"humidity_pct"`
	WindSpeedKMH *float64  `json:"wind_speed_kmh"`
	WeatherCode  *int32    `json:"weather_code"`
	IsDay        *bool     `json:"is_day"`
}

type ListResponse struct {
	Lakes []LakeView `json:"lakes"`
}

//encore:api public method=GET path=/lakes
func List(ctx context.Context) (*ListResponse, error) {
	rows, err := queries.LatestReadingPerLakePerAdapter(ctx)
	if err != nil {
		return nil, fmt.Errorf("query latest readings: %w", err)
	}

	var allLakes []adapters.Lake
	for _, a := range registered {
		allLakes = append(allLakes, a.Lakes()...)
	}

	return &ListResponse{Lakes: buildView(allLakes, rows, time.Now())}, nil
}

// buildView merges the catalog (deduped across adapters) with the latest
// readings, joining per-adapter rows for the same lake into one
// LatestReading.
func buildView(lakes []adapters.Lake, readings []LatestReadingPerLakePerAdapterRow, now time.Time) []LakeView {
	bySlug := make(map[string][]LatestReadingPerLakePerAdapterRow)
	for _, r := range readings {
		bySlug[r.LakeSlug] = append(bySlug[r.LakeSlug], r)
	}

	seen := make(map[string]bool, len(lakes))
	out := make([]LakeView, 0, len(lakes))
	for _, l := range lakes {
		if seen[l.Slug] {
			continue
		}
		seen[l.Slug] = true

		view := LakeView{
			Slug:   l.Slug,
			Name:   l.Name,
			Region: l.Region,
			Lat:    l.Lat,
			Lon:    l.Lon,
		}
		if rows, ok := bySlug[l.Slug]; ok {
			view.Latest = mergeReadings(rows, now)
		}
		out = append(out, view)
	}
	return out
}

// mergeReadings combines rows from one or more adapters into a single
// LatestReading. The "primary" row (sensor when present, otherwise the
// freshest) defines the timestamp/age/stale flag. Sensor-source fields
// (water/air/humidity) come from the sensor row. Weather-source fields
// (wind/weather code/is_day) are merged from a weather row only when the
// primary is still fresh — older than stalenessThreshold means we don't
// blend fresh weather with old water; show the sensor reading as-is.
func mergeReadings(rows []LatestReadingPerLakePerAdapterRow, now time.Time) *LatestReading {
	if len(rows) == 0 {
		return nil
	}

	var sensor, weather *LatestReadingPerLakePerAdapterRow
	for i := range rows {
		r := &rows[i]
		if r.WaterTempC != nil {
			if sensor == nil || r.MeasuredAt.After(sensor.MeasuredAt) {
				sensor = r
			}
		}
		if r.WindSpeedKmh != nil || r.WeatherCode != nil || r.IsDay != nil {
			if weather == nil || r.MeasuredAt.After(weather.MeasuredAt) {
				weather = r
			}
		}
	}

	primary := sensor
	if primary == nil {
		primary = weather
	}
	if primary == nil {
		primary = &rows[0]
		for i := range rows[1:] {
			if rows[i+1].MeasuredAt.After(primary.MeasuredAt) {
				primary = &rows[i+1]
			}
		}
	}

	age := now.Sub(primary.MeasuredAt)
	out := &LatestReading{
		Adapter:    primary.Adapter,
		MeasuredAt: primary.MeasuredAt,
		AgeSeconds: int64(age.Seconds()),
		Stale:      age > stalenessThreshold,
	}

	if sensor != nil {
		out.WaterTempC = sensor.WaterTempC
		out.AirTempC = sensor.AirTempC
		out.HumidityPct = sensor.HumidityPct
	}
	// Only blend weather when the primary reading is still fresh. Old water
	// data with fresh weather would be misleading.
	if weather != nil && !out.Stale {
		out.WindSpeedKMH = weather.WindSpeedKmh
		out.WeatherCode = weather.WeatherCode
		out.IsDay = weather.IsDay
		if out.AirTempC == nil {
			out.AirTempC = weather.AirTempC
		}
		if out.HumidityPct == nil {
			out.HumidityPct = weather.HumidityPct
		}
	}

	return out
}
