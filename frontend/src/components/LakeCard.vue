<script setup lang="ts">
import { computed } from "vue";
import type { Lake } from "../types";
import { weatherEmoji, relativeTime, windArrow, windCompass } from "../weather";

const props = defineProps<{ lake: Lake }>();

const water = computed(() => props.lake.latest?.water ?? null);
const weather = computed(() => props.lake.latest?.weather ?? null);

// Air and humidity always come from openmeteo. The wachplan buoy also
// reports these, but the sensor housing biases air several degrees warm
// vs. ambient — and mixing sources under a single "openmeteo" footer
// label was misleading.
const airC = computed(() => weather.value?.air_temp_c ?? null);
const humidity = computed(() => weather.value?.humidity_pct ?? null);

function fmt(n: number | null | undefined, digits = 1): string {
  return n == null ? "—" : n.toFixed(digits);
}
</script>

<template>
  <div
    class="rounded-lg border border-slate-200 bg-white p-5 shadow-sm hover:shadow-md transition-shadow"
  >
    <div class="flex items-start justify-between mb-4">
      <div>
        <h2 class="text-lg font-semibold text-slate-900">{{ lake.name }}</h2>
        <p class="text-xs text-slate-500 mt-0.5">{{ lake.region }}</p>
      </div>
      <span
        class="text-3xl leading-none"
        :title="`weather code ${weather?.weather_code ?? ''}`"
      >
        {{ weatherEmoji(weather?.weather_code ?? null, weather?.is_day ?? null) }}
      </span>
    </div>

    <div v-if="water || weather" class="space-y-3">
      <!-- WATER section: big temp + per-section staleness -->
      <div>
        <div class="flex items-center justify-between">
          <p class="text-xs uppercase tracking-wide text-slate-500">Water</p>
          <span
            v-if="water?.stale"
            class="px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 text-xs font-medium"
          >
            stale
          </span>
        </div>
        <p class="text-4xl font-bold text-sky-700">
          {{ fmt(water?.temp_c ?? null, 1)
          }}<span class="text-xl text-slate-400">°C</span>
        </p>
        <p v-if="water" class="text-xs text-slate-500 mt-0.5">
          {{ relativeTime(water.age_seconds) }} · {{ water.adapter }}
        </p>
        <p v-else class="text-xs text-slate-400 mt-0.5">no sensor</p>
      </div>

      <!-- WEATHER section: secondary stats from openmeteo, always-on -->
      <div class="grid grid-cols-3 gap-2 text-sm pt-1 border-t border-slate-100">
        <div>
          <p class="text-xs text-slate-500">Air</p>
          <p class="font-medium text-slate-700">{{ fmt(airC, 1) }}°C</p>
        </div>
        <div>
          <p class="text-xs text-slate-500">Humidity</p>
          <p class="font-medium text-slate-700">{{ fmt(humidity, 0) }}%</p>
        </div>
        <div>
          <p class="text-xs text-slate-500">Wind</p>
          <p
            class="font-medium text-slate-700"
            :title="weather?.wind_direction_deg != null ? `from ${windCompass(weather.wind_direction_deg)} (${weather.wind_direction_deg}°)` : ''"
          >
            <span class="text-slate-400 mr-0.5">{{ windArrow(weather?.wind_direction_deg ?? null) }}</span>
            {{ fmt(weather?.wind_speed_kmh ?? null, 0) }} km/h
          </p>
        </div>
      </div>

      <p v-if="weather" class="text-xs text-slate-400">
        weather: {{ relativeTime(weather.age_seconds) }} · {{ weather.adapter
        }}<span v-if="weather.stale"> (stale)</span>
      </p>
    </div>

    <div v-else class="text-sm text-slate-400 py-4">
      No reading yet. Waiting for the next poll.
    </div>
  </div>
</template>
