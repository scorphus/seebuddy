-- name: InsertReading :exec
INSERT INTO readings (
    lake_slug, adapter, measured_at,
    water_temp_c, air_temp_c, humidity_pct,
    wind_speed_kmh, wind_direction_deg, weather_code, is_day, raw_id
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9, $10, $11
)
ON CONFLICT (lake_slug, adapter, measured_at) DO NOTHING;

-- name: LatestReadingPerLakePerAdapter :many
SELECT DISTINCT ON (lake_slug, adapter)
    lake_slug, adapter, measured_at,
    water_temp_c, air_temp_c, humidity_pct,
    wind_speed_kmh, wind_direction_deg, weather_code, is_day, raw_id, fetched_at
FROM readings
ORDER BY lake_slug, adapter, measured_at DESC;
