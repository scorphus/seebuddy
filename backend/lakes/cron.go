package lakes

import (
	"context"

	"encore.dev/cron"
	"encore.dev/rlog"

	"github.com/scorphus/muenchner-see-buddy/backend/adapters"
	"github.com/scorphus/muenchner-see-buddy/backend/adapters/generic"
	"github.com/scorphus/muenchner-see-buddy/backend/adapters/gkd"
	"github.com/scorphus/muenchner-see-buddy/backend/adapters/wachplan"
)

// registered is the list of every adapter the central cron iterates over.
// Adding a new adapter = +1 import and +1 entry here.
var registered = []adapters.Adapter{
	wachplan.Adapter{},
	gkd.Adapter{},
	generic.Adapter{},
}

var _ = cron.NewJob("poll", cron.JobConfig{
	Title:    "Poll all adapters",
	Every:    5 * cron.Minute,
	Endpoint: Poll,
})

//encore:api
func Poll(ctx context.Context) error {
	pollAdapters(ctx, registered, storeReading)
	return nil
}

// pollAdapters drives every adapter through one Tick and stores any readings
// each one returns. A failing adapter logs and is skipped — the others run.
func pollAdapters(ctx context.Context, all []adapters.Adapter, store func(context.Context, adapters.LakeReading) error) {
	for _, a := range all {
		readings, err := a.Tick(ctx)
		if err != nil {
			rlog.Error("adapter tick", "adapter", a.ID(), "err", err)
			continue
		}
		for _, r := range readings {
			if err := store(ctx, r); err != nil {
				rlog.Error("store reading", "adapter", a.ID(), "lake", r.Slug, "err", err)
			}
		}
	}
}
