<template>
  <div class="h-screen flex overflow-hidden bg-white dark:bg-wechat-bg-dark">
    <div
      v-if="isMobile && showSidebar"
      class="fixed inset-0 bg-black bg-opacity-50 z-10 lg:hidden"
      @click="showSidebar = false"
    />

    <aside
      class="flex flex-col bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-700 flex-shrink-0 z-20"
      :class="isMobile ? (showSidebar ? 'fixed inset-y-0 left-0 w-full' : 'hidden') : 'w-80'"
    >
      <div class="flex items-center justify-between px-5 py-4 border-b border-gray-200 dark:border-gray-700">
        <h1 class="text-xl font-bold text-gray-800 dark:text-gray-100">RainYi</h1>
        <button
          class="p-2 rounded-lg text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="router.push('/settings')"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
        </button>
      </div>

      <div class="px-4 py-3">
        <div class="relative">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"/>
            <line x1="21" y1="21" x2="16.65" y2="16.65"/>
          </svg>
          <input
            v-model="searchQuery"
            type="text"
            placeholder="搜索会话"
            class="w-full pl-10 pr-4 py-2 text-sm bg-gray-100 dark:bg-gray-800 text-gray-700 dark:text-gray-300 rounded-xl border-none outline-none focus:ring-2 focus:ring-wechat-green/30 transition-all"
          />
        </div>
      </div>

      <div class="flex-1 overflow-y-auto">
        <div
          v-for="conv in filteredConversations"
          :key="conv.id"
          class="flex items-center gap-3 px-4 py-3 cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors border-b border-gray-100 dark:border-gray-800"
          :class="conv.id === chatStore.currentConversationId ? 'bg-gray-100 dark:bg-gray-800' : ''"
          @click="selectConversation(conv)"
        >
          <div class="flex-shrink-0 w-12 h-12 rounded-full overflow-hidden">
            <img
              v-if="conv.ai_avatar"
              :src="conv.ai_avatar"
              alt="avatar"
              class="w-full h-full object-cover"
              @error="(conv as any).ai_avatar = ''"
            />
            <div
              v-else
              class="w-full h-full bg-wechat-green bg-opacity-10 flex items-center justify-center text-wechat-green text-lg font-medium"
            >
              {{ (conv.title || 'R').charAt(0) }}
            </div>
          </div>
          <div class="flex-1 min-w-0">
            <div class="flex items-center justify-between">
              <h3 class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">{{ conv.ai_nickname || conv.title || '会话' }}</h3>
              <span class="text-xs text-gray-400 flex-shrink-0 ml-2">{{ formatConversationTime(conv.updated_at) }}</span>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 truncate">
              {{ conv.last_message?.content ? truncate(conv.last_message.content, 40) : '暂无消息' }}
            </p>
          </div>
        </div>

        <div v-if="filteredConversations.length === 0" class="flex flex-col items-center justify-center py-16 text-gray-400 dark:text-gray-500">
          <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 mb-3" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
          </svg>
          <p class="text-sm">{{ searchQuery ? '没有找到匹配的会话' : '暂无会话，开始一段新对话吧' }}</p>
        </div>
      </div>

      <div class="px-4 py-3 border-t border-gray-200 dark:border-gray-700">
        <button
          class="w-full flex items-center justify-center gap-2 py-2.5 text-sm font-medium text-wechat-green hover:bg-wechat-green/5 rounded-xl transition-colors border border-wechat-green/30"
          @click="handleNewConversation"
        >
          <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <line x1="12" y1="5" x2="12" y2="19"/>
            <line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          新建会话
        </button>
      </div>
    </aside>

    <main
      class="flex-1 flex flex-col min-w-0"
      :class="isMobile && showSidebar ? 'hidden' : ''"
    >
      <template v-if="chatStore.currentConversationId">
        <div class="flex items-center justify-between px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-900">
          <div class="flex items-center gap-3">
            <button
              v-if="isMobile"
              class="p-1 rounded-lg text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
              @click="showSidebar = true"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-6 h-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="19" y1="12" x2="5" y2="12"/>
                <polyline points="12 19 5 12 12 5"/>
              </svg>
            </button>
            <div class="flex items-center gap-3">
              <div class="w-9 h-9 rounded-full overflow-hidden flex-shrink-0">
                <img
                  v-if="!aiAvatarError"
                  :src="chatStore.aiAvatar"
                  :alt="chatStore.aiNickname"
                  class="w-full h-full object-cover"
                  @error="aiAvatarError = true"
                />
                <div
                  v-else
                  class="w-full h-full bg-purple-500 flex items-center justify-center text-white text-sm font-medium"
                >
                  {{ chatStore.aiNickname.charAt(0) }}
                </div>
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <h2 class="text-base font-semibold text-gray-800 dark:text-gray-100">{{ chatStore.aiNickname }}</h2>
                  <span class="text-xs px-2 py-0.5 rounded-full bg-purple-100 text-purple-600 dark:bg-purple-900/30 dark:text-purple-400">
                    {{ chatStore.currentPersonaName }}
                  </span>
                </div>
                <p class="text-xs text-gray-400">{{ chatStore.wsConnected ? '在线' : '连接中...' }}</p>
              </div>
            </div>
          </div>
          <button
            class="p-2 rounded-lg text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            @click="router.push('/settings')"
          >
            <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
          </button>
        </div>

        <div
          ref="messagesContainer"
          class="flex-1 overflow-y-auto px-4 py-4 space-y-1 bg-gray-50 dark:bg-gray-800/50"
        >
          <div ref="loadMoreSentinel" class="h-4 flex items-center justify-center">
            <span v-if="chatStore.isLoading" class="text-xs text-gray-400">加载中...</span>
          </div>

          <template v-for="(msg, idx) in chatStore.messages" :key="idx">
            <TimeStamp
              v-if="shouldShowTimestamp(msg, idx)"
              :time="msg.created_at"
            />
            <ChatBubble
              :message="msg"
              :isAI="msg.role === 'assistant'"
              :aiNickname="chatStore.aiNickname"
              :aiAvatar="chatStore.aiAvatar"
              :userNickname="userStore.username"
              :userAvatar="userStore.userAvatar"
            />
          </template>

          <div v-if="chatStore.isStreaming" class="flex items-start gap-3 mb-4">
            <div class="flex-shrink-0">
              <div class="w-10 h-10 rounded-full bg-purple-500 flex items-center justify-center text-white text-sm font-medium">
                {{ chatStore.aiNickname.charAt(0) }}
              </div>
            </div>
            <div class="flex flex-col">
              <span class="text-xs text-gray-400 mb-1 px-1">{{ chatStore.aiNickname }}</span>
              <div class="bg-white dark:bg-wechat-bubble-ai-dark text-gray-800 dark:text-wechat-text-dark rounded-2xl rounded-bl-sm px-4 py-3 shadow-sm inline-flex items-center gap-1.5">
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
                <span class="typing-dot"></span>
                <span class="text-xs text-gray-400 ml-2">对方正在输入...</span>
              </div>
            </div>
          </div>

          <div ref="bottomSentinel" />
        </div>

        <div class="bg-white dark:bg-gray-900 border-t border-gray-200 dark:border-gray-700 px-4 py-3">
          <div class="flex items-end gap-3">
            <div class="flex items-center gap-1 pb-1">
              <button
                class="flex items-center justify-center w-9 h-9 rounded-lg text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
                @click="showEmojiPicker = true"
              >
                <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <circle cx="12" cy="12" r="10"/>
                  <path d="M8 14s1.5 2 4 2 4-2 4-2"/>
                  <line x1="9" y1="9" x2="9.01" y2="9"/>
                  <line x1="15" y1="9" x2="15.01" y2="9"/>
                </svg>
              </button>
              <VoiceButton />
            </div>

            <div class="flex-1 min-w-0">
              <textarea
                ref="textareaRef"
                v-model="inputContent"
                class="w-full resize-none bg-gray-100 dark:bg-gray-800 text-gray-800 dark:text-gray-200 rounded-xl px-4 py-2.5 text-sm outline-none focus:ring-2 focus:ring-wechat-green/30 transition-all leading-relaxed"
                rows="1"
                placeholder="输入消息..."
                :disabled="chatStore.isStreaming"
                @input="autoGrow"
                @keydown="onKeydown"
              />
            </div>

            <button
              class="flex-shrink-0 flex items-center justify-center w-10 h-10 rounded-xl transition-colors"
              :class="inputContent.trim() && !chatStore.isStreaming ? 'bg-wechat-green hover:bg-wechat-green-dark text-white' : 'bg-gray-200 dark:bg-gray-700 text-gray-400 dark:text-gray-500 cursor-not-allowed'"
              :disabled="!inputContent.trim() || chatStore.isStreaming"
              @click="sendMessage"
            >
              <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <line x1="22" y1="2" x2="11" y2="13"/>
                <polygon points="22 2 15 22 11 13 2 9 22 2"/>
              </svg>
            </button>
          </div>
        </div>
      </template>

      <template v-else>
        <div class="flex-1 flex flex-col items-center justify-center bg-gray-50 dark:bg-gray-800/50 text-gray-400 dark:text-gray-500">
          <div class="w-24 h-24 rounded-full bg-wechat-green/10 flex items-center justify-center mb-6">
            <svg xmlns="http://www.w3.org/2000/svg" class="w-12 h-12 text-wechat-green" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/>
            </svg>
          </div>
          <h3 class="text-lg font-medium text-gray-600 dark:text-gray-400 mb-2">欢迎使用 RainYi</h3>
          <p class="text-sm text-center max-w-xs">选择或创建一个会话，开始你的情感陪伴之旅</p>
        </div>
      </template>
    </main>

    <EmojiPicker
      :visible="showEmojiPicker"
      @close="showEmojiPicker = false"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { useChatStore } from '../store/chat'
