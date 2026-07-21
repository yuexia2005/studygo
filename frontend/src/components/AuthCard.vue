<template>
  <div class="auth-card">
    <div class="brand">
      <span class="logo">&#9670;</span>
      <span>视频中心</span>
    </div>
    <div class="divider"></div>

    <!-- Login -->
    <form v-if="mode === 'login'" @submit.prevent="handleLogin" class="auth-form">
      <div class="input-group">
        <label>用户名</label>
        <input v-model="loginUser" placeholder="请输入用户名" />
      </div>
      <div class="input-group">
        <label>密码</label>
        <input v-model="loginPass" type="password" placeholder="请输入密码" />
      </div>
      <button class="btn primary" type="submit" :disabled="loading">
        {{ loading ? '...' : '登录' }}
      </button>
      <div class="switch" @click="mode = 'register'">注册账号 &#8594;</div>
    </form>

    <!-- Register -->
    <form v-else @submit.prevent="handleRegister" class="auth-form">
      <div class="input-group">
        <label>用户名</label>
        <input v-model="regUser" placeholder="请输入用户名" />
      </div>
      <div class="input-group">
        <label>密码</label>
        <input v-model="regPass" type="password" placeholder="密码 (至少6位)" />
      </div>
      <button class="btn primary" type="submit" :disabled="loading">
        {{ loading ? '...' : '注册' }}
      </button>
      <div class="switch" @click="mode = 'login'">&#8592; 返回登录</div>
    </form>

    <p class="error" v-if="error">{{ error }}</p>
  </div>
</template>

<script setup>
import { ref } from 'vue'

const props = defineProps({
  loginFn: Function,
  registerFn: Function,
})

const mode = ref('login')
const loginUser = ref('')
const loginPass = ref('')
const regUser = ref('')
const regPass = ref('')
const loading = ref(false)
const error = ref('')

async function handleLogin() {
  error.value = ''
  if (!loginUser.value || !loginPass.value) { error.value = '请填写完整'; return }
  loading.value = true
  try {
    await props.loginFn(loginUser.value, loginPass.value)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  error.value = ''
  if (!regUser.value || !regPass.value || regPass.value.length < 6) { error.value = '密码至少6位'; return }
  loading.value = true
  try {
    await props.registerFn(regUser.value, regPass.value)
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>
