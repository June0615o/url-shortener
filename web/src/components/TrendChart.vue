<script setup>
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { LineChart } from 'echarts/charts'
import { GridComponent, TooltipComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([LineChart, GridComponent, TooltipComponent, CanvasRenderer])

const props = defineProps({ data: { type: Array, default: () => [] } })

const option = computed(() => ({
  tooltip: { trigger: 'axis' },
  grid: { left: 40, right: 20, top: 10, bottom: 30 },
  xAxis: {
    type: 'category',
    data: props.data.map(d => new Date(d.time).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })),
  },
  yAxis: { type: 'value', minInterval: 1 },
  series: [{
    data: props.data.map(d => d.clicks),
    type: 'line',
    smooth: true,
    areaStyle: { opacity: 0.15 },
    lineStyle: { color: '#409eff' },
    itemStyle: { color: '#409eff' },
  }],
}))
</script>

<template>
  <v-chart v-if="data.length" :option="option" style="height:300px" autoresize />
  <div v-else style="text-align:center;color:#909399;padding:60px">暂无数据</div>
</template>
