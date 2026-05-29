export interface User {
  id: number
  username: string
  email: string
  avatar: string
  created_at: string
  updated_at: string
}

export interface Conversation {
  id: number
  user_id: number
  title: string
  ai_nickname: string
  ai_avatar: string
  persona_id: number | null
  created_at: string
  updated_at: string
  avatar?: string | null
  last_message?: { content: string; role: string; created_at: string } | null
}

export interface Message {
  id: number
  conversation_id: number
  role: 'user' | 'assistant'
  content: string
  created_at: string
  is_deleted: boolean
}

export interface Persona {
  id: number
  user_id: number
  name: string
  nickname: string
  description: string
  dir_name: string
  avatar: string
  is_active: boolean
  created_at: string
  updated_at: string
  is_built_in?: boolean
  file_count?: number
}

export interface LoginResponse {
  token: string
  user: User
}

export interface RegisterResponse {
  token: string
  user: User
}

export interface ProfileResponse {
  user: User
}

export interface ConversationsResponse {
  conversations: Conversation[]
}

export interface MessagesResponse {
  messages: Message[]
  total: number
}

export interface ConversationResponse {
  conversation: Conversation
}

export interface PersonaFromServer {
  id: number
  user_id: number
  name: string
  nickname: string
  description: string
  dir_name: string
  avatar: string
  is_active: boolean
  created_at: string
  updated_at: string
  is_built_in?: boolean
  file_count?: number
}

export interface PersonasResponse {
  personas: PersonaFromServer[]
}

export interface PersonaResponse {
  persona: PersonaFromServer
  persona_files: PersonaFile[]
}

export interface PersonaFile {
  id: number
  persona_id: number
  file_name: string
  minio_path: string
  priority: number
  module_category: string
  file_size: number
  created_at: string
}

export interface FileRecord {
  id: number
  user_id: number
  file_type: string
  reference_id: number
  reference_type: string
  original_name: string
  storage_path: string
  url: string
  size: number
  mime_type: string
  created_at: string
}
