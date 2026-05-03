// Package gkd adapts the Gewässerkundlicher Dienst Bayern (gkd.bayern.de),
// the official Bavarian state water-temperature service operated by the
// Landesamt für Umwelt, to the seebudy adapter contract. It
// scrapes the public HTML mini-table at /messwerte for each station because
// GKD's structured endpoints (`/webservices/`, `/downloadcenter/`) are
// disallowed by their robots.txt. See docs/INVESTIGATION_GKD.md for the
// reasoning.
//
// Attribution required by CC BY 4.0:
//
//	Datenquelle: Bayerisches Landesamt für Umwelt, www.lfu.bayern.de
package gkd

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/scorphus/seebudy/backend/adapters"
)

const id = "gkd"

// period mirrors GKD's 15-minute reporting cadence. Going faster wastes
// upstream load with no fresher data.
const period = 15 * time.Minute

// politeDelay sits between successive station GETs to keep us well under any
// reasonable rate limit. With 6 stations this adds up to ~12 s per Tick.
const politeDelay = 2 * time.Second

var db = sqldb.Driver[*pgxpool.Pool](sqldb.NewDatabase("gkd", sqldb.DatabaseConfig{
	Migrations: "./migrations",
}))

var queries = New(db)

// gkdLake links a catalog lake to its GKD station. SensorID encodes the URL
// path segment used to locate the station: "{basin}/{slug}-{id}".
var lakes = []adapters.Lake{
	{Slug: "ammersee", Name: "Ammersee", Region: "munich-southwest",
		Lat: 47.997458, Lon: 11.165848,
		SensorID: "isar/stegen-16602008"},
	{Slug: "pilsensee", Name: "Pilsensee", Region: "munich-southwest",
		Lat: 48.024241, Lon: 11.189486,
		SensorID: "isar/pilsensee-16628055"},
	{Slug: "schliersee", Name: "Schliersee", Region: "munich-south",
		Lat: 47.731727, Lon: 11.863647,
		SensorID: "isar/schliersee-18222008"},
	{Slug: "starnberger", Name: "Starnberger See", Region: "munich-south",
		Lat: 47.995538, Lon: 11.345372,
		SensorID: "isar/starnberg-16663002"},
	{Slug: "tegernsee", Name: "Tegernsee", Region: "munich-south",
		Lat: 47.699499, Lon: 11.758345,
		SensorID: "isar/gmund_tegernsee-18201303"},
	{Slug: "woerthsee", Name: "Wörthsee", Region: "munich-southwest",
		Lat: 48.067428, Lon: 11.199795,
		SensorID: "isar/woerthsee-16651003"},
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
	for i, l := range lakes {
		if i > 0 {
			select {
			case <-ctx.Done():
				return out, ctx.Err()
			case <-time.After(politeDelay):
			}
		}

		stationID, stationSlug, stationBasin, err := splitSensorID(l.SensorID)
		if err != nil {
			rlog.Error("gkd parse sensor_id", "lake", l.Slug, "sensor_id", l.SensorID, "err", err)
			continue
		}

		readings, _, err := fetch(ctx, l.SensorID)
		if err != nil {
			rlog.Error("gkd fetch", "lake", l.Slug, "err", err)
			continue
		}
		if len(readings) == 0 {
			rlog.Warn("gkd no readings", "lake", l.Slug)
			continue
		}

		// readings are ordered newest-first as rendered by GKD. We persist
		// every parsed row (idempotent via UNIQUE) so a backfill captures
		// any rows we may have missed since the last poll.
		var newest *parsedReading
		for j := range readings {
			r := &readings[j]
			if _, err := queries.InsertRaw(ctx, InsertRawParams{
				StationID:    stationID,
				StationSlug:  stationSlug,
				StationBasin: stationBasin,
				StationName:  stationDisplayName(stationSlug),
				WaterTempC:   r.WaterTempC,
				MeasuredAt:   r.MeasuredAt,
				RawRowHtml:   ptrString(r.RawHTML),
			}); err != nil && !errors.Is(err, pgx.ErrNoRows) {
				rlog.Error("gkd insert raw", "lake", l.Slug, "measured_at", r.MeasuredAt, "err", err)
				continue
			}
			if newest == nil || r.MeasuredAt.After(newest.MeasuredAt) {
				newest = r
			}
		}

		if newest == nil || newest.WaterTempC == nil {
			continue
		}

		out = append(out, adapters.LakeReading{
			Lake:       l,
			Adapter:    id,
			MeasuredAt: newest.MeasuredAt,
			WaterTempC: newest.WaterTempC,
			// GKD provides only water temperature; air/wind/etc. come from
			// the openmeteo adapter via the merge in lakes/conditions.go.
		})
	}
	return out, nil
}

// splitSensorID parses our SensorID convention "{basin}/{slug}-{id}" into
// components.
func splitSensorID(sensorID string) (gkdID, slug, basin string, err error) {
	slash := strings.Index(sensorID, "/")
	if slash < 0 {
		return "", "", "", fmt.Errorf("missing basin/slug separator")
	}
	basin = sensorID[:slash]
	rest := sensorID[slash+1:]
	dash := strings.LastIndex(rest, "-")
	if dash < 0 {
		return "", "", "", fmt.Errorf("missing slug-id separator")
	}
	slug = rest[:dash]
	gkdID = rest[dash+1:]
	if basin == "" || slug == "" || gkdID == "" {
		return "", "", "", fmt.Errorf("empty component")
	}
	return gkdID, slug, basin, nil
}

// stationDisplayName turns the URL slug into a Title-case display name.
// Underscores become spaces and the first letter is capitalised. Good enough
// for the few stations we cover; revisit if a name looks ugly.
func stationDisplayName(slug string) string {
	if slug == "" {
		return slug
	}
	clean := strings.ReplaceAll(slug, "_", " ")
	parts := strings.Fields(clean)
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
