<template>
  <div class="model-association">
    <div class="flex justify-between items-center mb-4">
      <h3 class="text-lg font-semibold">Model Associations</h3>
      <button
        @click="showAddDialog = true"
        class="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 transition"
      >
        Add Model
      </button>
    </div>

    <div v-if="loading" class="text-center py-8">
      <div class="animate-spin inline-block w-8 h-8 border-4 border-current border-t-transparent rounded-full"></div>
    </div>

    <div v-else-if="error" class="text-red-600 py-4">
      {{ error }}
    </div>

    <div v-else-if="models.length === 0" class="text-gray-500 py-8 text-center">
      No models associated. Add a model to get started.
    </div>

    <table v-else class="min-w-full divide-y divide-gray-200">
      <thead class="bg-gray-50">
        <tr>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Model</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Input Cost</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Output Cost</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Effective Date</th>
          <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Status</th>
          <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
        </tr>
      </thead>
      <tbody class="bg-white divide-y divide-gray-200">
        <tr v-for="model in models" :key="model.id" class="hover:bg-gray-50">
          <td class="px-6 py-4 whitespace-nowrap">
            <div class="text-sm font-medium text-gray-900">{{ model.model_id }}</div>
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
            ${{ model.input_cost.toFixed(4) }} / 1K tokens
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
            ${{ model.output_cost.toFixed(4) }} / 1K tokens
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
            {{ formatDate(model.effective_date) }}
          </td>
          <td class="px-6 py-4 whitespace-nowrap">
            <span
              :class="{
                'px-2 inline-flex text-xs leading-5 font-semibold rounded-full': true,
                'bg-green-100 text-green-800': model.is_active,
                'bg-red-100 text-red-800': !model.is_active,
              }"
            >
              {{ model.is_active ? 'Active' : 'Inactive' }}
            </span>
          </td>
          <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
            <button
              @click="editModel(model)"
              class="text-blue-600 hover:text-blue-900 mr-3"
            >
              Edit
            </button>
            <button
              @click="removeModel(model.id)"
              class="text-red-600 hover:text-red-900"
            >
              Remove
            </button>
          </td>
        </tr>
      </tbody>
    </table>

    <!-- Add/Edit Model Dialog -->
    <div v-if="showAddDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 max-w-md w-full">
        <h3 class="text-lg font-semibold mb-4">
          {{ editingModel ? 'Edit Model Cost' : 'Add Model' }}
        </h3>
        <form @submit.prevent="saveModel">
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Model ID</label>
            <input
              v-model="modelForm.model_id"
              type="text"
              required
              :disabled="!!editingModel"
              class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:bg-gray-100"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Input Cost ($/1K tokens)</label>
            <input
              v-model.number="modelForm.input_cost"
              type="number"
              step="0.0001"
              min="0"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 mb-1">Output Cost ($/1K tokens)</label>
            <input
              v-model.number="modelForm.output_cost"
              type="number"
              step="0.0001"
              min="0"
              required
              class="w-full px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
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
import type { SupplierModel } from '@/types/models'

interface Props {
  supplierId: string
}

const props = defineProps<Props>()

const models = ref<SupplierModel[]>([])
const loading = ref(true)
const error = ref('')
const showAddDialog = ref(false)
const saving = ref(false)
const successMessage = ref('')
const editingModel = ref<SupplierModel | null>(null)

const modelForm = ref({
  model_id: '',
  input_cost: 0,
  output_cost: 0,
})

const loadModels = async () => {
  loading.value = true
  error.value = ''
  try {
    models.value = await suppliersApi.listModels(props.supplierId)
  } catch (err: any) {
    error.value = err.message || 'Failed to load models'
  } finally {
    loading.value = false
  }
}

const editModel = (model: SupplierModel) => {
  editingModel.value = model
  modelForm.value = {
    model_id: model.model_id,
    input_cost: model.input_cost,
    output_cost: model.output_cost,
  }
  showAddDialog.value = true
}

const saveModel = async () => {
  saving.value = true
  try {
    if (editingModel.value) {
      await suppliersApi.updateModelCost(props.supplierId, editingModel.value.id, {
        input_cost: modelForm.value.input_cost,
        output_cost: modelForm.value.output_cost,
      })
      successMessage.value = 'Model cost updated successfully'
    } else {
      await suppliersApi.addModel(props.supplierId, modelForm.value)
      successMessage.value = 'Model added successfully'
    }
    closeDialog()
    setTimeout(() => successMessage.value = '', 3000)
    await loadModels()
  } catch (err: any) {
    error.value = err.message || 'Failed to save model'
  } finally {
    saving.value = false
  }
}

const removeModel = async (modelId: string) => {
  if (!confirm('Are you sure you want to remove this model association?')) {
    return
  }
  try {
    await suppliersApi.removeModel(props.supplierId, modelId)
    successMessage.value = 'Model removed successfully'
    setTimeout(() => successMessage.value = '', 3000)
    await loadModels()
  } catch (err: any) {
    error.value = err.message || 'Failed to remove model'
  }
}

const closeDialog = () => {
  showAddDialog.value = false
  editingModel.value = null
  modelForm.value = { model_id: '', input_cost: 0, output_cost: 0 }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString()
}

onMounted(() => {
  loadModels()
})
</script>
