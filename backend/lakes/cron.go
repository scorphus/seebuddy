package lakes

import (
	"context"

	"encore.dev/cron"
	"encore.dev/rlog"

	"github.com/scorphus/seebudy/backend/adapters"
	"github.com/scorphus/seebudy/backend/adapters/gkd"
	"github.com/scorphus/seebudy/backend/adapters/openmeteo"
	"github.com/scorphus/seebudy/backend/adapters/wachplan"
)

// catalog is every adapter exposed only for its catalog metadata (Lakes()).
// Tick is no longer on the Adapter interface because each adapter package is
// its own Encore service so Encore provisions its database; cross-service
// calls go through the package-level //encore:api Tick endpoints called
// directly from Poll below.
var catalog = []adapters.Adapter{
	wachplan.Adapter{},
	gkd.Adapter{},
	openmeteo.Adapter{},
}

var _ = cron.NewJob("poll", cron.JobConfig{
	Title:    "Poll all adapters",
	Every:    5 * cron.Minute,
	Endpoint: Poll,
})

// tickFn produces readings for one adapter. We can't store //encore:api
// endpoints as function values in a registry — Encore rejects that with
// E1387 — so Poll calls each adapter's Tick by name and we hand the closures
// to pollAdapters as ordinary Go values.
type tickFn func(context.Context) (*adapters.TickResponse, error)

type registeredEntry struct {
	id   string
	tick tickFn
}

//encore:api
func Poll(ctx context.Context) error {
	pollAdapters(ctx, []registeredEntry{
		{id: "wachplan", tick: func(ctx context.Context) (*adapters.TickResponse, error) { return wachplan.Tick(ctx) }},
		{id: "gkd", tick: func(ctx context.Context) (*adapters.TickResponse, error) { return gkd.Tick(ctx) }},
		{id: "openmeteo", tick: func(ctx context.Context) (*adapters.TickResponse, error) { return openmeteo.Tick(ctx) }},
	}, storeReading)
	return nil
}

// pollAdapters drives every adapter through one Tick and stores any readings
// each one returns. A failing adapter logs and is skipped — the others run.
func pollAdapters(ctx context.Context, all []registeredEntry, store func(context.Context, adapters.LakeReading) error) {
	for _, e := range all {
		resp, err := e.tick(ctx)
		if err != nil {
			rlog.Error("adapter tick", "adapter", e.id, "err", err)
			continue
		}
		if resp == nil {
			continue
		}
		for _, r := range resp.Readings {
			if err := store(ctx, r); err != nil {
				rlog.Error("store reading", "adapter", e.id, "lake", r.Slug, "err", err)
			}
		}
	}
}
