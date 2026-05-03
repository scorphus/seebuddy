export default {
  async scheduled(event, env, ctx) {
    const res = await fetch(env.POLL_URL, {
      method: "POST",
      headers: { "X-Poll-Token": env.POLL_TOKEN },
    });
    if (!res.ok) {
      throw new Error(`poll failed: ${res.status} ${await res.text()}`);
    }
  },
};
