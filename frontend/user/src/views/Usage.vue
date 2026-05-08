<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usageApi } from '@/api/usage'
import type { UsageRecord } from '@/types/models'
import UsageChart from '@/components/charts/UsageChart.vue'
import CostChart from '@/components/charts/CostChart.vue'

const loading = ref(true)
const records = ref<UsageRecord[]>([])

const chartData = ref({
  dates: [] as string[],
  requests: [] as number[],
  tokens: [] as number[],
  costs: [] as number[],
})

const stats = ref({
  totalRequests: 0,
  totalTokens: 0,
  totalCost: 0,
})

onMounted(async () => {
  await fetchData()
})

const fetchData = async () => {
  loading.value = true
  try {
    const response = await usageApi.list({ limit: 30 })
    records.value = response.data

    // 按日期聚合数据
    const dailyData = new Map<string, { requests: number; tokens: number; cost: number }>()

    for (const record of records.value) {
      const date = new Date(record.created_at).toLocaleDateString('zh-CN')
      if (!dailyData.has(date)) {
        dailyData.set(date, { requests: 0, tokens: 0, cost: 0 })
      }
      const day = dailyData.get(date)!
      day.requests++
      day.tokens += record.total_tokens
      day.cost += record.selling_price
    }

    const sortedDates = Array.from(dailyData.keys()).slice(-30)

    chartData.value = {
      dates: sortedDates,
      requests: sortedDates.map(d => dailyData.get(d)!.requests),
      tokens: sortedDates.map(d => dailyData.get(d)!.tokens),
      costs: sortedDates.map(d => dailyData.get(d)!.cost),
    }

    stats.value = {
      totalRequests: records.value.length,
      totalTokens: records.value.reduce((sum, r) => sum + r.total_tokens, 0),
      totalCost: records.value.reduce((sum, r) => sum + r.selling_price, 0),
    }
  } catch (err) {
    console.error('Failed to fetch usage:', err)
  } finally {
    loading.value = false
  }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString('zh-CN')
}
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-slate-900">使用统计</h1>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      加载中...
    </div>

    <div v-else class="space-y-6">
      <!-- 统计卡片 -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">总请求数</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">
            {{ stats.totalRequests.toLocaleString() }}
          </p>
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">总 Token</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">
            {{ stats.totalTokens.toLocaleString() }}
          </p>
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <p class="text-sm text-slate-600">总费用</p>
          <p class="text-2xl font-bold text-slate-900 mt-1">
            ¥{{ stats.totalCost.toFixed(2) }}
          </p>
        </div>
      </div>

      <!-- 图表 -->
      <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div class="bg-white rounded-xl shadow-sm p-6">
          <h2 class="text-lg font-semibold text-slate-900 mb-4">使用趋势</h2>
          <UsageChart :data="chartData" />
        </div>
        <div class="bg-white rounded-xl shadow-sm p-6">
          <h2 class="text-lg font-semibold text-slate-900 mb-4">费用趋势</h2>
          <CostChart :data="{ dates: chartData.dates, costs: chartData.costs }" />
        </div>
      </div>

      <!-- 使用记录 -->
      <div class="bg-white rounded-xl shadow-sm overflow-hidden">
        <div class="px-6 py-4 border-b border-slate-200">
          <h2 class="text-lg font-semibold text-slate-900">最近记录</h2>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full">
            <thead class="bg-slate-50">
              <tr>
                <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">时间</th>
                <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">模型</th>
                <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">Token</th>
                <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">费用</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-200">
              <tr v-if="records.length === 0">
                <td colspan="4" class="px-6 py-8 text-center text-slate-500">
                  暂无使用记录
                </td>
              </tr>
              <tr v-for="record in records.slice(0, 10)" :key="record.id" class="hover:bg-slate-50">
                <td class="px-6 py-4 text-sm text-slate-600">
                  {{ formatDate(record.created_at) }}
                </td>
                <td class="px-6 py-4 text-sm text-slate-900">{{ record.model_id }}</td>
                <td class="px-6 py-4 text-sm text-slate-600">
                  {{ record.total_tokens.toLocaleString() }}
                </td>
                <td class="px-6 py-4 text-sm text-slate-900">
                  ¥{{ record.selling_price.toFixed(4) }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
