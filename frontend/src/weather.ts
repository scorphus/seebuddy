// WMO weather codes → emoji, with day/night variants for the clear-ish codes.
// Reference: https://open-meteo.com/en/docs (search "weather_code")

const day: Record<number, string> = {
  0: "☀️",
  1: "🌤️",
  2: "⛅",
  3: "☁️",
  45: "🌫️",
  48: "🌫️",
  51: "🌦️",
  53: "🌦️",
  55: "🌦️",
  56: "🌦️",
  57: "🌦️",
  61: "🌧️",
  63: "🌧️",
  65: "🌧️",
  66: "🌧️",
  67: "🌧️",
  71: "🌨️",
  73: "🌨️",
  75: "🌨️",
  77: "🌨️",
  80: "🌧️",
  81: "🌧️",
  82: "🌧️",
  85: "🌨️",
  86: "🌨️",
  95: "⛈️",
  96: "⛈️",
  99: "⛈️",
};

// Codes whose emoji changes at night. The rest (overcast, fog, rain, snow,
// thunder) look the same regardless of light.
const night: Record<number, string> = {
  0: "🌙",
  1: "🌙",
  2: "☁️",
};

export function weatherEmoji(code: number | null, isDay: boolean | null): string {
  if (code === null) return "";
  if (isDay === false && night[code]) return night[code];
  return day[code] ?? "";
}

export function relativeTime(seconds: number): string {
  if (seconds < 60) return `${seconds}s ago`;
  const min = Math.round(seconds / 60);
  if (min < 60) return `${min} min ago`;
  const h = Math.round(min / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.round(h / 24);
  return `${d}d ago`;
}
