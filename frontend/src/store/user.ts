import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authAPI, userAPI } from '../api'
import type { User } from '../types/api'

export const useUserStore = defineStore('user', () => {
  const token = ref(localStorage.getItem('token') || '')
  const userInfo = ref<User | null>(null)

  const isLoggedIn = computed(() => !!token.value)
  const userId = computed(() => userInfo.value?.id || null)
  const username = computed(() => userInfo.value?.username || '用户')
  const userAvatar = computed(() => userInfo.value?.avatar || '/static/default-avatar.svg')

  async function login(email: string, password: string) {
    const res = await authAPI.login(email, password)
    token.value = res.token
    userInfo.value = res.user
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
    return res
  }

  async function register(username: string, email: string, password: string) {
    const res = await authAPI.register(username, email, password)
    token.value = res.token
    userInfo.value = res.user
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
    return res
  }

  async function fetchProfile() {
    const res = await userAPI.getProfile()
    userInfo.value = res.user
    localStorage.setItem('user', JSON.stringify(res.user))
    return res
  }

  async function updateProfile(data: Record<string, unknown>) {
    const res = await userAPI.updateProfile(data)
    userInfo.value = res.user
    localStorage.setItem('user', JSON.stringify(res.user))
    return res
  }

  function logout() {
    token.value = ''
    userInfo.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('user')
  }

  function restoreSession() {
    const saved = localStorage.getItem('user')
    if (saved) {
      try {
        userInfo.value = JSON.parse(saved)
      } catch {
        userInfo.value = null
      }
    }
  }

  return {
    token,
    userInfo,
    isLoggedIn,
    userId,
    username,
    userAvatar,
    login,
    register,
    fetchProfile,
    updateProfile,
    logout,
    restoreSession,
  }
})
