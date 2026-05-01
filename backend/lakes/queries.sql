-- name: InsertReading :exec
INSERT INTO readings (
    lake_slug, adapter, measured_at,
    water_temp_c, air_temp_c, humidity_pct,
    wind_speed_kmh, weather_code, raw_id
) VALUES (
    $1, $2, $3,
    $4, $5, $6,
    $7, $8, $9
)
ON CONFLICT (lake_slug, adapter, measured_at) DO NOTHING;

-- name: LatestReadingPerLake :many
SELECT DISTINCT ON (lake_slug)
    lake_slug, adapter, measured_at,
    water_temp_c, air_temp_c, humidity_pct,
    wind_speed_kmh, weather_code, raw_id, fetched_at
FROM readings
ORDER BY lake_slug, measured_at DESC;
