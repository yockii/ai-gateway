<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { suppliersApi } from '@/api/suppliers'
import type { Supplier } from '@/types/models'
import { Plus, Edit2, Trash2, Zap } from 'lucide-vue-next'

const suppliers = ref<Supplier[]>([])
const loading = ref(false)
const showDialog = ref(false)
const editingSupplier = ref<Supplier | null>(null)
const testingId = ref<string | null>(null)
const formData = ref({ name: '', display_name: '', provider: '' })

onMounted(async () => {
  await fetchSuppliers()
})

const fetchSuppliers = async () => {
  loading.value = true
  try {
    suppliers.value = await suppliersApi.list()
  } catch (err) {
    console.error('Failed to fetch suppliers:', err)
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  editingSupplier.value = null
  formData.value = { name: '', display_name: '', provider: '' }
  showDialog.value = true
}

const openEditDialog = (supplier: Supplier) => {
  editingSupplier.value = supplier
  formData.value = { name: supplier.name, display_name: supplier.display_name, provider: supplier.provider }
  showDialog.value = true
}

const save = async () => {
  try {
    if (editingSupplier.value) {
      await suppliersApi.update(editingSupplier.value.id, formData.value)
    } else {
      await suppliersApi.create(formData.value)
    }
    await fetchSuppliers()
    showDialog.value = false
  } catch (err: any) {
    alert(err.message || '保存失败')
  }
}

const testConnection = async (supplier: Supplier) => {
  testingId.value = supplier.id
  try {
    await suppliersApi.testConnection(supplier.id)
    alert('连接成功！')
  } catch (err: any) {
    alert(err.message || '连接失败')
  } finally {
    testingId.value = null
  }
}

const deleteSupplier = async (supplier: Supplier) => {
  if (!confirm(`确定要删除供应商 ${supplier.display_name} 吗？`)) return
  try {
    await suppliersApi.delete(supplier.id)
    await fetchSuppliers()
  } catch (err: any) {
    alert(err.message || '删除失败')
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">供应商管理</h1>
      <button @click="openCreateDialog" class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
        <Plus :size="18" />
        添加供应商
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">加载中...</div>

    <div v-else class="bg-white rounded-xl shadow-sm overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">显示名称</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">供应商名称</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">提供商</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">状态</th>
            <th class="text-right px-6 py-3 text-sm font-medium text-slate-600">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="supplier in suppliers" :key="supplier.id" class="hover:bg-slate-50">
            <td class="px-6 py-4 text-sm text-slate-900">{{ supplier.display_name }}</td>
            <td class="px-6 py-4 text-sm text-slate-600 font-mono">{{ supplier.name }}</td>
            <td class="px-6 py-4 text-sm text-slate-600">{{ supplier.provider }}</td>
            <td class="px-6 py-4">
              <span :class="[
                'px-2 py-1 text-xs font-medium rounded-full',
                supplier.is_active ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600',
              ]">
                {{ supplier.is_active ? '活跃' : '禁用' }}
              </span>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center justify-end gap-2">
                <button
                  @click="testConnection(supplier)"
                  :disabled="testingId === supplier.id"
                  class="p-2 rounded hover:bg-yellow-100 text-yellow-600"
                  title="测试连接"
                >
                  <Zap :size="16" />
                </button>
                <button @click="openEditDialog(supplier)" class="p-2 rounded hover:bg-slate-100 text-slate-600">
                  <Edit2 :size="16" />
                </button>
                <button @click="deleteSupplier(supplier)" class="p-2 rounded hover:bg-red-100 text-red-600">
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
        <h2 class="text-xl font-bold mb-4">{{ editingSupplier ? '编辑供应商' : '添加供应商' }}</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">显示名称</label>
            <input v-model="formData.display_name" type="text" class="w-full px-4 py-2 border border-slate-300 rounded-lg" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">供应商名称</label>
            <input v-model="formData.name" type="text" class="w-full px-4 py-2 border border-slate-300 rounded-lg font-mono" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">提供商</label>
            <input v-model="formData.provider" type="text" class="w-full px-4 py-2 border border-slate-300 rounded-lg" placeholder="e.g., openai, anthropic" />
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
