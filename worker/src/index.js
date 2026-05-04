// The worker is intentionally dumb about which stations to fetch — it asks
// the backend at /gkd/stations on each cycle so the catalog only lives in
// backend/adapters/gkd/gkd.go. The base URL and User-Agent stay here because
// they describe how the worker talks to the upstream, not what to fetch.
const GKD_BASE = "https://www.gkd.bayern.de/de/seen/wassertemperatur/";
const GKD_USER_AGENT = "seebuddy/0.1 (+https://github.com/scorphus/seebuddy)";

async function listStations(env) {
  const res = await fetch(env.STATIONS_URL);
  if (!res.ok) {
    throw new Error(`stations failed: ${res.status} ${await res.text()}`);
  }
  const body = await res.json();
  return body.stations.map((s) => s.sensor_id);
}

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

async function runCycle(env) {
  const stations = await listStations(env);
  // Run all station ingests concurrently and don't let one failure block
  // the rest. Each result is logged; the poll trigger fires after.
  const results = await Promise.allSettled(
    stations.map((s) => ingestStation(env, s)),
  );
  const summary = [];
  for (const r of results) {
    if (r.status === "rejected") {
      console.error("gkd ingest threw", r.reason);
      summary.push({ ok: false, error: String(r.reason) });
    } else {
      if (!r.value.ok) console.error("gkd ingest failed", r.value);
      summary.push(r.value);
    }
  }

  const poll = await fetch(env.POLL_URL, {
    method: "POST",
    headers: { "X-Poll-Token": env.POLL_TOKEN },
  });
  if (!poll.ok) {
    throw new Error(`poll failed: ${poll.status} ${await poll.text()}`);
  }
  return summary;
}

export default {
  async scheduled(event, env, ctx) {
    await runCycle(env);
  },

  // fetch lets us manually trigger one cycle without waiting for the cron.
  // Authorised by the same shared token so the endpoint isn't an open DoS.
  async fetch(request, env, ctx) {
    if (request.headers.get("X-Poll-Token") !== env.POLL_TOKEN) {
      return new Response("forbidden", { status: 403 });
    }
    const summary = await runCycle(env);
    return Response.json({ summary });
  },
};
