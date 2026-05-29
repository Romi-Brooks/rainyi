<template>
  <div class="h-screen flex flex-col bg-gray-50 dark:bg-gray-900">
    <header class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-4 py-3 flex items-center gap-3 shrink-0">
      <button
        class="w-8 h-8 flex items-center justify-center rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors text-gray-600 dark:text-gray-300"
        @click="goBack"
      >
        <svg xmlns="http://www.w3.org/2000/svg" class="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M19 12H5"/><polyline points="12 19 5 12 12 5"/>
        </svg>
      </button>
      <h1 class="text-lg font-semibold text-gray-800 dark:text-gray-100">设置</h1>
    </header>

    <div class="flex-1 overflow-y-auto">
      <div class="max-w-2xl mx-auto px-4 py-4 space-y-4">

        <!-- Toast -->
        <Transition name="toast">
          <div v-if="showSuccess"
            class="fixed top-4 left-1/2 -translate-x-1/2 z-50 bg-green-500 text-white text-sm px-4 py-2 rounded-full shadow-lg">
            {{ successMessage }}
          </div>
        </Transition>

        <!-- AI 配置 -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm overflow-hidden">
          <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-700">
            <h2 class="text-base font-semibold text-gray-800 dark:text-gray-100">AI 配置</h2>
            <p class="text-xs text-gray-400 mt-0.5">修改 AI 昵称和头像</p>
          </div>

          <div class="px-5 py-4 space-y-4">
            <div>
              <label class="text-xs font-medium text-gray-500 dark:text-gray-400">AI 头像</label>
              <div class="flex items-center gap-3 mt-2">
                <img v-if="aiAvatarUrl"
                  :src="aiAvatarUrl"
                  class="w-12 h-12 rounded-full object-cover border border-gray-200 dark:border-gray-600"
                  @error="aiAvatarUrl = ''"
                />
                <div v-else class="w-12 h-12 rounded-full bg-purple-100 dark:bg-purple-900/30 flex items-center justify-center text-purple-600 dark:text-purple-400 text-sm font-medium">
                  {{ (aiNickname || 'R').charAt(0) }}
                </div>
                <label class="text-xs text-wechat-green cursor-pointer hover:text-wechat-green-dark transition-colors flex items-center gap-1 border border-wechat-green/30 rounded-lg px-3 py-1.5">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
                  </svg>
                  <span>{{ uploadingAvatarFor === 'ai' ? '上传中...' : '上传头像' }}</span>
                  <input type="file" accept="image/*" class="hidden" :disabled="uploadingAvatarFor !== null" @change="uploadAiAvatar($event)" />
                </label>
              </div>
            </div>

            <div>
              <label class="text-xs font-medium text-gray-500 dark:text-gray-400">AI 昵称</label>
              <div class="flex gap-2 mt-1">
                <input v-model="newNickname"
                  class="flex-1 text-sm border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-750 rounded-lg px-3 py-2 outline-none focus:border-wechat-green dark:text-gray-200"
                  placeholder="输入新昵称..."
                  @keyup.enter="applyNickname"
                />
                <button
                  class="text-sm text-white bg-wechat-green hover:bg-wechat-green-dark rounded-lg px-4 py-2 font-medium transition-colors"
                  @click="applyNickname"
                >应用</button>
              </div>
              <p class="text-xs text-gray-400 mt-1 ml-1">当前昵称: <span class="text-gray-600 dark:text-gray-300">{{ aiNickname || chatStore.currentConversation?.ai_nickname || 'RainYi' }}</span></p>
            </div>

            <div>
              <label class="text-xs font-medium text-gray-500 dark:text-gray-400">当前人格</label>
              <div class="mt-1.5 flex flex-wrap gap-1.5">
                <button
                  v-for="p in personas"
                  :key="p.id"
                  class="text-xs px-2.5 py-1 rounded-full border transition-colors"
                  :class="chatStore.currentConversation?.persona_id === p.id
                    ? 'bg-purple-600 text-white border-purple-600'
                    : 'bg-white dark:bg-gray-750 text-gray-600 dark:text-gray-400 border-gray-200 dark:border-gray-600 hover:border-purple-300 dark:hover:border-purple-500'"
                  @click="switchPersona(p.id)"
                >{{ p.name }}</button>
                <button
                  class="text-xs px-2.5 py-1 rounded-full border transition-colors"
                  :class="!chatStore.currentConversation?.persona_id
                    ? 'bg-purple-600 text-white border-purple-600'
                    : 'bg-white dark:bg-gray-750 text-gray-500 dark:text-gray-500 border-gray-200 dark:border-gray-600 hover:border-gray-300'"
                  @click="switchPersona(null)"
                >默认</button>
              </div>
              <p class="text-xs text-amber-500 dark:text-amber-400 mt-1.5 ml-1 flex items-center gap-1">
                <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
                </svg>
                <span>切换人格可能导致上下文断裂或对话风格不一致，建议在新会话中切换</span>
              </p>
            </div>

            <div class="pt-1">
              <button
                class="w-full text-sm text-orange-500 border border-orange-200 dark:border-orange-800 rounded-lg py-2 hover:bg-orange-50 dark:hover:bg-orange-900/20 transition-colors font-medium"
                @click="clearCurrentChat"
              >清空当前会话</button>
            </div>
          </div>
        </div>

        <!-- 人格管理 -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm overflow-hidden">
          <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-700 flex items-center justify-between">
            <div>
              <h2 class="text-base font-semibold text-gray-800 dark:text-gray-100">人格管理</h2>
              <p class="text-xs text-gray-400 mt-0.5">管理 AI 角色和技能文件</p>
            </div>
            <button
              class="text-xs text-wechat-green border border-wechat-green rounded-lg px-3 py-1.5 hover:bg-wechat-green/5 transition-colors font-medium"
              @click="loadFromDirectory"
            >从本地加载</button>
          </div>

          <div class="divide-y divide-gray-100 dark:divide-gray-700">
            <div v-for="persona in personas" :key="persona.id">
              <div
                class="px-5 py-3 flex items-center justify-between cursor-pointer hover:bg-gray-50 dark:hover:bg-gray-750 transition-colors"
                @click="togglePersonaExpand(persona.id)"
              >
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2">
                    <p class="text-sm font-medium text-gray-800 dark:text-gray-200 truncate">{{ persona.name }}</p>
                    <span v-if="persona.is_built_in" class="inline-block text-xs text-wechat-green bg-wechat-green/10 rounded-full px-2 py-0.5">内置</span>
                  </div>
                  <p class="text-xs text-gray-400 truncate mt-0.5">{{ persona.description || '暂无描述' }}</p>
                  <p class="text-xs text-gray-400 mt-0.5">技能节点: {{ persona.skill_node_count || 0 }}</p>
                </div>
                <div class="flex items-center gap-2 shrink-0">
                  <button v-if="!persona.is_built_in"
                    class="text-xs text-red-400 hover:text-red-600 transition-colors px-2 py-1"
                    @click.stop="deletePersona(persona.id)"
                  >删除</button>
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 text-gray-400" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    :class="expandedPersonaId === persona.id ? 'rotate-180' : ''">
                    <polyline points="6 9 12 15 18 9"/>
                  </svg>
                </div>
              </div>

              <!-- Expanded: skill nodes -->
              <div v-if="expandedPersonaId === persona.id" class="px-5 pb-4 space-y-2">
                <div v-if="skillNodesMap[persona.id]?.length" class="space-y-1.5">
                  <div v-for="sn in skillNodesMap[persona.id]" :key="sn.id"
                    class="bg-gray-50 dark:bg-gray-750 rounded-lg px-3 py-2">
                    <div class="flex items-center justify-between">
                      <div class="flex-1 min-w-0">
                        <p class="text-xs font-medium text-gray-700 dark:text-gray-300 truncate">{{ sn.name }}</p>
                        <p class="text-xs text-gray-400 truncate">{{ sn.file_name }} · 优先级 {{ sn.priority }}</p>
                      </div>
                      <button v-if="!persona.is_built_in"
                        class="text-xs text-red-400 hover:text-red-600 transition-colors shrink-0 ml-2"
                        @click="deleteSkillNode(persona.id, sn.id)"
                      >删除</button>
                    </div>
                    <!-- KVs under this node -->
                    <div v-if="sn.kvs?.length" class="mt-1.5 pl-2 border-l-2 border-gray-200 dark:border-gray-600 space-y-1">
                      <div v-for="kv in sn.kvs" :key="kv.id" class="text-xs">
                        <span class="text-gray-500 dark:text-gray-400 font-medium">{{ kv.key }}: </span>
                        <span class="text-gray-600 dark:text-gray-400">{{ truncateText(kv.value, 60) }}</span>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-else class="text-xs text-gray-400 py-1">暂无技能文件</div>

                <!-- Upload button (only for non-built-in) -->
                <div v-if="!persona.is_built_in" class="flex items-center gap-2 pt-1">
                  <label class="flex-1 flex items-center gap-2 text-xs text-wechat-green cursor-pointer border border-dashed border-wechat-green/30 rounded-lg px-3 py-2 hover:bg-wechat-green/5 transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" class="w-4 h-4 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
                    </svg>
                    <span>{{ uploadingPersonaId === persona.id ? '上传中...' : '上传技能文件 (.md)' }}</span>
                    <input
                      type="file"
                      accept=".md"
                      multiple
                      class="hidden"
                      :disabled="uploadingPersonaId === persona.id"
                      @change="uploadSkillFiles(persona.id, $event)"
                    />
                  </label>
                </div>
              </div>
            </div>
          </div>

          <!-- Create new persona -->
          <div class="px-5 py-4 border-t border-gray-100 dark:border-gray-700">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2">创建新人格</p>
            <div class="flex gap-2 mb-2">
              <input v-model="newPersonaName"
                class="flex-1 text-sm border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-750 rounded-lg px-3 py-2 outline-none focus:border-wechat-green dark:text-gray-200"
                placeholder="人格名称"
              />
              <button
                class="text-sm text-white bg-wechat-green hover:bg-wechat-green-dark rounded-lg px-4 py-2 font-medium transition-colors"
                @click="createPersona"
              >创建</button>
            </div>
            <input v-model="newPersonaDesc"
              class="w-full text-sm border border-gray-200 dark:border-gray-600 bg-gray-50 dark:bg-gray-750 rounded-lg px-3 py-2 outline-none focus:border-wechat-green dark:text-gray-200"
              placeholder="人格描述（可选）"
            />
          </div>
        </div>

        <!-- Feature info -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm overflow-hidden">
          <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-700">
            <h2 class="text-base font-semibold text-gray-800 dark:text-gray-100">功能说明</h2>
            <p class="text-xs text-gray-400 mt-0.5">功能说明</p>
          </div>

          <div class="divide-y divide-gray-100 dark:divide-gray-700">
            <div class="px-5 py-4">
              <div class="flex items-center gap-2 mb-2">
                <p class="text-sm font-medium text-gray-800 dark:text-gray-200">表情发送</p>
                <span class="inline-block text-xs text-wechat-green bg-wechat-green/10 rounded-full px-2 py-0.5">当前支持</span>
              </div>
              <p class="text-xs text-gray-400">功能入口已预留，暂未填充具体表情</p>
              <p class="text-xs text-gray-400 mt-1">后续版本：支持表情配置、自定义表情上传</p>
            </div>

            <div class="px-5 py-4">
              <div class="flex items-center gap-2 mb-2">
                <p class="text-sm font-medium text-gray-800 dark:text-gray-200">语音输入/输出</p>
                <span class="inline-block text-xs text-wechat-green bg-wechat-green/10 rounded-full px-2 py-0.5">当前支持</span>
              </div>
              <p class="text-xs text-gray-400">功能入口已预留</p>
              <p class="text-xs text-gray-400 mt-1">后续版本：接入 TTS 语音输出、语音输入转文字</p>
            </div>

            <div class="px-5 py-4">
              <div class="flex items-center gap-2 mb-2">
                <p class="text-sm font-medium text-gray-800 dark:text-gray-200">关键事件记忆</p>
                <span class="inline-block text-xs text-wechat-green bg-wechat-green/10 rounded-full px-2 py-0.5">当前支持</span>
              </div>
              <p class="text-xs text-gray-400">清空记录后重置上下文</p>
              <p class="text-xs text-gray-400 mt-1">后续版本：清除记录后仍保留核心信息</p>
            </div>

            <div class="px-5 py-4">
              <div class="flex items-center gap-2 mb-2">
                <p class="text-sm font-medium text-gray-800 dark:text-gray-200">选择性清除</p>
                <span class="inline-block text-xs text-gray-400 bg-gray-100 dark:bg-gray-700 rounded-full px-2 py-0.5">暂不支持</span>
              </div>
              <p class="text-xs text-gray-400">暂不支持</p>
              <p class="text-xs text-gray-400 mt-1">后续版本：支持选择性清除聊天记录</p>
            </div>
          </div>
        </div>

        <!-- User info -->
        <div class="bg-white dark:bg-gray-800 rounded-2xl shadow-sm overflow-hidden">
          <div class="px-5 py-4 border-b border-gray-100 dark:border-gray-700">
            <h2 class="text-base font-semibold text-gray-800 dark:text-gray-100">用户信息</h2>
            <p class="text-xs text-gray-400 mt-0.5">用户信息</p>
          </div>

          <div class="px-5 py-4 space-y-3">
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">用户名</span>
              <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ userStore.username }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">邮箱</span>
              <span class="text-sm font-medium text-gray-800 dark:text-gray-200">{{ userStore.userInfo?.email || '未设置' }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-sm text-gray-500 dark:text-gray-400">头像</span>
              <div class="flex items-center gap-2">
                <img :src="userStore.userAvatar" class="w-8 h-8 rounded-full object-cover border border-gray-200 dark:border-gray-600"
                  @error="(userStore as any).userInfo = { ...(userStore.userInfo || {}), avatar: '' }"
                />
                <label class="text-xs text-wechat-green cursor-pointer hover:text-wechat-green-dark transition-colors flex items-center gap-1">
                  <svg xmlns="http://www.w3.org/2000/svg" class="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/>
                  </svg>
                  <span>{{ uploadingAvatarFor === 'user' ? '上传中...' : '更换' }}</span>
                  <input type="file" accept="image/*" class="hidden" :disabled="uploadingAvatarFor !== null" @change="uploadUserAvatar($event)" />
                </label>
              </div>
            </div>
            <div class="pt-2">
              <button
                class="w-full text-sm text-red-500 border border-red-300 dark:border-red-700 rounded-lg py-2.5 hover:bg-red-50 dark:hover:bg-red-900/20 transition-colors font-medium"
                @click="handleLogout"
              >退出登录</button>
            </div>
          </div>
        </div>

        <div class="h-6" />
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useChatStore } from '../store/chat'
import { useUserStore } from '../store/user'
import { useThemeStore } from '../store/theme'
import { personaAPI } from '../api'
import type { PersonaFromServer, SkillNode } from '../types/api'

