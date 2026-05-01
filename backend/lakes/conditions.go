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
}

type ListResponse struct {
	Lakes []LakeView `json:"lakes"`
}

//encore:api public method=GET path=/lakes
func List(ctx context.Context) (*ListResponse, error) {
	rows, err := queries.LatestReadingPerLake(ctx)
	if err != nil {
		return nil, fmt.Errorf("query latest readings: %w", err)
	}

	var allLakes []adapters.Lake
	for _, a := range registered {
		allLakes = append(allLakes, a.Lakes()...)
	}

	return &ListResponse{Lakes: buildView(allLakes, rows, time.Now())}, nil
}

// buildView merges the catalog of lakes coming from adapters with the latest
// reading per lake_slug. Stale flag is set when the reading is older than
// stalenessThreshold relative to now.
func buildView(lakes []adapters.Lake, readings []LatestReadingPerLakeRow, now time.Time) []LakeView {
	bySlug := make(map[string]LatestReadingPerLakeRow, len(readings))
	for _, r := range readings {
		bySlug[r.LakeSlug] = r
	}
	out := make([]LakeView, 0, len(lakes))
	for _, l := range lakes {
		view := LakeView{
			Slug:   l.Slug,
			Name:   l.Name,
			Region: l.Region,
			Lat:    l.Lat,
			Lon:    l.Lon,
		}
		if r, ok := bySlug[l.Slug]; ok {
			age := now.Sub(r.MeasuredAt)
			view.Latest = &LatestReading{
				Adapter:      r.Adapter,
				MeasuredAt:   r.MeasuredAt,
				AgeSeconds:   int64(age.Seconds()),
				Stale:        age > stalenessThreshold,
				WaterTempC:   r.WaterTempC,
				AirTempC:     r.AirTempC,
				HumidityPct:  r.HumidityPct,
				WindSpeedKMH: r.WindSpeedKmh,
				WeatherCode:  r.WeatherCode,
			}
		}
		out = append(out, view)
	}
	return out
}
