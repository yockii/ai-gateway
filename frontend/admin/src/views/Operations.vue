<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { monitoringApi } from '@/api/monitoring'
import { suppliersApi } from '@/api/suppliers'
import type { SystemMetrics, Supplier, FailoverEvent, SupplierFailureEvent } from '@/types/models'

const router = useRouter()
const metrics = ref<SystemMetrics>({
  qps: 0,
  avg_latency: 0,
  error_rate: 0,
  active_users: 0,
  total_requests: 0,
  cpu_usage: 0,
  memory_usage: 0,
})

const suppliers = ref<Supplier[]>([])
const supplierHealth = ref<Record<string, { is_healthy: boolean; latency: number; last_check: string }>>({})
const failoverEvents = ref<FailoverEvent[]>([])
const failureEvents = ref<SupplierFailureEvent[]>([])
const alerts = ref<Array<{ type: string; message: string; created_at: string }>>([])

const loading = ref(false)
let metricsInterval: ReturnType<typeof setInterval> | null = null
let healthInterval: ReturnType<typeof setInterval> | null = null
let eventSource: EventSource | null = null

onMounted(async () => {
  await fetchData()
  startIntervals()
  setupSSE()
})

onUnmounted(() => {
  if (metricsInterval) clearInterval(metricsInterval)
  if (healthInterval) clearInterval(healthInterval)
  if (eventSource) eventSource.close()
})

const fetchData = async () => {
  if (document.hidden) return
  loading.value = true
  try {
    metrics.value = await monitoringApi.getMetrics()
    await loadSuppliers()
    await loadEvents()
  } catch (err) {
    console.error('Failed to fetch metrics:', err)
  } finally {
    loading.value = false
  }
}

const loadSuppliers = async () => {
  try {
    suppliers.value = await suppliersApi.list()
    // Load health status for each supplier
    for (const supplier of suppliers.value) {
      try {
        const health = await suppliersApi.getHealth(supplier.id)
        supplierHealth.value[supplier.id] = {
          is_healthy: health.is_healthy,
          latency: health.latency,
          last_check: health.timestamp,
        }
      } catch (err) {
        supplierHealth.value[supplier.id] = {
          is_healthy: false,
          latency: 0,
          last_check: new Date().toISOString(),
        }
      }
    }
  } catch (err) {
    console.error('Failed to load suppliers:', err)
  }
}

const loadEvents = async () => {
  try {
    // Load recent failover and failure events
    // These endpoints should be implemented in the backend
    // For now, using mock data
    failoverEvents.value = []
    failureEvents.value = []
  } catch (err) {
    console.error('Failed to load events:', err)
  }
}

const startIntervals = () => {
  metricsInterval = setInterval(fetchData, 30000)
  healthInterval = setInterval(loadSuppliers, 30000)
}

const setupSSE = () => {
  try {
    // Setup SSE connection for real-time health updates
    // Falls back to polling if SSE is not available
    eventSource = new EventSource('/api/v1/admin/suppliers/health/stream')

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.supplier_id) {
          const oldHealth = supplierHealth.value[data.supplier_id]
          supplierHealth.value[data.supplier_id] = {
            is_healthy: data.is_healthy,
            latency: data.latency,
            last_check: data.last_check,
          }

          // Show alert on health status change
          if (oldHealth) {
            if (oldHealth.is_healthy && !data.is_healthy) {
              showAlert('error', `Supplier ${data.supplier_id} is now unhealthy`)
            } else if (!oldHealth.is_healthy && data.is_healthy) {
              showAlert('success', `Supplier ${data.supplier_id} is now healthy`)
            }
          }
        }
      } catch (err) {
        console.error('Failed to parse SSE data:', err)
      }
    }

    eventSource.onerror = () => {
      console.warn('SSE connection failed, falling back to polling')
      if (eventSource) {
        eventSource.close()
        eventSource = null
      }
    }
  } catch (err) {
    console.warn('SSE not available, using polling only')
  }
}

const showAlert = (type: string, message: string) => {
  alerts.value.unshift({
    type,
    message,
    created_at: new Date().toISOString(),
  })
  // Keep only last 10 alerts
  if (alerts.value.length > 10) {
    alerts.value = alerts.value.slice(0, 10)
  }
}

const getMetricColor = (value: number, threshold: number) => {
  if (value >= threshold) return 'text-red-600'
  if (value >= threshold * 0.8) return 'text-yellow-600'
  return 'text-green-600'
}

const getLatencyColor = (latency: number) => {
  if (latency >= 500) return 'text-red-600'
  if (latency >= 100) return 'text-yellow-600'
  return 'text-green-600'
}

const goToSupplier = (supplierId: string) => {
  router.push(`/suppliers/${supplierId}`)
}

