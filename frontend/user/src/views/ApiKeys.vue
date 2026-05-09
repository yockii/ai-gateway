<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { keysApi } from '@/api/keys'
import type { UserAPIKey } from '@/types/models'
import { Plus, Trash2, Power, EyeOff, Copy, Check } from 'lucide-vue-next'

const keys = ref<UserAPIKey[]>([])
const loading = ref(false)
const showDialog = ref(false)
const showDeleteDialog = ref(false)
const selectedKeyId = ref('')
const newKeyName = ref('')
const newKeyQuota = ref<number>()
const newKeyConcurrency = ref<number>()
const createdKey = ref<{ key: string; data: UserAPIKey } | null>(null)
const copied = ref(false)
const disablingKeyId = ref('')

onMounted(async () => {
  await fetchKeys()
})

const fetchKeys = async () => {
  loading.value = true
  try {
    keys.value = await keysApi.list()
  } catch (err) {
    console.error('Failed to fetch keys:', err)
  } finally {
    loading.value = false
  }
}

const openCreateDialog = () => {
  newKeyName.value = ''
  newKeyQuota.value = undefined
  newKeyConcurrency.value = undefined
  createdKey.value = null
  showDialog.value = true
}

const handleCreate = async () => {
  try {
    createdKey.value = await keysApi.create({
      name: newKeyName.value,
      quota_daily: newKeyQuota.value,
      concurrency_limit: newKeyConcurrency.value,
    })
    await fetchKeys()
  } catch (err: any) {
    alert(err.message || '创建失败')
  }
}

const copyKey = async () => {
  if (createdKey.value) {
    await navigator.clipboard.writeText(createdKey.value.key)
    copied.value = true
    setTimeout(() => (copied.value = false), 2000)
  }
}

const closeDialog = () => {
  showDialog.value = false
  createdKey.value = null
}

const confirmDelete = (id: string) => {
  selectedKeyId.value = id
  showDeleteDialog.value = true
}

const handleDelete = async () => {
  try {
    await keysApi.delete(selectedKeyId.value)
    await fetchKeys()
    showDeleteDialog.value = false
  } catch (err: any) {
    alert(err.message || '删除失败')
  }
}

