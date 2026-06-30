<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'
import { ElMessageBox } from 'element-plus'

const router = useRouter()
const links = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const loading = ref(false)

async function fetchLinks() {
  loading.value = true
  try {
    const res = await api.get('/links', { params: { page: page.value, page_size: pageSize.value } })
    links.value = res.data
    total.value = res.total
  } finally { loading.value = false }
}

async function del(link) {
  try {
    await ElMessageBox.confirm(`确定删除短链接 "${link.short_code}"？`, '确认删除', { type: 'warning' })
    await api.delete(`/links/${link.short_code}`)
    fetchLinks()
  } catch {}
}

function copyUrl(shortUrl) {
  navigator.clipboard.writeText(shortUrl)
}

onMounted(fetchLinks)
</script>

<template>
  <div>
    <div style="display:flex;justify-content:space-between;align-items:center;margin-bottom:16px">
      <h3>链接管理</h3>
      <el-button type="primary" @click="router.push('/links/create')">创建链接</el-button>
    </div>
    <el-table :data="links" v-loading="loading" stripe>
      <el-table-column prop="short_code" label="短码" width="120" />
      <el-table-column prop="original_url" label="原始URL" show-overflow-tooltip />
      <el-table-column prop="title" label="标题" show-overflow-tooltip />
      <el-table-column prop="click_count" label="点击" width="80" />
      <el-table-column prop="created_at" label="创建时间" width="170">
        <template #default="{ row }">{{ new Date(row.created_at).toLocaleString('zh-CN') }}</template>
      </el-table-column>
      <el-table-column label="操作" width="260" fixed="right">
        <template #default="{ row }">
          <el-button size="small" @click="copyUrl(row.short_url)">复制</el-button>
          <el-button size="small" type="info" @click="router.push(`/links/${row.short_code}`)">详情</el-button>
          <el-button size="small" type="danger" @click="del(row)">删除</el-button>
        </template>
      </el-table-column>
    </el-table>
    <div style="margin-top:16px;text-align:right">
      <el-pagination v-model:current-page="page" :page-size="pageSize" :total="total"
        layout="total, prev, pager, next" @current-change="fetchLinks" />
    </div>
  </div>
</template>
