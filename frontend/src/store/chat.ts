import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { conversationAPI, personaAPI } from '../api'
import type { Conversation, Message, Persona } from '../types/api'
import {
  getLocalMessages,
  saveLocalMessages,
  appendLocalMessage,
  clearLocalMessages,
} from '../utils/storage'

export const useChatStore = defineStore('chat', () => {
  const conversations = ref<Conversation[]>([])
  const currentConversationId = ref<number | null>(null)
  const messages = ref<Message[]>([])
  const isLoading = ref(false)
  const isStreaming = ref(false)
  const streamingContent = ref('')
  const totalMessages = ref(0)
  const ws = ref<WebSocket | null>(null)
  const wsConnected = ref(false)
  const currentPersona = ref<Persona | null>(null)
  const liveAiAvatar = ref('')
  const liveAiNickname = ref('')

  const currentConversation = computed<Conversation | null>(() => {
    return conversations.value.find(c => c.id === currentConversationId.value) || null
  })

  const aiNickname = computed(() => liveAiNickname.value || currentConversation.value?.ai_nickname || 'RainYi')
  const aiAvatar = computed(() => {
    return liveAiAvatar.value || currentConversation.value?.ai_avatar || '/static/default-avatar.svg'
  })

  const currentPersonaName = computed(() => {
    return currentPersona.value?.name || '默认'
  })

  async function fetchConversations() {
    const res = await conversationAPI.getConversations()
    conversations.value = res.conversations
    return res
  }

  async function selectConversation(id: number) {
    currentConversationId.value = id
    liveAiAvatar.value = ''
    liveAiNickname.value = ''
    messages.value = []
    totalMessages.value = 0
    await loadMessages(id)
    await fetchCurrentPersona()
  }

  async function loadMessages(convId: number, limit = 50, offset = 0) {
    if (!convId) return

    const cached = await getLocalMessages(convId)
    if (cached.length > 0 && offset === 0) {
      messages.value = cached
      totalMessages.value = cached.length
      return
    }

    isLoading.value = true
    try {
      const res = await conversationAPI.getMessages(convId, limit, offset)
      if (offset === 0) {
        messages.value = res.messages || []
        if (res.messages?.length) {
          saveLocalMessages(convId, res.messages)
        }
      } else {
        const merged = [...(res.messages || []), ...messages.value]
        messages.value = merged
      }
      totalMessages.value = res.total || 0
    } finally {
      isLoading.value = false
    }
  }

  async function loadMoreMessages() {
    if (messages.value.length >= totalMessages.value) return
    const offset = messages.value.length
    const convId = currentConversationId.value
    if (!convId) return
    isLoading.value = true
    try {
      const res = await conversationAPI.getMessages(convId, 50, offset)
      messages.value = [...(res.messages || []), ...messages.value]
    } finally {
      isLoading.value = false
    }
  }

  async function clearMessages(convId?: number) {
    const id = convId || currentConversationId.value
    if (!id) return
    await conversationAPI.clearMessages(id)
    messages.value = []
    totalMessages.value = 0
    clearLocalMessages(id)
  }

  async function updateConversationConfig(convId: number, config: Record<string, unknown>) {
    const res = await conversationAPI.updateConfig(convId, config)
    const idx = conversations.value.findIndex(c => c.id === convId)
    if (idx !== -1) {
      conversations.value[idx] = res.conversation
    }
    if (config.ai_avatar) {
      liveAiAvatar.value = config.ai_avatar as string
    }
    if (config.ai_nickname) {
      liveAiNickname.value = config.ai_nickname as string
    }
    return res
  }

  async function fetchCurrentPersona() {
    const convId = currentConversationId.value
    if (!convId) {
      currentPersona.value = null
      return
    }
    try {
      const res = await personaAPI.getConversationPersona(convId)
      currentPersona.value = res.persona || null
    } catch {
      currentPersona.value = null
    }
  }

  async function setConversationPersona(personaId: number | null) {
    const convId = currentConversationId.value
    if (!convId) return
    const personaIdVal = personaId || null
    const res = await personaAPI.setConversationPersona(convId, personaIdVal)
    const idx = conversations.value.findIndex(c => c.id === convId)
    if (idx !== -1) {
      conversations.value[idx] = res.conversation
    }
    currentPersona.value = personaIdVal ? await personaAPI.getPersona(personaIdVal).then(r => r.persona) : null
  }

  async function createConversation(title = '情感陪伴') {
    const res = await conversationAPI.createConversation(title)
    conversations.value.unshift(res.conversation)
    return res.conversation
  }

  function connectWebSocket() {
    const token = localStorage.getItem('token')
    if (!token) return

    const wsUrl = import.meta.env.VITE_WS_URL
    if (wsUrl) {
      ws.value = new WebSocket(`${wsUrl}?token=${token}`)
    } else {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      const host = window.location.host
      ws.value = new WebSocket(`${protocol}//${host}/api/ws/chat?token=${token}`)
    }

    ws.value.onopen = () => {
      wsConnected.value = true
    }

    ws.value.onmessage = (event: MessageEvent) => {
      try {
        const data = JSON.parse(event.data) as { type: string; content?: string }

        switch (data.type) {
          case 'ai_start':
            isStreaming.value = true
            streamingContent.value = ''
            break
          case 'stream':
            streamingContent.value += data.content
            break
          case 'complete':
            isStreaming.value = false
            const newMsg: Message = {
              id: Date.now(),
              conversation_id: currentConversationId.value!,
              role: 'assistant',
              content: streamingContent.value,
              created_at: new Date().toISOString(),
              is_deleted: false,
            }
            messages.value.push(newMsg)
            if (currentConversationId.value) {
              appendLocalMessage(currentConversationId.value, newMsg)
            }
            streamingContent.value = ''
            break
          case 'error':
            isStreaming.value = false
            streamingContent.value = ''
            console.error('AI error:', data.content)
            break
        }
      } catch (e) {
        console.error('WS parse error:', e)
      }
    }

    ws.value.onclose = () => {
      wsConnected.value = false
      isStreaming.value = false
    }

    ws.value.onerror = (err: Event) => {
      console.error('WS error:', err)
      wsConnected.value = false
    }
  }

  function sendMessage(content: string) {
    if (!ws.value || ws.value.readyState !== WebSocket.OPEN) {
      console.error('WebSocket not connected')
      return
    }

    if (!currentConversationId.value) return

    const userMsg: Message = {
      id: Date.now(),
      conversation_id: currentConversationId.value,
      role: 'user',
      content: content,
      created_at: new Date().toISOString(),
      is_deleted: false,
    }

    messages.value.push(userMsg)
    appendLocalMessage(currentConversationId.value, userMsg)

    ws.value.send(JSON.stringify({
      conversation_id: currentConversationId.value,
      content: content,
    }))
  }

  function disconnectWebSocket() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    wsConnected.value = false
  }

  return {
    conversations,
    currentConversationId,
    messages,
    isLoading,
    isStreaming,
    streamingContent,
    totalMessages,
    wsConnected,
    currentConversation,
    aiNickname,
    aiAvatar,
    currentPersona,
    currentPersonaName,
    fetchConversations,
    selectConversation,
    loadMessages,
    loadMoreMessages,
    clearMessages,
    updateConversationConfig,
    fetchCurrentPersona,
    setConversationPersona,
    createConversation,
    connectWebSocket,
    sendMessage,
    disconnectWebSocket,
  }
})
