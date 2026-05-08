<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { usersApi } from '@/api/users'
import type { User } from '@/types/models'
import { Search, Eye, Edit2, Ban, Check } from 'lucide-vue-next'

const users = ref<User[]>([])
const loading = ref(false)
const searchQuery = ref('')
const showDetailDialog = ref(false)
const selectedUser = ref<User | null>(null)

onMounted(async () => {
  await fetchUsers()
})

const fetchUsers = async () => {
  loading.value = true
  try {
    const response = await usersApi.list({ search: searchQuery.value })
    users.value = response.data
  } catch (err) {
    console.error('Failed to fetch users:', err)
  } finally {
    loading.value = false
  }
}

const viewUser = (user: User) => {
  selectedUser.value = user
  showDetailDialog.value = true
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
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">状态</th>
            <th class="text-left px-6 py-3 text-sm font-medium text-slate-600">注册时间</th>
            <th class="text-right px-6 py-3 text-sm font-medium text-slate-600">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-200">
          <tr v-for="user in users" :key="user.id" class="hover:bg-slate-50">
            <td class="px-6 py-4 text-sm text-slate-900">{{ user.email }}</td>
            <td class="px-6 py-4 text-sm text-slate-900">{{ user.name }}</td>
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
                <button @click="viewUser(user)" class="p-2 rounded hover:bg-slate-100 text-slate-600">
                  <Eye :size="16" />
                </button>
                <button @click="toggleStatus(user)" class="p-2 rounded hover:bg-slate-100 text-slate-600">
                  <Ban v-if="user.is_active" :size="16" />
                  <Check v-else :size="16" />
                </button>
                <button @click="deleteUser(user)" class="p-2 rounded hover:bg-red-100 text-red-600">
                  <Ban :size="16" />
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <div v-if="showDetailDialog && selectedUser" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white rounded-xl p-6 w-full max-w-md">
        <h2 class="text-xl font-bold mb-4">用户详情</h2>
        <div class="space-y-2 text-sm">
          <p><span class="text-slate-600">ID:</span> {{ selectedUser.id }}</p>
          <p><span class="text-slate-600">邮箱:</span> {{ selectedUser.email }}</p>
          <p><span class="text-slate-600">姓名:</span> {{ selectedUser.name }}</p>
          <p><span class="text-slate-600">用户组:</span> {{ selectedUser.user_group_id }}</p>
          <p><span class="text-slate-600">状态:</span> {{ selectedUser.is_active ? '活跃' : '禁用' }}</p>
        </div>
        <button @click="showDetailDialog = false" class="mt-4 w-full px-4 py-2 bg-slate-100 rounded-lg">关闭</button>
      </div>
    </div>
  </div>
</template>
