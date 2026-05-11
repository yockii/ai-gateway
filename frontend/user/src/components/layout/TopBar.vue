<script setup lang="ts">
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useLayoutStore } from '@/stores/layout'
import { Menu, LogOut, User } from 'lucide-vue-next'

const router = useRouter()
const authStore = useAuthStore()
const layoutStore = useLayoutStore()

const handleLogout = async () => {
  await authStore.logout()
  router.push('/')
}
</script>

<template>
  <header class="h-16 bg-white border-b border-slate-200 flex items-center justify-between px-4">
    <div class="flex items-center gap-4">
      <button
        @click="layoutStore.openMobileSidebar"
        class="lg:hidden p-2 hover:bg-slate-100 rounded"
      >
        <Menu :size="24" />
      </button>
      <h1 class="text-lg font-semibold text-slate-800">
        {{ $route.meta.title || '控制台' }}
      </h1>
    </div>

    <div class="flex items-center gap-4">
      <div class="flex items-center gap-2">
        <div class="w-8 h-8 bg-blue-500 rounded-full flex items-center justify-center">
          <User :size="16" class="text-white" />
        </div>
        <span class="text-sm text-slate-600 hidden sm:block">
          {{ authStore.user?.name || '用户' }}
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
