-- name: InsertRaw :one
INSERT INTO openmeteo_raw (
    lake_slug, measured_at,
    temperature_2m_c, relative_humidity_2m_pct, wind_speed_10m_kmh, wind_direction_10m_deg,
    weather_code, is_day,
    raw_payload
) VALUES (
    $1, $2,
    $3, $4, $5, $6,
    $7, $8,
    $9
)
ON CONFLICT (lake_slug, measured_at) DO NOTHING
RETURNING id;

-- name: MaxFetchedAt :one
SELECT COALESCE(MAX(fetched_at), '1970-01-01 00:00:00+00'::TIMESTAMPTZ)::TIMESTAMPTZ AS max_fetched_at FROM openmeteo_raw;
