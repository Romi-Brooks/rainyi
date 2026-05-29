<template>
  <div
    v-if="visible"
    class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50"
    @click.self="emit('close')"
  >
    <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-xl w-80 p-6">
      <h3 class="text-lg font-semibold text-center text-gray-800 dark:text-gray-200 mb-4">
        选择表情
      </h3>
      <p class="text-sm text-gray-500 dark:text-gray-400 text-center mb-6">
        表情功能正在开发中，后续版本将支持
      </p>
      <div class="flex justify-center">
        <button
          class="btn-primary px-8"
          @click="emit('close')"
        >
          关闭
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>

<style scoped>
</style>
