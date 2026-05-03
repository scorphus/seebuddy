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

// Adapter is the interface every data source implements for catalog metadata.
// Tick is no longer part of the interface because each adapter owns its own
// Encore-provisioned database, and Encore only configures a database for the
// service that declares it. To get its own service context (and therefore a
// usable database pool), each adapter exposes its Tick as an //encore:api
// endpoint with the TickFunc signature below; the lakes cron pairs each
// Adapter with its TickFunc and drives them per cycle.
type Adapter interface {
	// ID returns a stable identifier ("wachplan", "gkd", "openmeteo", ...).
	ID() string

	// Lakes returns the metadata of every lake this adapter covers.
	// The returned slice is the source of truth for those lakes; the
	// lakes service uses it to render the catalog regardless of whether a
	// reading has been persisted yet.
	Lakes() []Lake
}

// TickResponse wraps the readings produced by one adapter cycle for transport
// across Encore's service boundary. Encore APIs require a struct response
// (not a slice), so we wrap.
type TickResponse struct {
	Readings []LakeReading `json:"readings"`
}

// TickFunc is the signature each adapter's //encore:api Tick endpoint
// satisfies. Lakes pairs it with the Adapter's metadata in the cron registry.
type TickFunc func(ctx context.Context) (*TickResponse, error)
