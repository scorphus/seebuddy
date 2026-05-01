# muenchner-see-buddy

Live water and weather conditions for swimmable lakes around Munich.

```sh
encore run
curl localhost:4000/lakes
```

A cron job polls each adapter periodically and stores readings. The
`/lakes` endpoint returns the catalog of known lakes plus the latest
reading for each.

## Adapters

- **wachplan**: water + air temperature + humidity from sensors run by
  [Wasserwacht München-West](https://sensors.mein-wachplan.de/) (Bernhard
  Rohloff). Currently covers Lußsee and Langwieder See.
- **generic**: lakes without a dedicated sensor. Pulls air temperature,
  humidity, wind, and weather code from [Open-Meteo](https://open-meteo.com/)
  using the lake's lat/lon.

Adding a new lake = edit the relevant adapter's lake list. Adding a new
data source = new adapter package + register in `lakes/cron.go`.

Personal project. BSD-3-Clause. No SLA — be nice to upstream services.
