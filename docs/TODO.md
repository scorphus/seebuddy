# TODO

## Candidate lakes to add

Coordinates need to be supplied per-lake (don't trust auto-lookups).

### Munich proper / city border (likely swim-friendly OWS)
- [ ] Hinterbrühler See — southwest, Fürstenried
- [ ] Heimstettener See — east
- [ ] Birkensee — north
- [ ] Echinger Klärweiher — north (verify swimming is allowed)
- [ ] Speichersee Ismaning — reservoir, north (verify swimming is allowed)

### Day trip — south / southeast
- [ ] Spitzingsee — alpine, near Schliersee
- [ ] Eibsee — base of Zugspitze
- [ ] Sylvensteinspeicher — alpine reservoir
- [ ] Osterseen — group of small lakes near Iffeldorf
- [ ] Staffelsee — near Murnau
- [ ] Riegsee — near Murnau

### Day trip — southwest / Allgäu
- [ ] Forggensee — Füssen
- [ ] Bannwaldsee — Schwangau

### Day trip — east
- [ ] Chiemsee — the big one (~85km)
- [ ] Simssee — near Rosenheim
- [ ] Waginger See + Tachinger See — southeast

## Frontend i18n

Goal: ship `en` (default), `de`, `pt` initially. Keep the backend
identifier-only — no translatable strings cross the API boundary.

### Stack
- [ ] Add `vue-i18n` (composition API mode, `legacy: false`)
- [ ] Locale files at `frontend/src/locales/{en,de,pt}.json`
- [ ] Detect initial locale from `navigator.language`, with manual override
      persisted in `localStorage` (`muenchner-see-buddy:locale`)
- [ ] Tiny locale switcher in the header (3 buttons or a `<select>`)

### Strings to extract (frontend only)
- [ ] Header: title tagline ("Live water + weather conditions for…"), Refresh button
- [ ] Card labels: Water / Air / Humidity / Wind
- [ ] Stale badge ("stale")
- [ ] Empty state ("No reading yet. Waiting for the next poll.")
- [ ] Footer ("weather: X ago · openmeteo (stale)")
- [ ] Error state ("Error loading lakes: …")
- [ ] Loading state ("Loading…")
- [ ] Tooltip: `from {compass} ({deg}°)` for wind direction

### Mappings (display, not translation)
- [ ] Adapter display names (proper nouns, not localized):
      `wachplan → Wasserwacht München-West`,
      `gkd → GKD Bayern`,
      `openmeteo → Open-Meteo`
- [ ] Region display names — translate per locale:
      `munich-west → West Munich / München West / Oeste de Munique`
- [ ] Lake names: keep as-is (German proper nouns) across all locales

### Native browser APIs (no extra deps)
- [ ] `relativeTime` → `Intl.RelativeTimeFormat(locale).format(-mins, 'minute')`
- [ ] Temperature/humidity numbers → `Intl.NumberFormat(locale, { maximumFractionDigits: 1 })`
- [ ] Measured-at hover tooltip → `Intl.DateTimeFormat(locale).format(date)`

### Backend touch points (minimal)
- [ ] Confirm the API still returns identifiers only: adapter IDs, lake
      slugs, region slugs, weather codes (numeric WMO). No human strings
      from the server.
- [ ] If we ever want server-side error messages translated, defer — pass
      an error code instead and let the frontend translate.

### Out of scope (for now)
- Server-side i18n (no use case yet — UI is only consumer)
- Pluralization beyond what `Intl.RelativeTimeFormat` and vue-i18n give
- Right-to-left support
- Adapter docs or attribution string translation
  (`Datenquelle: Bayerisches Landesamt für Umwelt, www.lfu.bayern.de` is a
  fixed attribution per the CC BY 4.0 license — leave as-is)
