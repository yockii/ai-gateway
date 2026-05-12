<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usersApi } from '@/api/users'
import { pricingApi } from '@/api/pricing'
import type { User, EnterprisePricing } from '@/types/models'
import { Search, Eye, Ban, Check, DollarSign, X, Plus, Trash2 } from 'lucide-vue-next'

const users = ref<User[]>([])
const loading = ref(false)
const searchQuery = ref('')
const showDetailDialog = ref(false)
const selectedUser = ref<User | null>(null)
const activeTab = ref<'basic' | 'pricing' | 'settings'>('basic')

// 用户定价相关
const userPricing = ref<EnterprisePricing[]>([])
const pricingLoading = ref(false)
const showAddPricingDialog = ref(false)
const newPricing = ref({
  model_id: '',
  input_price: 0,
  output_price: 0,
  effective_date: new Date().toISOString().split('T')[0],
})

// 可用模型列表
const availableModels = ref([
  { id: 'gpt-4', name: 'GPT-4' },
  { id: 'gpt-3.5-turbo', name: 'GPT-3.5 Turbo' },
  { id: 'claude-3-opus', name: 'Claude 3 Opus' },
  { id: 'claude-3-sonnet', name: 'Claude 3 Sonnet' },
])

onMounted(async () => {
  await fetchUsers()
})

const fetchUsers = async () => {
  loading.value = true
  try {
    users.value = await usersApi.list({ search: searchQuery.value })
  } catch (err) {
    console.error('Failed to fetch users:', err)
  } finally {
    loading.value = false
  }
}

const viewUser = (user: User) => {
  selectedUser.value = user
  activeTab.value = 'basic'
  showDetailDialog.value = true
}

const viewUserPricing = (user: User) => {
  selectedUser.value = user
  activeTab.value = 'pricing'
  showDetailDialog.value = true
  fetchUserPricing()
}

const fetchUserPricing = async () => {
  if (!selectedUser.value) return
  pricingLoading.value = true
  try {
    const allPricing = await pricingApi.listEnterprise()
    userPricing.value = allPricing.filter((p: EnterprisePricing) => p.customer_id === selectedUser.value?.id)
  } catch (err) {
    console.error('Failed to fetch user pricing:', err)
  } finally {
    pricingLoading.value = false
  }
}

const openAddPricingDialog = () => {
  newPricing.value = {
    model_id: '',
    input_price: 0,
    output_price: 0,
    effective_date: new Date().toISOString().split('T')[0],
  }
  showAddPricingDialog.value = true
}

const addPricing = async () => {
  if (!selectedUser.value) return
  try {
    await pricingApi.createEnterprise({
      customer_id: selectedUser.value.id,
      customer_name: selectedUser.value.name,
      model_id: newPricing.value.model_id,
      input_price: newPricing.value.input_price,
      output_price: newPricing.value.output_price,
      effective_date: newPricing.value.effective_date,
    })
    showAddPricingDialog.value = false
    await fetchUserPricing()
  } catch (err: any) {
    alert(err.message || '添加定价失败')
  }
}

const deletePricing = async (pricing: EnterprisePricing) => {
  if (!confirm(`确定要删除 ${pricing.model_id} 的定价配置吗？`)) return
  try {
    await pricingApi.deleteEnterprise(pricing.id)
    await fetchUserPricing()
  } catch (err: any) {
    alert(err.message || '删除定价失败')
  }
}

const toggleStatus = async (user: User) => {
  try {
    await usersApi.toggleStatus(user.id)
    await fetchUsers()
  } catch (err: any) {
    alert(err.message || '操作失败')
  }
}

