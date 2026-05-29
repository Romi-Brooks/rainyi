<template>
  <div class="min-h-screen bg-gray-100 dark:bg-wechat-bg-dark flex items-center justify-center p-4">
    <div class="w-full max-w-md">
      <div class="text-center mb-8">
        <h1 class="text-4xl font-bold text-wechat-green mb-2">RainYi</h1>
        <p class="text-sm text-gray-500 dark:text-wechat-text-secondary-dark">你的专属情感陪伴机器人</p>
      </div>

      <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-lg p-6 sm:p-8">
        <div class="flex mb-6 bg-gray-100 dark:bg-gray-700 rounded-xl p-1">
          <button
            class="flex-1 py-2 text-sm font-medium rounded-lg transition-all"
            :class="activeTab === 'login' ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
            @click="activeTab = 'login'"
          >
            登录
          </button>
          <button
            class="flex-1 py-2 text-sm font-medium rounded-lg transition-all"
            :class="activeTab === 'register' ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm' : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'"
            @click="activeTab = 'register'"
          >
            注册
          </button>
        </div>

        <form v-if="activeTab === 'login'" @submit.prevent="handleLogin">
          <div class="space-y-4">
            <div>
              <input
                ref="loginEmailRef"
                v-model="loginForm.email"
                type="email"
                placeholder="邮箱"
                class="input-field w-full"
                :class="{ 'border-red-400 focus:border-red-400': loginErrors.email }"
                @input="loginErrors.email = ''"
              />
              <p v-if="loginErrors.email" class="text-red-500 text-xs mt-1 ml-1">{{ loginErrors.email }}</p>
            </div>
            <div>
              <input
                v-model="loginForm.password"
                type="password"
                placeholder="密码"
                class="input-field w-full"
                :class="{ 'border-red-400 focus:border-red-400': loginErrors.password }"
                @input="loginErrors.password = ''"
              />
              <p v-if="loginErrors.password" class="text-red-500 text-xs mt-1 ml-1">{{ loginErrors.password }}</p>
            </div>
          </div>

          <p v-if="loginError" class="text-red-500 text-sm text-center mt-4">{{ loginError }}</p>

          <button
            type="submit"
            class="btn-primary w-full mt-6 py-3"
            :disabled="loginLoading"
          >
            <span v-if="loginLoading" class="flex items-center justify-center gap-2">
              <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
              登录中...
            </span>
            <span v-else>登录</span>
          </button>
        </form>

        <form v-if="activeTab === 'register'" @submit.prevent="handleRegister">
          <div class="space-y-4">
            <div>
              <input
                ref="registerUsernameRef"
                v-model="registerForm.username"
                type="text"
                placeholder="用户名"
                class="input-field w-full"
                :class="{ 'border-red-400 focus:border-red-400': registerErrors.username }"
                @input="registerErrors.username = ''"
              />
              <p v-if="registerErrors.username" class="text-red-500 text-xs mt-1 ml-1">{{ registerErrors.username }}</p>
            </div>
            <div>
              <input
                v-model="registerForm.email"
                type="email"
                placeholder="邮箱"
                class="input-field w-full"
                :class="{ 'border-red-400 focus:border-red-400': registerErrors.email }"
                @input="registerErrors.email = ''"
              />
              <p v-if="registerErrors.email" class="text-red-500 text-xs mt-1 ml-1">{{ registerErrors.email }}</p>
            </div>
            <div>
              <input
                v-model="registerForm.password"
                type="password"
                placeholder="密码"
                class="input-field w-full"
                :class="{ 'border-red-400 focus:border-red-400': registerErrors.password }"
                @input="registerErrors.password = ''"
              />
              <p v-if="registerErrors.password" class="text-red-500 text-xs mt-1 ml-1">{{ registerErrors.password }}</p>
            </div>
          </div>

          <p v-if="registerError" class="text-red-500 text-sm text-center mt-4">{{ registerError }}</p>

          <button
            type="submit"
            class="btn-primary w-full mt-6 py-3"
            :disabled="registerLoading"
          >
            <span v-if="registerLoading" class="flex items-center justify-center gap-2">
              <span class="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></span>
              注册中...
            </span>
            <span v-else>注册</span>
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../store/user'

const router = useRouter()
const userStore = useUserStore()

const activeTab = ref('login')

const loginEmailRef = ref<HTMLInputElement | null>(null)
const registerUsernameRef = ref<HTMLInputElement | null>(null)

const loginForm = ref({ email: '', password: '' })
const loginErrors = ref({ email: '', password: '' })
const loginError = ref('')
const loginLoading = ref(false)

const registerForm = ref({ username: '', email: '', password: '' })
const registerErrors = ref({ username: '', email: '', password: '' })
const registerError = ref('')
const registerLoading = ref(false)

function validateEmail(email: string) {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)
}

function validateLogin() {
  let valid = true
  if (!loginForm.value.email) {
    loginErrors.value.email = '请输入邮箱'
    valid = false
  } else if (!validateEmail(loginForm.value.email)) {
    loginErrors.value.email = '邮箱格式不正确'
    valid = false
  }
  if (!loginForm.value.password) {
    loginErrors.value.password = '请输入密码'
    valid = false
  }
  return valid
}

function validateRegister() {
  let valid = true
  if (!registerForm.value.username) {
    registerErrors.value.username = '请输入用户名'
    valid = false
  }
  if (!registerForm.value.email) {
    registerErrors.value.email = '请输入邮箱'
    valid = false
  } else if (!validateEmail(registerForm.value.email)) {
    registerErrors.value.email = '邮箱格式不正确'
    valid = false
  }
  if (!registerForm.value.password) {
    registerErrors.value.password = '请输入密码'
    valid = false
  }
  return valid
}

async function handleLogin() {
  loginError.value = ''
  if (!validateLogin()) return

  loginLoading.value = true
  try {
    await userStore.login(loginForm.value.email, loginForm.value.password)
    router.push('/chat')
  } catch (err) {
    loginError.value = (err as Error).message || '登录失败，请检查邮箱和密码'
  } finally {
    loginLoading.value = false
  }
}

async function handleRegister() {
  registerError.value = ''
  if (!validateRegister()) return

  registerLoading.value = true
  try {
    await userStore.register(registerForm.value.username, registerForm.value.email, registerForm.value.password)
    router.push('/chat')
  } catch (err) {
    registerError.value = (err as Error).message || '注册失败，请稍后重试'
  } finally {
    registerLoading.value = false
  }
}
</script>

<style scoped>
</style>
