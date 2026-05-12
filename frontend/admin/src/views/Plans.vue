<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { plansApi } from '@/api/plans'
import type { MembershipTier } from '@/types/models'
import { Plus, Edit2, Trash2, DollarSign, Save, X } from 'lucide-vue-next'

// 套餐列表
const tiers = ref<MembershipTier[]>([])
const loading = ref(false)
const showTierDialog = ref(false)
const showPricingDialog = ref(false)
const editingTier = ref<MembershipTier | null>(null)
const selectedTier = ref<MembershipTier | null>(null)

// 套餐表单
const formData = ref({ name: '', display_name: '', level: 1 })

// 可用模型
const availableModels = ref([
  { id: 'gpt-4', name: 'GPT-4', type: 'chat' },
  { id: 'gpt-3.5-turbo', name: 'GPT-3.5 Turbo', type: 'chat' },
  { id: 'claude-3-opus', name: 'Claude 3 Opus', type: 'chat' },
  { id: 'claude-3-sonnet', name: 'Claude 3 Sonnet', type: 'chat' },
  { id: 'text-embedding-ada-002', name: 'Ada Embedding', type: 'embedding' },
])

// 套餐模型定价
interface TierModelPricing {
  model_id: string
  model_name: string
  input_price: number
  output_price: number
  is_available: boolean
}

const tierPricing = ref<TierModelPricing[]>([])

// 初始化定价数据
const initializePricing = () => {
  tierPricing.value = availableModels.value.map(model => ({
    model_id: model.id,
    model_name: model.name,
    input_price: getDefaultPrice(model.id, 'input'),
    output_price: getDefaultPrice(model.id, 'output'),
    is_available: true,
  }))
}

// 获取默认价格
const getDefaultPrice = (modelId: string, type: 'input' | 'output') => {
  const defaults: Record<string, { input: number; output: number }> = {
    'gpt-4': { input: 0.03, output: 0.06 },
    'gpt-3.5-turbo': { input: 0.0015, output: 0.002 },
    'claude-3-opus': { input: 0.015, output: 0.075 },
    'claude-3-sonnet': { input: 0.003, output: 0.015 },
    'text-embedding-ada-002': { input: 0.0001, output: 0 },
  }
  return defaults[modelId]?.[type] || 0
}

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
  showTierDialog.value = true
}

const openEditDialog = (tier: MembershipTier) => {
  editingTier.value = tier
  formData.value = { name: tier.name, display_name: tier.display_name, level: tier.level }
  showTierDialog.value = true
}

