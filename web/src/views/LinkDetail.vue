<script setup>
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '../api'
import TrendChart from '../components/TrendChart.vue'
import DevicePie from '../components/DevicePie.vue'
import QrcodeModal from '../components/QrcodeModal.vue'

const route = useRoute()
const router = useRouter()
const code = route.params.code
const link = ref(null)
const stats = ref({ total_clicks: 0, unique_ips: 0, trend: [], geo: [], referrers: [] })

onMounted(async () => {
  try {
    const [l, s] = await Promise.all([
      api.get(`/links/${code}`),
      api.get(`/links/${code}/stats`).catch(() => null),
    ])
    link.value = l
    if (s) stats.value = s
  } catch {}
})

function copyUrl(url) { navigator.clipboard.writeText(url) }
</script>

<template>
  <div v-if="link">
    <el-page-header @back="router.push('/links')" content="链接详情" style="margin-bottom:16px" />
    <el-descriptions :column="2" border>
      <el-descriptions-item label="短码">{{ link.short_code }}</el-descriptions-item>
      <el-descriptions-item label="短链接">
        <el-link :href="link.short_url" target="_blank">{{ link.short_url }}</el-link>
      </el-descriptions-item>
      <el-descriptions-item label="原始URL" :span="2">
        <el-link :href="link.original_url" target="_blank">{{ link.original_url }}</el-link>
      </el-descriptions-item>
      <el-descriptions-item label="标题">{{ link.title || '-' }}</el-descriptions-item>
      <el-descriptions-item label="点击量">{{ link.click_count }}</el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ new Date(link.created_at).toLocaleString('zh-CN') }}</el-descriptions-item>
      <el-descriptions-item label="过期时间">{{ link.expire_at ? new Date(link.expire_at).toLocaleString('zh-CN') : '永不过期' }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="link.is_active ? 'success' : 'danger'">{{ link.is_active ? '活跃' : '已禁用' }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="跳转类型">{{ link.redirect_type }}</el-descriptions-item>
    </el-descriptions>
    <div style="margin-top:16px">
      <el-button @click="copyUrl(link.short_url)">📋 复制</el-button>
      <QrcodeModal :url="link.short_url" />
    </div>

    <el-divider />
    <h4>点击统计</h4>
    <el-row :gutter="16" style="margin-bottom:16px">
      <el-col :span="8"><el-card><div style="text-align:center"><div style="color:#909399">总点击</div><div style="font-size:24px;font-weight:bold;color:#409eff">{{ stats.total_clicks }}</div></div></el-card></el-col>
      <el-col :span="8"><el-card><div style="text-align:center"><div style="color:#909399">独立访客</div><div style="font-size:24px;font-weight:bold;color:#67c23a">{{ stats.unique_ips }}</div></div></el-card></el-col>
    </el-row>
    <el-row :gutter="16">
      <el-col :span="14"><el-card><template #header>点击趋势</template><TrendChart :data="stats.trend || []" /></el-card></el-col>
      <el-col :span="10"><el-card><template #header>来源域名</template>
        <el-table :data="stats.referrers || []" size="small" max-height="300">
          <el-table-column prop="referer" label="来源" show-overflow-tooltip />
          <el-table-column prop="clicks" label="点击" width="80" />
        </el-table>
      </el-card></el-col>
    </el-row>
  </div>
</template>
