<template>
  <div v-if="loading" class="flex flex-col items-center justify-center py-8 gap-3">
    <div class="w-8 h-8 border-4 border-gray-200 dark:border-gray-700 border-t-wechat-green rounded-full animate-spin"></div>
    <span class="text-sm text-gray-400">{{ text }}</span>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  loading: boolean
  text: string
}>()
</script>

<style scoped>
</style>
