<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { monitoringApi } from '@/api/monitoring'
import type { SystemMetrics } from '@/types/models'

const metrics = ref<SystemMetrics>({
  qps: 0,
  avg_latency: 0,
  error_rate: 0,
  active_users: 0,
  total_requests: 0,
  cpu_usage: 0,
  memory_usage: 0,
})
const loading = ref(false)
let intervalId: number | null = null

onMounted(async () => {
  await fetchData()
  intervalId = window.setInterval(fetchData, 30000)
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
})

const fetchData = async () => {
  if (document.hidden) return
  loading.value = true
  try {
    metrics.value = await monitoringApi.getMetrics()
  } catch (err) {
    console.error('Failed to fetch metrics:', err)
  } finally {
    loading.value = false
  }
}

const getMetricColor = (value: number, threshold: number) => {
  if (value >= threshold) return 'text-red-600'
  if (value >= threshold * 0.8) return 'text-yellow-600'
  return 'text-green-600'
}
</script>

<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto space-y-6">
      <h1 class="text-2xl font-bold text-white">运维大屏</h1>

      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
          <p class="text-slate-400 text-sm">实时 QPS</p>
          <p class="text-4xl font-bold text-white mt-2">{{ metrics.qps }}</p>
        </div>
        <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
          <p class="text-slate-400 text-sm">平均延迟</p>
          <p :class="['text-4xl font-bold mt-2', getMetricColor(metrics.avg_latency, 500)]">
            {{ metrics.avg_latency }}<span class="text-lg">ms</span>
          </p>
        </div>
        <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
          <p class="text-slate-400 text-sm">错误率</p>
          <p :class="['text-4xl font-bold mt-2', getMetricColor(metrics.error_rate * 100, 1)]">
            {{ (metrics.error_rate * 100).toFixed(2) }}<span class="text-lg">%</span>
          </p>
        </div>
        <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
          <p class="text-slate-400 text-sm">活跃用户</p>
          <p class="text-4xl font-bold text-white mt-2">{{ metrics.active_users }}</p>
        </div>
      </div>

      <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
        <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
          <p class="text-slate-400 text-sm mb-4">CPU 使用率</p>
          <div class="flex items-center gap-4">
            <div class="flex-1 h-4 bg-slate-700 rounded-full overflow-hidden">
              <div
                class="h-full bg-blue-500 transition-all"
                :class="{
                  'bg-green-500': metrics.cpu_usage < 60,
                  'bg-yellow-500': metrics.cpu_usage >= 60 && metrics.cpu_usage < 80,
                  'bg-red-500': metrics.cpu_usage >= 80,
                }"
                :style="{ width: `${metrics.cpu_usage}%` }"
              />
            </div>
            <span class="text-2xl font-bold text-white">{{ metrics.cpu_usage }}%</span>
          </div>
        </div>

        <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
          <p class="text-slate-400 text-sm mb-4">内存使用率</p>
          <div class="flex items-center gap-4">
            <div class="flex-1 h-4 bg-slate-700 rounded-full overflow-hidden">
              <div
                class="h-full bg-blue-500 transition-all"
                :class="{
                  'bg-green-500': metrics.memory_usage < 60,
                  'bg-yellow-500': metrics.memory_usage >= 60 && metrics.memory_usage < 80,
                  'bg-red-500': metrics.memory_usage >= 80,
                }"
                :style="{ width: `${metrics.memory_usage}%` }"
              />
            </div>
            <span class="text-2xl font-bold text-white">{{ metrics.memory_usage }}%</span>
          </div>
        </div>
      </div>

      <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
        <p class="text-slate-400 text-sm mb-4">累计请求数</p>
        <p class="text-5xl font-bold text-white">{{ metrics.total_requests.toLocaleString() }}</p>
      </div>
    </div>
  </div>
</template>
