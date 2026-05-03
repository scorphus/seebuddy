// Package gkd adapts the Gewässerkundlicher Dienst Bayern (gkd.bayern.de),
// the official Bavarian state water-temperature service operated by the
// Landesamt für Umwelt, to the seebuddy adapter contract. It
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
	"crypto/subtle"
	"errors"
	"fmt"
	"strings"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/scorphus/seebuddy/backend/adapters"
)

const id = "gkd"

var db = sqldb.Driver[*pgxpool.Pool](sqldb.NewDatabase("gkd", sqldb.DatabaseConfig{
	Migrations: "./migrations",
}))

var queries = New(db)

var secrets struct {
	// PollToken authorises the Cloudflare Worker that proxies GKD HTML
	// into /gkd/ingest. We share the same secret as lakes.PollExternal so
	// there's a single token to rotate.
	PollToken string
}

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

// Tick emits the latest stored reading per GKD station. It does not contact
// gkd.bayern.de — that upstream silently drops Encore Cloud's egress IPs.
// Fetching is delegated to the Cloudflare Worker, which POSTs raw HTML into
// Ingest below, and Tick reads from gkd_raw to build LakeReadings.
//
//encore:api
func Tick(ctx context.Context) (*adapters.TickResponse, error) {
	rows, err := queries.LatestPerStation(ctx)
	if err != nil {
		return nil, fmt.Errorf("gkd latest per station: %w", err)
	}
	byStation := make(map[string]LatestPerStationRow, len(rows))
	for _, r := range rows {
		byStation[r.StationID] = r
	}

	out := make([]adapters.LakeReading, 0, len(lakes))
	for _, l := range lakes {
		stationID, _, _, err := splitSensorID(l.SensorID)
		if err != nil {
			rlog.Error("gkd parse sensor_id", "lake", l.Slug, "sensor_id", l.SensorID, "err", err)
			continue
		}
		row, ok := byStation[stationID]
		if !ok || row.WaterTempC == nil {
			continue
		}
		out = append(out, adapters.LakeReading{
			Lake:       l,
			Adapter:    id,
			MeasuredAt: row.MeasuredAt,
			WaterTempC: row.WaterTempC,
		})
	}
	return &adapters.TickResponse{Readings: out}, nil
}

// IngestParams carries one station's HTML mini-table plus the shared poll
// token. SensorID follows our "{basin}/{slug}-{id}" convention so the server
// can derive station components without trusting the caller to split them.
type IngestParams struct {
	Token    string `header:"X-Poll-Token"`
	SensorID string `json:"sensor_id"`
	HTML     string `json:"html"`
}

// IngestResponse reports how many newly parsed rows landed in gkd_raw. Rows
// already present (UNIQUE conflict) are silently ignored so the worker can
// re-post the same HTML without bookkeeping.
type IngestResponse struct {
	Inserted int `json:"inserted"`
	Parsed   int `json:"parsed"`
}

// Ingest accepts a single station's rendered HTML from a non-blocked egress
// (the Cloudflare Worker), parses the Wassertemperatur table, and stores
// every row. It is the only writer to gkd_raw.
//
//encore:api public method=POST path=/gkd/ingest
func Ingest(ctx context.Context, p *IngestParams) (*IngestResponse, error) {
	if subtle.ConstantTimeCompare([]byte(p.Token), []byte(secrets.PollToken)) != 1 {
		return nil, &errs.Error{Code: errs.PermissionDenied, Message: "invalid token"}
	}
	stationID, stationSlug, stationBasin, err := splitSensorID(p.SensorID)
	if err != nil {
		return nil, &errs.Error{Code: errs.InvalidArgument, Message: fmt.Sprintf("bad sensor_id: %v", err)}
	}
	parsed, err := parseTable([]byte(p.HTML))
	if err != nil {
		return nil, fmt.Errorf("parse: %w", err)
	}

	resp := &IngestResponse{Parsed: len(parsed)}
	for j := range parsed {
		r := &parsed[j]
		id, err := queries.InsertRaw(ctx, InsertRawParams{
			StationID:    stationID,
			StationSlug:  stationSlug,
			StationBasin: stationBasin,
			StationName:  stationDisplayName(stationSlug),
			WaterTempC:   r.WaterTempC,
			MeasuredAt:   r.MeasuredAt,
			RawRowHtml:   ptrString(r.RawHTML),
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			rlog.Error("gkd insert raw", "sensor_id", p.SensorID, "measured_at", r.MeasuredAt, "err", err)
			continue
		}
		if id > 0 {
			resp.Inserted++
		}
	}
	return resp, nil
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
