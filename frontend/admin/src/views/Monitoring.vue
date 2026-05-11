<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick } from 'vue'
import { monitoringApi } from '@/api/monitoring'
import { suppliersApi } from '@/api/suppliers'
import type { SystemMetrics, Alert, Supplier, HealthCheckHistory } from '@/types/models'
import * as echarts from 'echarts'

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
const suppliers = ref<Supplier[]>([])
const supplierHealthHistory = ref<Record<string, HealthCheckHistory[]>>({})
const failureEventCount = ref(0)
const failoverEventCount = ref(0)

const loading = ref(false)
const healthChartRef = ref<HTMLElement | null>(null)
let healthChart: echarts.ECharts | null = null
let intervalId: ReturnType<typeof setInterval> | null = null
let eventSource: EventSource | null = null

onMounted(async () => {
  await fetchData()
  await initHealthChart()
  intervalId = window.setInterval(fetchData, 30000)
  setupSSE()

  // Handle window resize
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  if (intervalId) clearInterval(intervalId)
  if (eventSource) eventSource.close()
  if (healthChart) {
    healthChart.dispose()
    healthChart = null
  }
  window.removeEventListener('resize', handleResize)
})

const fetchData = async () => {
  if (document.hidden) return
  loading.value = true
  try {
    metrics.value = await monitoringApi.getMetrics()
    alerts.value = await monitoringApi.getAlerts()
    await loadSuppliers()
  } catch (err) {
    console.error('Failed to fetch monitoring data:', err)
  } finally {
    loading.value = false
  }
}

const loadSuppliers = async () => {
  try {
    suppliers.value = await suppliersApi.list()
    // Load health history for each supplier
    for (const supplier of suppliers.value) {
      try {
        const history = await suppliersApi.getHealthHistory(supplier.id, 100)
        supplierHealthHistory.value[supplier.id] = history
      } catch (err) {
        console.error(`Failed to load health history for ${supplier.id}:`, err)
        supplierHealthHistory.value[supplier.id] = []
      }
    }
    // Update chart after loading data
    await nextTick()
    updateHealthChart()
  } catch (err) {
    console.error('Failed to load suppliers:', err)
  }
}

const initHealthChart = async () => {
  await nextTick()
  if (!healthChartRef.value) return

  healthChart = echarts.init(healthChartRef.value)
  updateHealthChart()
}

