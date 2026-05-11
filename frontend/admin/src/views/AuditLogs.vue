<template>
  <div class="audit-logs">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">Audit Logs</h1>
    </div>

    <!-- Filters -->
    <div class="bg-white shadow rounded-lg p-4 mb-6">
      <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-5 gap-4">
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Entity Type</label>
          <select
            v-model="filters.entity_type"
            class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All</option>
            <option value="supplier">Supplier</option>
            <option value="pricing">Pricing</option>
            <option value="api_key">API Key</option>
            <option value="model">Model</option>
            <option value="user">User</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Action</label>
          <select
            v-model="filters.action"
            class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="">All</option>
            <option value="create">Create</option>
            <option value="update">Update</option>
            <option value="delete">Delete</option>
            <option value="rotate">Rotate</option>
            <option value="set_primary">Set Primary</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Time Range</label>
          <select
            v-model="timeRange"
            class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          >
            <option value="1h">Last 1 Hour</option>
            <option value="today">Today</option>
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="custom">Custom</option>
          </select>
        </div>
        <div>
          <label class="block text-sm font-medium text-gray-700 mb-1">Keyword</label>
          <input
            v-model="filters.keyword"
            type="text"
            placeholder="Search entity ID or admin..."
            class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <div class="flex items-end">
          <button
            @click="applyFilters"
            class="w-full px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
          >
            Search
          </button>
        </div>
      </div>
    </div>

    <!-- Logs Table -->
    <div class="bg-white shadow rounded-lg overflow-hidden">
      <div v-if="loading" class="text-center py-8">
        <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent rounded-full"></div>
      </div>

      <div v-else-if="error" class="text-red-600 py-4">
        {{ error }}
      </div>

      <div v-else-if="logs.length === 0" class="text-gray-500 py-8 text-center">
        No audit logs found matching your filters.
      </div>

      <div v-else>
        <table class="min-w-full divide-y divide-gray-200">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Timestamp</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Admin</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entity</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Entity ID</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Action</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Changes</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">IP Address</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
            </tr>
          </thead>
          <tbody class="bg-white divide-y divide-gray-200">
            <tr v-for="log in logs" :key="log.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ formatDateTime(log.timestamp) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                {{ log.admin_name }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  :class="{
                    'px-2 inline-flex text-xs leading-5 font-semibold rounded-full': true,
                    'bg-blue-100 text-blue-800': log.entity_type === 'supplier',
                    'bg-green-100 text-green-800': log.entity_type === 'pricing',
                    'bg-yellow-100 text-yellow-800': log.entity_type === 'api_key',
                    'bg-purple-100 text-purple-800': log.entity_type === 'model',
                    'bg-gray-100 text-gray-800': log.entity_type === 'user',
                  }"
                >
                  {{ formatEntityType(log.entity_type) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-mono">
                {{ truncateId(log.entity_id) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span
                  :class="{
                    'px-2 inline-flex text-xs leading-5 font-semibold rounded-full': true,
                    'bg-green-100 text-green-800': log.action === 'create',
                    'bg-blue-100 text-blue-800': log.action === 'update',
                    'bg-red-100 text-red-800': log.action === 'delete',
                    'bg-orange-100 text-orange-800': log.action === 'rotate',
                    'bg-purple-100 text-purple-800': log.action === 'set_primary',
                  }"
                >
                  {{ formatAction(log.action) }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ getChangesSummary(log.changes) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ log.ip_address }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button
                  @click="viewDetail(log)"
                  class="text-blue-600 hover:text-blue-900"
                >
                  View
                </button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- Pagination -->
        <div class="bg-gray-50 px-6 py-4 border-t border-gray-200 flex items-center justify-between">
          <div class="text-sm text-gray-700">
            Showing {{ (pagination.page - 1) * pagination.page_size + 1 }} to
            {{ Math.min(pagination.page * pagination.page_size, pagination.total) }} of
            {{ pagination.total }} results
          </div>
          <div class="flex gap-2">
            <button
              @click="prevPage"
              :disabled="pagination.page <= 1"
              class="px-3 py-1 border border-gray-300 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Previous
            </button>
            <span class="px-3 py-1 text-gray-700">
              Page {{ pagination.page }} of {{ pagination.total_pages }}
            </span>
            <button
              @click="nextPage"
              :disabled="pagination.page >= pagination.total_pages"
              class="px-3 py-1 border border-gray-300 rounded hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Next
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Detail Dialog -->
    <div v-if="showDetailDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-2xl w-full max-h-[80vh] overflow-y-auto">
        <div class="flex justify-between items-center mb-4">
          <h3 class="text-lg font-semibold">Audit Log Detail</h3>
          <button
            @click="showDetailDialog = false"
            class="text-gray-400 hover:text-gray-600"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div v-if="selectedLog" class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-xs font-medium text-gray-500">Timestamp</label>
              <p class="text-sm">{{ formatDateTime(selectedLog.timestamp) }}</p>
            </div>
            <div>
              <label class="text-xs font-medium text-gray-500">Admin</label>
              <p class="text-sm">{{ selectedLog.admin_name }} ({{ selectedLog.admin_id }})</p>
            </div>
            <div>
              <label class="text-xs font-medium text-gray-500">Entity Type</label>
              <p class="text-sm">{{ formatEntityType(selectedLog.entity_type) }}</p>
            </div>
            <div>
              <label class="text-xs font-medium text-gray-500">Entity ID</label>
              <p class="text-sm font-mono">{{ selectedLog.entity_id }}</p>
            </div>
            <div>
              <label class="text-xs font-medium text-gray-500">Action</label>
              <p class="text-sm">{{ formatAction(selectedLog.action) }}</p>
            </div>
            <div>
              <label class="text-xs font-medium text-gray-500">IP Address</label>
              <p class="text-sm">{{ selectedLog.ip_address }}</p>
            </div>
          </div>

          <div v-if="selectedLog.changes.before || selectedLog.changes.after">
            <label class="text-xs font-medium text-gray-500 block mb-2">Changes</label>
            <div class="grid grid-cols-2 gap-4">
              <div v-if="selectedLog.changes.before" class="bg-red-50 p-3 rounded">
                <p class="text-xs font-medium text-red-800 mb-2">Before</p>
                <pre class="text-xs text-red-700 overflow-x-auto">{{ JSON.stringify(selectedLog.changes.before, null, 2) }}</pre>
              </div>
              <div v-if="selectedLog.changes.after" class="bg-green-50 p-3 rounded">
                <p class="text-xs font-medium text-green-800 mb-2">After</p>
                <pre class="text-xs text-green-700 overflow-x-auto">{{ JSON.stringify(selectedLog.changes.after, null, 2) }}</pre>
              </div>
            </div>
          </div>

          <div>
            <label class="text-xs font-medium text-gray-500">User Agent</label>
            <p class="text-xs text-gray-600 break-all">{{ selectedLog.user_agent }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { auditApi, type AuditLogFilter } from '@/api/audit'
import type { AuditLog } from '@/types/models'

const logs = ref<AuditLog[]>([])
const loading = ref(true)
const error = ref('')
const showDetailDialog = ref(false)
const selectedLog = ref<AuditLog | null>(null)

const filters = ref<AuditLogFilter>({
  entity_type: '',
  action: '',
  keyword: '',
  page: 1,
  page_size: 20,
})

const timeRange = ref('today')
const pagination = ref({
  page: 1,
  page_size: 20,
  total: 0,
  total_pages: 0,
})

const loadLogs = async () => {
  loading.value = true
  error.value = ''
  try {
    applyTimeRange()
    const response = await auditApi.listLogs(filters.value)
    logs.value = response.data
    pagination.value = response.pagination
  } catch (err: any) {
    error.value = err.message || 'Failed to load audit logs'
  } finally {
    loading.value = false
  }
}

const applyTimeRange = () => {
  const now = new Date()

  switch (timeRange.value) {
    case '1h':
      filters.value.start_time = new Date(now.getTime() - 60 * 60 * 1000).toISOString()
      filters.value.end_time = now.toISOString()
      break
    case 'today':
      filters.value.start_time = new Date(now.setHours(0, 0, 0, 0)).toISOString()
      filters.value.end_time = new Date().toISOString()
      break
    case '7d':
      filters.value.start_time = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString()
      filters.value.end_time = new Date().toISOString()
      break
    case '30d':
      filters.value.start_time = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString()
      filters.value.end_time = new Date().toISOString()
      break
  }
}

const applyFilters = () => {
  filters.value.page = 1
  loadLogs()
}

const prevPage = () => {
  if (pagination.value.page > 1) {
    filters.value.page!--
    loadLogs()
  }
}

const nextPage = () => {
  if (pagination.value.page < pagination.value.total_pages) {
    filters.value.page!++
    loadLogs()
  }
}

const viewDetail = (log: AuditLog) => {
  selectedLog.value = log
  showDetailDialog.value = true
}

const formatDateTime = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}

const formatEntityType = (type: string) => {
  const map: Record<string, string> = {
    supplier: 'Supplier',
    pricing: 'Pricing',
    api_key: 'API Key',
    model: 'Model',
    user: 'User',
  }
  return map[type] || type
}

const formatAction = (action: string) => {
  const map: Record<string, string> = {
    create: 'Create',
    update: 'Update',
    delete: 'Delete',
    rotate: 'Rotate',
    set_primary: 'Set Primary',
  }
  return map[action] || action
}

const truncateId = (id: string) => {
  if (id.length <= 12) return id
  return id.substring(0, 8) + '...'
}

const getChangesSummary = (changes: { before?: Record<string, any>; after?: Record<string, any> }) => {
  if (!changes.before && !changes.after) return '-'

  const keys = new Set([
    ...Object.keys(changes.before || {}),
    ...Object.keys(changes.after || {}),
  ])

  const changedKeys = Array.from(keys).filter(key => {
    const before = JSON.stringify(changes.before?.[key])
    const after = JSON.stringify(changes.after?.[key])
    return before !== after
  })

  if (changedKeys.length === 0) return 'No changes'
  if (changedKeys.length <= 2) return changedKeys.join(', ')
  return `${changedKeys.length} fields changed`
}

onMounted(() => {
  loadLogs()
})
</script>
