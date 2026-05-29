<template>
  <div :class="['h-full', themeStore.themeClass]">
    <div class="h-full bg-wechat-bg dark:bg-wechat-bg-dark text-wechat-text dark:text-wechat-text-dark transition-colors duration-200">
      <router-view />
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { useThemeStore } from './store/theme'
import { useUserStore } from './store/user'

const themeStore = useThemeStore()
const userStore = useUserStore()

onMounted(() => {
  themeStore.initTheme()
  userStore.restoreSession()
})
</script>

<style>
html, body, #app {
  height: 100%;
  margin: 0;
  padding: 0;
}
</style>
