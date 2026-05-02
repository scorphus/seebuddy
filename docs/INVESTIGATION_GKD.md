# Investigation: gkd.bayern.de

**Date:** 2026-05-02
**Investigator:** muenchner-see-buddy/0.1
**Total requests made:** ~14 (well under 30-page cap)
**User-Agent used:** `muenchner-see-buddy/0.1 (+https://github.com/scorphus/muenchner-see-buddy)`

Adapter v0.4 of `muenchner-see-buddy`. Covers the large alpine/foothill lakes via the Gewässerkundlicher Dienst Bayern (official, LfU).

## Executive Summary

GKD does **not** expose a JSON or CSV endpoint we are allowed to use — the `graphik.php` chart generator and the `/downloadcenter/` exports both live under `/webservices/` and `/downloadcenter/`, which are explicitly disallowed in `robots.txt`.

The good news: **the public HTML pages contain everything we need**, are fresh (latest readings within ~30 min for active stations), and are explicitly licensed CC BY 4.0. Two HTML pages give us two flavours of access:

1. `…/messwerte/tabelle` — full **7-day** week, ~700 rows of 15-minute readings, perfect for backfilling.
2. `…/messwerte` (the diagram page) — a small **inline 7-row table** with the most recent readings, perfect for cheap polling. The table is server-rendered into the HTML — no XHR needed.

Of the 6 candidate lakes, **5 have a GKD station**. **Kochelsee has no GKD water-temperature station** at all (only a river-level pegel `kochel-16407002` on the Loisach, which carries no temperature timeseries) — we need an alternative source for it (see `INVESTIGATION_OTHER_SOURCES.md`).

The `/seen/wassertemperatur` index page bonuses us with a JavaScript blob (`LfUMap.init({...})`) that lists every Bavarian lake-temperature station with `lat`, `lon`, last reading, and last value — useful as a one-shot inventory and freshness probe, no per-station GETs required.

## URL Structure

```
https://www.gkd.bayern.de/de/seen/wassertemperatur/{flussgebiet}/{slug}-{id}                # Stammdaten (master data)
https://www.gkd.bayern.de/de/seen/wassertemperatur/{flussgebiet}/{slug}-{id}/messwerte       # Diagram + inline mini-table (7 rows)
https://www.gkd.bayern.de/de/seen/wassertemperatur/{flussgebiet}/{slug}-{id}/messwerte/tabelle  # Full 7-day table (~700 rows)
https://www.gkd.bayern.de/de/seen/wassertemperatur/{flussgebiet}/{slug}-{id}/jahreswerte    # Year aggregate
https://www.gkd.bayern.de/de/seen/wassertemperatur/{flussgebiet}/{slug}-{id}/gesamtzeitraum # Full record
https://www.gkd.bayern.de/de/seen/wassertemperatur/{flussgebiet}/{slug}-{id}/download       # → /downloadcenter (DO NOT use, robots-blocked)
```

## 1. Hidden API / XHR — none usable

Inspected the diagram page `…/starnberg-16663002/messwerte` for any data URL.

- **No XHR/fetch endpoints.** The chart is a server-rendered PNG, the table is rendered server-side in the same HTML.
- The PNG comes from `https://www.gkd.bayern.de/webservices/graphik.php?statnr={id}&thema=gkd.woche&wert=wassertemperatur&begin={DD.MM.YYYY}&end={DD.MM.YYYY}` — a chart image, not data.
- Both `/webservices/` and `/downloadcenter/enqueue_download` are **disallowed by `/robots.txt`** (verified, see §5). We don't probe them.

**Conclusion:** No JSON API, no CSV, no clean XHR. We scrape the HTML.

## 2. Query parameters on `/messwerte` and `/messwerte/tabelle`

Verified by fetching with different querystrings against `…/starnberg-16663002`:

| Param                     | Effect on `/messwerte/tabelle`                            | Effect on `/messwerte` |
| ------------------------- | --------------------------------------------------------- | ---------------------- |
| (none)                    | Last 7 days, ~700 rows                                    | Last 7 days, but **only 7 rows** rendered in HTML table |
| `?beginn=DD.MM.YYYY&ende=DD.MM.YYYY` | Window honored, returns rows in that window     | Same window, still only 7 rows shown                    |
| `?zr=tag`                 | No visible change (still ~700 rows / 7 days)              | Reduces tail, still ~7 rows                              |
| `?zr=monat`               | No visible change                                         | Same — still 7 rows                                      |
| `?dir=prev&start=DD.MM.YYYY` | Pagination: returns the week ending at `start` (used by "Zurückblättern" link) | Same                                                     |

Date format is **`DD.MM.YYYY`** (German), e.g. `25.04.2026`. The form on the diagram page uses `name="beginn"` and `name="ende"` — confirm those are the canonical names.

