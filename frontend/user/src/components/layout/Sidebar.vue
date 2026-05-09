<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useLayoutStore } from '@/stores/layout'
import {
  LayoutDashboard,
  KeyRound,
  TrendingUp,
  FileText,
  Settings,
  CreditCard,
} from 'lucide-vue-next'

const layoutStore = useLayoutStore()
const route = useRoute()

const navItems = computed(() => [
  { name: '控制台概览', path: '/dashboard', icon: LayoutDashboard },
  { name: 'API Key 管理', path: '/api-keys', icon: KeyRound },
  { name: '使用统计', path: '/usage', icon: TrendingUp },
  { name: '账单查询', path: '/bills', icon: FileText },
  { name: '套餐购买', path: '/plans', icon: CreditCard },
  { name: '个人设置', path: '/settings', icon: Settings },
])

const isActive = (path: string) => route.path === path
</script>

<template>
  <aside
    :class="[
      'bg-slate-900 text-white transition-all duration-300 flex flex-col',
      layoutStore.sidebarCollapsed ? 'w-16' : 'w-60',
    ]"
  >
    <div class="h-16 flex items-center justify-between px-4 border-b border-slate-700">
      <span v-if="!layoutStore.sidebarCollapsed" class="text-xl font-bold">AI Gateway</span>
      <button
        v-else
        @click="layoutStore.toggleSidebar"
        class="p-2 hover:bg-slate-800 rounded"
      >
        <LayoutDashboard :size="24" />
      </button>
    </div>

    <nav class="flex-1 py-4">
      <router-link
        v-for="item in navItems"
        :key="item.path"
        :to="item.path"
        :class="[
          'flex items-center px-4 py-3 mx-2 rounded-lg transition-colors',
          isActive(item.path)
            ? 'bg-blue-600 text-white'
            : 'text-slate-300 hover:bg-slate-800 hover:text-white',
        ]"
      >
        <component :is="item.icon" :size="20" />
        <span v-if="!layoutStore.sidebarCollapsed" class="ml-3">{{ item.name }}</span>
      </router-link>
    </nav>
  </aside>
</template>