const formatTime = (timeStr: string) => {
  const date = new Date(timeStr)
  const now = new Date()
  const diff = now.getTime() - date.getTime()

  if (diff < 60000) {
    return 'Just now'
  } else if (diff < 3600000) {
    const minutes = Math.floor(diff / 60000)
    return `${minutes}m ago`
  } else {
    return date.toLocaleTimeString()
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-900 p-6">
    <div class="max-w-7xl mx-auto space-y-6">
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-white">运维大屏</h1>
        <button @click="fetchData" class="px-4 py-2 bg-slate-700 text-white rounded-lg hover:bg-slate-600">
          刷新
        </button>
      </div>

      <!-- Real-time Alerts -->
      <div v-if="alerts.length > 0" class="space-y-2">
        <div
          v-for="alert in alerts.slice(0, 5)"
          :key="alert.created_at"
          :class="[
            'p-3 rounded-lg border',
            alert.type === 'error' ? 'bg-red-900/50 border-red-700 text-red-200' : 'bg-green-900/50 border-green-700 text-green-200',
          ]"
        >
          <div class="flex items-center justify-between">
            <span class="font-medium">{{ alert.message }}</span>
            <span class="text-sm opacity-75">{{ formatTime(alert.created_at) }}</span>
          </div>
        </div>
      </div>

      <!-- System Metrics -->
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
                class="h-full transition-all"
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
                class="h-full transition-all"
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
        <p class="text-5xl font-bold text-white">{{ (metrics.total_requests || 0).toLocaleString() }}</p>
      </div>

      <!-- Supplier Health Status Cards -->
      <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
        <h2 class="text-xl font-semibold text-white mb-4">供应商健康状态</h2>
        <div v-if="suppliers.length === 0" class="text-center py-8 text-slate-500">
          暂无供应商数据
        </div>
        <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          <div
            v-for="supplier in suppliers"
            :key="supplier.id"
            @click="goToSupplier(supplier.id)"
            class="bg-slate-700 rounded-lg p-4 border border-slate-600 hover:border-blue-500 cursor-pointer transition-colors"
          >
            <div class="flex items-center justify-between mb-3">
              <h3 class="text-lg font-semibold text-white">{{ supplier.display_name }}</h3>
              <div
                :class="{
                  'w-3 h-3 rounded-full': true,
                  'bg-green-500': supplierHealth[supplier.id]?.is_healthy,
                  'bg-red-500': !supplierHealth[supplier.id]?.is_healthy,
                  'animate-pulse': !supplierHealth[supplier.id]?.is_healthy,
                }"
              />
            </div>
            <div class="space-y-2 text-sm">
              <div class="flex justify-between">
                <span class="text-slate-400">Provider:</span>
                <span class="text-white">{{ supplier.provider }}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-400">Status:</span>
                <span :class="supplierHealth[supplier.id]?.is_healthy ? 'text-green-400' : 'text-red-400'">
                  {{ supplierHealth[supplier.id]?.is_healthy ? 'Healthy' : 'Unhealthy' }}
                </span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-400">Latency:</span>
                <span :class="['font-medium', getLatencyColor(supplierHealth[supplier.id]?.latency || 0)]">
                  {{ supplierHealth[supplier.id]?.latency || 0 }}ms
                </span>
              </div>
              <div class="flex justify-between">
                <span class="text-slate-400">Last Check:</span>
                <span class="text-slate-300">
                  {{ supplierHealth[supplier.id] ? formatTime(supplierHealth[supplier.id].last_check) : 'N/A' }}
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Recent Events -->
      <div class="bg-slate-800 rounded-xl p-6 border border-slate-700">
        <h2 class="text-xl font-semibold text-white mb-4">最近事件</h2>
        <div v-if="failoverEvents.length === 0 && failureEvents.length === 0" class="text-center py-8 text-slate-500">
          暂无事件记录
        </div>
        <div v-else class="space-y-3">
          <div
            v-for="event in [...failoverEvents, ...failureEvents].slice(0, 10)"
            :key="event.id"
            class="bg-slate-700 rounded-lg p-4 border border-slate-600"
          >
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div
                  :class="{
                    'w-2 h-2 rounded-full': true,
                    'bg-orange-500': 'reason' in event,
                    'bg-red-500': 'error_type' in event,
                  }"
                />
                <span class="text-white">
                  <template v-if="'reason' in event">
                    故障转移: {{ event.from_supplier_id }} → {{ event.to_supplier_id }}
                  </template>
                  <template v-else>
                    失败事件: {{ event.supplier_id }} - {{ event.error_type }}
                  </template>
                </span>
              </div>
              <span class="text-sm text-slate-400">{{ formatTime(event.timestamp) }}</span>
            </div>
            <p v-if="'reason' in event" class="text-sm text-slate-400 mt-2">{{ event.reason }}</p>
            <p v-else class="text-sm text-slate-400 mt-2">{{ event.error_msg }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
