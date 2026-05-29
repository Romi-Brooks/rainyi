import type {
  LoginResponse,
  RegisterResponse,
  ProfileResponse,
  ConversationsResponse,
  MessagesResponse,
  ConversationResponse,
  PersonasResponse,
  PersonaResponse,
  FileRecord,
} from '../types/api'

const BASE_URL = import.meta.env.VITE_API_URL || '/api'

function getToken(): string {
  return localStorage.getItem('token') || ''
}

interface RequestOptions extends Omit<RequestInit, 'headers'> {
  headers?: Record<string, string>
}

async function request<T>(url: string, options: RequestOptions = {}): Promise<T> {
  const config: RequestInit = {
    headers: {
      'Content-Type': 'application/json',
      ...(options.headers || {}),
    },
    ...options,
  }

  const token = getToken()
  if (token) {
    config.headers = {
      ...(config.headers as Record<string, string>),
      'Authorization': `Bearer ${token}`,
    }
  }

  try {
    const response = await fetch(`${BASE_URL}${url}`, config)
    const data = await response.json()

    if (!response.ok) {
      throw new Error(data.error || `请求失败 (${response.status})`)
    }

    return data as T
  } catch (error) {
    if (error instanceof TypeError && (error as Error).message.includes('fetch')) {
      throw new Error('网络连接失败，请检查服务器是否启动')
    }
    throw error
  }
}

export const authAPI = {
  login: (email: string, password: string) =>
    request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  register: (username: string, email: string, password: string) =>
    request<RegisterResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ username, email, password }),
    }),
}

export const conversationAPI = {
  getConversations: () =>
    request<ConversationsResponse>('/conversations'),

  getMessages: (convId: number, limit = 50, offset = 0) =>
    request<MessagesResponse>(`/conversations/${convId}/messages?limit=${limit}&offset=${offset}`),

  clearMessages: (convId: number) =>
    request<void>(`/conversations/${convId}/messages`, {
      method: 'DELETE',
    }),

  updateConfig: (convId: number, config: Record<string, unknown>) =>
    request<ConversationResponse>(`/conversations/${convId}/config`, {
      method: 'PUT',
      body: JSON.stringify(config),
    }),

  createConversation: (title: string) =>
    request<ConversationResponse>('/conversations', {
      method: 'POST',
      body: JSON.stringify({ title }),
    }),
}

export const userAPI = {
  getProfile: () =>
    request<ProfileResponse>('/user/profile'),

  updateProfile: (data: Record<string, unknown>) =>
    request<ProfileResponse>('/user/profile', {
      method: 'PUT',
      body: JSON.stringify(data),
    }),
}

export const personaAPI = {
  getPersonas: () =>
    request<PersonasResponse>('/personas'),

  getPersona: (id: number) =>
    request<PersonaResponse>(`/personas/${id}`),

  createPersona: (data: Record<string, unknown>) =>
    request<PersonaResponse>('/personas', {
      method: 'POST',
      body: JSON.stringify(data),
    }),

  updatePersona: (id: number, data: Record<string, unknown>) =>
    request<PersonaResponse>(`/personas/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    }),

  deletePersona: (id: number) =>
    request<void>(`/personas/${id}`, {
      method: 'DELETE',
    }),

  uploadSkillFile: (personaId: number, files: File[]) => {
    const formData = new FormData()
    for (const file of files) {
      formData.append('file', file)
    }
    return request<{ message: string; uploaded?: string[]; errors?: string[] }>(`/personas/${personaId}/files`, {
      method: 'POST',
      headers: {},
      body: formData,
    })
  },

  getDebugPrompt: (convId: number) =>
    request<{ system_prompt: string }>(`/personas/${convId}/debug`),

  deleteSkillFile: (personaId: number, nodeId: number) =>
    request<void>(`/personas/${personaId}/files/${nodeId}`, {
      method: 'DELETE',
    }),

  loadFromDirectory: () =>
    request<{ message: string; count: number }>('/personas/load', {
      method: 'POST',
    }),

  getConversationPersona: (convId: number) =>
    request<PersonaResponse>(`/conversations/${convId}/persona`),

  setConversationPersona: (convId: number, personaId: number | null) =>
    request<ConversationResponse>(`/conversations/${convId}/persona`, {
      method: 'PUT',
      body: JSON.stringify({ persona_id: personaId }),
    }),

  uploadAvatar: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request<{ message: string; file: FileRecord }>('/upload/avatar', {
      method: 'POST',
      headers: {},
      body: formData,
    })
  },

  uploadAIAvatar: (file: File, conversationId: number) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('conversation_id', String(conversationId))
    return request<{ message: string; file: FileRecord }>('/upload/avatar/ai', {
      method: 'POST',
      headers: {},
      body: formData,
    })
  },

  uploadPersonaAvatar: (personaId: number, file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request<{ message: string; avatar: string }>(`/personas/${personaId}/avatar`, {
      method: 'POST',
      headers: {},
      body: formData,
    })
  },

  uploadImage: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request<{ message: string; file: FileRecord }>('/upload/image', {
      method: 'POST',
      headers: {},
      body: formData,
    })
  },
}
