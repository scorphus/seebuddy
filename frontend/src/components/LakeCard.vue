<script setup lang="ts">
import type { Lake } from "../types";
import { weatherEmoji, relativeTime } from "../weather";

defineProps<{ lake: Lake }>();

function fmt(n: number | null, digits = 1): string {
  return n === null ? "—" : n.toFixed(digits);
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
      <span class="text-3xl leading-none" :title="`weather code ${lake.latest?.weather_code ?? ''}`">
        {{ weatherEmoji(lake.latest?.weather_code ?? null) }}
      </span>
    </div>

    <div v-if="lake.latest" class="space-y-3">
      <div>
        <p class="text-xs uppercase tracking-wide text-slate-500">Water</p>
        <p class="text-4xl font-bold text-sky-700">
          {{ fmt(lake.latest.water_temp_c, 1) }}<span class="text-xl text-slate-400">°C</span>
        </p>
      </div>

      <div class="grid grid-cols-3 gap-2 text-sm pt-1 border-t border-slate-100">
        <div>
          <p class="text-xs text-slate-500">Air</p>
          <p class="font-medium text-slate-700">{{ fmt(lake.latest.air_temp_c, 1) }}°C</p>
        </div>
        <div>
          <p class="text-xs text-slate-500">Humidity</p>
          <p class="font-medium text-slate-700">{{ fmt(lake.latest.humidity_pct, 0) }}%</p>
        </div>
        <div>
          <p class="text-xs text-slate-500">Wind</p>
          <p class="font-medium text-slate-700">{{ fmt(lake.latest.wind_speed_kmh, 0) }} km/h</p>
        </div>
      </div>

      <div class="flex items-center justify-between pt-2 text-xs">
        <span class="text-slate-500">
          {{ relativeTime(lake.latest.age_seconds) }} · {{ lake.latest.adapter }}
        </span>
        <span
          v-if="lake.latest.stale"
          class="px-2 py-0.5 rounded-full bg-amber-100 text-amber-800 font-medium"
        >
          stale
        </span>
      </div>
    </div>

    <div v-else class="text-sm text-slate-400 py-4">
      No reading yet. Waiting for the next poll.
    </div>
  </div>
</template>
