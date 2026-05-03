CREATE TABLE readings (
    id                  BIGSERIAL PRIMARY KEY,
    lake_slug           TEXT NOT NULL,
    adapter             TEXT NOT NULL,
    measured_at         TIMESTAMPTZ NOT NULL,
    water_temp_c        NUMERIC(5,2),
    air_temp_c          NUMERIC(5,2),
    humidity_pct        NUMERIC(5,2),
    wind_speed_kmh      NUMERIC(5,2),
    wind_direction_deg  INTEGER,
    weather_code        INTEGER,
    is_day              BOOLEAN,
    raw_id              BIGINT,
    fetched_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (lake_slug, adapter, measured_at)
);

CREATE INDEX idx_readings_lake_measured ON readings (lake_slug, measured_at DESC);
