<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { modelsApi } from '@/api/models'
import type { ExternalModel } from '@/types/models'
import { Plus, Power, Edit2, Trash2 } from 'lucide-vue-next'

const models = ref<ExternalModel[]>([])
const loading = ref(false)
const showDialog = ref(false)
const editingModel = ref<ExternalModel | null>(null)
const formData = ref({ name: '', display_name: '', model_type: '' })

onMounted(async () => {
  await fetchModels()
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

const openCreateDialog = () => {
  editingModel.value = null
  formData.value = { name: '', display_name: '', model_type: '' }
  showDialog.value = true
}

const openEditDialog = (model: ExternalModel) => {
  editingModel.value = model
  formData.value = { name: model.name, display_name: model.display_name, model_type: model.model_type }
  showDialog.value = true
}

const save = async () => {
  try {
    if (editingModel.value) {
      await modelsApi.update(editingModel.value.id, formData.value)
    } else {
      await modelsApi.create(formData.value)
    }
    await fetchModels()
    showDialog.value = false
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

const getTypeBadge = (type: string) => {
  const map: Record<string, string> = {
    chat: 'bg-blue-100 text-blue-700',
    completion: 'bg-green-100 text-green-700',
    embedding: 'bg-purple-100 text-purple-700',
  }
  return map[type] || 'bg-slate-100 text-slate-600'
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
                {{ model.model_type }}
              </span>
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

    <div v-if="showDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
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
            </select>
          </div>
          <div class="flex gap-3">
            <button @click="showDialog = false" class="flex-1 px-4 py-2 bg-slate-100 rounded-lg">取消</button>
            <button @click="save" class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg">保存</button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
