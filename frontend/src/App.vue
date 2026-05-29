<template>
  <div :class="['h-full', themeStore.themeClass]">
    <div class="h-full bg-wechat-bg dark:bg-wechat-bg-dark text-wechat-text dark:text-wechat-text-dark transition-colors duration-200 safe-area">
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
.safe-area {
  padding-top: env(safe-area-inset-top, 0px);
  padding-bottom: env(safe-area-inset-bottom, 0px);
  padding-left: env(safe-area-inset-left, 0px);
  padding-right: env(safe-area-inset-right, 0px);
}
</style>
