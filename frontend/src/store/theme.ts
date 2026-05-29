import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const isDark = ref(localStorage.getItem('darkMode') === 'true')

  const themeClass = computed(() => isDark.value ? 'dark' : '')

  function toggleTheme() {
    isDark.value = !isDark.value
    localStorage.setItem('darkMode', isDark.value.toString())
    applyTheme()
  }

  function applyTheme() {
    if (isDark.value) {
      document.documentElement.classList.add('dark')
    } else {
      document.documentElement.classList.remove('dark')
    }
  }

  function initTheme() {
    applyTheme()
  }

  return {
    isDark,
    themeClass,
    toggleTheme,
    applyTheme,
    initTheme,
  }
})