const router = useRouter()
const chatStore = useChatStore()
const userStore = useUserStore()
const themeStore = useThemeStore()

const newNickname = ref('')
const showSuccess = ref(false)
const successMessage = ref('')
const personas = ref<PersonaFromServer[]>([])
const skillNodesMap = ref<Record<number, SkillNode[]>>({})
const expandedPersonaId = ref<number | null>(null)
const newPersonaName = ref('')
const newPersonaDesc = ref('')
const uploadingPersonaId = ref<number | null>(null)
const uploadingAvatarFor = ref<string | null>(null)
const aiAvatarUrl = ref('')
const aiNickname = ref('')

const currentConversation = computed(() => chatStore.currentConversation)

function goBack() {
  router.back()
}

function showToast(msg: string) {
  successMessage.value = msg
  showSuccess.value = true
  setTimeout(() => {
    showSuccess.value = false
  }, 2000)
}

function truncateText(text: string, maxLen: number): string {
  if (!text) return ''
  return text.length > maxLen ? text.substring(0, maxLen) + '...' : text
}

async function applyNickname() {
  const name = newNickname.value.trim()
  if (!name || !chatStore.currentConversationId) return
  try {
    await chatStore.updateConversationConfig(chatStore.currentConversationId, { ai_nickname: name })
    aiNickname.value = name
    newNickname.value = ''
    showToast('昵称已更新')
  } catch (e) {
    showToast('更新失败，请重试')
  }
}


