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
  description: string
  created_at: string
  updated_at: string
  is_built_in?: boolean
  skill_node_count?: number
}

export interface SkillNode {
  id: number
  persona_id: number
  parent_id: number | null
  name: string
  description: string
  file_name: string
  content: string
  priority: number
  created_at: string
  kvs: SkillKV[]
}

export interface SkillKV {
  id: number
  skill_node_id: number
  key: string
  value: string
  sort_order: number
  created_at: string
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
  description: string
  created_at: string
  updated_at: string
  is_built_in?: boolean
  skill_node_count?: number
}

export interface PersonasResponse {
  personas: PersonaFromServer[]
}

export interface PersonaResponse {
  persona: PersonaFromServer
  skill_nodes: SkillNode[]
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
