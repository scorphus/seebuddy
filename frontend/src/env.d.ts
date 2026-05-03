/// <reference types="vite/client" />

interface ImportMetaEnv {
  /**
   * Base URL of the Encore backend. Empty in dev (Vite proxy handles it);
   * set to e.g. `https://staging-seebuddy-y3mi.encr.app` in
   * production via the host's build env vars.
   */
  readonly VITE_API_BASE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
