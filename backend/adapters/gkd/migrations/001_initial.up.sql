CREATE TABLE gkd_raw (
    id            BIGSERIAL PRIMARY KEY,
    station_id    TEXT NOT NULL,           -- e.g. '16663002'
    station_slug  TEXT NOT NULL,           -- e.g. 'starnberg'
    station_basin TEXT NOT NULL,           -- URL segment, e.g. 'isar', 'iller_lech'
    station_name  TEXT NOT NULL,           -- display name as shown on GKD
    water_temp_c  NUMERIC(5,2),            -- nullable; sensor gaps render as '—' on GKD
    measured_at   TIMESTAMPTZ NOT NULL,    -- parsed from German DD.MM.YYYY HH:MM in Europe/Berlin, stored UTC
    raw_row_html  TEXT,                    -- outer HTML of the <tr>, for forensics
    fetched_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (station_id, measured_at)
);

CREATE INDEX idx_gkd_raw_station_fetched ON gkd_raw (station_id, fetched_at DESC);
