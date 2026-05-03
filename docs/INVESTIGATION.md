# Investigation: sensors.mein-wachplan.de

**Date:** 2026-05-01
**Investigator:** seebudy-research/0.1
**Total requests made:** ~40

## Executive Summary

**We have structured JSON!** The sensors.mein-wachplan.de server exposes a JSON
API at `/json/?sensor_id=<id>` with CORS enabled (`Access-Control-Allow-Origin: *`).
Two sensor IDs were discovered: `laws` (Langwieder See) and `luss` (Lußsee).
Both return real-time LoRaWAN sensor data via The Things Network (TTN). No OCR
needed — we can build a clean HTTP+JSON adapter.

## JSON API

### Endpoint

```
GET https://sensors.mein-wachplan.de/json/?sensor_id={sensor_id}
```

- **CORS:** `Access-Control-Allow-Origin: *` (can be called from anywhere)
- **Content-Type:** `application/json`
- **Authentication:** None required
- **Error response:** `{ "error": "bad sensor id" }` (HTTP 200)

### Known Sensor IDs

| Sensor ID | App ID                      | Device ID (EUI)                | Location       |
| --------- | --------------------------- | ------------------------------ | -------------- |
| `laws`    | `wt01`                      | `eui-a84041f441890ac0-lws-neu` | Langwieder See |
| `luss`    | `wassertemperatur-luss-see` | `eui-a840414281840166`         | Lußsee         |

### Response Schema

```json
{
  "id": "172310",
  "datetime": "2026-05-01 14:14:11",
  "app_id": "wt01",
  "dev_id": "eui-a84041f441890ac0-lws-neu",
  "ttn_timestamp": "2026-05-01T12:14:11.475189348Z",
  "gtw_id": "wasserwacht-muenchen-west",
  "gtw_rssi": "-97",
  "gtw_snr": "9",
  "dev_raw_payload": "zAUFJwIGAQV+f/8=",
  "dev_value_1": "13.19",
  "dev_value_2": "14.06",
  "dev_value_3": "51.8",
  "dev_value_4": "3.077"
}
```

### Field Mapping (inferred from PNG cross-reference)

| Field           | Meaning (laws)                 | Meaning (luss)              | Unit |
| --------------- | ------------------------------ | --------------------------- | ---- |
| `dev_value_1`   | Air temperature                | Air temperature (at Lußsee) | °C   |
| `dev_value_2`   | Water temperature (Langwieder) | Water temperature (Lußsee)  | °C   |
| `dev_value_3`   | Humidity                       | Humidity                    | %    |
| `dev_value_4`   | Battery voltage (likely)       | Battery voltage (likely)    | V    |
| `datetime`      | Measurement time (CET/CEST)    | Measurement time (CET/CEST) | —    |
| `ttn_timestamp` | TTN receive time (UTC)         | TTN receive time (UTC)      | —    |

**Cross-reference with PNG (2026-05-01):**

- PNG "Temperatur: 13.11 °C" ≈ `laws.dev_value_1 = 13.19` (slight time diff)
- PNG "Wasser Langwieder See: 14.06 °C" = `laws.dev_value_2 = 14.06` (exact)
- PNG "Luftfeuchtigkeit: 52.5 %" ≈ `laws.dev_value_3 = 51.8` (slight time diff)
- PNG "Wasser Luss See: 14.25 °C" = `luss.dev_value_2 = 14.25` (exact)

### Infrastructure

- **LoRaWAN gateway:** `wasserwacht-muenchen-west`
- **Network:** The Things Network (TTN)
- **Server:** nginx on `185.26.156.137` (shared with mein-wachplan.de)
- **Made by:** Bernhard Rohloff (Wasserwacht München-West volunteer)

## Update Frequency

- Sensor readings appear to arrive every ~10-15 minutes (based on timestamp gaps)
- No `Last-Modified`, `ETag`, or `Cache-Control` headers on responses
- Same data returned on repeated requests within seconds (not re-queried per request)

## Endpoint Probe Results

| Path            | Status | Content-Type     | Notes                          |
| --------------- | ------ | ---------------- | ------------------------------ |
| `/`             | 200    | image/png        | Dashboard PNG (300x150)        |
| `/json/`        | 200    | application/json | API endpoint (needs sensor_id) |
| `/robots.txt`   | 404    | text/html        | Not found                      |
| `/api`          | 404    | text/html        | Not found                      |
| `/api/v1`       | 404    | text/html        | Not found                      |
| `/sensors`      | 404    | text/html        | Not found                      |
| `/data`         | 404    | text/html        | Not found                      |
| All other paths | 404    | text/html        | Not found                      |

## Related Services

| Domain                      | Purpose                                     |
| --------------------------- | ------------------------------------------- |
| `mein-wachplan.de`          | Documentation / landing page                |
| `app.mein-wachplan.de`      | Vue.js PWA for duty planning                |
| `services.mein-wachplan.de` | Backend API (`/api/v1/app`, `/api/v1/auth`) |
| `sensors.mein-wachplan.de`  | Sensor data (our target)                    |

## PNG Analysis

- **Dimensions:** 300 x 150 px
- **Color:** 1-bit colormap (palette mode), cyan text on dark background
- **DPI:** 96
- **No metadata** revealing generation tool (no exiftool available, PIL shows minimal info)
- **OCR:** Not needed since JSON API exists, but would be feasible given the simple layout

## Recommendation

### Adapter Strategy: HTTP + JSON (direct)

1. **Poll** `/json/?sensor_id=laws` and `/json/?sensor_id=luss` every 10 minutes
2. **Parse** `dev_value_2` as water temperature, `dev_value_1` as air temperature
3. **Use** `datetime` as measurement timestamp (CET/CEST timezone)
4. **Deduplicate** by `id` field (auto-incrementing reading ID)
5. **Set** `User-Agent` to identify the project

### No OCR needed

The JSON API provides exactly the data we need with CORS support and no
authentication. This is the ideal scenario.

### Polite usage

- Poll no more frequently than every 10 minutes (matching sensor cadence)
- Use conditional requests if headers become available in the future
- Identify with a proper User-Agent
- Consider reaching out to Bernhard Rohloff to let him know about the project