const updateHealthChart = () => {
  if (!healthChart) return

  // Prepare data for chart
  const colors: string[] = [
    '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
    '#ec4899', '#06b6d4', '#84cc16', '#f97316', '#6366f1',
  ]

  const series: any[] = []
  suppliers.value.forEach((supplier, index) => {
    const history = supplierHealthHistory.value[supplier.id] || []
    if (history.length === 0) return

    series.push({
      name: supplier.display_name,
      type: 'line',
      smooth: true,
      data: history.map((h) => ({
        name: new Date(h.timestamp).toLocaleTimeString(),
        value: [h.timestamp, h.latency],
      })),
      itemStyle: {
        color: colors[index % colors.length],
      },
      lineStyle: {
        width: 2,
      },
      symbol: 'circle',
      symbolSize: 6,
    })
  })

  const option = {
    title: {
      text: '供应商健康趋势',
      left: 'center',
      textStyle: {
        color: '#1e293b',
        fontSize: 16,
        fontWeight: 600,
      },
    },
    tooltip: {
      trigger: 'axis',
      formatter: (params: any) => {
        if (!Array.isArray(params) || params.length === 0) return ''
        const time = new Date(params[0].value[0]).toLocaleString()
        let result = `<div class="font-medium mb-1">${time}</div>`
        params.forEach((param: any) => {
          const status = param.value[1] >= 500 ? '🔴' : param.value[1] >= 100 ? '🟡' : '🟢'
          result += `<div class="flex items-center gap-2">
            <span class="w-3 h-3 rounded-full" style="background-color: ${param.color}"></span>
            <span>${param.seriesName}: ${status} ${param.value[1]}ms</span>
          </div>`
        })
        return result
      },
    },
    legend: {
      bottom: 10,
      type: 'scroll',
      textStyle: {
        color: '#64748b',
      },
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '15%',
      top: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'time',
      axisLabel: {
        color: '#64748b',
        formatter: (value: number) => {
          const date = new Date(value)
          return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}`
        },
      },
      splitLine: {
        lineStyle: {
          color: '#e2e8f0',
        },
      },
    },
    yAxis: {
      type: 'value',
      name: '延迟 (ms)',
      nameTextStyle: {
        color: '#64748b',
      },
      axisLabel: {
        color: '#64748b',
      },
      splitLine: {
        lineStyle: {
          color: '#e2e8f0',
        },
      },
    },
    series,
  }

  healthChart.setOption(option, true)
}

const handleResize = () => {
  if (healthChart) {
    healthChart.resize()
  }
}

const setupSSE = () => {
  try {
    // Setup SSE connection for real-time health updates
    eventSource = new EventSource('/api/v1/admin/suppliers/health/stream')

    eventSource.onmessage = async (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.supplier_id && supplierHealthHistory.value[data.supplier_id]) {
          // Add new data point to history
          supplierHealthHistory.value[data.supplier_id].push({
            id: `${data.supplier_id}-${Date.now()}`,
            supplier_id: data.supplier_id,
            is_healthy: data.is_healthy,
            latency: data.latency,
            error: data.error || '',
            timestamp: data.last_check,
          })

          // Keep only last 100 points
          if (supplierHealthHistory.value[data.supplier_id].length > 100) {
            supplierHealthHistory.value[data.supplier_id] = supplierHealthHistory.value[data.supplier_id].slice(-100)
          }

          // Update chart
          updateHealthChart()
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

const getAlertColor = (type: string) => {
  const map: Record<string, string> = {
    error: 'bg-red-100 text-red-700 border-red-200',
    warning: 'bg-yellow-100 text-yellow-700 border-yellow-200',
    info: 'bg-blue-100 text-blue-700 border-blue-200',
  }
  return map[type] || 'bg-slate-100 text-slate-600'
}

const getMetricColor = (value: number, threshold: number) => {
  if (value >= threshold) return 'text-red-600'
  if (value >= threshold * 0.8) return 'text-yellow-600'
  return 'text-green-600'
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
      <!-- System Metrics -->
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

      <!-- Event Statistics -->
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div class="bg-white rounded-xl shadow-sm p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-slate-600">失败事件</p>
              <p :class="['text-2xl font-bold mt-1', failureEventCount > 0 ? 'text-red-600' : 'text-slate-900']">
                {{ failureEventCount }}
              </p>
            </div>
            <div class="w-12 h-12 bg-red-100 rounded-full flex items-center justify-center">
              <svg class="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
          </div>
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <div class="flex items-center justify-between">
            <div>
              <p class="text-sm text-slate-600">故障转移</p>
              <p :class="['text-2xl font-bold mt-1', failoverEventCount > 0 ? 'text-orange-600' : 'text-slate-900']">
                {{ failoverEventCount }}
              </p>
            </div>
            <div class="w-12 h-12 bg-orange-100 rounded-full flex items-center justify-center">
              <svg class="w-6 h-6 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7h12m0 0l-4-4m4 4l-4 4m0 6H4m0 0l4 4m-4-4l4-4" />
              </svg>
            </div>
          </div>
        </div>
      </div>

      <!-- Health Trend Chart -->
      <div class="bg-white rounded-xl shadow-sm p-6">
        <h2 class="text-lg font-semibold text-slate-900 mb-4">供应商健康趋势</h2>
        <div ref="healthChartRef" class="w-full h-96"></div>
        <p v-if="suppliers.length === 0" class="text-center py-8 text-slate-500">
          暂无供应商数据
        </p>
      </div>

      <!-- Alerts List -->
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
