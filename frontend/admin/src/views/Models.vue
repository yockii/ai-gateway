<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { modelsApi } from '@/api/models'
import { suppliersApi } from '@/api/suppliers'
import type { ExternalModel, Supplier } from '@/types/models'
import { Plus, Power, Edit2, Trash2, Route, Save, X, ChevronDown, ChevronUp } from 'lucide-vue-next'

const models = ref<ExternalModel[]>([])
const suppliers = ref<Supplier[]>([])
const loading = ref(false)
const showEditDialog = ref(false)
const showRouteDialog = ref(false)
const editingModel = ref<ExternalModel | null>(null)
const selectedModel = ref<ExternalModel | null>(null)

// 表单数据
const formData = ref({ name: '', display_name: '', model_type: '' })

// 模型路由配置
interface ModelRoute {
  supplier_id: string
  supplier_name: string
  is_active: boolean
  priority: number
  input_cost: number
  output_cost: number
}

const modelRoutes = ref<ModelRoute[]>([])
const expandedRoutes = ref<Set<string>>(new Set())

onMounted(async () => {
  await Promise.all([fetchModels(), fetchSuppliers()])
})

const fetchModels = async () => {
  loading.value = true
  try {
    models.value = await modelsApi.list()
  } catch (err) {
    console.error('Failed to fetch models:', err)
  } finally {
    loading.value = false
  }
}

const fetchSuppliers = async () => {
  try {
    suppliers.value = await suppliersApi.list()
  } catch (err) {
    console.error('Failed to fetch suppliers:', err)
  }
}

const openCreateDialog = () => {
  editingModel.value = null
  formData.value = { name: '', display_name: '', model_type: '' }
  showEditDialog.value = true
}

const openEditDialog = (model: ExternalModel) => {
  editingModel.value = model
  formData.value = { name: model.name, display_name: model.display_name, model_type: model.model_type }
  showEditDialog.value = true
}

const save = async () => {
  try {
    if (editingModel.value) {
      await modelsApi.update(editingModel.value.id, formData.value)
    } else {
      await modelsApi.create(formData.value)
    }
    await fetchModels()
    showEditDialog.value = false
  } catch (err: any) {
    alert(err.message || '保存失败')
  }
}

const toggleStatus = async (model: ExternalModel) => {
  try {
    await modelsApi.toggleStatus(model.id)
    await fetchModels()
  } catch (err: any) {
    alert(err.message || '操作失败')
  }
}

const deleteModel = async (model: ExternalModel) => {
  if (!confirm(`确定要删除模型 ${model.display_name} 吗？`)) return
  try {
    await modelsApi.delete(model.id)
    await fetchModels()
  } catch (err: any) {
    alert(err.message || '删除失败')
  }
}

const openRouteDialog = async (model: ExternalModel) => {
  selectedModel.value = model
  await loadModelRoutes(model.id)
  showRouteDialog.value = true
}

const loadModelRoutes = async (modelId: string) => {
  // 初始化所有供应商为可用路由
  modelRoutes.value = suppliers.value.map(supplier => ({
    supplier_id: supplier.id,
    supplier_name: supplier.display_name || supplier.name,
    is_active: supplier.is_active,
    priority: 1,
    input_cost: 0,
    output_cost: 0,
  }))

  // TODO: 从后端获取实际的路由配置
  // const routes = await modelsApi.getModelRoutes(modelId)
  // modelRoutes.value = routes
}

const saveRoutes = async () => {
  if (!selectedModel.value) return
  // TODO: 调用 API 保存路由配置
  console.log('Saving routes for model:', selectedModel.value.id, modelRoutes.value)
  alert('路由配置已保存（演示）')
  showRouteDialog.value = false
}

const toggleExpand = (supplierId: string) => {
  if (expandedRoutes.value.has(supplierId)) {
    expandedRoutes.value.delete(supplierId)
  } else {
    expandedRoutes.value.add(supplierId)
  }
}

const getTypeBadge = (type: string) => {
  const map: Record<string, string> = {
    chat: 'bg-blue-100 text-blue-700',
    completion: 'bg-green-100 text-green-700',
    embedding: 'bg-purple-100 text-purple-700',
    image: 'bg-orange-100 text-orange-700',
  }
  return map[type] || 'bg-slate-100 text-slate-600'
}

const getTypeName = (type: string) => {
  const map: Record<string, string> = {
    chat: '对话',
    completion: '补全',
    embedding: '嵌入',
    image: '图像',
  }
  return map[type] || type
}

