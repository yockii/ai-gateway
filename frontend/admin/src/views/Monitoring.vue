<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { monitoringApi } from '@/api/monitoring'
import type { SystemMetrics, Alert } from '@/types/models'

const metrics = ref<SystemMetrics>({
  qps: 0,
  avg_latency: 0,
  error_rate: 0,
  active_users: 0,
  total_requests: 0,
  cpu_usage: 0,
  memory_usage: 0,
})
const alerts = ref<Alert[]>([])
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
  loading.value = true
  try {
    metrics.value = await monitoringApi.getMetrics()
    alerts.value = await monitoringApi.getAlerts()
  } catch (err) {
    console.error('Failed to fetch monitoring data:', err)
  } finally {
    loading.value = false
  }
}

const getAlertColor = (type: string) => {
  const map: Record<string, string> = {
    error: 'bg-red-100 text-red-700 border-red-200',
    warning: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    info: 'bg-blue-100 text-blue-700 border-blue-200',
  }
  return map[type] || 'bg-slate-100 text-slate-600'
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">系统监控</h1>
      <button @click="fetchData" class="px-4 py-2 bg-slate-100 rounded-lg hover:bg-slate-200">
        刷新
      </button>
    </div>

    <div v-if="loading && !metrics.qps" class="text-center py-12 text-slate-500">加载中...</div>

    <div v-else class="space-y-6">
      <!-- 系统指标 -->
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">QPS</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">{{ metrics.qps }}</p>
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">平均延迟</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">{{ metrics.avg_latency }}ms</p>
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">错误率</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">{{ (metrics.error_rate * 100).toFixed(2) }}%</p>
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">活跃用户</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">{{ metrics.active_users }}</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">CPU 使用率</p>
          <div class="mt-2">
            <div class="h-2 bg-slate-200 rounded-full overflow-hidden">
              <div class="h-full bg-blue-600 transition-all" :style="{ width: `${metrics.cpu_usage}%` }" />
            </div>
            <p class="text-sm text-slate-600 mt-1">{{ metrics.cpu_usage }}%</p>
          </div>
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">内存使用率</p>
          <div class="mt-2">
            <div class="h-2 bg-slate-200 rounded-full overflow-hidden">
              <div class="h-full bg-green-600 transition-all" :style="{ width: `${metrics.memory_usage}%` }" />
            </div>
            <p class="text-sm text-slate-600 mt-1">{{ metrics.memory_usage }}%</p>
          </div>
        </div>
      </div>

      <!-- 告警列表 -->
      <div class="bg-white rounded-xl shadow-sm p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">最近告警</h2>
        <div v-if="alerts.length === 0" class="text-center py-8 text-slate-500">暂无告警</div>
        <div v-else class="space-y-3">
          <div
            v-for="alert in alerts"
            :key="alert.id"
            :class="['p-4 rounded-lg border', getAlertColor(alert.type)]"
          >
            <div class="flex items-center justify-between">
              <p class="font-medium">{{ alert.message }}</p>
              <span class="text-sm">{{ new Date(alert.created_at).toLocaleString() }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
