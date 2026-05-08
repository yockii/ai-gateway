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
} from 'lucide-vue-next'

const layoutStore = useLayoutStore()
const route = useRoute()

const navItems = computed(() => [
  { name: '用户管理', path: '/admin/users', icon: Users },
  { name: '模型管理', path: '/admin/models', icon: Cpu },
  { name: '供应商管理', path: '/admin/suppliers', icon: Truck },
  { name: '套餐管理', path: '/admin/plans', icon: CreditCard },
  { name: '系统监控', path: '/admin/monitoring', icon: Activity },
  { name: '运维大屏', path: '/admin/operations', icon: BarChart3 },
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
    <div class="h-16 flex items-center px-4 border-b border-slate-700">
      <span v-if="!layoutStore.sidebarCollapsed" class="text-xl font-bold">管理后台</span>
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
