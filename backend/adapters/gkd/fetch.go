package gkd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const userAgent = "seebudy/0.1 (+https://github.com/scorphus/seebudy)"
const baseURL = "https://www.gkd.bayern.de/de/seen/wassertemperatur/"

var httpClient = &http.Client{Timeout: 15 * time.Second}

// europeBerlin is GKD's reporting timezone. The HTML renders timestamps as
// `DD.MM.YYYY HH:MM` without a TZ suffix; they are local Bavarian wall-clock
// time, which is Europe/Berlin (CET/CEST with DST).
var europeBerlin = mustLoadLocation("Europe/Berlin")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(fmt.Sprintf("load location %q: %v", name, err))
	}
	return loc
}

// parsedReading is one row extracted from the GKD table.
type parsedReading struct {
	MeasuredAt time.Time
	WaterTempC *float64
	RawHTML    string
}

// fetch GETs the /messwerte mini-table for a station and returns up to ~7
// most recent readings parsed from the HTML. The raw response body is
// returned alongside for forensics.
func fetch(ctx context.Context, sensorID string) ([]parsedReading, []byte, error) {
	url := baseURL + sensorID + "/messwerte"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("http: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, fmt.Errorf("read body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, body, fmt.Errorf("upstream status %d", resp.StatusCode)
	}

	readings, err := parseTable(body)
	return readings, body, err
}

// parseTable finds the Wassertemperatur table inside the rendered HTML and
// extracts every data row. Rows where the temperature cell isn't a parseable
// decimal (sensor gaps render as `–`) are skipped.
func parseTable(body []byte) ([]parsedReading, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	// Pick the first table.tblsort whose caption mentions Wassertemperatur.
	// (Defensive — the page only has one in practice, but the heuristic
	// guards against future layout changes.)
	var tbl *goquery.Selection
	doc.Find("table.tblsort").EachWithBreak(func(_ int, t *goquery.Selection) bool {
		caption := strings.ToLower(t.Find("caption").Text())
		header := strings.ToLower(t.Find("thead").Text())
		if strings.Contains(caption, "wassertemperatur") || strings.Contains(header, "wassertemperatur") {
			tbl = t
			return false
		}
		return true
	})
	if tbl == nil {
		return nil, fmt.Errorf("wassertemperatur table not found")
	}

	var out []parsedReading
	tbl.Find("tbody tr").Each(func(_ int, tr *goquery.Selection) {
		cells := tr.Find("td")
		if cells.Length() < 2 {
			return
		}
		dateStr := strings.TrimSpace(cells.Eq(0).Text())
		tempStr := strings.TrimSpace(cells.Eq(1).Text())
		if dateStr == "" {
			return
		}

		measuredAt, err := time.ParseInLocation("02.01.2006 15:04", dateStr, europeBerlin)
		if err != nil {
			return
		}

		var temp *float64
		if tempStr != "" && tempStr != "–" && tempStr != "-" {
			normalized := strings.Replace(tempStr, ",", ".", 1)
			if v, err := strconv.ParseFloat(normalized, 64); err == nil {
				temp = &v
			}
		}

		raw, _ := goquery.OuterHtml(tr)

		out = append(out, parsedReading{
			MeasuredAt: measuredAt.UTC(),
			WaterTempC: temp,
			RawHTML:    strings.TrimSpace(raw),
		})
	})

	return out, nil
}
