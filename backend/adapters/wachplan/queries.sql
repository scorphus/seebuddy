-- name: InsertRaw :one
INSERT INTO wachplan_raw (
    sensor_id, upstream_id, app_id, dev_id,
    gtw_id, gtw_rssi, gtw_snr,
    air_temp_c, water_temp_c, humidity_pct, battery_v,
    measured_at, raw_payload
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13
)
ON CONFLICT (sensor_id, upstream_id) DO NOTHING
RETURNING id;

-- name: MaxFetchedAt :one
SELECT COALESCE(MAX(fetched_at), '1970-01-01 00:00:00+00'::TIMESTAMPTZ)::TIMESTAMPTZ AS max_fetched_at FROM wachplan_raw;