import { useUserStore } from '../store/user'
import { useThemeStore } from '../store/theme'
import ChatBubble from '../components/ChatBubble.vue'
import TimeStamp from '../components/TimeStamp.vue'
import EmojiPicker from '../components/EmojiPicker.vue'
import VoiceButton from '../components/VoiceButton.vue'
import { formatTime, formatConversationTime, truncate, sanitizeHtml } from '../utils'
import type { Conversation, Message } from '../types/api'

const router = useRouter()
const chatStore = useChatStore()
const userStore = useUserStore()
const themeStore = useThemeStore()

const searchQuery = ref('')
const inputContent = ref('')
const showSidebar = ref(true)
const showEmojiPicker = ref(false)
const aiAvatarError = ref(false)

const isMobile = ref(window.innerWidth < 1024)

const messagesContainer = ref<HTMLElement | null>(null)
const textareaRef = ref<HTMLTextAreaElement | null>(null)
const loadMoreSentinel = ref<HTMLElement | null>(null)
const bottomSentinel = ref<HTMLElement | null>(null)

let loadMoreObserver: IntersectionObserver | null = null
let isAutoScrolling = false

const filteredConversations = computed(() => {
  if (!searchQuery.value) return chatStore.conversations
  const q = searchQuery.value.toLowerCase()
  return chatStore.conversations.filter(c =>
    (c.title || '').toLowerCase().includes(q)
  )
})

