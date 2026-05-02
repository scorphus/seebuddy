// Package generic adapts Open-Meteo to the muenchner-see-buddy adapter
// contract. It covers lakes that have no dedicated water-temperature sensor
// but for which we still want air conditions (and which we'll fall back to
// when no specialized adapter applies).
package generic

import (
	"context"
	"errors"
	"fmt"
	"time"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/scorphus/muenchner-see-buddy/backend/adapters"
	"github.com/scorphus/muenchner-see-buddy/backend/openmeteo"
)

const id = "generic"

// period matches Open-Meteo's typical refresh cadence.
const period = 15 * time.Minute

var db = sqldb.Driver[*pgxpool.Pool](sqldb.NewDatabase("generic", sqldb.DatabaseConfig{
	Migrations: "./migrations",
}))

var queries = New(db)

var lakes = []adapters.Lake{
	{
		Slug:   "ammersee",
		Name:   "Ammersee",
		Region: "munich-southwest",
		Lat:    47.997458,
		Lon:    11.165848,
	},
	{
		Slug:   "fasanerie",
		Name:   "Fasaneriesee",
		Region: "munich-north",
		Lat:    48.204173,
		Lon:    11.529536,
	},
	{
		Slug:   "feldmochinger",
		Name:   "Feldmochinger See",
		Region: "munich-north",
		Lat:    48.213918,
		Lon:    11.514862,
	},
	{
		Slug:   "feringasee",
		Name:   "Feringasee",
		Region: "munich-northeast",
		Lat:    48.194947,
		Lon:    11.671528,
	},
	{
		Slug:   "germeringer",
		Name:   "Germeringer See",
		Region: "munich-west",
		Lat:    48.137298,
		Lon:    11.343992,
	},
	{
		Slug:   "karlsfelder",
		Name:   "Karlsfelder See",
		Region: "munich-north",
		Lat:    48.236857,
		Lon:    11.468377,
	},
	{
		Slug:   "kochelsee",
		Name:   "Kochelsee",
		Region: "munich-south",
		Lat:    47.653429,
		Lon:    11.356356,
	},
	{
		Slug:   "langwieder",
		Name:   "Langwieder See",
		Region: "munich-west",
		Lat:    48.195959,
		Lon:    11.417537,
	},
	{
		Slug:   "lusssee",
		Name:   "Lußsee",
		Region: "munich-west",
		Lat:    48.196504,
		Lon:    11.417984,
	},
	{
		Slug:   "olchinger",
		Name:   "Olchinger See",
		Region: "munich-northwest",
		Lat:    48.208980,
		Lon:    11.357034,
	},
	{
		Slug:   "pilsensee",
		Name:   "Pilsensee",
		Region: "munich-southwest",
		Lat:    48.024241,
		Lon:    11.189486,
	},
	{
		Slug:   "regattaanlage",
		Name:   "Regattaanlage Oberschleißheim",
		Region: "munich-north",
		Lat:    48.248238,
		Lon:    11.523698,
	},
	{
		Slug:   "riemer",
		Name:   "Riemer See",
		Region: "munich-east",
		Lat:    48.126119,
		Lon:    11.705874,
	},
	{
		Slug:   "schliersee",
		Name:   "Schliersee",
		Region: "munich-south",
		Lat:    47.731727,
		Lon:    11.863647,
	},
	{
		Slug:   "starnberger",
		Name:   "Starnberger See",
		Region: "munich-south",
		Lat:    47.995538,
		Lon:    11.345372,
	},
	{
		Slug:   "tegernsee",
		Name:   "Tegernsee",
		Region: "munich-south",
		Lat:    47.699499,
		Lon:    11.758345,
	},
	{
		Slug:   "walchensee",
		Name:   "Walchensee",
		Region: "munich-south",
		Lat:    47.615885,
		Lon:    11.344804,
	},
	{
		Slug:   "wesslinger",
		Name:   "Weßlinger See",
		Region: "munich-southwest",
		Lat:    48.072887,
		Lon:    11.251091,
	},
	{
		Slug:   "woerthsee",
		Name:   "Wörthsee",
		Region: "munich-southwest",
		Lat:    48.067428,
		Lon:    11.199795,
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
		w, raw, err := openmeteo.FetchWeather(ctx, l.Lat, l.Lon)
		if err != nil {
			rlog.Error("openmeteo fetch", "lake", l.Slug, "err", err)
			continue
		}

		rawID, err := queries.InsertRaw(ctx, InsertRawParams{
			LakeSlug:              l.Slug,
			MeasuredAt:            w.MeasuredAt,
			Temperature2mC:        w.TempC,
			RelativeHumidity2mPct: w.HumidityPct,
			WindSpeed10mKmh:       w.WindKMH,
			WindDirection10mDeg:   w.WindDirectionDeg,
			WeatherCode:           w.WeatherCode,
			IsDay:                 w.IsDay,
			RawPayload:            raw,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			// already have this measurement
			continue
		}
		if err != nil {
			rlog.Error("generic insert raw", "lake", l.Slug, "err", err)
			continue
		}

		out = append(out, adapters.LakeReading{
			Lake:             l,
			Adapter:          id,
			MeasuredAt:       w.MeasuredAt,
			AirTempC:         w.TempC,
			HumidityPct:      w.HumidityPct,
			WindSpeedKMH:     w.WindKMH,
			WindDirectionDeg: w.WindDirectionDeg,
			WeatherCode:      w.WeatherCode,
			IsDay:            w.IsDay,
			RawID:            rawID,
		})
	}
	return out, nil
}
