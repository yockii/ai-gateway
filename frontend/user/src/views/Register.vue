<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { z } from 'zod'

const router = useRouter()
const authStore = useAuthStore()

const name = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const error = ref('')
const loading = ref(false)

const registerSchema = z.object({
  name: z.string().min(2, '姓名至少需要 2 个字符').max(50, '姓名最多 50 个字符'),
  email: z.string().email('请输入有效的邮箱地址'),
  password: z.string().min(8, '密码至少需要 8 个字符'),
  confirmPassword: z.string(),
}).refine((data) => data.password === data.confirmPassword, {
  message: '两次输入的密码不一致',
  path: ['confirmPassword'],
})

const isValid = computed(() => {
  const result = registerSchema.safeParse({
    name: name.value,
    email: email.value,
    password: password.value,
    confirmPassword: confirmPassword.value,
  })
  return result.success
})

const handleRegister = async () => {
  error.value = ''
  const result = registerSchema.safeParse({
    name: name.value,
    email: email.value,
    password: password.value,
    confirmPassword: confirmPassword.value,
  })

  if (!result.success) {
    error.value = result.error.issues[0].message
    return
  }

  loading.value = true
  try {
    await authStore.register(email.value, password.value, name.value)
    router.push('/dashboard')
  } catch (err: any) {
    error.value = err.message || '注册失败，请稍后重试'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-50 flex items-center justify-center p-4">
    <div class="w-full max-w-md bg-white rounded-xl shadow-sm p-8">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold text-slate-900">AI Gateway</h1>
        <p class="text-slate-600 mt-2">创建新账户</p>
      </div>

      <form @submit.prevent="handleRegister" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">姓名</label>
          <input
            v-model="name"
            type="text"
            placeholder="您的姓名"
            class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">邮箱</label>
          <input
            v-model="email"
            type="email"
            placeholder="your@email.com"
            class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">密码</label>
          <input
            v-model="password"
            type="password"
            placeholder="至少 8 个字符"
            class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>

        <div>
          <label class="block text-sm font-medium text-slate-700 mb-1">确认密码</label>
          <input
            v-model="confirmPassword"
            type="password"
            placeholder="再次输入密码"
            class="w-full px-4 py-2 border border-slate-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>

        <div v-if="error" class="text-red-500 text-sm">
          {{ error }}
        </div>

        <button
          type="submit"
          :disabled="loading || !isValid"
          class="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-slate-300 disabled:cursor-not-allowed transition-colors"
        >
          {{ loading ? '注册中...' : '注册' }}
        </button>
      </form>

      <p class="text-center text-slate-600 text-sm mt-6">
        已有账号？
        <router-link to="/login" class="text-blue-600 hover:underline">
          登录
        </router-link>
      </p>
    </div>
  </div>
</template>
