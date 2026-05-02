// Package adapters defines the contract every data-source adapter implements.
//
// An adapter knows which lakes it covers, how often to poll its upstream,
// how to fetch and parse readings, and how to persist its own raw payloads.
// The lakes service iterates the registered adapters from a cron job and
// stores the normalized LakeReading values they return into the readings
// table.
package adapters

import (
	"context"
	"time"
)

// Lake is the metadata an adapter knows about a lake it covers.
type Lake struct {
	Slug     string
	Name     string
	Region   string
	Lat      float64
	Lon      float64
	SensorID string // adapter-specific upstream identifier, empty when not applicable
}

// LakeReading is one normalized observation produced by an adapter for a lake.
// Fields not provided by the upstream are left as nil.
type LakeReading struct {
	Lake
	Adapter          string
	MeasuredAt       time.Time
	WaterTempC       *float64
	AirTempC         *float64
	HumidityPct      *float64
	WindSpeedKMH     *float64
	WindDirectionDeg *int32 // compass degrees the wind blows FROM (0=N, 90=E, 180=S, 270=W)
	WeatherCode      *int32
	IsDay            *bool
	RawID            int64 // pointer into the adapter's own raw table; 0 if not stored
}

// Adapter is the interface every data source implements.
type Adapter interface {
	// ID returns a stable identifier ("wachplan", "generic", ...).
	ID() string

	// Lakes returns the metadata of every lake this adapter covers.
	// The returned slice is the source of truth for those lakes; lakes
	// service uses it to render the catalog regardless of whether a
	// reading has been persisted yet.
	Lakes() []Lake

	// Tick is called frequently by the central cron. It returns an empty
	// slice when the adapter's own period has not yet elapsed since the
	// last successful fetch; otherwise it fetches, persists raw, and
	// returns one LakeReading per lake successfully read.
	Tick(ctx context.Context) ([]LakeReading, error)
}
