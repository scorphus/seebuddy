package lakes

import (
	"context"
	"errors"
	"testing"

	"github.com/scorphus/seebuddy/backend/adapters"
)

func entry(id string, readings []adapters.LakeReading, err error) registeredEntry {
	return registeredEntry{
		id: id,
		tick: func(_ context.Context) (*adapters.TickResponse, error) {
			if err != nil {
				return nil, err
			}
			return &adapters.TickResponse{Readings: readings}, nil
		},
	}
}

func TestPollAdapters_OneFailsOthersContinue(t *testing.T) {
	t.Parallel()

	failing := entry("boom", nil, errors.New("upstream down"))
	working := entry("ok", []adapters.LakeReading{
		{Lake: adapters.Lake{Slug: "lake-1"}, Adapter: "ok"},
		{Lake: adapters.Lake{Slug: "lake-2"}, Adapter: "ok"},
	}, nil)

	var stored []adapters.LakeReading
	store := func(_ context.Context, r adapters.LakeReading) error {
		stored = append(stored, r)
		return nil
	}

	pollAdapters(context.Background(), []registeredEntry{failing, working}, store)

	if len(stored) != 2 {
		t.Fatalf("stored %d readings, want 2", len(stored))
	}
	if stored[0].Slug != "lake-1" || stored[1].Slug != "lake-2" {
		t.Errorf("stored slugs = [%q, %q], want [lake-1, lake-2]", stored[0].Slug, stored[1].Slug)
	}
}

func TestPollAdapters_StoreErrorDoesNotStopOthers(t *testing.T) {
	t.Parallel()

	a := entry("a", []adapters.LakeReading{
		{Lake: adapters.Lake{Slug: "first"}, Adapter: "a"},
		{Lake: adapters.Lake{Slug: "second"}, Adapter: "a"},
	}, nil)

	var attempts []string
	store := func(_ context.Context, r adapters.LakeReading) error {
		attempts = append(attempts, r.Slug)
		if r.Slug == "first" {
			return errors.New("db conflict")
		}
		return nil
	}

	pollAdapters(context.Background(), []registeredEntry{a}, store)

	if len(attempts) != 2 {
		t.Fatalf("attempts = %v, want both readings to be attempted", attempts)
	}
}
