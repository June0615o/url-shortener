<script setup>
import { computed } from 'vue'
import VChart from 'vue-echarts'
import { use } from 'echarts/core'
import { PieChart } from 'echarts/charts'
import { TooltipComponent, LegendComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'

use([PieChart, TooltipComponent, LegendComponent, CanvasRenderer])

const props = defineProps({ data: { type: Array, default: () => [] } })

const option = computed(() => ({
  tooltip: { trigger: 'item' },
  legend: { bottom: 0 },
  series: [{
    type: 'pie',
    radius: ['40%', '70%'],
    center: ['50%', '45%'],
    data: props.data.map(d => ({ name: d.device || 'Unknown', value: d.clicks })),
    label: { show: false },
    emphasis: { label: { show: true } },
  }],
}))
</script>

<template>
  <v-chart v-if="data.length" :option="option" style="height:280px" autoresize />
  <div v-else style="text-align:center;color:#909399;padding:60px">暂无数据</div>
</template>
