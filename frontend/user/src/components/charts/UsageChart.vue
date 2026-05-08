<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import type { EChartsOption } from 'echarts'

interface ChartData {
  dates: string[]
  requests: number[]
  tokens: number[]
}

const props = defineProps<{
  data: ChartData
}>()

const chartRef = ref<HTMLDivElement>()
let chartInstance: echarts.ECharts | null = null

const initChart = () => {
  if (!chartRef.value) return

  chartInstance = echarts.init(chartRef.value)
  updateChart()
}

const updateChart = () => {
  if (!chartInstance) return

  const option: EChartsOption = {
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
    },
    legend: {
      data: ['Token 使用量', '请求数'],
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: props.data.dates,
    },
    yAxis: [
      {
        type: 'value',
        name: 'Token 使用量',
        position: 'left',
      },
      {
        type: 'value',
        name: '请求数',
        position: 'right',
      },
    ],
    series: [
      {
        name: 'Token 使用量',
        type: 'bar',
        data: props.data.tokens,
        itemStyle: { color: '#3b82f6' },
      },
      {
        name: '请求数',
        type: 'line',
        yAxisIndex: 1,
        data: props.data.requests,
        itemStyle: { color: '#10b981' },
        smooth: true,
      },
    ],
  }

  chartInstance.setOption(option)
}

onMounted(() => {
  initChart()
  window.addEventListener('resize', () => chartInstance?.resize())
})

onUnmounted(() => {
  chartInstance?.dispose()
  window.removeEventListener('resize', () => chartInstance?.resize())
})

watch(() => props.data, updateChart, { deep: true })
</script>

<template>
  <div ref="chartRef" class="w-full h-80" />
</template>
