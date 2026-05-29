<template>
  <div v-if="isAI" class="flex items-start gap-3 mb-4">
    <div class="flex-shrink-0">
      <img
        v-if="!avatarError"
        :src="aiAvatar"
        :alt="aiNickname"
        class="w-10 h-10 rounded-full object-cover"
        @error="avatarError = true"
      />
      <div
        v-else
        class="w-10 h-10 rounded-full flex items-center justify-center text-white text-sm font-medium bg-purple-500"
      >
        {{ aiNickname?.charAt(0) || 'R' }}
      </div>
    </div>
    <div class="flex flex-col max-w-[70%]">
      <span class="text-xs text-gray-400 mb-1 px-1">{{ aiNickname }}</span>
      <div class="chat-bubble-ai"><template v-for="(seg, idx) in parsedContent" :key="idx"><span v-if="seg.type === 'text'">{{ seg.text }}</span><em v-else class="emotional-tag">{{ seg.text }}</em></template></div>
      <span v-if="message?.created_at" class="text-xs text-gray-400 mt-1 px-1">
        {{ formatTime(message.created_at) }}
      </span>
    </div>
  </div>

  <div v-else class="flex items-start gap-3 mb-4" style="width: 100%; justify-content: flex-end;">
    <div class="flex flex-col max-w-[70%]">
      <span class="text-xs text-gray-400 mb-1 px-1" style="text-align: right;">我</span>
      <div class="chat-bubble-user" style="display: inline-block; max-width: 100%;">{{ message?.content }}</div>
      <span v-if="message?.created_at" class="text-xs text-gray-400 mt-1 px-1" style="text-align: right;">
        {{ formatTime(message.created_at) }}
      </span>
    </div>
    <div class="flex-shrink-0">
      <img
        v-if="!avatarError"
        :src="userAvatar"
        :alt="userNickname"
        class="w-10 h-10 rounded-full object-cover"
        @error="avatarError = true"
      />
      <div
        v-else
        class="w-10 h-10 rounded-full flex items-center justify-center text-white text-sm font-medium bg-wechat-green"
      >
        {{ userNickname?.charAt(0) || '我' }}
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { formatTime } from '../utils'
import type { Message } from '../types/api'

interface ContentSegment {
  type: 'text' | 'emotional'
  text: string
}

const props = defineProps<{
  message: Message
  isAI: boolean
  aiNickname: string
  aiAvatar: string
  userNickname: string
  userAvatar: string
}>()

const emotionalRegex = /<emotional>([\s\S]*?)<\/emotional>/g

const parsedContent = computed(() => {
  const content = props.message?.content || ''
  const segments: ContentSegment[] = []
  let lastIndex = 0
  let match: RegExpExecArray | null = null

  while ((match = emotionalRegex.exec(content)) !== null) {
    if (match.index > lastIndex) {
      segments.push({
        type: 'text',
        text: content.slice(lastIndex, match.index),
      })
    }
    segments.push({
      type: 'emotional',
      text: match[1],
    })
    lastIndex = match.index + match[0].length
  }

  if (lastIndex < content.length) {
    segments.push({
      type: 'text',
      text: content.slice(lastIndex),
    })
  }

  return segments.length > 0 ? segments : [{ type: 'text', text: content }]
})

const avatarError = ref(false)

watch(() => props.aiAvatar, () => {
  avatarError.value = false
})
</script>

<style scoped>
.emotional-tag {
  font-style: italic;
  opacity: 0.65;
  font-size: 0.9em;
}
</style>
