<template>
  <div v-if="time" class="flex justify-center mb-4">
    <span class="text-xs text-gray-400 bg-gray-100 dark:bg-gray-800 px-3 py-1 rounded-full">
      {{ formatTime(time) }}
    </span>
  </div>
</template>

<script setup lang="ts">
import { formatTime } from '../utils'

defineProps<{
  time: string
}>()
</script>

<style scoped>
</style>