const formatPrice = (price: number) => {
  return `$${price.toFixed(4)}`
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">模型管理</h1>
      <button @click="openCreateDialog" class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
        <Plus :size="18" />
        添加模型
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">加载中...</div>

    <div v-else class="bg-white rounded-xl shadow-sm overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">显示名称</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">模型名称</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">类型</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">路由配置</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">状态</th>
            <th class="text-right px-6 py-3 text-sm font-medium text-slate-600">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="model in models" :key="model.id" class="hover:bg-slate-50">
            <td class="px-6 py-4 text-sm text-slate-900">{{ model.display_name }}</td>
            <td class="px-6 py-4 text-sm text-slate-600 font-mono">{{ model.name }}</td>
            <td class="px-6 py-4">
              <span :class="['px-2 py-1 text-xs font-medium rounded-full', getTypeBadge(model.model_type)]">
                {{ getTypeName(model.model_type) }}
              </span>
            </td>
            <td class="px-6 py-4">
              <button
                @click="openRouteDialog(model)"
                class="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800"
              >
                <Route :size="14" />
                配置路由
              </button>
            </td>
            <td class="px-6 py-4">
              <span :class="[
                'px-2 py-1 text-xs font-medium rounded-full',
                model.is_active ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600',
              ]">
                {{ model.is_active ? '启用' : '禁用' }}
              </span>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center justify-end gap-2">
                <button @click="openEditDialog(model)" class="p-2 rounded hover:bg-slate-100 text-slate-600">
                  <Edit2 :size="16" />
                </button>
                <button @click="toggleStatus(model)" class="p-2 rounded hover:bg-slate-100 text-slate-600">
                  <Power :size="16" />
                </button>
                <button @click="deleteModel(model)" class="p-2 rounded hover:bg-red-100 text-red-600">
                  <Trash2 :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 模型编辑对话框 -->
    <div v-if="showEditDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h2 class="text-xl font-bold mb-4">{{ editingModel ? '编辑模型' : '添加模型' }}</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">显示名称</label>
            <input v-model="formData.display_name" type="text" class="w-full px-4 py-2 border border-slate-300 rounded-lg" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">模型名称</label>
            <input v-model="formData.name" type="text" class="w-full px-4 py-2 border border-slate-300 rounded-lg font-mono" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">类型</label>
            <select v-model="formData.model_type" class="w-full px-4 py-2 border border-slate-300 rounded-lg">
              <option value="chat">chat</option>
              <option value="completion">completion</option>
              <option value="embedding">embedding</option>
              <option value="image">image</option>
            </select>
          </div>
          <div class="flex gap-3">
            <button @click="showEditDialog = false" class="flex-1 px-4 py-2 bg-slate-100 rounded-lg">取消</button>
            <button @click="save" class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg">保存</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 路由配置对话框 -->
    <div v-if="showRouteDialog && selectedModel" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl w-full max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
        <!-- 头部 -->
        <div class="p-6 border-b border-slate-200 flex items-center justify-between">
          <div>
            <h2 class="text-xl font-bold">模型路由配置</h2>
            <p class="text-sm text-slate-500">{{ selectedModel.display_name }} ({{ selectedModel.name }})</p>
          </div>
          <button @click="showRouteDialog = false" class="p-2 rounded hover:bg-slate-100">
            <X :size="20" />
          </button>
        </div>

        <!-- 路由配置列表 -->
        <div class="p-6 overflow-y-auto flex-1 space-y-3">
          <div class="text-sm text-slate-500 mb-4">
            配置该模型可以路由到的供应商。优先级数字越小，优先级越高。
          </div>

          <div
            v-for="route in modelRoutes.sort((a, b) => a.priority - b.priority)"
            :key="route.supplier_id"
            class="border border-slate-200 rounded-lg overflow-hidden"
          >
            <!-- 路由项头部 -->
            <div
              @click="toggleExpand(route.supplier_id)"
              class="flex items-center justify-between p-4 bg-slate-50 cursor-pointer hover:bg-slate-100"
            >
              <div class="flex items-center gap-3">
                <input
                  v-model="route.is_active"
                  type="checkbox"
                  class="w-5 h-5 text-blue-600 rounded"
                  @click.stop
                />
                <span class="font-medium">{{ route.supplier_name }}</span>
                <span v-if="route.is_active" class="text-xs bg-green-100 text-green-700 px-2 py-1 rounded-full">
                  活跃
                </span>
              </div>
              <div class="flex items-center gap-4">
                <div class="flex items-center gap-2 text-sm text-slate-600">
                  <span>优先级:</span>
                  <input
                    v-model.number="route.priority"
                    type="number"
                    min="1"
                    max="100"
                    class="w-16 px-2 py-1 border border-slate-300 rounded text-center"
                    @click.stop
                  />
                </div>
                <button>
                  <ChevronDown v-if="!expandedRoutes.has(route.supplier_id)" :size="18" />
                  <ChevronUp v-else :size="18" />
                </button>
              </div>
            </div>

            <!-- 展开的成本配置 -->
            <div v-if="expandedRoutes.has(route.supplier_id)" class="p-4 border-t border-slate-200 bg-white">
              <div class="grid grid-cols-2 gap-4">
                <div>
                  <label class="block text-sm font-medium text-slate-700 mb-1">输入成本 ($/1K tokens)</label>
                  <input
                    v-model.number="route.input_cost"
                    type="number"
                    step="0.0001"
                    min="0"
                    class="w-full px-3 py-2 border border-slate-300 rounded"
                    :disabled="!route.is_active"
                  />
                </div>
                <div>
                  <label class="block text-sm font-medium text-slate-700 mb-1">输出成本 ($/1K tokens)</label>
                  <input
                    v-model.number="route.output_cost"
                    type="number"
                    step="0.0001"
                    min="0"
                    class="w-full px-3 py-2 border border-slate-300 rounded"
                    :disabled="!route.is_active"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- 路由说明 -->
          <div class="p-4 bg-blue-50 rounded-lg">
            <h4 class="text-sm font-medium text-blue-900 mb-2">路由说明</h4>
            <ul class="text-sm text-blue-700 space-y-1">
              <li>• 勾选"活跃"启用该供应商的路由</li>
              <li>• 优先级数字越小，该供应商越优先被使用</li>
              <li>• 当高优先级供应商不可用时，自动切换到下一优先级供应商</li>
              <li>• 成本配置用于计算利润和选择最优供应商</li>
            </ul>
          </div>
        </div>

        <!-- 底部操作栏 -->
        <div class="p-6 border-t border-slate-200 flex items-center justify-between">
          <div class="text-sm text-slate-500">
            活跃路由: {{ modelRoutes.filter(r => r.is_active).length }} / {{ modelRoutes.length }}
          </div>
          <div class="flex gap-3">
            <button @click="showRouteDialog = false" class="px-6 py-2 bg-slate-100 rounded-lg">
              取消
            </button>
            <button @click="saveRoutes" class="flex items-center gap-2 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
              <Save :size="16" />
              保存配置
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
