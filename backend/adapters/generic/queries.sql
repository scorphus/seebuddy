-- name: InsertRaw :one
INSERT INTO generic_raw (
    lake_slug, measured_at,
    temperature_2m_c, relative_humidity_2m_pct, wind_speed_10m_kmh, weather_code,
    raw_payload
) VALUES (
    $1, $2,
    $3, $4, $5, $6,
    $7
)
ON CONFLICT (lake_slug, measured_at) DO NOTHING
RETURNING id;

-- name: MaxFetchedAt :one
SELECT COALESCE(MAX(fetched_at), '1970-01-01 00:00:00+00'::TIMESTAMPTZ)::TIMESTAMPTZ AS max_fetched_at FROM generic_raw;
