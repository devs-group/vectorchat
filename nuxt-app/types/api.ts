// Chatbot types
export interface ChatbotCreateRequest {
  name: string
  description: string
  model_name: string
  system_instructions: string
  max_tokens: number
  temperature_param: number
}

export interface ChatbotResponse {
  id: string
  name: string
  description: string
  model_name: string
  system_instructions: string
  max_tokens: number
  temperature_param: number
  user_id: string
  created_at: string
  updated_at: string
}

// Chat message types
export interface ChatMessageRequest {
  query: string
}

export interface ChatResponse {
  chat_id: string
  message: string
  context: string
}

// File types
export interface FileUploadResponse {
  chat_id: string
  filename: string
  size: number
}

export interface ChatFile {
  filename: string
  size: number
  updated_at: string
}

export interface ChatFilesResponse {
  files: ChatFile[]
}

// User types
export interface User {
  id: string
  name: string
  email: string
  provider: string
  created_at: string
  updated_at: string
}

export interface UserResponse {
  user: User
}

// API Key types
export interface APIKey {
  id: string
  key: string
  user_id: string
  created_at: string
  expires_at: string
  revoked_at: string | null
}

export interface APIKeyResponse {
  api_key: APIKey
}

export interface APIKeysResponse {
  api_keys: APIKey[]
}

// Session types
export interface SessionResponse {
  user: User
}

// Generic API response
export interface APIResponse {
  message?: string
  error?: string
  data?: any
} 