const deleteUser = async (user: User) => {
  if (!confirm(`确定要删除用户 ${user.name} 吗？`)) return
  try {
    await usersApi.delete(user.id)
    await fetchUsers()
  } catch (err: any) {
    alert(err.message || '删除失败')
  }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

const formatPrice = (price: number) => {
  return `$${price.toFixed(4)}`
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">用户管理</h1>
      <div class="relative">
        <Search :size="18" class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400" />
        <input
          v-model="searchQuery"
          @keyup.enter="fetchUsers"
          type="text"
          placeholder="搜索邮箱或姓名"
          class="pl-10 pr-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none w-64"
        />
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">加载中...</div>

    <div v-else class="bg-white rounded-xl shadow-sm overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">邮箱</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">姓名</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">用户组</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">状态</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">注册时间</th>
            <th class="text-right px-6 py-3 text-sm font-medium text-slate-600">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50">
            <td class="px-6 py-4 text-sm text-slate-900">{{ user.email }}</td>
            <td class="px-6 py-4 text-sm text-slate-900">{{ user.name }}</td>
            <td class="px-6 py-4 text-sm text-slate-600">{{ user.user_group_id || '-' }}</td>
            <td class="px-6 py-4">
              <span
                :class="[
                  'px-2 py-1 text-xs font-medium rounded-full',
                  user.is_active ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600',
                ]"
              >
                {{ user.is_active ? '活跃' : '禁用' }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm text-slate-600">{{ formatDate(user.created_at) }}</td>
            <td class="px-6 py-4">
              <div class="flex items-center justify-end gap-2">
                <button @click="viewUser(user)" class="p-2 rounded hover:bg-slate-100 text-slate-600" title="查看详情">
                  <Eye :size="16" />
                </button>
                <button @click="viewUserPricing(user)" class="p-2 rounded hover:bg-slate-100 text-green-600" title="定价配置">
                  <DollarSign :size="16" />
                </button>
                <button @click="toggleStatus(user)" class="p-2 rounded hover:bg-slate-100 text-slate-600">
                  <Ban v-if="user.is_active" :size="16" />
                  <Check v-else :size="16" />
                </button>
                <button @click="deleteUser(user)" class="p-2 rounded hover:bg-red-100 text-red-600" title="删除用户">
                  <Trash2 :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 用户详情对话框 -->
    <div v-if="showDetailDialog && selectedUser" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl w-full max-w-3xl max-h-[80vh] overflow-hidden flex flex-col">
        <!-- 头部 -->
        <div class="p-6 border-b border-slate-200 flex items-center justify-between">
          <div>
            <h2 class="text-xl font-bold">用户详情</h2>
            <p class="text-sm text-slate-500">{{ selectedUser.email }}</p>
          </div>
          <button @click="showDetailDialog = false" class="p-2 rounded hover:bg-slate-100">
            <X :size="20" />
          </button>
        </div>

        <!-- 标签页 -->
        <div class="border-b border-slate-200">
          <nav class="flex">
            <button
              @click="activeTab = 'basic'"
              :class="[
                'px-6 py-3 text-sm font-medium border-b-2 transition-colors',
                activeTab === 'basic' ? 'border-blue-500 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700',
              ]"
            >
              基本信息
            </button>
            <button
              @click="activeTab = 'pricing'"
              :class="[
                'px-6 py-3 text-sm font-medium border-b-2 transition-colors',
                activeTab === 'pricing' ? 'border-blue-500 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700',
              ]"
            >
              定价配置
            </button>
            <button
              @click="activeTab = 'settings'"
              :class="[
                'px-6 py-3 text-sm font-medium border-b-2 transition-colors',
                activeTab === 'settings' ? 'border-blue-500 text-blue-600' : 'border-transparent text-slate-500 hover:text-slate-700',
              ]"
            >
              设置
            </button>
          </nav>
        </div>

        <!-- 内容 -->
        <div class="p-6 overflow-y-auto flex-1">
          <!-- 基本信息 -->
          <div v-if="activeTab === 'basic'" class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="text-sm text-slate-500">用户 ID</label>
                <p class="font-mono text-sm">{{ selectedUser.id }}</p>
              </div>
              <div>
                <label class="text-sm text-slate-500">邮箱</label>
                <p>{{ selectedUser.email }}</p>
              </div>
              <div>
                <label class="text-sm text-slate-500">姓名</label>
                <p>{{ selectedUser.name }}</p>
              </div>
              <div>
                <label class="text-sm text-slate-500">用户组</label>
                <p>{{ selectedUser.user_group_id || '未设置' }}</p>
              </div>
              <div>
                <label class="text-sm text-slate-500">状态</label>
                <p>
                  <span
                    :class="[
                      'px-2 py-1 text-xs font-medium rounded-full',
                      selectedUser.is_active ? 'bg-green-100 text-green-700' : 'bg-slate-100 text-slate-600',
                    ]"
                  >
                    {{ selectedUser.is_active ? '活跃' : '禁用' }}
                  </span>
                </p>
              </div>
              <div>
                <label class="text-sm text-slate-500">注册时间</label>
                <p>{{ formatDate(selectedUser.created_at) }}</p>
              </div>
            </div>
          </div>

          <!-- 定价配置 -->
          <div v-if="activeTab === 'pricing'" class="space-y-4">
            <div class="flex items-center justify-between">
              <h3 class="text-lg font-semibold">企业定价配置</h3>
              <button
                @click="openAddPricingDialog"
                class="flex items-center gap-2 px-3 py-2 bg-blue-600 text-white rounded-lg text-sm hover:bg-blue-700"
              >
                <Plus :size="16" />
                添加定价
              </button>
            </div>

            <div v-if="pricingLoading" class="text-center py-8 text-slate-500">加载中...</div>

            <div v-else-if="userPricing.length === 0" class="text-center py-8 text-slate-400">
              暂无企业定价配置
            </div>

            <div v-else class="space-y-2">
              <div
                v-for="pricing in userPricing"
                :key="pricing.id"
                class="flex items-center justify-between p-4 bg-slate-50 rounded-lg"
              >
                <div class="flex-1">
                  <p class="font-medium">{{ pricing.model_id }}</p>
                  <p class="text-sm text-slate-500">
                    输入: {{ formatPrice(pricing.input_price) }} /
                    输出: {{ formatPrice(pricing.output_price) }}
                  </p>
                </div>
                <button
                  @click="deletePricing(pricing)"
                  class="p-2 rounded hover:bg-red-100 text-red-600"
                >
                  <X :size="16" />
                </button>
              </div>
            </div>
          </div>

          <!-- 设置 -->
          <div v-if="activeTab === 'settings'" class="space-y-4">
            <h3 class="text-lg font-semibold">用户设置</h3>
            <div class="space-y-4">
              <div>
                <label class="block text-sm font-medium text-slate-700 mb-2">会员等级</label>
                <select class="w-full px-4 py-2 border border-slate-300 rounded-lg">
                  <option value="free">免费版</option>
                  <option value="pro">专业版</option>
                  <option value="enterprise">企业版</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-slate-700 mb-2">用户组</label>
                <select class="w-full px-4 py-2 border border-slate-300 rounded-lg">
                  <option value="">默认用户组</option>
                  <option value="vip">VIP 用户组</option>
                  <option value="enterprise">企业用户组</option>
                </select>
              </div>
              <button class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">
                保存设置
              </button>
            </div>
          </div>
        </div>

        <!-- 底部 -->
        <div class="p-6 border-t border-slate-200">
          <button @click="showDetailDialog = false" class="w-full px-4 py-2 bg-slate-100 rounded-lg">
            关闭
          </button>
        </div>
      </div>
    </div>

    <!-- 添加定价对话框 -->
    <div v-if="showAddPricingDialog" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h2 class="text-xl font-bold mb-4">添加企业定价</h2>
        <div class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">模型</label>
            <select v-model="newPricing.model_id" class="w-full px-4 py-2 border border-slate-300 rounded-lg">
              <option value="">请选择模型</option>
              <option v-for="model in availableModels" :key="model.id" :value="model.id">
                {{ model.name }}
              </option>
            </select>
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">输入价格 ($/1K tokens)</label>
            <input
              v-model.number="newPricing.input_price"
              type="number"
              step="0.0001"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">输出价格 ($/1K tokens)</label>
            <input
              v-model.number="newPricing.output_price"
              type="number"
              step="0.0001"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">生效日期</label>
            <input
              v-model="newPricing.effective_date"
              type="date"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg"
            />
          </div>
          <div class="flex gap-3">
            <button @click="showAddPricingDialog = false" class="flex-1 px-4 py-2 bg-slate-100 rounded-lg">
              取消
            </button>
            <button @click="addPricing" class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg">
              添加
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
