package lakes

import (
	"context"
	"crypto/subtle"
	"sync"
	"time"

	"encore.dev/beta/errs"
	"encore.dev/cron"
	"encore.dev/rlog"

	"github.com/scorphus/seebudy/backend/adapters"
	"github.com/scorphus/seebudy/backend/adapters/gkd"
	"github.com/scorphus/seebudy/backend/adapters/openmeteo"
	"github.com/scorphus/seebudy/backend/adapters/wachplan"
)

var secrets struct {
	// PollToken is the shared secret the external trigger (Cloudflare Worker)
	// presents in the X-Poll-Token header. Set via `encore secret set PollToken`.
	PollToken string
}

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

// PollParams carries the shared-secret header the external trigger
// (Cloudflare Worker) sends to authorize a poll cycle.
type PollParams struct {
	Token string `header:"X-Poll-Token"`
}

// PollExternal is the public, secret-protected wrapper around Poll. The Encore
// free tier caps cron jobs at one execution per hour, so a Cloudflare Worker
// hits this endpoint every 15 minutes for fresher data; the internal cron
// remains as an hourly fallback.
//
//encore:api public method=POST path=/lakes/poll
func PollExternal(ctx context.Context, p *PollParams) error {
	if subtle.ConstantTimeCompare([]byte(p.Token), []byte(secrets.PollToken)) != 1 {
		return &errs.Error{Code: errs.PermissionDenied, Message: "invalid token"}
	}
	return Poll(ctx)
}

// pollAdapters drives every adapter through one Tick concurrently and stores
// any readings each one returns. A failing adapter logs and is skipped — the
// others run. Each adapter has its own database and Tick is idempotent, so
// fan-out is safe; pgxpool and rlog are concurrency-safe.
func pollAdapters(ctx context.Context, all []registeredEntry, store func(context.Context, adapters.LakeReading) error) {
	totalStart := time.Now()
	var wg sync.WaitGroup
	for _, e := range all {
		wg.Add(1)
		go func(e registeredEntry) {
			defer wg.Done()
			startedAt := time.Now()
			rlog.Info("adapter tick start", "adapter", e.id, "elapsed_ms_since_total_start", time.Since(totalStart).Milliseconds())
			resp, err := e.tick(ctx)
			tickDur := time.Since(startedAt)
			if err != nil {
				rlog.Error("adapter tick", "adapter", e.id, "tick_ms", tickDur.Milliseconds(), "err", err)
				return
			}
			rlog.Info("adapter tick done", "adapter", e.id, "tick_ms", tickDur.Milliseconds(), "readings", func() int {
				if resp == nil {
					return 0
				}
				return len(resp.Readings)
			}())
			if resp == nil {
				return
			}
			for _, r := range resp.Readings {
				if err := store(ctx, r); err != nil {
					rlog.Error("store reading", "adapter", e.id, "lake", r.Slug, "err", err)
				}
			}
			rlog.Info("adapter done", "adapter", e.id, "total_ms", time.Since(startedAt).Milliseconds())
		}(e)
	}
	wg.Wait()
	rlog.Info("pollAdapters done", "total_ms", time.Since(totalStart).Milliseconds())
}
