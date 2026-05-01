<script setup lang="ts">
import { ref, onMounted } from "vue";
import type { Lake, ListResponse } from "./types";
import LakeCard from "./components/LakeCard.vue";

const lakes = ref<Lake[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const res = await fetch("/lakes");
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const body = (await res.json()) as ListResponse;
    lakes.value = body.lakes;
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    loading.value = false;
  }
}

onMounted(load);
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <header class="bg-white border-b border-slate-200">
      <div class="max-w-6xl mx-auto px-6 py-5 flex items-center justify-between">
        <div>
          <h1 class="text-xl font-bold text-slate-900">Münchner See Buddy</h1>
          <p class="text-xs text-slate-500">Live water + weather conditions for swim-friendly Munich lakes</p>
        </div>
        <button
          @click="load"
          :disabled="loading"
          class="text-sm px-3 py-1.5 rounded border border-slate-300 hover:bg-slate-100 disabled:opacity-50"
        >
          {{ loading ? "..." : "Refresh" }}
        </button>
      </div>
    </header>

    <main class="max-w-6xl mx-auto px-6 py-8">
      <div v-if="error" class="rounded-lg bg-red-50 border border-red-200 p-4 text-red-800 text-sm">
        Error loading lakes: {{ error }}
      </div>

      <div v-else-if="loading && lakes.length === 0" class="text-slate-500">Loading...</div>

      <div
        v-else
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        <LakeCard v-for="lake in lakes" :key="lake.slug" :lake="lake" />
      </div>
    </main>
  </div>
</template>
