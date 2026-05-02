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

// LatestReading splits the observation into two independent sections so each
// can carry its own freshness. A lake can have just a sensor (Water), just
// weather (Weather), or both — and one being stale doesn't hide the other.
type LatestReading struct {
	Water   *WaterReading   `json:"water"`
	Weather *WeatherReading `json:"weather"`
}

// WaterReading is what a sensor adapter (wachplan, gkd) reports. Some sensors
// also surface air/humidity (wachplan does); others don't (gkd is water-only).
type WaterReading struct {
	Adapter     string    `json:"adapter"`
	MeasuredAt  time.Time `json:"measured_at"`
	AgeSeconds  int64     `json:"age_seconds"`
	Stale       bool      `json:"stale"`
	TempC       *float64  `json:"temp_c"`
	AirTempC    *float64  `json:"air_temp_c,omitempty"`
	HumidityPct *float64  `json:"humidity_pct,omitempty"`
}

// WeatherReading is the ambient observation from openmeteo (the generic
// adapter). Always available for any lake in the catalog.
type WeatherReading struct {
	Adapter          string    `json:"adapter"`
	MeasuredAt       time.Time `json:"measured_at"`
	AgeSeconds       int64     `json:"age_seconds"`
	Stale            bool      `json:"stale"`
	AirTempC         *float64  `json:"air_temp_c"`
	HumidityPct      *float64  `json:"humidity_pct"`
	WindSpeedKMH     *float64  `json:"wind_speed_kmh"`
	WindDirectionDeg *int32    `json:"wind_direction_deg"`
	WeatherCode      *int32    `json:"weather_code"`
	IsDay            *bool     `json:"is_day"`
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
// readings, producing one Water section (the freshest sensor row) and one
// Weather section (the freshest row carrying weather fields) per lake.
// Either section may be nil when no data is available for it.
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
			view.Latest = splitReadings(rows, now)
		}
		out = append(out, view)
	}
	return out
}

// splitReadings picks the best sensor row (has water_temp_c) and the best
// weather row (has wind/weather_code/is_day) and returns them as separate
// sections.
func splitReadings(rows []LatestReadingPerLakePerAdapterRow, now time.Time) *LatestReading {
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

	if sensor == nil && weather == nil {
		return nil
	}

	out := &LatestReading{}
	if sensor != nil {
		age := now.Sub(sensor.MeasuredAt)
		out.Water = &WaterReading{
			Adapter:     sensor.Adapter,
			MeasuredAt:  sensor.MeasuredAt,
			AgeSeconds:  int64(age.Seconds()),
			Stale:       age > stalenessThreshold,
			TempC:       sensor.WaterTempC,
			AirTempC:    sensor.AirTempC,
			HumidityPct: sensor.HumidityPct,
		}
	}
	if weather != nil {
		age := now.Sub(weather.MeasuredAt)
		out.Weather = &WeatherReading{
			Adapter:          weather.Adapter,
			MeasuredAt:       weather.MeasuredAt,
			AgeSeconds:       int64(age.Seconds()),
			Stale:            age > stalenessThreshold,
			AirTempC:         weather.AirTempC,
			HumidityPct:      weather.HumidityPct,
			WindSpeedKMH:     weather.WindSpeedKmh,
			WindDirectionDeg: weather.WindDirectionDeg,
			WeatherCode:      weather.WeatherCode,
			IsDay:            weather.IsDay,
		}
	}
	return out
}
