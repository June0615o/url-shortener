<script setup>
import { ref, onMounted } from 'vue'
import api from '../api'
import TrendChart from '../components/TrendChart.vue'
import DevicePie from '../components/DevicePie.vue'

const overview = ref({ total_links: 0, total_clicks: 0, today_clicks: 0, active_links: 0, expired_links: 0, avg_clicks_per_day: 0 })
const trend = ref([])
const geo = ref([])
const devices = ref([])
const loading = ref(true)

onMounted(async () => {
  try {
    const [ov, td, gd, dd] = await Promise.all([
      api.get('/dashboard/overview'),
      api.get('/dashboard/trend?hours=24'),
      api.get('/dashboard/geo'),
      api.get('/dashboard/devices'),
    ])
    overview.value = ov
    trend.value = td
    geo.value = gd
    devices.value = dd
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-loading="loading">
    <h3>数据看板</h3>
    <el-row :gutter="16">
      <el-col :span="4" v-for="card in [
        { label: '总链接', value: overview.total_links, color: '#409eff' },
        { label: '总点击', value: overview.total_clicks, color: '#67c23a' },
        { label: '今日点击', value: overview.today_clicks, color: '#e6a23c' },
        { label: '活跃链接', value: overview.active_links, color: '#909399' },
        { label: '已过期', value: overview.expired_links, color: '#f56c6c' },
        { label: '日均点击', value: overview.avg_clicks_per_day?.toFixed(1), color: '#409eff' },
      ]" :key="card.label" style="margin-bottom:16px">
        <el-card shadow="hover">
          <div style="text-align:center">
            <div style="color:#909399;font-size:14px">{{ card.label }}</div>
            <div :style="{color:card.color,fontSize:'28px',fontWeight:'bold',marginTop:'8px'}">
              {{ card.value ?? 0 }}
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16">
      <el-col :span="14">
        <el-card><template #header>点击趋势 (24小时)</template>
          <TrendChart :data="trend" />
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card><template #header>设备分布</template>
          <DevicePie :data="devices" />
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="16" style="margin-top:16px">
      <el-col :span="24">
        <el-card>
          <template #header>访问来源 Top 20</template>
          <el-table :data="geo" size="small" max-height="400">
            <el-table-column prop="country" label="国家/地区" />
            <el-table-column prop="clicks" label="点击量" sortable />
          </el-table>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
