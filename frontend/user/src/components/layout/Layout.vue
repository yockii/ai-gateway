<script setup lang="ts">
import { useLayoutStore } from '@/stores/layout'
import Sidebar from './Sidebar.vue'
import TopBar from './TopBar.vue'

const layoutStore = useLayoutStore()
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <div class="flex">
      <!-- 侧边栏 -->
      <Sidebar />

      <!-- 主内容区域 -->
      <div class="flex-1 flex flex-col min-h-screen">
        <TopBar />

        <!-- 移动端侧边栏遮罩 -->
        <div
          v-if="layoutStore.mobileSidebarOpen"
          @click="layoutStore.closeMobileSidebar"
          class="fixed inset-0 bg-black/50 z-40 lg:hidden"
        />

        <!-- 移动端侧边栏 -->
        <div
          v-if="layoutStore.mobileSidebarOpen"
          class="fixed inset-y-0 left-0 w-60 bg-slate-900 z-50 lg:hidden"
        >
          <div class="flex justify-end p-4">
            <button
              @click="layoutStore.closeMobileSidebar"
              class="text-white hover:bg-slate-800 p-2 rounded"
            >
              <X :size="24" />
            </button>
          </div>
          <div class="px-2">
            <Sidebar />
          </div>
        </div>

        <!-- 页面内容 -->
        <main class="flex-1 p-6">
          <router-view />
        </main>
      </div>
    </div>
  </div>
</template>
