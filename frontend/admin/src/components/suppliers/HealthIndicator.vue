<template>
  <div class="health-indicator">
    <div class="flex items-center gap-4">
      <!-- Health Status Dot -->
      <div class="flex items-center gap-2">
        <div
          :class="{
            'w-3 h-3 rounded-full': true,
            'bg-green-500': healthStatus.is_healthy,
            'bg-red-500': !healthStatus.is_healthy,
            'animate-pulse': loading,
          }"
        ></div>
        <span class="text-sm font-medium">
          {{ healthStatus.is_healthy ? 'Healthy' : 'Unhealthy' }}
        </span>
      </div>

      <!-- Latency -->
      <div class="flex items-center gap-2">
        <span class="text-sm text-gray-500">Latency:</span>
        <span
          :class="{
            'text-sm font-medium': true,
            'text-green-600': healthStatus.latency < 100,
            'text-yellow-600': healthStatus.latency >= 100 && healthStatus.latency < 500,
            'text-red-600': healthStatus.latency >= 500,
          }"
        >
          {{ healthStatus.latency }}ms
        </span>
      </div>

      <!-- Last Check -->
      <div class="text-sm text-gray-500">
        Last check: {{ formatTime(healthStatus.last_check) }}
      </div>

      <!-- Refresh Button -->
      <button
        @click="refreshHealth"
        :disabled="loading"
        class="p-1 hover:bg-gray-100 rounded transition"
        title="Refresh health status"
      >
        <svg
          :class="{ 'animate-spin': loading }"
          class="w-4 h-4"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
          />
        </svg>
      </button>
    </div>

    <!-- Error Message -->
    <div v-if="!healthStatus.is_healthy && healthStatus.error" class="mt-2 text-sm text-red-600">
      {{ healthStatus.error }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { suppliersApi } from '@/api/suppliers'
import type { HealthCheckResult, HealthStatus } from '@/types/models'

interface Props {
  supplierId: string
  autoConnect?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  autoConnect: true,
})

const loading = ref(false)
const healthStatus = ref<HealthStatus>({
  is_healthy: false,
  latency: 0,
  last_check: new Date().toISOString(),
})

let pollInterval: ReturnType<typeof setInterval> | null = null

const loadHealthStatus = async () => {
  loading.value = true
  try {
    const result = await suppliersApi.getHealth(props.supplierId)
    healthStatus.value = {
      is_healthy: result.is_healthy,
      latency: result.latency,
      last_check: result.timestamp,
    }
  } catch (err: any) {
    console.error('Failed to load health status:', err)
    healthStatus.value.is_healthy = false
    healthStatus.value.error = err.message || 'Failed to check health'
  } finally {
    loading.value = false
  }
}

const refreshHealth = () => {
  loadHealthStatus()
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

const startPolling = () => {
  if (pollInterval) return

  loadHealthStatus()
  pollInterval = setInterval(() => {
    loadHealthStatus()
  }, 30000) // Poll every 30 seconds
}

const stopPolling = () => {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}

onMounted(() => {
  if (props.autoConnect) {
    startPolling()
  } else {
    loadHealthStatus()
  }
})

onUnmounted(() => {
  stopPolling()
})

defineExpose({
  refresh: refreshHealth,
})
</script>
