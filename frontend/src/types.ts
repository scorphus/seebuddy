// Mirrors the response shape of GET /lakes from the lakes service.

export interface WaterReading {
  adapter: string;
  measured_at: string;
  age_seconds: number;
  stale: boolean;
  temp_c: number | null;
  air_temp_c?: number | null;
  humidity_pct?: number | null;
}

export interface WeatherReading {
  adapter: string;
  measured_at: string;
  age_seconds: number;
  stale: boolean;
  air_temp_c: number | null;
  humidity_pct: number | null;
  wind_speed_kmh: number | null;
  weather_code: number | null;
  is_day: boolean | null;
}

export interface Latest {
  water: WaterReading | null;
  weather: WeatherReading | null;
}

export interface Lake {
  slug: string;
  name: string;
  region: string;
  lat: number;
  lon: number;
  latest: Latest | null;
}

export interface ListResponse {
  lakes: Lake[];
}
