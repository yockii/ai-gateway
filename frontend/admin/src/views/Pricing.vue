<template>
  <div class="pricing-management">
    <div class="flex justify-between items-center mb-6">
      <h1 class="text-2xl font-bold">Pricing Management</h1>
    </div>

    <!-- Tabs -->
    <div class="border-b border-gray-200 mb-6">
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

    <!-- Enterprise Pricing Tab -->
    <div v-if="activeTab === 'enterprise'">
      <div class="flex justify-between items-center mb-4">
        <div class="flex gap-2">
          <input
            v-model="searchKeyword"
            type="text"
            placeholder="Search by customer or model..."
            class="px-4 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>
        <button
          @click="openCreateDialog"
          class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
        >
          Add Pricing
        </button>
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
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Customer</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Model</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Input Price</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Output Price</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Profit Margin</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Effective</th>
            <th class="px-6 py-3 text-center text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
          </tr>
        </thead>
        <tbody class="bg-white divide-y divide-gray-200">
          <tr v-for="pricing in filteredPricings" :key="pricing.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 whitespace-nowrap">
              <div class="text-sm font-medium text-gray-900">{{ pricing.customer_name }}</div>
              <div class="text-xs text-gray-500">{{ pricing.customer_id }}</div>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 font-mono">
              {{ pricing.model_id }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 text-right">
              ${{ pricing.input_price.toFixed(4) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900 text-right">
              ${{ pricing.output_price.toFixed(4) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-right">
              <span
                :class="{
                  'font-medium': true,
                  'text-green-600': pricing.min_profit_margin >= 0.1,
                  'text-yellow-600': pricing.min_profit_margin >= 0.05 && pricing.min_profit_margin < 0.1,
                  'text-red-600': pricing.min_profit_margin < 0.05,
                }"
              >
                {{ (pricing.min_profit_margin * 100).toFixed(1) }}%
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDate(pricing.effective_date) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-center">
              <span
                :class="{
                  'px-2 inline-flex text-xs leading-5 font-semibold rounded-full': true,
                  'bg-green-100 text-green-800': pricing.is_active,
                  'bg-red-100 text-red-800': !pricing.is_active,
                }"
              >
                {{ pricing.is_active ? 'Active' : 'Inactive' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
              <button
                @click="openEditDialog(pricing)"
                class="text-blue-600 hover:text-blue-900 mr-3"
              >
                Edit
              </button>
              <button
                @click="deletePricing(pricing.id)"
                class="text-red-600 hover:text-red-900"
              >
                Delete
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Group Pricing Tab (Placeholder) -->
    <div v-else-if="activeTab === 'group'" class="text-center py-12 text-gray-500">
      <svg class="w-16 h-16 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p class="text-lg font-medium">Coming Soon</p>
      <p class="text-sm">User group pricing management will be available in a future release.</p>
    </div>

    <!-- Membership Discount Tab (Placeholder) -->
    <div v-else-if="activeTab === 'membership'" class="text-center py-12 text-gray-500">
      <svg class="w-16 h-16 mx-auto mb-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
      </svg>
      <p class="text-lg font-medium">Coming Soon</p>
      <p class="text-sm">Membership discount configuration will be available in a future release.</p>
    </div>

    <!-- Create/Edit Dialog -->
    <div v-if="showDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-semibold mb-4">
          {{ editingPricing ? 'Edit Pricing' : 'Create Pricing' }}
        </h3>
        <form @submit.prevent="savePricing">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Customer ID *</label>
              <input
                v-model="formData.customer_id"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Customer Name *</label>
              <input
                v-model="formData.customer_name"
                type="text"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Model ID *</label>
              <input
                v-model="formData.model_id"
                type="text"
                required
                placeholder="e.g., gpt-4"
                class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
              />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Input Price ($)</label>
                <input
                  v-model.number="formData.input_price"
                  type="number"
                  step="0.0001"
                  min="0"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">Output Price ($)</label>
                <input
                  v-model.number="formData.output_price"
                  type="number"
                  step="0.0001"
                  min="0"
                  required
                  class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">
                Min Profit Margin (%)
              </label>
              <input
                v-model.number="formData.min_profit_margin"
                type="number"
                step="0.01"
                min="0"
                max="100"
                required
                class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p class="text-xs text-gray-500 mt-1">Recommended: 10% or higher</p>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Effective Date *</label>
              <input
                v-model="formData.effective_date"
                type="date"
                required
                :min="todayDate"
                class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">Expiry Date (Optional)</label>
              <input
                v-model="formData.expiry_date"
                type="date"
                class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div v-if="formErrors.length > 0" class="bg-red-50 border border-red-200 rounded p-3">
              <p class="text-sm text-red-600" v-for="error in formErrors" :key="error">{{ error }}</p>
            </div>
            <div class="flex justify-end gap-2">
              <button
                type="button"
                @click="closeDialog"
                class="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded transition"
              >
                Cancel
              </button>
              <button
                type="submit"
                :disabled="saving"
                class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition disabled:opacity-50"
              >
                {{ saving ? 'Saving...' : 'Save' }}
              </button>
            </div>
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
import { ref, computed, onMounted } from 'vue'
import { pricingApi } from '@/api/pricing'
import type { EnterprisePricing } from '@/types/models'

const tabs = [
  { key: 'enterprise', label: 'Enterprise Pricing' },
  { key: 'group', label: 'Group Pricing' },
  { key: 'membership', label: 'Membership Discounts' },
]

const activeTab = ref('enterprise')
const pricings = ref<EnterprisePricing[]>([])
const loading = ref(true)
const error = ref('')
const searchKeyword = ref('')
const showDialog = ref(false)
const saving = ref(false)
const successMessage = ref('')
const editingPricing = ref<EnterprisePricing | null>(null)
const formErrors = ref<string[]>([])

const formData = ref({
  customer_id: '',
  customer_name: '',
  model_id: '',
  input_price: 0,
  output_price: 0,
  min_profit_margin: 10,
  effective_date: new Date().toISOString().split('T')[0],
  expiry_date: '',
  is_active: true,
})

const todayDate = computed(() => {
  return new Date().toISOString().split('T')[0]
})

const filteredPricings = computed(() => {
  if (!searchKeyword.value) return pricings.value
  const keyword = searchKeyword.value.toLowerCase()
  return pricings.value.filter(
    (p) =>
      p.customer_name.toLowerCase().includes(keyword) ||
      p.customer_id.toLowerCase().includes(keyword) ||
      p.model_id.toLowerCase().includes(keyword)
  )
})

const loadPricings = async () => {
  loading.value = true
  error.value = ''
  try {
    pricings.value = await pricingApi.listEnterprise()
  } catch (err: any) {
    error.value = err.message || 'Failed to load pricings'
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  editingPricing.value = null
  formData.value = {
    customer_id: '',
    customer_name: '',
    model_id: '',
    input_price: 0,
    output_price: 0,
    min_profit_margin: 10,
    effective_date: new Date().toISOString().split('T')[0],
    expiry_date: '',
    is_active: true,
  }
  formErrors.value = []
  showDialog.value = true
}

const openEditDialog = (pricing: EnterprisePricing) => {
  editingPricing.value = pricing
  formData.value = {
    customer_id: pricing.customer_id,
    customer_name: pricing.customer_name,
    model_id: pricing.model_id,
    input_price: pricing.input_price,
    output_price: pricing.output_price,
    min_profit_margin: pricing.min_profit_margin * 100,
    effective_date: pricing.effective_date.split('T')[0],
    expiry_date: pricing.expiry_date ? pricing.expiry_date.split('T')[0] : '',
    is_active: pricing.is_active,
  }
  formErrors.value = []
  showDialog.value = true
}

const validateForm = (): boolean => {
  formErrors.value = []

  if (formData.value.input_price <= 0) {
    formErrors.value.push('Input price must be greater than 0')
  }
  if (formData.value.output_price <= 0) {
    formErrors.value.push('Output price must be greater than 0')
  }
  if (formData.value.min_profit_margin < 0 || formData.value.min_profit_margin > 100) {
    formErrors.value.push('Profit margin must be between 0 and 100')
  }
  const effectiveDate = new Date(formData.value.effective_date)
  const today = new Date()
  today.setHours(0, 0, 0, 0)
  if (effectiveDate < today) {
    formErrors.value.push('Effective date cannot be in the past')
  }

  return formErrors.value.length === 0
}

const savePricing = async () => {
  if (!validateForm()) {
    return
  }

  saving.value = true
  try {
    const data = {
      ...formData.value,
      min_profit_margin: formData.value.min_profit_margin / 100,
      effective_date: new Date(formData.value.effective_date).toISOString(),
      expiry_date: formData.value.expiry_date ? new Date(formData.value.expiry_date).toISOString() : undefined,
    }

    if (editingPricing.value) {
      await pricingApi.updateEnterprise(editingPricing.value.id, data)
      successMessage.value = 'Pricing updated successfully'
    } else {
      await pricingApi.createEnterprise(data)
      successMessage.value = 'Pricing created successfully'
    }

    closeDialog()
    setTimeout(() => successMessage.value = '', 3000)
    await loadPricings()
  } catch (err: any) {
    formErrors.value = [err.message || 'Failed to save pricing']
  } finally {
    saving.value = false
  }
}

const deletePricing = async (id: string) => {
  if (!confirm('Are you sure you want to delete this pricing?')) {
    return
  }
  try {
    await pricingApi.deleteEnterprise(id)
    successMessage.value = 'Pricing deleted successfully'
    setTimeout(() => successMessage.value = '', 3000)
    await loadPricings()
  } catch (err: any) {
    error.value = err.message || 'Failed to delete pricing'
  }
}

const closeDialog = () => {
  showDialog.value = false
  editingPricing.value = null
  formErrors.value = []
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => {
  loadPricings()
})
</script>
