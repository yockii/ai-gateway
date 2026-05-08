<script setup lang="ts">
import { useAuthStore } from '@/stores/auth'
import { LogOut, User } from 'lucide-vue-next'

const authStore = useAuthStore()

const handleLogout = async () => {
  await authStore.logout()
}
</script>

<template>
  <header class="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-6">
    <h1 class="text-lg font-semibold text-slate-800">
      {{ $route.meta.title || '管理控制台' }}
    </h1>

    <div class="flex items-center gap-4">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 bg-purple-500 rounded-full flex items-center justify-center">
          <User :size="16" class="text-white" />
        </div>
        <span class="text-sm text-slate-600">
          {{ authStore.admin?.name || '管理员' }}
        </span>
      </div>
      <button
        @click="handleLogout"
        class="p-2 hover:bg-slate-100 rounded text-slate-600"
        title="退出登录"
      >
        <LogOut :size="20" />
      </button>
    </div>
  </header>
</template>
