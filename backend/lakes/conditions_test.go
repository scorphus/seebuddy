package lakes

import (
	"testing"
	"time"

	"github.com/scorphus/muenchner-see-buddy/backend/adapters"
)

func TestBuildView_NoReadings(t *testing.T) {
	t.Parallel()

	lakes := []adapters.Lake{
		{Slug: "a", Name: "Lake A"},
		{Slug: "b", Name: "Lake B"},
	}
	views := buildView(lakes, nil, time.Now())

	if len(views) != 2 {
		t.Fatalf("got %d views, want 2", len(views))
	}
	for _, v := range views {
		if v.Latest != nil {
			t.Errorf("lake %q: Latest = %+v, want nil", v.Slug, v.Latest)
		}
	}
}

func TestBuildView_FreshAndStale(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 1, 15, 0, 0, 0, time.UTC)

	lakes := []adapters.Lake{
		{Slug: "fresh", Name: "Fresh Lake"},
		{Slug: "stale", Name: "Stale Lake"},
		{Slug: "empty", Name: "No Reading Lake"},
	}
	temp := 14.56
	rows := []LatestReadingPerLakeRow{
		{LakeSlug: "fresh", Adapter: "wachplan", MeasuredAt: now.Add(-30 * time.Minute), WaterTempC: &temp},
		{LakeSlug: "stale", Adapter: "wachplan", MeasuredAt: now.Add(-3 * time.Hour), WaterTempC: &temp},
	}
	views := buildView(lakes, rows, now)

	if len(views) != 3 {
		t.Fatalf("got %d views, want 3", len(views))
	}

	bySlug := map[string]LakeView{}
	for _, v := range views {
		bySlug[v.Slug] = v
	}

	if v := bySlug["fresh"]; v.Latest == nil || v.Latest.Stale {
		t.Errorf("fresh: stale=%v, want false", v.Latest)
	}
	if v := bySlug["fresh"]; v.Latest.AgeSeconds != int64((30 * time.Minute).Seconds()) {
		t.Errorf("fresh: AgeSeconds = %d, want %d", v.Latest.AgeSeconds, int64((30 * time.Minute).Seconds()))
	}
	if v := bySlug["stale"]; v.Latest == nil || !v.Latest.Stale {
		t.Errorf("stale: Latest = %+v, want stale=true", v.Latest)
	}
	if v := bySlug["empty"]; v.Latest != nil {
		t.Errorf("empty: Latest = %+v, want nil", v.Latest)
	}
}
