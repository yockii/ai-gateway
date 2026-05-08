<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { billsApi } from '@/api/bills'
import type { Bill } from '@/types/models'
import { FileText, Download, Eye } from 'lucide-vue-next'

const bills = ref<Bill[]>([])
const loading = ref(false)
const showDetailDialog = ref(false)
const selectedBill = ref<Bill | null>(null)
const exporting = ref<string | null>(null)

onMounted(async () => {
  await fetchBills()
})

const fetchBills = async () => {
  loading.value = true
  try {
    bills.value = await billsApi.list()
  } catch (err) {
    console.error('Failed to fetch bills:', err)
  } finally {
    loading.value = false
  }
}

const viewDetail = async (bill: Bill) => {
  try {
    selectedBill.value = await billsApi.getById(bill.id)
    showDetailDialog.value = true
  } catch (err) {
    console.error('Failed to fetch bill detail:', err)
  }
}

const exportBill = async (billId: string, format: 'pdf' | 'csv') => {
  exporting.value = `${billId}-${format}`
  try {
    const result = await billsApi.export(billId, format)
    window.open(result.download_url, '_blank')
  } catch (err: any) {
    alert(err.message || '导出失败')
  } finally {
    exporting.value = null
  }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleDateString('zh-CN')
}

const getStatusClass = (status: string) => {
  switch (status) {
    case 'paid':
      return 'bg-green-100 text-green-700'
    case 'pending':
      return 'bg-yellow-100 text-yellow-700'
    case 'overdue':
      return 'bg-red-100 text-red-700'
    default:
      return 'bg-slate-100 text-slate-600'
  }
}

const getStatusText = (status: string) => {
  switch (status) {
    case 'paid':
      return '已支付'
    case 'pending':
      return '待支付'
    case 'overdue':
      return '逾期'
    default:
      return status
  }
}
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-slate-900">账单查询</h1>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      加载中...
    </div>

    <div v-else-if="bills.length === 0" class="bg-white rounded-xl shadow-sm p-12 text-center">
      <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <FileText :size="32" class="text-slate-400" />
      </div>
      <h3 class="text-lg font-semibold text-slate-900 mb-2">暂无账单</h3>
      <p class="text-slate-600">开始使用 API 服务后，账单将在这里显示</p>
    </div>

    <div v-else class="bg-white rounded-xl shadow-sm overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">账单期</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">总费用</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">请求数</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">状态</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">创建时间</th>
            <th class="text-right px-6 py-3 text-sm font-medium text-slate-600">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="bill in bills" :key="bill.id" class="hover:bg-slate-50">
            <td class="px-6 py-4 text-sm text-slate-900">{{ bill.period }}</td>
            <td class="px-6 py-4 text-sm font-medium text-slate-900">
              ¥{{ bill.total_revenue.toFixed(2) }}
            </td>
            <td class="px-6 py-4 text-sm text-slate-600">
              {{ bill.total_requests.toLocaleString() }}
            </td>
            <td class="px-6 py-4">
              <span
                :class="['px-2 py-1 text-xs font-medium rounded-full', getStatusClass(bill.status)]"
              >
                {{ getStatusText(bill.status) }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm text-slate-600">
              {{ formatDate(bill.created_at) }}
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center justify-end gap-2">
                <button
                  @click="viewDetail(bill)"
                  class="p-2 rounded hover:bg-slate-100 text-slate-600"
                  title="查看详情"
                >
                  <Eye :size="16" />
                </button>
                <button
                  @click="exportBill(bill.id, 'pdf')"
                  :disabled="exporting === `${bill.id}-pdf`"
                  class="p-2 rounded hover:bg-slate-100 text-slate-600"
                  title="导出 PDF"
                >
                  <Download :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 账单详情对话框 -->
    <div
      v-if="showDetailDialog && selectedBill"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    >
      <div class="bg-white rounded-xl shadow-lg w-full max-w-2xl max-h-[90vh] overflow-auto p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="text-xl font-bold text-slate-900">账单详情</h2>
          <button
            @click="showDetailDialog = false"
            class="p-2 hover:bg-slate-100 rounded"
          >
            ✕
          </button>
        </div>

        <div class="space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div>
              <p class="text-sm text-slate-600">账单期</p>
              <p class="font-medium text-slate-900">{{ selectedBill.period }}</p>
            </div>
            <div>
              <p class="text-sm text-slate-600">状态</p>
              <p
                :class="[
                  'font-medium',
                  selectedBill.status === 'paid' ? 'text-green-600' : 'text-yellow-600',
                ]"
              >
                {{ getStatusText(selectedBill.status) }}
              </p>
            </div>
          </div>

          <div class="grid grid-cols-4 gap-4">
            <div>
              <p class="text-sm text-slate-600">总请求数</p>
              <p class="font-medium text-slate-900">
                {{ selectedBill.total_requests.toLocaleString() }}
              </p>
            </div>
            <div>
              <p class="text-sm text-slate-600">总 Token</p>
              <p class="font-medium text-slate-900">
                {{ selectedBill.items.reduce((sum, i) => sum + i.total_tokens, 0).toLocaleString() }}
              </p>
            </div>
            <div>
              <p class="text-sm text-slate-600">总成本</p>
              <p class="font-medium text-slate-900">
                ¥{{ selectedBill.total_cost.toFixed(2) }}
              </p>
            </div>
            <div>
              <p class="text-sm text-slate-600">总收入</p>
              <p class="font-medium text-slate-900">
                ¥{{ selectedBill.total_revenue.toFixed(2) }}
              </p>
            </div>
          </div>

          <div>
            <h3 class="font-medium text-slate-900 mb-2">模型明细</h3>
            <table class="w-full text-sm">
              <thead class="bg-slate-50">
                <tr>
                  <th class="text-left px-3 py-2 text-slate-600">模型</th>
                  <th class="text-right px-3 py-2 text-slate-600">请求数</th>
                  <th class="text-right px-3 py-2 text-slate-600">Token</th>
                  <th class="text-right px-3 py-2 text-slate-600">费用</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-slate-200">
                <tr v-for="item in selectedBill.items" :key="item.model_id">
                  <td class="px-3 py-2 text-slate-900">{{ item.model_id }}</td>
                  <td class="px-3 py-2 text-slate-600 text-right">
                    {{ item.request_count.toLocaleString() }}
                  </td>
                  <td class="px-3 py-2 text-slate-600 text-right">
                    {{ item.total_tokens.toLocaleString() }}
                  </td>
                  <td class="px-3 py-2 text-slate-900 text-right">
                    ¥{{ item.total_revenue.toFixed(2) }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="flex gap-3 pt-4 border-t">
            <button
              @click="exportBill(selectedBill.id, 'pdf')"
              :disabled="exporting === `${selectedBill.id}-pdf`"
              class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-slate-300"
            >
              导出 PDF
            </button>
            <button
              @click="exportBill(selectedBill.id, 'csv')"
              :disabled="exporting === `${selectedBill.id}-csv`"
              class="flex-1 px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200 disabled:bg-slate-300"
            >
              导出 CSV
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
