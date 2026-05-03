// GKD station sensor IDs in our "{basin}/{slug}-{station_id}" convention.
// Mirrors backend/adapters/gkd/gkd.go's `lakes` list. Kept in the worker
// because the worker fetches gkd.bayern.de directly: the Encore Cloud egress
// IPs are silently filtered there, so the worker is the only place that can
// reach the upstream.
const GKD_STATIONS = [
  "isar/stegen-16602008",
  "isar/pilsensee-16628055",
  "isar/schliersee-18222008",
  "isar/starnberg-16663002",
  "isar/gmund_tegernsee-18201303",
  "isar/woerthsee-16651003",
];

const GKD_BASE = "https://www.gkd.bayern.de/de/seen/wassertemperatur/";
const GKD_USER_AGENT = "seebuddy/0.1 (+https://github.com/scorphus/seebuddy)";

async function ingestStation(env, sensorID) {
  const upstream = await fetch(GKD_BASE + sensorID + "/messwerte", {
    headers: { "User-Agent": GKD_USER_AGENT },
  });
  if (!upstream.ok) {
    return { sensorID, ok: false, stage: "fetch", status: upstream.status };
  }
  const html = await upstream.text();
  const ingest = await fetch(env.INGEST_URL, {
    method: "POST",
    headers: {
      "X-Poll-Token": env.POLL_TOKEN,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ sensor_id: sensorID, html }),
  });
  if (!ingest.ok) {
    return { sensorID, ok: false, stage: "ingest", status: ingest.status, body: await ingest.text() };
  }
  return { sensorID, ok: true };
}

export default {
  async scheduled(event, env, ctx) {
    // Run all station ingests concurrently and don't let one failure block
    // the rest. Each result is logged; the poll trigger fires after.
    const results = await Promise.allSettled(
      GKD_STATIONS.map((s) => ingestStation(env, s)),
    );
    for (const r of results) {
      if (r.status === "rejected") {
        console.error("gkd ingest threw", r.reason);
      } else if (!r.value.ok) {
        console.error("gkd ingest failed", r.value);
      }
    }

    const poll = await fetch(env.POLL_URL, {
      method: "POST",
      headers: { "X-Poll-Token": env.POLL_TOKEN },
    });
    if (!poll.ok) {
      throw new Error(`poll failed: ${poll.status} ${await poll.text()}`);
    }
  },
};
