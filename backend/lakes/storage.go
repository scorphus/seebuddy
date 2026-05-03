package lakes

import (
	"context"
	"fmt"

	"github.com/scorphus/seebuddy/backend/adapters"
)

// storeReading inserts one LakeReading into the readings table. Idempotent
// via the (lake_slug, adapter, measured_at) UNIQUE constraint.
func storeReading(ctx context.Context, r adapters.LakeReading) error {
	var rawID *int64
	if r.RawID != 0 {
		rawID = &r.RawID
	}
	err := queries.InsertReading(ctx, InsertReadingParams{
		LakeSlug:         r.Slug,
		Adapter:          r.Adapter,
		MeasuredAt:       r.MeasuredAt,
		WaterTempC:       r.WaterTempC,
		AirTempC:         r.AirTempC,
		HumidityPct:      r.HumidityPct,
		WindSpeedKmh:     r.WindSpeedKMH,
		WindDirectionDeg: r.WindDirectionDeg,
		WeatherCode:      r.WeatherCode,
		IsDay:            r.IsDay,
		RawID:            rawID,
	})
	if err != nil {
		return fmt.Errorf("insert reading: %w", err)
	}
	return nil
}
