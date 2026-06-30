<script setup>
import { ref } from 'vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const isRegister = ref(false)
const form = ref({ username: '', password: '', email: '' })
const loading = ref(false)

async function submit() {
  loading.value = true
  try {
    if (isRegister.value) {
      await auth.register(form.value)
    } else {
      await auth.login({ username: form.value.username, password: form.value.password })
    }
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2 style="text-align:center;margin-bottom:24px">
        {{ isRegister ? '注册' : '登录' }} — URL Shortener
      </h2>
      <el-form @submit.prevent="submit" label-position="top">
        <el-form-item label="用户名">
          <el-input v-model="form.username" placeholder="请输入用户名" required />
        </el-form-item>
        <el-form-item v-if="isRegister" label="邮箱">
          <el-input v-model="form.email" placeholder="选填" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" required show-password />
        </el-form-item>
        <el-button type="primary" native-type="submit" :loading="loading" style="width:100%">
          {{ isRegister ? '注册' : '登录' }}
        </el-button>
      </el-form>
      <p style="text-align:center;margin-top:16px">
        <el-button link @click="isRegister = !isRegister">
          {{ isRegister ? '已有账号？去登录' : '没有账号？去注册' }}
        </el-button>
      </p>
    </el-card>
  </div>
</template>

<style scoped>
.login-container { display: flex; justify-content: center; align-items: center; min-height: 80vh; }
.login-card { width: 100%; max-width: 400px; }
</style>
