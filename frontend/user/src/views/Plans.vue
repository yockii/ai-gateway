<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { plansApi } from '@/api/plans'
import type { MembershipPlan, UserMembership } from '@/types/models'
import { Check, Crown } from 'lucide-vue-next'

const plans = ref<MembershipPlan[]>([])
const currentMembership = ref<UserMembership | null>(null)
const loading = ref(false)
const subscribing = ref<string | null>(null)
const selectedInterval = ref<'monthly' | 'yearly'>('monthly')

onMounted(async () => {
  await fetchData()
})

const fetchData = async () => {
  loading.value = true
  try {
    plans.value = await plansApi.list()
    currentMembership.value = await plansApi.getCurrent()
  } catch (err) {
    console.error('Failed to fetch plans:', err)
  } finally {
    loading.value = false
  }
}

const subscribe = async (planId: string) => {
  subscribing.value = planId
  try {
    const result = await plansApi.subscribe(planId, selectedInterval.value)
    // TODO: 处理 Stripe 支付
    alert('支付功能待实现')
  } catch (err: any) {
    alert(err.message || '订阅失败')
  } finally {
    subscribing.value = null
  }
}

const displayPrice = computed(() => (plan: MembershipPlan) => {
  return selectedInterval.value === 'yearly'
    ? plan.price_yearly
    : plan.price_monthly
})

const billingText = computed(() => {
  return selectedInterval.value === 'yearly' ? '/年' : '/月'
})

const isCurrentPlan = (planId: string) => {
  return currentMembership.value?.plan_id === planId && currentMembership.value.status === 'active'
}

const yearlySavings = (plan: MembershipPlan) => {
  const monthlyTotal = plan.price_monthly * 12
  return Math.round((1 - plan.price_yearly / monthlyTotal) * 100)
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">套餐购买</h1>
      <div class="flex items-center gap-2 bg-slate-100 rounded-lg p-1">
        <button
          @click="selectedInterval = 'monthly'"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            selectedInterval === 'monthly'
              ? 'bg-white text-slate-900 shadow-sm'
              : 'text-slate-600',
          ]"
        >
          按月付费
        </button>
        <button
          @click="selectedInterval = 'yearly'"
          :class="[
            'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
            selectedInterval === 'yearly'
              ? 'bg-white text-slate-900 shadow-sm'
              : 'text-slate-600',
          ]"
        >
          按年付费
          <span class="text-green-600 text-xs">省 20%</span>
        </button>
      </div>
    </div>

    <div v-if="loading" class="text-center py-12 text-slate-500">
      加载中...
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <div
        v-for="plan in plans"
        :key="plan.id"
        :class="[
          'bg-white rounded-xl shadow-sm p-6 relative',
          isCurrentPlan(plan.id) ? 'ring-2 ring-blue-600' : '',
        ]"
      >
        <div
          v-if="isCurrentPlan(plan.id)"
          class="absolute -top-3 left-1/2 -translate-x-1/2 bg-blue-600 text-white text-xs font-medium px-3 py-1 rounded-full"
        >
          当前套餐
        </div>

        <div class="text-center mb-6">
          <div class="w-12 h-12 bg-blue-100 rounded-full flex items-center justify-center mx-auto mb-3">
            <Crown :size="24" class="text-blue-600" />
          </div>
          <h3 class="text-lg font-semibold text-slate-900">{{ plan.name }}</h3>
          <p class="text-sm text-slate-600 mt-1">{{ plan.description }}</p>
        </div>

        <div class="text-center mb-6">
          <span class="text-3xl font-bold text-slate-900">
            ¥{{ displayPrice(plan) }}
          </span>
          <span class="text-slate-600">{{ billingText }}</span>
          <div
            v-if="selectedInterval === 'yearly' && plan.price_yearly > 0"
            class="text-sm text-green-600 mt-1"
          >
            节省 {{ yearlySavings(plan) }}%
          </div>
        </div>

        <ul class="space-y-3 mb-6">
          <li v-for="feature in plan.features" :key="feature" class="flex items-start gap-2 text-sm">
            <Check :size="16" class="text-green-600 mt-0.5 flex-shrink-0" />
            <span class="text-slate-600">{{ feature }}</span>
          </li>
        </ul>

        <button
          @click="subscribe(plan.id)"
          :disabled="subscribing === plan.id || isCurrentPlan(plan.id)"
          :class="[
            'w-full py-2 rounded-lg font-medium transition-colors',
            isCurrentPlan(plan.id)
              ? 'bg-slate-100 text-slate-600 cursor-not-allowed'
              : 'bg-blue-600 text-white hover:bg-blue-700 disabled:bg-slate-300',
          ]"
        >
          {{ subscribing === plan.id ? '处理中...' : isCurrentPlan(plan.id) ? '已订阅' : '订阅' }}
        </button>
      </div>
    </div>

    <div v-if="currentMembership" class="bg-white rounded-xl shadow-sm p-6">
      <h2 class="text-lg font-semibold text-slate-900 mb-4">当前订阅</h2>
      <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
        <div>
          <p class="text-slate-600">状态</p>
          <p class="font-medium text-slate-900">
            {{ currentMembership.status === 'active' ? '活跃' : '已过期' }}
          </p>
        </div>
        <div>
          <p class="text-slate-600">当前周期</p>
          <p class="font-medium text-slate-900">
            {{ new Date(currentMembership.current_period_start).toLocaleDateString() }} -
            {{ new Date(currentMembership.current_period_end).toLocaleDateString() }}
          </p>
        </div>
        <div>
          <p class="text-slate-600">自动续费</p>
          <p class="font-medium text-slate-900">
            {{ currentMembership.auto_renew ? '已开启' : '已关闭' }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>