**Practical takeaway:**
- For routine polling, hit `…/messwerte` — it's ~22 KB and contains the freshest readings (7 rows × 15 min ≈ last 1h45m).
- For backfill / catch-up after downtime, hit `…/messwerte/tabelle` — ~78 KB and gives the full week.
- For arbitrary historical windows, add `?beginn=…&ende=…` on either page.

### Freshness sanity check (2026-05-02 ~17:30 UTC test)

Pulled `…/starnberg-16663002/messwerte/tabelle`: latest row was **02.05.2026 17:45** (today, ~live). The earlier note about "07.04–14.04" stale data was specific to whatever fetch was made in the original plan — not a systemic issue. **The data is fresh.**

Caveat: not all stations are equally fresh. Tegernsee (`gmund_tegernsee-18201303`) on the same day showed last row at **02.05.2026 11:00** — about 6 h stale. Possible sensor / transmission issue at that station; flag it in the adapter and rely on staleness checks per-station, not a global SLA.

## 3. Confirmed station table

Pulled the index page `…/de/seen/wassertemperatur` once. Embedded inside the HTML is a `LfUMap.init({"pointer": [...]})` JS call with the full list of 21 Bavarian lake-temperature stations, including coordinates and last readings. This is the single source of truth.

Stations relevant to the Munich area:

| Lake             | Station name        | `gkd_id`     | `gkd_slug`         | `gkd_basin`  | lat       | lon       | Sample reading (2026-05-02) |
| ---------------- | ------------------- | ------------ | ------------------ | ------------ | --------- | --------- | ---------------------------- |
| Starnberger See  | Starnberg           | `16663002`   | `starnberg`        | `isar`       | 47.9973   | 11.3493   | 12.6 °C @ 17:30             |
| Ammersee         | Stegen              | `16602008`   | `stegen`           | `isar`       | 48.0764   | 11.1309   | 13.1 °C @ 17:30             |
| Ammersee (extra) | Ammerseeboje        | `16601050`   | `ammerseeboje`     | `isar`       | 47.9811   | 11.1227   | 13.4 °C @ 20.10.2025 (offline?) |
| Chiemsee         | Stock               | `18400503`   | `stock`            | `isar`       | 47.8612   | 12.3658   | 15.4 °C @ 17:30             |
| Tegernsee        | Gmund_Tegernsee     | `18201303`   | `gmund_tegernsee`  | `isar`       | 47.7478   | 11.7354   | 10.8 °C @ 11:00 (stale ~6 h) |
| **Kochelsee**    | — (no station)      | —            | —                  | —            | —         | —         | **NOT COVERED BY GKD**      |
| Bodensee         | Lindau              | `20001001`   | `lindau`           | `iller_lech` | 47.5446   | 9.6850    | 14.2 °C @ 17:30             |

Bonus stations from the same source that may be of interest later:

| Lake          | Station          | `gkd_id`   | `gkd_slug`         | `gkd_basin`  | lat     | lon     |
| ------------- | ---------------- | ---------- | ------------------ | ------------ | ------- | ------- |
| Pilsensee     | Pilsensee        | `16628055` | `pilsensee`        | `isar`       | 48.0243 | 11.1895 |
| Wörthsee      | Wörthsee         | `16651003` | `woerthsee`        | `isar`       | 48.0512 | 11.1617 |
| Schliersee    | Schliersee       | `18222008` | `schliersee`       | `isar`       | 47.7306 | 11.8680 |
| Waginger See  | Buchwinkel       | `18682507` | `buchwinkel`       | `isar`       | 47.9318 | 12.7772 |
| Königssee     | Königssee        | `18624806` | `koenigssee`       | `inn`        | 47.5872 | 12.9902 |
| Abtsdorfer S. | Seethal          | `18673955` | `seethal`          | `inn`        | 47.9120 | 12.9104 |
| Weitsee       | Seegatterl       | `18486509` | `seegatterl`       | `isar`       | 47.6870 | 12.5687 |

Note: `gkd_basin` here is the URL-path segment (`isar`, `iller_lech`, `inn`, `kelheim`, `main_unten`, `passau`). It corresponds to the Flussgebiet / region grouping used by the GKD navigation, not strictly to the river the lake drains into (e.g. Bodensee → `iller_lech`, Tegernsee → `isar`).

### Verified Stammdaten URL

Spot-checked `…/de/seen/wassertemperatur/isar/starnberg-16663002` — returns 200 with `<title>Wassertemperatur: Stammdaten Starnberg / StarnbergerSee</title>`. The page contains UTM (ETRS89/UTM 32N) Ostwert/Nordwert and Geländehöhe; the LfUMap.init JSON in the index page already carries WGS84 lat/lon, so we don't need the per-station Stammdaten fetch for coordinates.

## 4. Coordinates source

**Use the `LfUMap.init({...})` JSON** embedded in `https://www.gkd.bayern.de/de/seen/wassertemperatur` (or `…/tabellen`) for catalog metadata (`lat`, `lon`, last `d`atetime, last `w`ater value). No per-station fetch required.

