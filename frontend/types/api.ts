export interface User {
  id: string;
  name: string;
  email: string;
  provider: string;
  created_at: string;
  updated_at: string;
}

export interface APIKey {
  id: string;
  key: string;
  name: string;
  user_id: string;
  created_at: string;
  expires_at: string;
  revoked_at: string | null;
}

export interface ChatbotCreateRequest {
  name: string;
  description: string;
  model_name: string;
  system_instructions: string;
  max_tokens: number;
  temperature_param: number;
}

export interface ChatbotResponse extends ChatbotCreateRequest {
  id: string;
  user_id: string;
  created_at: string;
  updated_at: string;
}

export type ListChatsResponse = {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  user_id: string;
  message_count: number;
  file_count: number;
}[];

export interface ChatMessageRequest {
  query: string;
}

export interface ChatResponse {
  chat_id: string;
  message: string;
  context: string;
}

export interface ChatFile {
  filename: string;
  size: number;
  updated_at: string;
}

export interface APIResponse {
  message: string;
  error: string;
  data: any;
}

export interface SessionResponse {
  user: User;
}

export interface APIKeyResponse {
  api_key: APIKey;
  message: string;
  plain_key: string;
}

export interface GenerateAPIKeyRequest {
  name: string;
  expires_at?: string;
}

export interface PaginationMetadata {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface APIKeysResponse {
  api_keys: APIKey[];
  pagination: PaginationMetadata;
}

export interface FileUploadResponse {
  chat_id: string;
  filename: string;
  size: number;
}

export interface ChatFilesResponse {
  files: ChatFile[];
}
