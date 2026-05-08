<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { z } from 'zod'

const authStore = useAuthStore()

const activeTab = ref('profile')
const loading = ref(false)
const message = ref('')

// 基本信息
const name = ref('')
const email = ref('')

// 修改密码
const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const profileSchema = z.object({
  name: z.string().min(2, '姓名至少需要 2 个字符').max(50, '姓名最多 50 个字符'),
  email: z.string().email('请输入有效的邮箱地址'),
})

const passwordSchema = z.object({
  currentPassword: z.string().min(1, '请输入当前密码'),
  newPassword: z.string().min(8, '新密码至少需要 8 个字符'),
  confirmPassword: z.string(),
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: '两次输入的密码不一致',
  path: ['confirmPassword'],
})

onMounted(() => {
  if (authStore.user) {
    name.value = authStore.user.name
    email.value = authStore.user.email
  }
})

const updateProfile = async () => {
  message.value = ''
  const result = profileSchema.safeParse({ name: name.value, email: email.value })

  if (!result.success) {
    message.value = result.error.errors[0].message
    return
  }

  loading.value = true
  try {
    // TODO: 调用更新用户信息 API
    authStore.user!.name = name.value
    authStore.user!.email = email.value
    message.value = '保存成功'
    setTimeout(() => (message.value = ''), 3000)
  } catch (err: any) {
    message.value = err.message || '保存失败'
  } finally {
    loading.value = false
  }
}

const changePassword = async () => {
  message.value = ''
  const result = passwordSchema.safeParse({
    currentPassword: currentPassword.value,
    newPassword: newPassword.value,
    confirmPassword: confirmPassword.value,
  })

  if (!result.success) {
    message.value = result.error.errors[0].message
    return
  }

  loading.value = true
  try {
    // TODO: 调用修改密码 API
    message.value = '密码修改成功'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    setTimeout(() => (message.value = ''), 3000)
  } catch (err: any) {
    message.value = err.message || '密码修改失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="space-y-6">
    <h1 class="text-2xl font-bold text-slate-900">个人设置</h1>

    <div v-if="message" :class="[
      'px-4 py-3 rounded-lg text-sm',
      message.includes('成功') ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'
    ]">
      {{ message }}
    </div>

    <div class="bg-white rounded-xl shadow-sm">
      <div class="border-b border-slate-200">
        <nav class="flex">
          <button
            v-for="tab in [
              { id: 'profile', name: '基本信息' },
              { id: 'security', name: '安全设置' },
            ]"
            :key="tab.id"
            @click="activeTab = tab.id"
            :class="[
              'px-6 py-4 text-sm font-medium border-b-2 -mb-px',
              activeTab === tab.id
                ? 'border-blue-600 text-blue-600'
                : 'border-transparent text-slate-600 hover:text-slate-900',
            ]"
          >
            {{ tab.name }}
          </button>
        </nav>
      </div>

      <div class="p-6">
        <!-- 基本信息 -->
        <div v-if="activeTab === 'profile'" class="space-y-4 max-w-md">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">姓名</label>
            <input
              v-model="name"
              type="text"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">邮箱</label>
            <input
              v-model="email"
              type="email"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <button
            @click="updateProfile"
            :disabled="loading"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-slate-300"
          >
            {{ loading ? '保存中...' : '保存' }}
          </button>
        </div>

        <!-- 安全设置 -->
        <div v-if="activeTab === 'security'" class="space-y-4 max-w-md">
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">当前密码</label>
            <input
              v-model="currentPassword"
              type="password"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">新密码</label>
            <input
              v-model="newPassword"
              type="password"
              placeholder="至少 8 个字符"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <div>
            <label class="block text-sm font-medium text-slate-700 mb-1">确认新密码</label>
            <input
              v-model="confirmPassword"
              type="password"
              placeholder="再次输入新密码"
              class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
            />
          </div>
          <button
            @click="changePassword"
            :disabled="loading"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-slate-300"
          >
            {{ loading ? '修改中...' : '修改密码' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