function shouldShowTimestamp(msg: Message, idx: number) {
  if (idx === 0) return true
  const prev = chatStore.messages[idx - 1]
  if (!prev || !prev.created_at || !msg.created_at) return true
  const diff = new Date(msg.created_at).getTime() - new Date(prev.created_at).getTime()
  return diff > 5 * 60 * 1000
}

function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault()
    sendMessage()
  }
}

function sendMessage() {
  const text = inputContent.value.trim()
  if (!text || chatStore.isStreaming) return
  const sanitized = sanitizeHtml(text)
  chatStore.sendMessage(sanitized)
  inputContent.value = ''
  autoGrow()
  scrollToBottom()
}

function autoGrow() {
  const el = textareaRef.value
  if (!el) return
  el.style.height = 'auto'
  const lineHeight = parseInt(getComputedStyle(el).lineHeight) || 20
  const maxHeight = lineHeight * 4 + 8
  el.style.height = Math.min(el.scrollHeight, maxHeight) + 'px'
}

function selectConversation(conv: Conversation) {
  if (conv.id === chatStore.currentConversationId) return
  chatStore.selectConversation(conv.id)
  aiAvatarError.value = false
  if (isMobile.value) {
    showSidebar.value = false
  }
  resetTextarea()
}

async function handleNewConversation() {
  try {
    const conv = await chatStore.createConversation()
    chatStore.selectConversation(conv.id)
    aiAvatarError.value = false
    if (isMobile.value) {
      showSidebar.value = false
    }
    resetTextarea()
  } catch (e) {
    console.error('创建会话失败:', e)
  }
}

