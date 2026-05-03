# Investigation: Other water-temperature sources for Munich-area lakes

**Date:** 2026-05-02
**Investigator:** seebudy/0.1
**User-Agent used:** `seebudy/0.1 (+https://github.com/scorphus/seebudy)`
**Total requests for Part B:** ~16 (cumulative for Parts A+B: ~30, well under budget).

## Why this scan

We already have:
- **Wachplan** (`sensors.mein-wachplan.de`) — JSON, 2 sensors (Lußsee, Langwieder See).
- **GKD Bayern** (`gkd.bayern.de`) — HTML, covers Starnberger See, Ammersee, Chiemsee, Tegernsee, Bodensee. **Does not cover Kochelsee.** See `INVESTIGATION_GKD.md`.
- **Open-Meteo Forecast** — air temperature only.

This document hunts for additional/redundant sources, especially for Kochelsee and as fallback when GKD stations go stale.

## Summary of candidates

| # | Source                                         | Coverage for our lakes                       | Format               | Worth building? |
| - | ---------------------------------------------- | -------------------------------------------- | -------------------- | --------------- |
| 1 | `nid.bayern.de` (Niedrigwasser-Informationsdienst) | Same as GKD (re-skin)                    | HTML                 | **No** — same backend |
| 2 | `hnd.bayern.de` (Hochwassernachrichtendienst)  | River pegels, no lake temperature             | HTML                 | **No** — wrong product |
| 3 | `pegelonline.wsv.de` (federal WSV REST API)    | Lake-adjacent rivers; **no lake temperatures** anywhere relevant | JSON REST | **No** — no temp data for our lakes |
| 4 | `opendata.dwd.de`                              | Air/weather only, no lake water temperature  | NetCDF/CSV index     | **No** — wrong product |
| 5 | Open-Meteo `marine-api` (`sea_surface_temperature`) | Returns null for inland lakes (verified Chiemsee) | JSON | **No** — null inland |
| 6 | Open-Meteo `forecast` `lake_water_temperature` variable | Variable not supported (HTTP 400) | JSON | **No** — does not exist |
| 7 | `wassertemperatur.org`                          | All 6 lakes have pages, but values are stale or `–°C` | HTML (WordPress)  | **No** — third-party scraper, low S/N |
| 8 | `chiemsee-schifffahrt.de`                       | Chiemsee — links out to `nid.bayern.de`       | HTML link            | **No** — link only |
| 9 | `tegernsee.com`                                | Editorial copy (climate averages)             | HTML                 | **No** — no live temp |
| 10 | `tourismus.prien.de`                           | None visible                                  | HTML                 | **No** — no temp |
| 11 | `starnbergersee.de`                            | None visible                                  | HTML                 | **No** — no temp |
| 12 | `seenschifffahrt.de` (Bayerische Seenschifffahrt) | None visible                              | HTML                 | **No** — no temp |
| 13 | `zwei-seen-land.de` (Kochel/Walchensee tourism) | None visible                                 | HTML                 | **No** — no temp |
| 14 | `wasserwacht.bayern` (DRK regional landing)     | No public sensor; sensors are run per-Ortsgruppe (e.g. Wasserwacht München-West, our Wachplan source) | HTML | **No** — landing only |
| 15 | `schliersee.de`                                 | None visible                                 | HTML                 | **No** — but Schliersee already in GKD |
| 16 | `lgl.bayern.de` (badewasser/badeseen subdomains) | DNS for `badewasser.lgl.bayern.de` and `badeseen.bayern.de` does not resolve | — | **Inconclusive — cannot reach** |

**Net new lake covered: zero.** Specifically, **no source found that fills the Kochelsee gap** in this round.

## Detailed findings

### 1. `nid.bayern.de` — Niedrigwasser-Informationsdienst

