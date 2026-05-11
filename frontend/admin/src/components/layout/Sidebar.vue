<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useLayoutStore } from '@/stores/layout'
import {
  Users,
  Cpu,
  Truck,
  CreditCard,
  Activity,
  BarChart3,
  FileText,
  DollarSign,
} from 'lucide-vue-next'

const layoutStore = useLayoutStore()
const route = useRoute()

const navItems = computed(() => [
  { name: '用户管理', path: '/users', icon: Users },
  { name: '模型管理', path: '/models', icon: Cpu },
  { name: '供应商管理', path: '/suppliers', icon: Truck },
  { name: '套餐管理', path: '/plans', icon: CreditCard },
  { name: '定价管理', path: '/pricing', icon: DollarSign },
  { name: '审计日志', path: '/audit', icon: FileText },
  { name: '系统监控', path: '/monitoring', icon: Activity },
  { name: '运维大屏', path: '/operations', icon: BarChart3 },
])

const isActive = (path: string) => {
  // Handle active state for routes with params (e.g., /suppliers/:id)
  if (path === '/suppliers' && route.path.startsWith('/suppliers/')) {
    return true
  }
  return route.path === path
}
</script>

<template>
  <aside
    :class="[
      'bg-slate-900 text-white transition-all duration-300 flex flex-col',
      layoutStore.sidebarCollapsed ? 'w-16' : 'w-60',
    ]"
  >
    <div class="h-16 flex items-center px-4 border-b border-slate-700">
      <span v-if="!layoutStore.sidebarCollapsed" class="text-xl font-bold">管理后台</span>
    </div>

    <nav class="flex-1 py-4 overflow-y-auto">
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
        :title="layoutStore.sidebarCollapsed ? item.name : ''"
      >
        <component :is="item.icon" :size="20" />
        <span v-if="!layoutStore.sidebarCollapsed" class="ml-3">{{ item.name }}</span>
      </router-link>
    </nav>

    <!-- Collapse toggle button -->
    <button
      @click="layoutStore.toggleSidebar"
      class="p-3 mx-2 mb-2 rounded-lg text-slate-400 hover:bg-slate-800 hover:text-white transition-colors"
      :title="layoutStore.sidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
    >
      <svg
        :class="{ 'rotate-180': layoutStore.sidebarCollapsed }"
        class="w-5 h-5 transition-transform"
        fill="none"
        stroke="currentColor"
        viewBox="0 0 24 24"
      >
        <path
          stroke-linecap="round"
          stroke-linejoin="round"
          stroke-width="2"
          d="M11 19l-7-7 7-7m8 14l-7-7 7-7"
        />
      </svg>
    </button>
  </aside>
</template>