async function clearCurrentChat() {
  if (!chatStore.currentConversationId) return
  const confirmed = window.confirm('确定要清空当前会话记录吗？此操作不可恢复。')
  if (!confirmed) return
  try {
    await chatStore.clearMessages(chatStore.currentConversationId)
    showToast('会话记录已清空')
  } catch (e) {
    showToast('清空失败，请重试')
  }
}

async function fetchPersonas() {
  try {
    const res = await personaAPI.getPersonas()
    personas.value = res.personas || []
  } catch (e) {
    console.error('获取人格列表失败:', e)
  }
}

async function fetchSkillNodes(personaId: number) {
  try {
    const res = await personaAPI.getPersona(personaId)
    skillNodesMap.value[personaId] = res.skill_nodes || []
  } catch (e) {
    console.error('获取技能节点失败:', e)
  }
}

function togglePersonaExpand(personaId: number) {
  if (expandedPersonaId.value === personaId) {
    expandedPersonaId.value = null
    return
  }
  expandedPersonaId.value = personaId
  if (!skillNodesMap.value[personaId]) {
    fetchSkillNodes(personaId)
  }
}

async function createPersona() {
  const name = newPersonaName.value.trim()
  if (!name) return
  try {
    await personaAPI.createPersona({ name, description: newPersonaDesc.value.trim() })
    newPersonaName.value = ''
    newPersonaDesc.value = ''
    showToast('人格创建成功')
    await fetchPersonas()
  } catch (e) {
    showToast('创建失败: ' + (e as Error).message)
  }
}