- URL: `https://www.nid.bayern.de/wassertemperatur/inn/stock-18400503` (example).
- Coverage: same `wassertemperatur` data as GKD, with identical station IDs and URL slugs (just a different sub-path on the same LfU CMS).
- Format: HTML with same `LfUMap.init` JS blob and same table layout.
- Update freq: same as GKD.
- License: assumed identical (LfU/CC BY 4.0).
- **Verdict:** Don't build a separate adapter — it's the same data. Useful only as a *redundant* mirror if `gkd.bayern.de` is down.

### 2. `hnd.bayern.de` — Hochwassernachrichtendienst

- Same backend family as GKD (LfU). Index lists pegels with **river water levels** and warning thresholds, no Wassertemperatur on lakes (e.g. `kochel-16407002` is a river-level station on the Loisach, not a lake-temp sensor).
- **Verdict:** Wrong product. Skip.

### 3. `pegelonline.wsv.de` — federal WSV REST API

- Endpoint: `https://www.pegelonline.wsv.de/webservices/rest-api/v2/stations.json`.
- 785 stations total. Filtered for our keywords (`starnberg, ammersee, chiemsee, tegernsee, kochel, bodensee, lindau, schliersee, münchen`).
- The only Bavarian-lake-relevant hit is **KONSTANZ** on Bodensee (`uuid: aa9179c1-17ef-4c61-a48a-74193fa7bfdf`). Its only timeseries is `W` (WASSERSTAND ROHDATEN, cm). **No water-temperature timeseries.**
- License: `Datenlizenz Deutschland - Namensnennung - Version 2.0` (federal).
- **Verdict:** Excellent JSON API but irrelevant for our use case. Skip.

### 4. `opendata.dwd.de`

- Top-level listing returns `climate_environment/`, `weather/`. Lake water temperature is not in DWD's product catalog (DWD focuses on air/weather/sea, not inland surface water).
- **Verdict:** Wrong product. Skip.

### 5. Open-Meteo Marine API (`marine-api.open-meteo.com`)

- Tested with Chiemsee coordinates (`lat=47.85, lon=12.36`) requesting `sea_surface_temperature`.
- HTTP 200, but **all 168 hourly values are `null`** — Open-Meteo Marine has no inland-lake coverage in Bavaria. (Returned `elevation: 545` confirms the grid cell is land.)
- **Verdict:** Skip.

### 6. Open-Meteo Forecast — `lake_water_temperature` variable

- Probed `https://api.open-meteo.com/v1/forecast?...&hourly=lake_water_temperature` → HTTP 400 with error "Cannot initialize ... from invalid String value lake_water_temperature".
- Variable is not in their schema.
- **Verdict:** Not a real Open-Meteo variable. Skip.

### 7. `wassertemperatur.org` (third-party WordPress aggregator)

- Pages exist for all 6 target lakes (`/starnberger-see/`, `/ammersee/`, `/chiemsee/`, `/tegernsee/`, `/kochelsee/`, `/bodensee/`).
- Sample fetched: `/kochelsee/`. Live block reads:
  > *"Aktuelle Wassertemperatur im Kochelsee: -°C — Aktuelle Wassertemperatur-Daten zur Badesaison 2026 wieder verfügbar!"*
  i.e., **the value is missing.** Last meaningful update: `2025-11-13` per JSON-LD.
- No documented API, no licensing info, runs on WordPress with Google Ads (`adsbygoogle`).
- **Verdict:** Low signal-to-noise; site is opportunistic SEO. Not trustworthy as upstream. Skip.

### 8. `chiemsee-schifffahrt.de` (Chiemsee ferry)

- Page links Wassertemperatur out to `nid.bayern.de/wassertemperatur/inn/stock-18400503` — **the same GKD station** (Stock, id 18400503) we already plan to scrape. No own data.
- **Verdict:** Skip — they consume the same source we do.

### 9. `tegernsee.com` (Tegernsee tourism)

- Has FAQ-style copy on average summer/winter temperatures ("up to 24°C in midsummer", "around 4°C in winter") but no current/live reading.
- **Verdict:** Editorial only. Skip.

