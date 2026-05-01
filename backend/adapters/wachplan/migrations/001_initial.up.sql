CREATE TABLE wachplan_raw (
    id           BIGSERIAL PRIMARY KEY,
    sensor_id    TEXT NOT NULL,                 -- 'luss', 'laws'
    upstream_id  TEXT NOT NULL,                 -- auto-incrementing "id" from upstream JSON
    app_id       TEXT NOT NULL,                 -- e.g. 'wassertemperatur-luss-see'
    dev_id       TEXT NOT NULL,                 -- LoRaWAN device EUI
    gtw_id       TEXT,
    gtw_rssi     INTEGER,
    gtw_snr      NUMERIC(5,1),
    air_temp_c   NUMERIC(5,2),                  -- dev_value_1
    water_temp_c NUMERIC(5,2),                  -- dev_value_2
    humidity_pct NUMERIC(5,2),                  -- dev_value_3
    battery_v    NUMERIC(5,3),                  -- dev_value_4
    measured_at  TIMESTAMPTZ NOT NULL,          -- ttn_timestamp (UTC)
    raw_payload  JSONB NOT NULL,                -- full upstream JSON, lossless
    fetched_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (sensor_id, upstream_id)
);

CREATE INDEX idx_wachplan_raw_sensor_fetched ON wachplan_raw (sensor_id, fetched_at DESC);