const save = async () => {
  try {
    if (editingTier.value) {
      await plansApi.updateTier(editingTier.value.id, formData.value)
    } else {
      await plansApi.createTier(formData.value)
    }
    await fetchTiers()
    showTierDialog.value = false
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

const openPricingDialog = (tier: MembershipTier) => {
  selectedTier.value = tier
  initializePricing()
  showPricingDialog.value = true
}

const savePricing = async () => {
  // TODO: 调用 API 保存定价
  console.log('Saving pricing for tier:', selectedTier.value?.id, tierPricing.value)
  alert('定价配置已保存（演示）')
  showPricingDialog.value = false
}

const formatPrice = (price: number) => {
  return `$${price.toFixed(4)}`
}

const getModelTypeBadge = (type: string) => {
  const map: Record<string, string> = {
    chat: 'bg-blue-100 text-blue-700',
    embedding: 'bg-purple-100 text-purple-700',
  }
  return map[type] || 'bg-slate-100 text-slate-600'
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
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">模型定价</th>
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
              <button
                @click="openPricingDialog(tier)"
                class="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800"
              >
                <DollarSign :size="14" />
                配置定价
              </button>
            </td>
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

    <!-- 套餐编辑对话框 -->
    <div v-if="showTierDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
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
            <button @click="showTierDialog = false" class="flex-1 px-4 py-2 bg-slate-100 rounded-lg">取消</button>
            <button @click="save" class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg">保存</button>
          </div>
        </div>
      </div>
    </div>

    <!-- 模型定价配置对话框 -->
    <div v-if="showPricingDialog && selectedTier" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl w-full max-w-4xl max-h-[80vh] overflow-hidden flex flex-col">
        <!-- 头部 -->
        <div class="p-6 border-b border-slate-200 flex items-center justify-between">
          <div>
            <h2 class="text-xl font-bold">套餐模型定价</h2>
            <p class="text-sm text-slate-500">{{ selectedTier.display_name }} ({{ selectedTier.name }})</p>
          </div>
          <button @click="showPricingDialog = false" class="p-2 rounded hover:bg-slate-100">
            <X :size="20" />
          </button>
        </div>

        <!-- 定价表格 -->
        <div class="p-6 overflow-y-auto flex-1">
          <table class="w-full">
            <thead class="bg-slate-50">
              <tr>
                <th class="text-left px-4 py-3 text-sm font-medium text-slate-600">模型</th>
                <th class="text-left px-4 py-3 text-sm font-medium text-slate-600">类型</th>
                <th class="text-right px-4 py-3 text-sm font-medium text-slate-600">输入价格 ($/1K)</th>
                <th class="text-right px-4 py-3 text-sm font-medium text-slate-600">输出价格 ($/1K)</th>
                <th class="text-center px-4 py-3 text-sm font-medium text-slate-600">可用</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200">
              <tr v-for="item in tierPricing" :key="item.model_id" class="hover:bg-slate-50">
                <td class="px-4 py-3 text-sm font-medium">{{ item.model_name }}</td>
                <td class="px-4 py-3">
                  <span :class="['px-2 py-1 text-xs font-medium rounded-full', getModelTypeBadge(availableModels.find(m => m.id === item.model_id)?.type || '')]">
                    {{ availableModels.find(m => m.id === item.model_id)?.type }}
                  </span>
                </td>
                <td class="px-4 py-3">
                  <input
                    v-model.number="item.input_price"
                    type="number"
                    step="0.0001"
                    min="0"
                    class="w-full px-3 py-2 border border-slate-300 rounded text-right text-sm"
                    :disabled="!item.is_available"
                  />
                </td>
                <td class="px-4 py-3">
                  <input
                    v-model.number="item.output_price"
                    type="number"
                    step="0.0001"
                    min="0"
                    class="w-full px-3 py-2 border border-slate-300 rounded text-right text-sm"
                    :disabled="!item.is_available"
                  />
                </td>
                <td class="px-4 py-3 text-center">
                  <input
                    v-model="item.is_available"
                    type="checkbox"
                    class="w-5 h-5 text-blue-600 rounded"
                  />
                </td>
              </tr>
            </tbody>
          </table>

          <!-- 价格说明 -->
          <div class="mt-6 p-4 bg-blue-50 rounded-lg">
            <h4 class="text-sm font-medium text-blue-900 mb-2">定价说明</h4>
            <ul class="text-sm text-blue-700 space-y-1">
              <li>• 价格单位为美元/1,000 tokens</li>
              <li>• 输入价格指用户发送的 prompt tokens</li>
              <li>• 输出价格指模型生成的 completion tokens</li>
              <li>• 取消勾选"可用"将隐藏该模型，用户无法使用</li>
            </ul>
          </div>
        </div>

        <!-- 底部操作栏 -->
        <div class="p-6 border-t border-slate-200 flex items-center justify-between">
          <div class="text-sm text-slate-500">
            共 {{ tierPricing.filter(p => p.is_available).length }} 个模型可用
          </div>
          <div class="flex gap-3">
            <button @click="showPricingDialog = false" class="px-6 py-2 bg-slate-100 rounded-lg">
              取消
            </button>
            <button @click="savePricing" class="flex items-center gap-2 px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
              <Save :size="16" />
              保存定价
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
