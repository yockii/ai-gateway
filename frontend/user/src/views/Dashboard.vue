<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { usageApi } from '@/api/usage'
import type { DashboardStats } from '@/types/api'
import { KeyRound, FileText, TrendingUp, DollarSign } from 'lucide-vue-next'

const authStore = useAuthStore()

const stats = ref<DashboardStats>({
  total_keys: 0,
  monthly_requests: 0,
  monthly_cost: 0,
  monthly_tokens: 0,
})
const loading = ref(true)

onMounted(async () => {
  try {
    stats.value = await usageApi.getStats()
  } catch (err) {
    console.error('Failed to fetch stats:', err)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-slate-900">
          欢迎回来，{{ authStore.user?.name || '用户' }}
        </h1>
        <p class="text-slate-600 mt-1">这是您的控制台概览</p>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      加载中...
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white rounded-xl shadow-sm p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-slate-600">API Key 数量</p>
            <p class="text-2xl font-bold text-slate-900 mt-1">{{ stats.total_keys }}</p>
          </div>
          <div class="p-3 bg-blue-100 rounded-lg">
            <KeyRound :size="24" class="text-blue-600" />
          </div>
        </div>
        <router-link
          to="/api-keys"
          class="text-sm text-blue-600 hover:underline mt-2 inline-block"
        >
          管理 API Key →
        </router-link>
      </div>

      <div class="bg-white rounded-xl shadow-sm p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-slate-600">本月请求数</p>
            <p class="text-2xl font-bold text-slate-900 mt-1">
              {{ stats.monthly_requests.toLocaleString() }}
            </p>
          </div>
          <div class="p-3 bg-green-100 rounded-lg">
            <TrendingUp :size="24" class="text-green-600" />
          </div>
        </div>
        <router-link
          to="/usage"
          class="text-sm text-blue-600 hover:underline mt-2 inline-block"
        >
          查看详情 →
        </router-link>
      </div>

      <div class="bg-white rounded-xl shadow-sm p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-slate-600">本月费用</p>
            <p class="text-2xl font-bold text-slate-900 mt-1">
              ¥{{ stats.monthly_cost.toFixed(2) }}
            </p>
          </div>
          <div class="p-3 bg-yellow-100 rounded-lg">
            <DollarSign :size="24" class="text-yellow-600" />
          </div>
        </div>
        <router-link
          to="/bills"
          class="text-sm text-blue-600 hover:underline mt-2 inline-block"
        >
          查看账单 →
        </router-link>
      </div>

      <div class="bg-white rounded-xl shadow-sm p-6">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-sm text-slate-600">本月 Token</p>
            <p class="text-2xl font-bold text-slate-900 mt-1">
              {{ stats.monthly_tokens.toLocaleString() }}
            </p>
          </div>
          <div class="p-3 bg-purple-100 rounded-lg">
            <FileText :size="24" class="text-purple-600" />
          </div>
        </div>
        <router-link
          to="/usage"
          class="text-sm text-blue-600 hover:underline mt-2 inline-block"
        >
          使用趋势 →
        </router-link>
      </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm p-6">
      <h2 class="text-lg font-semibold text-slate-900 mb-4">快速操作</h2>
      <div class="flex flex-wrap gap-3">
        <router-link
          to="/api-keys"
          class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          创建 API Key
        </router-link>
        <router-link
          to="/bills"
          class="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 transition-colors"
        >
          查看账单
        </router-link>
        <router-link
          to="/usage"
          class="px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 transition-colors"
        >
          使用统计
        </router-link>
      </div>
    </div>
  </div>
</template>
