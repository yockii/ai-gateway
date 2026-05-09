<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { plansApi } from '@/api/plans'
import type { MembershipTier } from '@/types/models'
import { Plus, Edit2, Trash2 } from 'lucide-vue-next'

const tiers = ref<MembershipTier[]>([])
const loading = ref(false)
const showDialog = ref(false)
const editingTier = ref<MembershipTier | null>(null)
const formData = ref({ name: '', display_name: '', level: 1 })

onMounted(async () => {
  await fetchTiers()
})

const fetchTiers = async () => {
  loading.value = true
  try {
    tiers.value = await plansApi.listTiers()
  } catch (err) {
    console.error('Failed to fetch tiers:', err)
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  editingTier.value = null
  formData.value = { name: '', display_name: '', level: 1 }
  showDialog.value = true
}

const openEditDialog = (tier: MembershipTier) => {
  editingTier.value = tier
  formData.value = { name: tier.name, display_name: tier.display_name, level: tier.level }
  showDialog.value = true
}

const save = async () => {
  try {
    if (editingTier.value) {
      await plansApi.updateTier(editingTier.value.id, formData.value)
    } else {
      await plansApi.createTier(formData.value)
    }
    await fetchTiers()
    showDialog.value = false
  } catch (err: any) {
    alert(err.message || '保存失败')
  }
}

const deleteTier = async (tier: MembershipTier) => {
  if (!confirm(`确定要删除套餐 ${tier.display_name} 吗？`)) return
  try {
    await plansApi.deleteTier(tier.id)
    await fetchTiers()
  } catch (err: any) {
    alert(err.message || '删除失败')
  }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">套餐管理</h1>
      <button @click="openCreateDialog" class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
        <Plus :size="18" />
        添加套餐
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">加载中...</div>

    <div v-else class="bg-white rounded-xl shadow-sm overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">显示名称</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">套餐名称</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">等级</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">状态</th>
            <th class="text-right px-6 py-3 text-sm font-medium text-slate-600">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="tier in tiers.sort((a: any, b: any) => a.level - b.level)" :key="tier.id" class="hover:bg-slate-50">
            <td class="px-6 py-4 text-sm text-slate-900">{{ tier.display_name }}</td>
            <td class="px-6 py-4 text-sm text-slate-600 font-mono">{{ tier.name }}</td>
            <td class="px-6 py-4 text-sm text-slate-600">Lv.{{ tier.level }}</td>
            <td class="px-6 py-4">
              <span :class="[
                'px-2 py-1 text-xs font-medium rounded-full',
                tier.is_active ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600',
              ]">
                {{ tier.is_active ? '启用' : '禁用' }}
              </span>
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center justify-end gap-2">
                <button @click="openEditDialog(tier)" class="p-2 rounded hover:bg-slate-100 text-slate-600">
                  <Edit2 :size="16" />
                </button>
                <button @click="deleteTier(tier)" class="p-2 rounded hover:bg-red-100 text-red-600">
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
        <h2 class="text-xl font-bold mb-4">{{ editingTier ? '编辑套餐' : '添加套餐' }}</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">显示名称</label>
            <input v-model="formData.display_name" type="text" class="w-full px-4 py-2 border border-slate-300 rounded-lg" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">套餐名称</label>
            <input v-model="formData.name" type="text" class="w-full px-4 py-2 border border-slate-300 rounded-lg font-mono" />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">等级</label>
            <input v-model.number="formData.level" type="number" class="w-full px-4 py-2 border border-slate-300 rounded-lg" />
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
