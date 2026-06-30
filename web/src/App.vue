<script setup>
import { useAuthStore } from './stores/auth'

const auth = useAuthStore()
</script>

<template>
  <div id="app-container">
    <el-container>
      <el-header v-if="auth.isLoggedIn" class="app-header">
        <div class="header-left">
          <h2 @click="$router.push('/dashboard')" style="cursor:pointer">🔗 URL Shortener</h2>
        </div>
        <div class="header-right">
          <el-button text @click="$router.push('/dashboard')">看板</el-button>
          <el-button text @click="$router.push('/links')">链接</el-button>
          <el-button text @click="$router.push('/links/create')">创建</el-button>
          <el-dropdown>
            <span class="user-info">{{ auth.username }}</span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="auth.logout()">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<style>
body { margin: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; }
#app-container { min-height: 100vh; background: #f5f7fa; }
.app-header {
  display: flex; justify-content: space-between; align-items: center;
  background: #fff; border-bottom: 1px solid #e4e7ed; padding: 0 24px; height: 60px;
}
.app-header h2 { margin: 0; color: #303133; font-size: 18px; }
.header-right { display: flex; align-items: center; gap: 8px; }
.user-info { cursor: pointer; color: #409eff; }
</style>