async function deletePersona(id: number) {
  const confirmed = window.confirm('确定要删除该人格及其所有技能节点吗？')
  if (!confirmed) return
  try {
    await personaAPI.deletePersona(id)
    showToast('人格已删除')
    await fetchPersonas()
  } catch (e) {
    showToast('删除失败: ' + (e as Error).message)
  }
}

async function uploadSkillFiles(personaId: number, event: Event) {
  const input = event.target as HTMLInputElement
  const files = input.files
  if (!files || files.length === 0) return

  uploadingPersonaId.value = personaId
  try {
    const fileArray = Array.from(files)
    const res = await personaAPI.uploadSkillFile(personaId, fileArray)
    const uploaded = res.uploaded || []
    const errs = res.errors || []
    const parts: string[] = [`成功上传 ${uploaded.length} 个文件`]
    if (errs.length > 0) {
      parts.push(`，${errs.length} 个失败`)
    }
    showToast(parts.join(''))
    await fetchSkillNodes(personaId)
  } catch (e) {
    showToast('上传失败: ' + (e as Error).message)
  } finally {
    uploadingPersonaId.value = null
  }
  input.value = ''
}

async function deleteSkillNode(personaId: number, nodeId: number) {
  const confirmed = window.confirm('确定要删除该技能节点吗？')
  if (!confirmed) return
  try {
    await personaAPI.deleteSkillFile(personaId, nodeId)
    showToast('技能节点已删除')
    await fetchSkillNodes(personaId)
  } catch (e) {
    showToast('删除失败: ' + (e as Error).message)
  }
}

