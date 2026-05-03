package lakes

import (
	"context"
	"errors"
	"testing"

	"github.com/scorphus/seebudy/backend/adapters"
)

type fakeAdapter struct {
	id       string
	err      error
	readings []adapters.LakeReading
}

func (f *fakeAdapter) ID() string             { return f.id }
func (f *fakeAdapter) Lakes() []adapters.Lake { return nil }
func (f *fakeAdapter) Tick(_ context.Context) ([]adapters.LakeReading, error) {
	return f.readings, f.err
}

func TestPollAdapters_OneFailsOthersContinue(t *testing.T) {
	t.Parallel()

	failing := &fakeAdapter{id: "boom", err: errors.New("upstream down")}
	working := &fakeAdapter{
		id: "ok",
		readings: []adapters.LakeReading{
			{Lake: adapters.Lake{Slug: "lake-1"}, Adapter: "ok"},
			{Lake: adapters.Lake{Slug: "lake-2"}, Adapter: "ok"},
		},
	}

	var stored []adapters.LakeReading
	store := func(_ context.Context, r adapters.LakeReading) error {
		stored = append(stored, r)
		return nil
	}

	pollAdapters(context.Background(), []adapters.Adapter{failing, working}, store)

	if len(stored) != 2 {
		t.Fatalf("stored %d readings, want 2", len(stored))
	}
	if stored[0].Slug != "lake-1" || stored[1].Slug != "lake-2" {
		t.Errorf("stored slugs = [%q, %q], want [lake-1, lake-2]", stored[0].Slug, stored[1].Slug)
	}
}

func TestPollAdapters_StoreErrorDoesNotStopOthers(t *testing.T) {
	t.Parallel()

	a := &fakeAdapter{
		id: "a",
		readings: []adapters.LakeReading{
			{Lake: adapters.Lake{Slug: "first"}, Adapter: "a"},
			{Lake: adapters.Lake{Slug: "second"}, Adapter: "a"},
		},
	}

	var attempts []string
	store := func(_ context.Context, r adapters.LakeReading) error {
		attempts = append(attempts, r.Slug)
		if r.Slug == "first" {
			return errors.New("db conflict")
		}
		return nil
	}

	pollAdapters(context.Background(), []adapters.Adapter{a}, store)

	if len(attempts) != 2 {
		t.Fatalf("attempts = %v, want both readings to be attempted", attempts)
	}
}