const toggleKey = async (key: UserAPIKey) => {
  disablingKeyId.value = key.id
  try {
    if (key.is_active) {
      await keysApi.disable(key.id)
    } else {
      await keysApi.enable(key.id)
    }
    await fetchKeys()
  } catch (err: any) {
    alert(err.message || '操作失败')
  } finally {
    disablingKeyId.value = ''
  }
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString('zh-CN')
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">API Key 管理</h1>
      <button
        @click="openCreateDialog"
        class="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
      >
        <Plus :size="18" />
        创建 API Key
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      加载中...
    </div>

    <div v-else-if="keys.length === 0" class="bg-white rounded-xl shadow-sm p-12 text-center">
      <div class="w-16 h-16 bg-slate-100 rounded-full flex items-center justify-center mx-auto mb-4">
        <EyeOff :size="32" class="text-slate-400" />
      </div>
      <h3 class="text-lg font-semibold text-slate-900 mb-2">暂无 API Key</h3>
      <p class="text-slate-600 mb-4">创建您的第一个 API Key 开始使用 AI 服务</p>
      <button
        @click="openCreateDialog"
        class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
      >
        创建 API Key
      </button>
    </div>

    <div v-else class="bg-white rounded-xl shadow-sm overflow-hidden">
      <table class="w-full">
        <thead class="bg-slate-50 border-b border-slate-200">
          <tr>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">名称</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">API Key</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">每日额度</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">状态</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">创建时间</th>
            <th class="text-right px-6 py-3 text-sm font-medium text-slate-600">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="key in keys" :key="key.id" class="hover:bg-slate-50">
            <td class="px-6 py-4 text-sm text-slate-900">{{ key.name }}</td>
            <td class="px-6 py-4">
              <code class="text-sm text-slate-600 bg-slate-100 px-2 py-1 rounded">
                sk-****...
              </code>
            </td>
            <td class="px-6 py-4 text-sm text-slate-600">
              {{ key.quota_daily ? key.quota_daily.toLocaleString() : '无限制' }}
            </td>
            <td class="px-6 py-4">
              <span
                :class="[
                  'px-2 py-1 text-xs font-medium rounded-full',
                  key.is_active
                    ? 'bg-green-100 text-green-700'
                    : 'bg-slate-100 text-slate-600',
                ]"
              >
                {{ key.is_active ? '活跃' : '禁用' }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm text-slate-600">
              {{ formatDate(key.created_at) }}
            </td>
            <td class="px-6 py-4">
              <div class="flex items-center justify-end gap-2">
                <button
                  @click="toggleKey(key)"
                  :disabled="disablingKeyId === key.id"
                  :class="[
                    'p-2 rounded',
                    key.is_active
                      ? 'hover:bg-yellow-100 text-yellow-600'
                      : 'hover:bg-green-100 text-green-600',
                  ]"
                  :title="key.is_active ? '禁用' : '启用'"
                >
                  <Power :size="16" />
                </button>
                <button
                  @click="confirmDelete(key.id)"
                  class="p-2 rounded hover:bg-red-100 text-red-600"
                  title="删除"
                >
                  <Trash2 :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 创建对话框 -->
    <div
      v-if="showDialog"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    >
      <div class="bg-white rounded-xl shadow-lg w-full max-w-md p-6">
        <h2 class="text-xl font-bold text-slate-900 mb-4">
          {{ createdKey ? 'API Key 创建成功' : '创建 API Key' }}
        </h2>

        <div v-if="!createdKey" class="space-y-4">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">名称</label>
            <input
              v-model="newKeyName"
              type="text"
              placeholder="例如：生产环境"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">
              每日额度（可选）
            </label>
            <input
              v-model.number="newKeyQuota"
              type="number"
              placeholder="留空表示无限制"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <div class="flex gap-3">
            <button
              @click="closeDialog"
              class="flex-1 px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200"
            >
              取消
            </button>
            <button
              @click="handleCreate"
              :disabled="!newKeyName"
              class="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed"
            >
              创建
            </button>
          </div>
        </div>

        <div v-else class="space-y-4">
          <div class="bg-amber-50 border border-amber-200 rounded-lg p-4">
            <p class="text-sm text-amber-800 mb-2">
              请立即复制此 API Key。关闭此窗口后将无法再次查看完整密钥。
            </p>
            <div class="flex items-center gap-2">
              <code class="flex-1 text-sm bg-white px-3 py-2 rounded border border-amber-200 break-all">
                {{ createdKey.key }}
              </code>
              <button
                @click="copyKey"
                class="p-2 bg-amber-100 text-amber-700 rounded hover:bg-amber-200"
              >
                <Check v-if="copied" :size="20" />
                <Copy v-else :size="20" />
              </button>
            </div>
          </div>
          <button
            @click="closeDialog"
            class="w-full px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            我已复制
          </button>
        </div>
      </div>
    </div>

    <!-- 删除确认对话框 -->
    <div
      v-if="showDeleteDialog"
      class="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4"
    >
      <div class="bg-white rounded-xl shadow-lg w-full max-w-sm p-6">
        <h2 class="text-xl font-bold text-slate-900 mb-2">确认删除</h2>
        <p class="text-slate-600 mb-4">确定要删除此 API Key 吗？此操作不可撤销。</p>
        <div class="flex gap-3">
          <button
            @click="showDeleteDialog = false"
            class="flex-1 px-4 py-2 bg-slate-100 text-slate-700 rounded-lg hover:bg-slate-200"
          >
            取消
          </button>
          <button
            @click="handleDelete"
            class="flex-1 px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700"
          >
            删除
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
