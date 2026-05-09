<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { z } from 'zod'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

const schema = z.object({
  email: z.string().email('请输入有效的邮箱地址'),
  password: z.string().min(1, '请输入密码'),
})

const isValid = computed(() => schema.safeParse({ email: email.value, password: password.value }).success)

const handleLogin = async () => {
  error.value = ''
  const result = schema.safeParse({ email: email.value, password: password.value })
  if (!result.success) {
    error.value = result.error.issues[0].message
    return
  }

  loading.value = true
  try {
    await authStore.login(email.value, password.value)
    router.push('/admin/users')
  } catch (err: any) {
    error.value = err.message || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="min-h-screen bg-slate-900 flex items-center justify-center p-4">
    <div class="w-full max-w-md bg-slate-800 rounded-xl p-8">
      <div class="text-center mb-8">
        <h1 class="text-2xl font-bold text-white">AI Gateway</h1>
        <p class="text-slate-400 mt-2">管理后台登录</p>
      </div>

      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-1">邮箱</label>
          <input
            v-model="email"
            type="email"
            class="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>
        <div>
          <label class="block text-sm font-medium text-slate-300 mb-1">密码</label>
          <input
            v-model="password"
            type="password"
            class="w-full px-4 py-2 bg-slate-700 border border-slate-600 rounded-lg text-white focus:ring-2 focus:ring-blue-500 focus:border-blue-500 outline-none"
          />
        </div>

        <div v-if="error" class="text-red-400 text-sm">{{ error }}</div>

        <button
          type="submit"
          :disabled="loading || !isValid"
          class="w-full bg-blue-600 text-white py-2 rounded-lg hover:bg-blue-700 disabled:bg-slate-600"
        >
          {{ loading ? '登录中...' : '登录' }}
        </button>
      </form>
    </div>
  </div>
</template>