function resetTextarea() {
  inputContent.value = ''
  autoGrow()
}

function scrollToBottom() {
  isAutoScrolling = true
  nextTick(() => {
    if (bottomSentinel.value) {
      bottomSentinel.value.scrollIntoView({ behavior: 'smooth' })
    }
    setTimeout(() => {
      isAutoScrolling = false
    }, 300)
  })
}

function checkMobile() {
  const wasMobile = isMobile.value
  isMobile.value = window.innerWidth < 1024
  if (!isMobile.value && wasMobile) {
    showSidebar.value = true
  }
}

watch(() => chatStore.messages.length, () => {
  if (!chatStore.isLoading) {
    scrollToBottom()
  }
})

watch(() => chatStore.streamingContent, () => {
  if (chatStore.isStreaming) {
    scrollToBottom()
  }
})

watch(() => chatStore.isStreaming, (val) => {
  if (!val) {
    nextTick(() => scrollToBottom())
  }
})

onMounted(async () => {
  window.addEventListener('resize', checkMobile)

  try {
    await chatStore.fetchConversations()
  } catch (e) {
    console.error('获取会话列表失败:', e)
  }

  if (chatStore.conversations.length === 0) {
    try {
      const conv = await chatStore.createConversation()
      await chatStore.selectConversation(conv.id)
    } catch (e) {
      console.error('创建默认会话失败:', e)
    }
  } else if (!chatStore.currentConversationId) {
    await chatStore.selectConversation(chatStore.conversations[0].id)
  }

  aiAvatarError.value = false

  chatStore.connectWebSocket()

  nextTick(() => {
    scrollToBottom()
    autoGrow()
  })

  loadMoreObserver = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting && !chatStore.isLoading && chatStore.messages.length > 0) {
      chatStore.loadMoreMessages()
    }
  }, { threshold: 0.1 })

  observeLoadMoreSentinel()
})

function observeLoadMoreSentinel() {
  nextTick(() => {
    if (loadMoreSentinel.value && loadMoreObserver) {
      loadMoreObserver.disconnect()
      loadMoreObserver.observe(loadMoreSentinel.value)
    }
  })
}

watch(() => chatStore.messages.length, () => {
  observeLoadMoreSentinel()
})

watch(() => chatStore.currentConversationId, () => {
  aiAvatarError.value = false
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
  if (loadMoreObserver) {
    loadMoreObserver.disconnect()
    loadMoreObserver = null
  }
  chatStore.disconnectWebSocket()
})
</script>

<style scoped>
textarea::-webkit-scrollbar {
  width: 4px;
}

textarea::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 2px;
}

.dark textarea::-webkit-scrollbar-thumb {
  background: #4a4a4a;
}
</style>
