-- name: InsertRaw :one
INSERT INTO gkd_raw (
    station_id, station_slug, station_basin, station_name,
    water_temp_c, measured_at, raw_row_html
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7
)
ON CONFLICT (station_id, measured_at) DO NOTHING
RETURNING id;

-- name: MaxFetchedAt :one
SELECT COALESCE(MAX(fetched_at), '1970-01-01 00:00:00+00'::TIMESTAMPTZ)::TIMESTAMPTZ AS max_fetched_at FROM gkd_raw;
