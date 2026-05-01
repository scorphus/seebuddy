// Mirrors the response shape of GET /lakes from the lakes service.
// Replace with the generated client (`just gen-client`) once it's wired.

export interface Reading {
  adapter: string;
  measured_at: string;
  age_seconds: number;
  stale: boolean;
  water_temp_c: number | null;
  air_temp_c: number | null;
  humidity_pct: number | null;
  wind_speed_kmh: number | null;
  weather_code: number | null;
}

export interface Lake {
  slug: string;
  name: string;
  region: string;
  lat: number;
  lon: number;
  latest: Reading | null;
}

export interface ListResponse {
  lakes: Lake[];
}
