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

-- name: LatestPerStation :many
SELECT DISTINCT ON (station_id)
    station_id, water_temp_c, measured_at
FROM gkd_raw
WHERE water_temp_c IS NOT NULL
ORDER BY station_id, measured_at DESC;
