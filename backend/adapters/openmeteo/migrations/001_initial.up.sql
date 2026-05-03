CREATE TABLE openmeteo_raw (
    id                       BIGSERIAL PRIMARY KEY,
    lake_slug                TEXT NOT NULL,
    measured_at              TIMESTAMPTZ NOT NULL,         -- from open-meteo current.time (UTC)
    temperature_2m_c         NUMERIC(5,2),
    relative_humidity_2m_pct NUMERIC(5,2),
    wind_speed_10m_kmh       NUMERIC(5,2),
    wind_direction_10m_deg   INTEGER,
    weather_code             INTEGER,
    is_day                   BOOLEAN,
    raw_payload              JSONB NOT NULL,
    fetched_at               TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (lake_slug, measured_at)
);

CREATE INDEX idx_openmeteo_raw_lake_fetched ON openmeteo_raw (lake_slug, fetched_at DESC);
