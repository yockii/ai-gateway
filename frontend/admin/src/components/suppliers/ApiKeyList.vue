<template>
  <div class="api-key-list">
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-lg font-semibold">API Keys</h3>
      <button
        @click="showCreateDialog = true"
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
      >
        Add API Key
      </button>
    </div>

    <div v-if="loading" class="text-center py-8">
      <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent rounded-full"></div>
    </div>

    <div v-else-if="error" class="text-red-600 py-4">
      {{ error }}
    </div>

    <div v-else-if="apiKeys.length === 0" class="text-gray-500 py-8 text-center">
      No API keys found. Create one to get started.
    </div>

    <table v-else class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Key Prefix</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Priority</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Usage</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Used</th>
          <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
        </tr>
      </thead>
      <tbody class="bg-white divide-y divide-gray-200">
        <tr v-for="key in apiKeys" :key="key.id" class="hover:bg-gray-50">
          <td class="px-6 py-4 whitespace-nowrap">
            <div class="flex items-center">
              <div>
                <div class="text-sm font-medium text-gray-900">{{ key.name }}</div>
                <div v-if="key.is_primary" class="text-xs text-blue-600">Primary Key</div>
              </div>
            </div>
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 font-mono">
            {{ key.key_prefix }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
            {{ key.priority }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap">
            <span
              :class="{
                'px-2 inline-flex text-xs leading-5 font-semibold rounded-full': true,
                'bg-green-100 text-green-800': key.is_active,
                'bg-red-100 text-red-800': !key.is_active,
              }"
            >
              {{ key.is_active ? 'Active' : 'Inactive' }}
            </span>
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
            {{ key.current_requests }} / {{ key.max_requests }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            {{ key.last_used_at ? formatDate(key.last_used_at) : 'Never' }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
            <button
              v-if="!key.is_primary"
              @click="setPrimaryKey(key.id)"
              class="text-blue-600 hover:text-blue-900 mr-3"
            >
              Set Primary
            </button>
            <button
              @click="rotateKey(key.id)"
              class="text-orange-600 hover:text-orange-900 mr-3"
            >
              Rotate
            </button>
            <button
              @click="deleteKey(key.id)"
              class="text-red-600 hover:text-red-900"
            >
              Delete
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Create API Key Dialog -->
    <div v-if="showCreateDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-semibold mb-4">Create API Key</h3>
        <form @submit.prevent="createApiKey">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
            <input
              v-model="newKey.name"
              type="text"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">API Key</label>
            <input
              v-model="newKey.api_key"
              type="password"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Priority</label>
            <input
              v-model.number="newKey.priority"
              type="number"
              class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Max Requests</label>
            <input
              v-model.number="newKey.max_requests"
              type="number"
              class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div class="flex justify-end gap-2">
            <button
              type="button"
              @click="showCreateDialog = false"
              class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded transition"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="creating"
              class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition disabled:opacity-50"
            >
              {{ creating ? 'Creating...' : 'Create' }}
            </button>
          </div>
        </form>
      </div>
    </div>

    <!-- Success Message -->
    <div v-if="successMessage" class="fixed bottom-4 right-4 bg-green-500 text-white px-4 py-2 rounded shadow-lg">
      {{ successMessage }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { suppliersApi } from '@/api/suppliers'
import type { SupplierApiKey } from '@/types/models'

interface Props {
  supplierId: string
}

const props = defineProps<Props>()

const apiKeys = ref<SupplierApiKey[]>([])
const loading = ref(true)
const error = ref('')
const showCreateDialog = ref(false)
const creating = ref(false)
const successMessage = ref('')

const newKey = ref({
  name: '',
  api_key: '',
  priority: 1,
  max_requests: 1000,
})

const loadApiKeys = async () => {
  loading.value = true
  error.value = ''
  try {
    apiKeys.value = await suppliersApi.listApiKeys(props.supplierId)
  } catch (err: any) {
    error.value = err.message || 'Failed to load API keys'
  } finally {
    loading.value = false
  }
}

const createApiKey = async () => {
  creating.value = true
  try {
    await suppliersApi.createApiKey(props.supplierId, newKey.value)
    showCreateDialog.value = false
    newKey.value = { name: '', api_key: '', priority: 1, max_requests: 1000 }
    successMessage.value = 'API Key created successfully'
    setTimeout(() => successMessage.value = '', 3000)
    await loadApiKeys()
  } catch (err: any) {
    error.value = err.message || 'Failed to create API key'
  } finally {
    creating.value = false
  }
}

const setPrimaryKey = async (keyId: string) => {
  try {
    await suppliersApi.setPrimary(props.supplierId, keyId)
    successMessage.value = 'Primary key set successfully'
    setTimeout(() => successMessage.value = '', 3000)
    await loadApiKeys()
  } catch (err: any) {
    error.value = err.message || 'Failed to set primary key'
  }
}

const rotateKey = async (keyId: string) => {
  if (!confirm('Are you sure you want to rotate this key? The old key will be invalidated.')) {
    return
  }
  try {
    await suppliersApi.rotateApiKey(props.supplierId, keyId)
    successMessage.value = 'Key rotated successfully'
    setTimeout(() => successMessage.value = '', 3000)
    await loadApiKeys()
  } catch (err: any) {
    error.value = err.message || 'Failed to rotate key'
  }
}

const deleteKey = async (keyId: string) => {
  if (!confirm('Are you sure you want to delete this key?')) {
    return
  }
  try {
    await suppliersApi.deleteApiKey(props.supplierId, keyId)
    successMessage.value = 'Key deleted successfully'
    setTimeout(() => successMessage.value = '', 3000)
    await loadApiKeys()
  } catch (err: any) {
    error.value = err.message || 'Failed to delete key'
  }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}

onMounted(() => {
  loadApiKeys()
})
</script>