async function switchPersona(personaId: number | null) {
  if (!chatStore.currentConversationId) {
    showToast('请先选择一个会话')
    return
  }
  try {
    await chatStore.setConversationPersona(personaId)
    showToast('人格已切换')
  } catch (e) {
    showToast('切换失败: ' + (e as Error).message)
  }
}

async function loadFromDirectory() {
  try {
    const res = await personaAPI.loadFromDirectory()
    showToast(res.message || '加载成功')
    await fetchPersonas()
  } catch (e) {
    showToast('加载失败: ' + (e as Error).message)
  }
}

async function uploadAiAvatar(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  if (!chatStore.currentConversationId) {
    showToast('请先选择一个会话')
    return
  }
  if (file.size > 1024 * 1024) {
    showToast('图片大小不能超过 1MB')
    input.value = ''
    return
  }
  uploadingAvatarFor.value = 'ai'
  try {
    const res = await personaAPI.uploadAvatar(file)
    await chatStore.updateConversationConfig(chatStore.currentConversationId, { ai_avatar: res.file.url })
    aiAvatarUrl.value = res.file.url
    showToast('AI 头像已更新')
  } catch (e) {
    showToast('上传失败: ' + (e as Error).message)
  } finally {
    uploadingAvatarFor.value = null
  }
  input.value = ''
}

async function uploadUserAvatar(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  uploadingAvatarFor.value = 'user'
  try {
    await personaAPI.uploadAvatar(file)
    await userStore.fetchProfile()
    showToast('头像已更新')
  } catch (e) {
    showToast('上传失败: ' + (e as Error).message)
  } finally {
    uploadingAvatarFor.value = null
  }
  input.value = ''
}

function handleLogout() {
  userStore.logout()
  router.push('/login')
}

onMounted(() => {
  fetchPersonas()
  if (chatStore.currentConversation?.ai_avatar) {
    aiAvatarUrl.value = chatStore.currentConversation.ai_avatar
  }
  if (chatStore.currentConversation?.ai_nickname) {
    aiNickname.value = chatStore.currentConversation.ai_nickname
  }
})
</script>