The blob has shape:
```js
LfUMap.init({"pointer": [
  {"p": "16663002", "n": "Starnberg",
   "uri": "https://www.gkd.bayern.de/de/seen/wassertemperatur/isar/starnberg-16663002",
   "lon": "11.3493", "lat": "47.9973",
   "g": "StarnbergerSee", "d": "02.05.2026 17:30", "w": "12,6", ...},
  ...
]})
```

Adapter idea: a small Go function that GETs the index once at startup, runs a regex like `LfUMap\.init\((\{.+?\})\);` on the body, `json.Unmarshal`s it, and yields a station catalog. Refresh weekly.

## 5. Robots.txt and licensing

### `/robots.txt` (fetched 2026-05-02)

GKD ships a thoughtful robots.txt (April 2026 version). Relevant rules under `User-agent: *`:

- **Disallowed:** `/webservices/`, `/de/downloadcenter/enqueue_download`, `/de/downloadcenter/download` (and their `/en/` counterparts).
- **Allowed:** everything else, including `/de/seen/wassertemperatur/...`.
- The file explicitly states that the public-data mission justifies access by search engines and AI crawlers; it only blocks server-load-intensive endpoints. Our adapter falls cleanly within "allowed".

### License (from `/de/impressum` → "Nutzungsbedingungen der Download-Daten")

> Die Download-Daten im GKD wurden von der Fachseite her mit der "Namensnennung 4.0 International (CC BY 4.0)" lizensiert. … Einzige Bedingung ist, dass der Nutzer die Daten entsprechend kennzeichnet. Der nötige Quellenvermerk soll wie folgt gut sichtbar dargestellt werden: "Datenquelle: Bayerisches Landesamt für Umwelt, www.lfu.bayern.de".

**Required attribution string (verbatim):**

> *Datenquelle: Bayerisches Landesamt für Umwelt, www.lfu.bayern.de*

(Note: this is **CC BY 4.0**, not "Datenlizenz Deutschland 2.0 — Namensnennung" as the original plan assumed. Update the project README accordingly.)

## 6. Recommendation

### Adapter strategy: HTML scraping with `goquery`

JSON-first is not available. The cleanest path is:

1. **Bootstrap once** — GET `/de/seen/wassertemperatur`, parse the `LfUMap.init(...)` argument, build a station catalog (id, slug, basin, lat/lon, gewässer name). Cache in DB; refresh weekly.
2. **Routine polling** — for each station of interest, GET `…/messwerte` every 15-20 minutes. Parse the `<table class="tblsort">` whose header contains `Wassertemperatur`. Each `<tr class="row|row2">` has `<td>DD.MM.YYYY HH:MM</td><td>…</td><td class="center">14,2</td>`.
3. **Backfill** — on first run / after downtime, GET `…/messwerte/tabelle` to grab the last 7 days in one shot. For older windows, add `?beginn=…&ende=…`.
4. **Parsing details** (critical):
   - Use `golang.org/x/net/html` or `github.com/PuerkitoBio/goquery`. **Never regex HTML.**
   - Date: `time.ParseInLocation("02.01.2006 15:04", v, europeBerlin)` — Europe/Berlin handles CET/CEST DST. Convert to UTC before storing.
   - Temperature: `strings.Replace(v, ",", ".", 1)` then `strconv.ParseFloat`.
   - Skip rows where the temperature cell is `–` / empty (sensor gaps).
5. **Polite client**:
   - User-Agent: `muenchner-see-buddy/0.1 (+https://github.com/scorphus/muenchner-see-buddy)`.
   - Sleep ≥ 2 s between GETs to `gkd.bayern.de`.
   - Conditional GET (`If-Modified-Since`) — they don't return `Last-Modified` on the HTML pages, so this is a no-op for now; revisit after observing in production.
   - Total polling: 5 lakes × 1 GET / 15 min = 480 GET/day to GKD. Acceptable.
6. **Per-station staleness flag** — store `last_seen_at`. If > 2 h old, fall back to a secondary source for that lake (Wachplan / OWS-tourism / …) — see `INVESTIGATION_OTHER_SOURCES.md`.

### Kochelsee

Not covered by GKD. Either:
- skip Kochelsee entirely (option for v0.4),
- or look at non-GKD candidates (best lead: tourism web cams with embedded sensor pop-ups; see other-sources doc).

### Mandatory README attribution

Add to project README:

> *Datenquelle: Bayerisches Landesamt für Umwelt, www.lfu.bayern.de* (CC BY 4.0)

## What we did **not** do (explicitly out of scope per robots.txt)

- Never fetched `/webservices/graphik.php` (chart-rendering endpoint).
- Never enqueued a download via `/de/downloadcenter/enqueue_download` or `/de/seen/.../download`.
- No POST/PUT/DELETE.
- No authenticated requests.
