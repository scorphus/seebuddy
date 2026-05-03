<script setup lang="ts">
import { ref, onMounted } from "vue";
import { VueDraggable } from "vue-draggable-plus";
import type { Lake, ListResponse } from "./types";
import LakeCard from "./components/LakeCard.vue";

const lakes = ref<Lake[]>([]);
const loading = ref(true);
const error = ref<string | null>(null);

const orderKey = "muenchner-see-buddy:lake-order";

function loadOrder(): string[] {
  try {
    const raw = localStorage.getItem(orderKey);
    return raw ? (JSON.parse(raw) as string[]) : [];
  } catch {
    return [];
  }
}

function saveOrder(slugs: string[]) {
  try {
    localStorage.setItem(orderKey, JSON.stringify(slugs));
  } catch {
    // localStorage may be disabled (private mode, etc.) — silently ignore.
  }
}

// applyOrder sorts the fetched lakes by the user's saved order. Slugs not
// in the saved list keep their fetched-order position at the end.
function applyOrder(fetched: Lake[]): Lake[] {
  const saved = loadOrder();
  if (saved.length === 0) return fetched;
  const rank = new Map(saved.map((s, i) => [s, i]));
  return [...fetched].sort((a, b) => {
    const ra = rank.get(a.slug) ?? Infinity;
    const rb = rank.get(b.slug) ?? Infinity;
    return ra - rb;
  });
}

// In dev, VITE_API_BASE_URL is unset and the empty string lets the Vite
// proxy forward /lakes to the local Encore daemon. In prod, the host
// (Cloudflare Pages, etc.) sets this to the deployed Encore URL.
const apiBase = import.meta.env.VITE_API_BASE_URL ?? "";

async function load() {
  loading.value = true;
  error.value = null;
  try {
    const res = await fetch(`${apiBase}/lakes`);
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const body = (await res.json()) as ListResponse;
    lakes.value = applyOrder(body.lakes);
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e);
  } finally {
    loading.value = false;
  }
}

function onReorder() {
  saveOrder(lakes.value.map((l) => l.slug));
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

      <VueDraggable
        v-else
        v-model="lakes"
        :animation="180"
        ghost-class="lake-ghost"
        chosen-class="lake-chosen"
        @end="onReorder"
        class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4"
      >
        <LakeCard
          v-for="lake in lakes"
          :key="lake.slug"
          :lake="lake"
          class="cursor-grab active:cursor-grabbing"
        />
      </VueDraggable>
    </main>
  </div>
</template>
