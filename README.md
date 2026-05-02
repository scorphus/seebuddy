# muenchner-see-buddy

Live water and weather conditions for swimmable lakes around Munich.

```sh
just backend     # encore run on :4000 (regenerates sqlc + TS client first)
just frontend    # vite dev server on :5173 (proxies /lakes to :4000)
```

First time only: `just frontend-install` to install the npm deps.

## What it does

A cron job polls each adapter periodically and stores readings in a
shared `readings` table. The `GET /lakes` endpoint returns the catalog
of known lakes plus the latest reading for each. The frontend renders
that as a grid of cards: water temperature in focus, air/humidity/wind
secondary, weather emoji from the WMO code, and a stale badge when the
reading is older than 90 minutes.

## Layout

```
backend/
├── lakes/                # only Encore service: catalog + cron + GET /lakes
├── adapters/
│   ├── adapter.go        # Adapter interface
│   ├── wachplan/         # Wasserwacht LoRaWAN sensors (Lußsee, Langwieder See)
│   └── generic/          # lakes without a dedicated sensor; uses openmeteo
└── openmeteo/            # shared lat/lon current-weather client
frontend/                 # Vue 3 + Vite + Tailwind
```

## Adapters

- **wachplan**: water + air temperature + humidity from sensors run by
  [Wasserwacht München-West](https://sensors.mein-wachplan.de/), built and
  operated by Bernhard Rohloff. Currently covers Lußsee and Langwieder See.
- **gkd**: water temperature scraped from
  [gkd.bayern.de](https://www.gkd.bayern.de/de/seen/wassertemperatur)
  (Gewässerkundlicher Dienst Bayern, the official Bavarian state service).
  Covers Ammersee, Pilsensee, Schliersee, Starnberger See, Tegernsee, and
  Wörthsee.
- **generic**: lakes without a dedicated sensor. Pulls air temperature,
  humidity, wind, and weather code from [Open-Meteo](https://open-meteo.com/)
  using the lake's lat/lon.

Adding a new lake = edit the relevant adapter's lake list.
Adding a new data source = new adapter package + register in
`backend/lakes/cron.go`.

## Other commands

```sh
just test                # encore test ./...
just lint                # go vet + gofmt check
just gen                 # gen-sqlc + gen-client
just gen-sqlc            # sqlc generate
just gen-client          # encore gen client → frontend/src/client.ts
```

## License & SLA

Personal project. BSD-3-Clause. No SLA — be nice to upstream services.

## Attribution

GKD water temperature data is provided under
[CC BY 4.0](https://creativecommons.org/licenses/by/4.0/):

> Datenquelle: Bayerisches Landesamt für Umwelt, www.lfu.bayern.de