### 10. `tourismus.prien.de` (Prien am Chiemsee)

- Searched body for `wassertemp` / temperature decimals → **no live water-temperature display.**
- **Verdict:** Skip.

### 11. `starnbergersee.de` (Starnberger See tourism)

- No `wassertemp` content found.
- **Verdict:** Skip.

### 12. `seenschifffahrt.de` (Bayerische Seenschifffahrt — state ferries on Starnberger See, Ammersee, Königssee, Tegernsee)

- No water-temperature display on landing page. (Site is large, 1.1 MB; deeper crawl would exceed budget.)
- **Verdict:** Skip for v0.4. Revisit if a per-lake page with temp is found via search.

### 13. `zwei-seen-land.de` (Kochel + Walchensee tourism)

- Tourist portal for Kochelsee/Walchensee. No live water-temperature reading.
- **Verdict:** Skip — but this **was** the most plausible Kochelsee candidate, so the gap remains open.

### 14. `wasserwacht.bayern` (Bavarian DRK Wasserwacht state landing)

- Public landing page only. Sub-URL `/gliederung/muenchen` returned 404. Sensors are operated by individual Ortsgruppen (München-West runs the Wachplan stack we already use).
- **Verdict:** Skip. Worth a separate, manual outreach effort: contact München-Ost (Pullach? Garching?) and Starnberger Wasserwacht to see if they run sensors with public endpoints — but that's a human task, not adapter work.

### 15. `schliersee.de` (Schliersee tourism)

- No `wassertemp` content. (Schliersee is already covered by GKD station `18222008`.)
- **Verdict:** Skip.

### 16. `lgl.bayern.de` (Bavarian State Office for Health & Food Safety)

- LGL is the responsible authority for **bathing-water quality** under the EU Bathing Water Directive — they sample microbiology + temperature at official Badegewässer.
- The plausible URLs `https://www.badewasser.lgl.bayern.de/` and `https://www.badeseen.bayern.de/` **do not resolve in DNS** at investigation time. The `lgl.bayern.de` main site does not surface a "badegewässer" link from the homepage. The Bavarian Ministry sub-URL `https://www.stmuv.bayern.de/themen/wasser/gewaesser/badegewaesser/` returns 404.
- LGL data, if accessible, would only update during the May–September bathing season and on a weekly cadence (sampling), not continuously — so probably not useful for OWS swimmers anyway.
- **Verdict:** Inconclusive (cannot reach). Probably the right URL has moved; worth a manual search later. Even if found, low cadence makes it secondary at best.

## Open questions / leads not pursued (under request budget)

- **Webcam scraping with sensor overlays.** Many alpine-lake webcams (e.g. Feratel, foto-webcam.eu) display a small text overlay with current air/water temperature. Last-resort OCR; high engineering cost, low data quality. Not pursued.
- **Bavarian Open Data portal** (`opendata.bayern.de` / `geoportal.bayern.de`). Not pursued in this round; could host static datasets but unlikely real-time.
- **Triathlon clubs / DLRG Bayern.** No structured public feeds; race-day water-temp is announced ad-hoc on Instagram/club sites. Not adapter-friendly.
- **Manual outreach to Wasserwacht Starnberg / Wasserwacht Bad Tölz-Wolfratshausen** about Kochelsee. Out of scope for an automated investigation but worth a human follow-up.

## Conclusion for v0.4

There is **no easy second source** for Munich-area lake water temperatures beyond GKD + Wachplan. The two complement each other almost perfectly:

- **GKD** = official, 5 of 6 target lakes, 15-min cadence, fresh, CC BY 4.0.
- **Wachplan** = community/Wasserwacht, 2 small Munich lakes (Lußsee + Langwieder See) **not covered by GKD**, JSON.
- **Air temp** = Open-Meteo for all lakes.

The remaining gap is Kochelsee — defer to v0.5 with a dedicated investigation (most likely path: contact Wasserwacht in person, or accept "no live data; show climatology only").
