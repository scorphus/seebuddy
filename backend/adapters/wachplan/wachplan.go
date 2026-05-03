// Package wachplan adapts the Wasserwacht München-West LoRaWAN sensor API
// (https://sensors.mein-wachplan.de) to the seebudy adapter
// contract. It owns the wachplan_raw table where every successful poll is
// stored verbatim alongside typed columns for forensics.
package wachplan

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/scorphus/seebudy/backend/adapters"
)

const id = "wachplan"

// period is how often this adapter is willing to hit the upstream. The TTN
// devices behind the upstream report every ~10–15 minutes, so polling more
// often is both wasteful and impolite.
const period = 15 * time.Minute

var db = sqldb.Driver[*pgxpool.Pool](sqldb.NewDatabase("wachplan", sqldb.DatabaseConfig{
	Migrations: "./migrations",
}))

var queries = New(db)

var lakes = []adapters.Lake{
	{
		Slug:     "langwieder",
		Name:     "Langwieder See",
		Region:   "munich-west",
		Lat:      48.195959,
		Lon:      11.417537,
		SensorID: "laws",
	},
	{
		Slug:     "lusssee",
		Name:     "Lußsee",
		Region:   "munich-west",
		Lat:      48.196504,
		Lon:      11.417984,
		SensorID: "luss",
	},
}

type Adapter struct{}

func (Adapter) ID() string { return id }

func (Adapter) Lakes() []adapters.Lake { return lakes }

func (Adapter) Tick(ctx context.Context) ([]adapters.LakeReading, error) {
	last, err := queries.MaxFetchedAt(ctx)
	if err != nil {
		return nil, fmt.Errorf("max fetched: %w", err)
	}
	if time.Since(last) < period {
		return nil, nil
	}

	out := make([]adapters.LakeReading, 0, len(lakes))
	for _, l := range lakes {
		parsed, raw, err := fetch(ctx, l.SensorID)
		if err != nil {
			rlog.Error("wachplan fetch", "sensor", l.SensorID, "err", err)
			continue
		}

		measuredAt, err := time.Parse(time.RFC3339Nano, parsed.TTNTimestamp)
		if err != nil {
			rlog.Error("wachplan parse timestamp", "sensor", l.SensorID, "err", err)
			continue
		}

		air := parseFloatPtr(parsed.DevValue1)
		water := parseFloatPtr(parsed.DevValue2)
		humidity := parseFloatPtr(parsed.DevValue3)
		battery := parseFloatPtr(parsed.DevValue4)
		rssi := parseIntPtr32(parsed.GtwRSSI)
		snr := parseFloatPtr(parsed.GtwSNR)

		var gtwID *string
		if parsed.GtwID != "" {
			gid := parsed.GtwID
			gtwID = &gid
		}

		rawID, err := queries.InsertRaw(ctx, InsertRawParams{
			SensorID:    l.SensorID,
			UpstreamID:  parsed.ID,
			AppID:       parsed.AppID,
			DevID:       parsed.DevID,
			GtwID:       gtwID,
			GtwRssi:     rssi,
			GtwSnr:      snr,
			AirTempC:    air,
			WaterTempC:  water,
			HumidityPct: humidity,
			BatteryV:    battery,
			MeasuredAt:  measuredAt,
			RawPayload:  raw,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			// upstream record already persisted on a previous run
			continue
		}
		if err != nil {
			rlog.Error("wachplan insert raw", "sensor", l.SensorID, "err", err)
			continue
		}

		out = append(out, adapters.LakeReading{
			Lake:        l,
			Adapter:     id,
			MeasuredAt:  measuredAt,
			WaterTempC:  water,
			AirTempC:    air,
			HumidityPct: humidity,
			RawID:       rawID,
		})
	}
	return out, nil
}
