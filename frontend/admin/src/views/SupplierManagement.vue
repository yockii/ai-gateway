<template>
  <div class="supplier-management">
    <!-- Supplier List View -->
    <div v-if="!selectedSupplier">
      <div class="flex justify-between items-center mb-6">
        <h1 class="text-2xl font-bold">Supplier Management</h1>
      </div>

      <div v-if="loading" class="text-center py-8">
        <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent rounded-full"></div>
      </div>

      <div v-else-if="error" class="text-red-600 py-4">
        {{ error }}
      </div>

      <table v-else class="min-w-full divide-y divide-gray-200">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Display Name</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Provider</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Health</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Models</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr
            v-for="supplier in suppliers"
            :key="supplier.id"
            @click="selectSupplier(supplier)"
            class="hover:bg-gray-50 cursor-pointer"
          >
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-gray-900">{{ supplier.name }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ supplier.display_name }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
              {{ supplier.provider }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <health-indicator :supplier-id="supplier.id" :auto-connect="false" />
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ supplierModelCount[supplier.id] || 0 }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span
                :class="{
                  'px-2 inline-flex text-xs leading-5 font-semibold rounded-full': true,
                  'bg-green-100 text-green-800': supplier.is_active,
                  'bg-red-100 text-red-800': !supplier.is_active,
                }"
              >
                {{ supplier.is_active ? 'Active' : 'Inactive' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button
                @click.stop="editSupplier(supplier)"
                class="text-blue-600 hover:text-blue-900"
              >
                Manage
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Supplier Detail View -->
    <div v-else>
      <!-- Breadcrumb -->
      <div class="flex items-center gap-2 mb-6">
        <button
          @click="selectedSupplier = null"
          class="text-blue-600 hover:text-blue-800 flex items-center gap-1"
        >
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
          </svg>
          Back to List
        </button>
        <span class="text-gray-400">/</span>
        <span class="text-gray-600">{{ selectedSupplier.display_name }}</span>
      </div>

      <!-- Supplier Info Card -->
      <div class="bg-white shadow rounded-lg p-6 mb-6">
        <div class="flex justify-between items-start">
          <div>
            <h2 class="text-xl font-bold text-gray-900">{{ selectedSupplier.display_name }}</h2>
            <p class="text-gray-500 mt-1">
              {{ selectedSupplier.provider }} • {{ selectedSupplier.name }}
            </p>
          </div>
          <div class="flex gap-2">
            <health-indicator :supplier-id="selectedSupplier.id" />
          </div>
        </div>
      </div>

      <!-- Tabs -->
      <div class="border-b border-gray-200">
        <nav class="flex gap-4 -mb-px">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key"
            :class="{
              'px-4 py-2 border-b-2 font-medium text-sm transition-colors': true,
              'border-blue-500 text-blue-600': activeTab === tab.key,
              'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300': activeTab !== tab.key,
            }"
          >
            {{ tab.label }}
          </button>
        </nav>
      </div>

      <!-- Tab Content -->
      <div class="mt-6">
        <!-- API Keys Tab -->
        <div v-if="activeTab === 'api-keys'">
          <api-key-list :supplier-id="selectedSupplier.id" />
        </div>

        <!-- Models Tab -->
        <div v-else-if="activeTab === 'models'">
          <model-association :supplier-id="selectedSupplier.id" />
        </div>

        <!-- Health Tab -->
        <div v-else-if="activeTab === 'health'">
          <div class="bg-white shadow rounded-lg p-6">
            <h3 class="text-lg font-semibold mb-4">Health Status</h3>
            <health-indicator :supplier-id="selectedSupplier.id" />
            <div class="mt-6">
              <h4 class="text-md font-medium mb-2">Health History</h4>
              <div ref="healthChart" class="w-full h-64"></div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, nextTick, watch } from 'vue'
import { suppliersApi } from '@/api/suppliers'
import type { Supplier } from '@/types/models'
import ApiKeyList from '@/components/suppliers/ApiKeyList.vue'
import ModelAssociation from '@/components/suppliers/ModelAssociation.vue'
import HealthIndicator from '@/components/suppliers/HealthIndicator.vue'
import * as echarts from 'echarts'

const suppliers = ref<Supplier[]>([])
const supplierModelCount = ref<Record<string, number>>({})
const loading = ref(true)
const error = ref('')
const selectedSupplier = ref<Supplier | null>(null)
const activeTab = ref('api-keys')
const healthChart = ref<HTMLElement | null>(null)

const tabs = [
  { key: 'api-keys', label: 'API Keys' },
  { key: 'models', label: 'Model Associations' },
  { key: 'health', label: 'Health Status' },
]

const loadSuppliers = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = await suppliersApi.list()
    suppliers.value = data.data || []

    // 修复 WR-02: 使用后端返回的 model_count 而不是 N+1 查询
    const modelCountMap: Record<string, number> = {}
    for (const supplier of suppliers.value) {
      if (typeof supplier.model_count === 'number') {
        modelCountMap[supplier.id] = supplier.model_count
      } else {
        // 降级：如果后端没有返回 model_count，设为 0
        modelCountMap[supplier.id] = 0
      }
    }
    supplierModelCount.value = modelCountMap
  } catch (err: any) {
    error.value = err.message || 'Failed to load suppliers'
  } finally {
    loading.value = false
  }
}

const selectSupplier = (supplier: Supplier) => {
  selectedSupplier.value = supplier
  activeTab.value = 'api-keys'
}

const editSupplier = (supplier: Supplier) => {
  selectSupplier(supplier)
}

const loadHealthChart = async () => {
  if (!selectedSupplier.value || activeTab.value !== 'health' || !healthChart.value) {
    return
  }

  try {
    const history = await suppliersApi.getHealthHistory(selectedSupplier.value.id, 50)

    const chart = echarts.init(healthChart.value)
    chart.setOption({
      title: {
        text: 'Health History',
        left: 'center',
      },
      tooltip: {
        trigger: 'axis',
      },
      xAxis: {
        type: 'category',
        data: history.map((h) => new Date(h.timestamp).toLocaleTimeString()),
      },
      yAxis: {
        type: 'value',
        name: 'Latency (ms)',
      },
      series: [
        {
          name: 'Latency',
          type: 'line',
          data: history.map((h) => h.latency),
          smooth: true,
          itemStyle: {
            color: '#3b82f6',
          },
        },
      ],
    })
  } catch (err: any) {
    console.error('Failed to load health chart:', err)
  }
}

onMounted(() => {
  loadSuppliers()
})

// 修复 WR-04: 使用 Vue watch API 替代不安全的轮询
watch(activeTab, (newTab) => {
  nextTick(() => {
    if (newTab === 'health') {
      loadHealthChart()
    }
  })
})
</script>
