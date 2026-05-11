<script setup lang="ts">
import { ArrowRight, Check, Zap, Shield, TrendingUp, Code } from 'lucide-vue-next'

const features = [
  {
    icon: Zap,
    title: '极速响应',
    description: '智能路由，毫秒级响应，确保您的应用流畅运行',
  },
  {
    icon: Shield,
    title: '安全可靠',
    description: '企业级安全防护，数据加密传输，保障您的 API 安全',
  },
  {
    icon: TrendingUp,
    title: '实时监控',
    description: '完整的调用监控和数据分析，随时掌握 API 使用情况',
  },
  {
    icon: Code,
    title: '简单易用',
    description: '统一 API 接口，一行代码即可接入多种 AI 模型',
  },
]

const models = [
  { name: 'GPT-4', provider: 'OpenAI', description: '最强大的语言模型' },
  { name: 'GPT-3.5', provider: 'OpenAI', description: '快速且经济的选择' },
  { name: 'Claude', provider: 'Anthropic', description: '长文本理解专家' },
  { name: '文心一言', provider: '百度', description: '中文理解更精准' },
]

const plans = [
  {
    name: '免费版',
    price: '¥0',
    period: '永久',
    features: ['10,000 Tokens/月', '基础模型访问', '社区支持', 'API 调用统计'],
    cta: '开始使用',
    highlight: false,
  },
  {
    name: '专业版',
    price: '¥99',
    period: '/月',
    features: ['100,000 Tokens/月', '所有模型访问', '优先处理', '邮件支持', '高级数据分析'],
    cta: '立即升级',
    highlight: true,
  },
  {
    name: '企业版',
    price: '¥999',
    period: '/月',
    features: ['无限 Tokens', '专属模型', 'SLA 保证', '7x24 技术支持', '定制化服务'],
    cta: '联系销售',
    highlight: false,
  },
]

const stats = [
  { value: '99.9%', label: '可用性' },
  { value: '<50ms', label: '平均响应' },
  { value: '10+', label: 'AI 模型' },
  { value: '1000+', label: '企业用户' },
]

const codeExample = `// 安装 SDK
npm install @ai-gateway/sdk

// 初始化
import AIGateway from '@ai-gateway/sdk'

const ai = new AIGateway('your-api-key')

// 调用模型
const response = await ai.chat.completions.create({
  model: 'gpt-4',
  messages: [{ role: 'user', content: 'Hello!' }]
})`

const scrollToSection = (id: string) => {
  document.getElementById(id)?.scrollIntoView({ behavior: 'smooth' })
}
</script>

