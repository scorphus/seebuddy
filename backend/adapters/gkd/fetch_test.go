package gkd

import (
	_ "embed"
	"testing"
	"time"
)

//go:embed testdata/starnberg-messwerte.html
var starnbergFixture []byte

func TestParseTable_Starnberg(t *testing.T) {
	t.Parallel()

	rows, err := parseTable(starnbergFixture)
	if err != nil {
		t.Fatalf("parseTable: %v", err)
	}
	if len(rows) != 7 {
		t.Fatalf("got %d rows, want 7 (mini-table)", len(rows))
	}

	// Newest row in the captured fixture: 02.05.2026 22:45 (Europe/Berlin),
	// 12.4 °C. CEST is UTC+2 in May, so UTC is 20:45.
	got := rows[0]
	wantUTC := time.Date(2026, 5, 2, 20, 45, 0, 0, time.UTC)
	if !got.MeasuredAt.Equal(wantUTC) {
		t.Errorf("MeasuredAt = %v, want %v", got.MeasuredAt, wantUTC)
	}
	if got.WaterTempC == nil || *got.WaterTempC != 12.4 {
		t.Errorf("WaterTempC = %v, want 12.4", got.WaterTempC)
	}
	if got.RawHTML == "" {
		t.Error("RawHTML is empty")
	}

	// All rows should have a parseable temperature in this fixture.
	for i, r := range rows {
		if r.WaterTempC == nil {
			t.Errorf("row %d: WaterTempC = nil", i)
		}
	}
}

func TestParseTable_NoTable(t *testing.T) {
	t.Parallel()

	_, err := parseTable([]byte(`<html><body><h1>404</h1></body></html>`))
	if err == nil {
		t.Fatal("expected error when wassertemperatur table missing")
	}
}

func TestParseTable_SkipsSensorGap(t *testing.T) {
	t.Parallel()

	// Synthetic table with one valid row and one gap row (em-dash).
	html := `<html><body>
<table class="tblsort">
<caption>Wassertemperatur Werte</caption>
<thead><tr><th>Datum</th><th class="center">Wassertemperatur [°C]</th></tr></thead>
<tbody>
<tr class="row"><td>02.05.2026 22:45</td><td class="center">14,2</td></tr>
<tr class="row2"><td>02.05.2026 22:30</td><td class="center">–</td></tr>
</tbody>
</table>
</body></html>`

	rows, err := parseTable([]byte(html))
	if err != nil {
		t.Fatalf("parseTable: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("got %d rows, want 2", len(rows))
	}
	if rows[0].WaterTempC == nil || *rows[0].WaterTempC != 14.2 {
		t.Errorf("row 0 WaterTempC = %v, want 14.2", rows[0].WaterTempC)
	}
	if rows[1].WaterTempC != nil {
		t.Errorf("row 1 WaterTempC = %v, want nil (sensor gap)", rows[1].WaterTempC)
	}
}

func TestSplitSensorID(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in      string
		gkdID   string
		slug    string
		basin   string
		wantErr bool
	}{
		{"isar/starnberg-16663002", "16663002", "starnberg", "isar", false},
		{"isar/gmund_tegernsee-18201303", "18201303", "gmund_tegernsee", "isar", false},
		{"iller_lech/lindau-20001001", "20001001", "lindau", "iller_lech", false},
		{"missing-separators", "", "", "", true},
		{"basin/no-id-", "", "", "", true},
		{"", "", "", "", true},
	}
	for _, c := range cases {
		gkdID, slug, basin, err := splitSensorID(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("splitSensorID(%q) err=%v, wantErr=%v", c.in, err, c.wantErr)
			continue
		}
		if !c.wantErr {
			if gkdID != c.gkdID || slug != c.slug || basin != c.basin {
				t.Errorf("splitSensorID(%q) = (%q, %q, %q), want (%q, %q, %q)",
					c.in, gkdID, slug, basin, c.gkdID, c.slug, c.basin)
			}
		}
	}
}

func TestStationDisplayName(t *testing.T) {
	t.Parallel()

	cases := map[string]string{
		"starnberg":       "Starnberg",
		"gmund_tegernsee": "Gmund Tegernsee",
		"":                "",
	}
	for in, want := range cases {
		if got := stationDisplayName(in); got != want {
			t.Errorf("stationDisplayName(%q) = %q, want %q", in, got, want)
		}
	}
}
