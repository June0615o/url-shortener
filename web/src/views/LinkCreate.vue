<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'
import { ElMessage } from 'element-plus'
import QrcodeModal from '../components/QrcodeModal.vue'

const router = useRouter()
const form = reactive({
  url: '', custom_code: '', title: '', description: '',
  expire_at: '', password: '', redirect_type: '302',
})
const loading = ref(false)
const result = ref(null)

async function create() {
  if (!form.url) return ElMessage.warning('请输入URL')
  loading.value = true
  try {
    const body = { url: form.url }
    if (form.custom_code) body.custom_code = form.custom_code
    if (form.title) body.title = form.title
    if (form.description) body.description = form.description
    if (form.expire_at) body.expire_at = new Date(form.expire_at).toISOString()
    if (form.password) body.password = form.password
    if (form.redirect_type) body.redirect_type = form.redirect_type

    result.value = await api.post('/links', body)
    ElMessage.success('创建成功！')
  } catch {} finally { loading.value = false }
}
</script>

<template>
  <div class="create-container">
    <h3>创建短链接</h3>
    <el-card style="max-width:640px">
      <el-form label-position="top">
        <el-form-item label="目标 URL *">
          <el-input v-model="form.url" placeholder="https://example.com/very/long/url" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="自定义短码 (可选)">
              <el-input v-model="form.custom_code" placeholder="留空自动生成" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="过期时间 (可选)">
              <el-date-picker v-model="form.expire_at" type="datetime" placeholder="永不过期" style="width:100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="标题 (可选)">
          <el-input v-model="form.title" placeholder="便于识别" />
        </el-form-item>
        <el-form-item label="描述 (可选)">
          <el-input v-model="form.description" type="textarea" :rows="2" placeholder="备注信息" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="访问密码 (可选)">
              <el-input v-model="form.password" type="password" placeholder="留空为公开访问" show-password />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="跳转类型">
              <el-select v-model="form.redirect_type" style="width:100%">
                <el-option label="302 临时重定向" value="302" />
                <el-option label="301 永久重定向" value="301" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-button type="primary" :loading="loading" @click="create" style="width:100%">创建短链接</el-button>
      </el-form>

      <div v-if="result" style="margin-top:20px;padding:16px;background:#f0f9ff;border-radius:8px">
        <p><strong>短链接：</strong>
          <el-link :href="result.short_url" target="_blank">{{ result.short_url }}</el-link>
        </p>
        <p><strong>短码：</strong>{{ result.short_code }}</p>
        <div style="margin-top:8px">
          <el-button size="small" @click="navigator.clipboard.writeText(result.short_url)">📋 复制链接</el-button>
          <QrcodeModal :url="result.short_url" />
          <el-button size="small" @click="router.push('/links')">返回列表</el-button>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
.create-container { max-width: 720px; }
</style>