<template>
  <div class="min-h-screen bg-white">
    <!-- 导航栏 -->
    <nav class="fixed top-0 left-0 right-0 z-50 bg-white/80 backdrop-blur-lg border-b border-slate-200">
      <div class="max-w-7xl mx-auto px-6 py-4 flex items-center justify-between">
        <div class="flex items-center gap-2">
          <div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg"></div>
          <span class="text-xl font-bold text-slate-900">AI Gateway</span>
        </div>
        <div class="hidden md:flex items-center gap-8">
          <a href="#features" @click.prevent="scrollToSection('features')" class="text-slate-600 hover:text-slate-900">功能特点</a>
          <a href="#models" @click.prevent="scrollToSection('models')" class="text-slate-600 hover:text-slate-900">支持模型</a>
          <a href="#pricing" @click.prevent="scrollToSection('pricing')" class="text-slate-600 hover:text-slate-900">价格方案</a>
          <a href="#docs" @click.prevent="scrollToSection('docs')" class="text-slate-600 hover:text-slate-900">开发文档</a>
        </div>
        <div class="flex items-center gap-3">
          <router-link to="/login" class="px-4 py-2 text-slate-600 hover:text-slate-900">登录</router-link>
          <router-link to="/register" class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700">注册</router-link>
        </div>
      </div>
    </nav>

    <!-- Hero 区域 -->
    <section class="pt-32 pb-20 px-6">
      <div class="max-w-7xl mx-auto text-center">
        <div class="inline-flex items-center gap-2 px-4 py-2 bg-blue-50 text-blue-700 rounded-full text-sm mb-6">
          <span class="w-2 h-2 bg-blue-500 rounded-full animate-pulse"></span>
          新一代 AI 模型网关
        </div>
        <h1 class="text-5xl md:text-6xl font-bold text-slate-900 mb-6">
          统一接入
          <span class="text-transparent bg-clip-text bg-gradient-to-r from-blue-600 to-purple-600">所有 AI 模型</span>
        </h1>
        <p class="text-xl text-slate-600 max-w-2xl mx-auto mb-8">
          一个 API Key，访问 GPT-4、Claude、文心一言等 10+ 主流 AI 模型。
          极简接入，稳定可靠，助力您的应用快速落地 AI 能力。
        </p>
        <div class="flex items-center justify-center gap-4">
          <router-link to="/register" class="flex items-center gap-2 px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-lg">
            免费开始
            <ArrowRight :size="20" />
          </router-link>
          <a href="#docs" @click.prevent="scrollToSection('docs')" class="px-6 py-3 border border-slate-300 rounded-lg hover:bg-slate-50 text-lg">
            查看文档
          </a>
        </div>

        <!-- 数据统计 -->
        <div class="grid grid-cols-2 md:grid-cols-4 gap-6 mt-16 max-w-3xl mx-auto">
          <div v-for="stat in stats" :key="stat.label" class="text-center">
            <div class="text-3xl font-bold text-slate-900">{{ stat.value }}</div>
            <div class="text-sm text-slate-600 mt-1">{{ stat.label }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- 功能特点 -->
    <section id="features" class="py-20 px-6 bg-slate-50">
      <div class="max-w-7xl mx-auto">
        <div class="text-center mb-12">
          <h2 class="text-3xl font-bold text-slate-900 mb-4">为什么选择 AI Gateway</h2>
          <p class="text-slate-600">强大的功能，简单的使用方式</p>
        </div>
        <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-6">
          <div v-for="feature in features" :key="feature.title" class="bg-white p-6 rounded-xl shadow-sm">
            <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
              <component :is="feature.icon" class="text-blue-600" :size="24" />
            </div>
            <h3 class="text-lg font-semibold text-slate-900 mb-2">{{ feature.title }}</h3>
            <p class="text-slate-600 text-sm">{{ feature.description }}</p>
          </div>
        </div>
      </div>
    </section>

    <!-- 支持的模型 -->
    <section id="models" class="py-20 px-6">
      <div class="max-w-7xl mx-auto">
        <div class="text-center mb-12">
          <h2 class="text-3xl font-bold text-slate-900 mb-4">支持的 AI 模型</h2>
          <p class="text-slate-600">一个平台，接入所有主流 AI 模型</p>
        </div>
        <div class="grid md:grid-cols-2 lg:grid-cols-4 gap-4">
          <div v-for="model in models" :key="model.name" class="border border-slate-200 rounded-xl p-4 hover:border-blue-300 hover:shadow-md transition">
            <div class="font-semibold text-slate-900">{{ model.name }}</div>
            <div class="text-sm text-slate-500 mb-2">{{ model.provider }}</div>
            <div class="text-sm text-slate-600">{{ model.description }}</div>
          </div>
        </div>
      </div>
    </section>

    <!-- 代码示例 -->
    <section id="docs" class="py-20 px-6 bg-slate-900">
      <div class="max-w-7xl mx-auto">
        <div class="grid lg:grid-cols-2 gap-12 items-center">
          <div>
            <h2 class="text-3xl font-bold text-white mb-4">极简开发体验</h2>
            <p class="text-slate-400 mb-6">只需几行代码，即可在您的应用中集成强大的 AI 能力。我们提供完善的 SDK 和文档。</p>
            <ul class="space-y-3">
              <li class="flex items-center gap-3 text-slate-300">
                <Check class="text-green-500" :size="20" />
                统一的 API 接口
              </li>
              <li class="flex items-center gap-3 text-slate-300">
                <Check class="text-green-500" :size="20" />
                多语言 SDK 支持
              </li>
              <li class="flex items-center gap-3 text-slate-300">
                <Check class="text-green-500" :size="20" />
                详细的开发文档
              </li>
              <li class="flex items-center gap-3 text-slate-300">
                <Check class="text-green-500" :size="20" />
                实时调试工具
              </li>
            </ul>
          </div>
          <div class="bg-slate-800 rounded-xl p-6 overflow-x-auto">
            <pre class="text-sm text-slate-300"><code>{{ codeExample }}</code></pre>
          </div>
        </div>
      </div>
    </section>

    <!-- 价格方案 -->
    <section id="pricing" class="py-20 px-6">
      <div class="max-w-7xl mx-auto">
        <div class="text-center mb-12">
          <h2 class="text-3xl font-bold text-slate-900 mb-4">灵活的价格方案</h2>
          <p class="text-slate-600">选择适合您的方案，随时升级</p>
        </div>
        <div class="grid md:grid-cols-3 gap-6 max-w-5xl mx-auto">
          <div v-for="plan in plans" :key="plan.name"
               :class="['border rounded-xl p-6', plan.highlight ? 'border-blue-500 shadow-lg relative' : 'border-slate-200']">
            <div v-if="plan.highlight" class="absolute top-0 left-1/2 -translate-x-1/2 -translate-y-1/2 px-3 py-1 bg-blue-600 text-white text-xs rounded-full">
              推荐
            </div>
            <h3 class="text-lg font-semibold text-slate-900 mb-2">{{ plan.name }}</h3>
            <div class="flex items-baseline gap-1 mb-4">
              <span class="text-3xl font-bold text-slate-900">{{ plan.price }}</span>
              <span class="text-slate-600">{{ plan.period }}</span>
            </div>
            <ul class="space-y-2 mb-6">
              <li v-for="feature in plan.features" :key="feature" class="flex items-center gap-2 text-sm text-slate-600">
                <Check class="text-green-500" :size="16" />
                {{ feature }}
              </li>
            </ul>
            <button :class="['w-full py-2 rounded-lg', plan.highlight ? 'bg-blue-600 text-white hover:bg-blue-700' : 'bg-slate-100 text-slate-900 hover:bg-slate-200']">
              {{ plan.cta }}
            </button>
          </div>
        </div>
      </div>
    </section>

    <!-- CTA -->
    <section class="py-20 px-6 bg-gradient-to-r from-blue-600 to-purple-600">
      <div class="max-w-3xl mx-auto text-center">
        <h2 class="text-3xl font-bold text-white mb-4">准备好开始了吗？</h2>
        <p class="text-blue-100 mb-8">立即注册，获得 10,000 免费 Tokens，体验 AI 带来的无限可能</p>
        <router-link to="/register" class="inline-flex items-center gap-2 px-8 py-3 bg-white text-blue-600 rounded-lg hover:bg-slate-100 text-lg font-semibold">
          免费注册
          <ArrowRight :size="20" />
        </router-link>
      </div>
    </section>

    <!-- 页脚 -->
    <footer class="py-12 px-6 bg-slate-900">
      <div class="max-w-7xl mx-auto">
        <div class="flex flex-col md:flex-row items-center justify-between gap-4">
          <div class="flex items-center gap-2">
            <div class="w-8 h-8 bg-gradient-to-br from-blue-500 to-purple-600 rounded-lg"></div>
            <span class="text-lg font-bold text-white">AI Gateway</span>
          </div>
          <div class="flex items-center gap-6 text-sm text-slate-400">
            <a href="#" class="hover:text-white">关于我们</a>
            <a href="#" class="hover:text-white">使用条款</a>
            <a href="#" class="hover:text-white">隐私政策</a>
            <a href="#" class="hover:text-white">联系我们</a>
          </div>
          <div class="text-sm text-slate-500">
            © 2024 AI Gateway. All rights reserved.
          </div>
        </div>
      </div>
    </footer>
  </div>
</template>
