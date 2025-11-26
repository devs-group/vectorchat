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
  client_id: string;
  name?: string | null;
  user_id: string;
  created_at: string;
  expires_at?: string | null;
}

export interface ChatbotCreateRequest {
  name: string;
  description: string;
  model_name: string;
  system_instructions: string;
  max_tokens: number;
  temperature_param: number;
  save_messages: boolean;
   use_max_tokens?: boolean;
  is_enabled?: boolean;
  shared_knowledge_base_ids?: string[];
}

export interface ChatbotResponse
  extends Omit<ChatbotCreateRequest, "is_enabled"> {
  id: string;
  user_id: string;
  is_enabled: boolean;
  created_at: string;
  updated_at: string;
  shared_knowledge_base_ids: string[];
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

export interface APIKeyCreateResponse {
  client_id: string;
  client_secret: string;
  name?: string | null;
  expires_at?: string | null;
  message: string;
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

export interface TextSource {
  id: string;
  title: string;
  size: number;
  uploaded_at: string;
}

export interface TextSourcesResponse {
  chat_id: string;
  sources: TextSource[];
}

export interface SharedKnowledgeBase {
  id: string;
  owner_id: string;
  name: string;
  description?: string | null;
  created_at: string;
  updated_at: string;
}

export interface SharedKnowledgeBaseCreateRequest {
  name: string;
  description?: string | null;
}

export interface SharedKnowledgeBaseUpdateRequest {
  name?: string;
  description?: string | null;
}

export interface SharedKnowledgeBaseListResponse {
  knowledge_bases: SharedKnowledgeBase[];
}

export interface SharedKnowledgeBaseFileUploadResponse {
  message: string;
  knowledge_base_id: string;
  file: string;
  filename?: string;
  size?: number;
}

export interface SharedKnowledgeBaseFilesResponse {
  knowledge_base_id: string;
  files: FileInfo[];
}

export interface SharedKnowledgeBaseTextSourcesResponse {
  knowledge_base_id: string;
  sources: TextSource[];
}

// Billing
export interface Plan {
  id: string;
  key: string;
  display_name: string;
  active: boolean;
  billing_interval: string; // day|week|month|year
  amount_cents: number;
  currency: string;
  plan_definition?: {
    features?: Record<string, any>;
    tags?: string[];
    [k: string]: any;
  } | null;
  created_at?: string;
  updated_at?: string;
}

export interface Subscription {
  id: string;
  customer_id: string;
  stripe_subscription_id: string;
  status: string;
  current_period_start?: string | null;
  current_period_end?: string | null;
  cancel_at_period_end: boolean;
  metadata?: Record<string, any> | null;
  created_at: string;
  updated_at: string;
}

// Conversations
export interface MessageDetails {
  id: string;
  chatbot_id: string;
  role: "user" | "assistant" | "system";
  content: string;
  created_at: string;
}

export interface ConversationListItemResponse {
  session_id: string;
  first_message_content: string;
  first_message_at: string;
  last_message_at: string;
}

export interface ConversationPaginationResponse {
  page: number;
  per_page: number;
  total_items: number;
  total_pages: number;
  has_next_page: boolean;
  has_prev_page: boolean;
  offset: number;
  requested_offset?: number;
  next_page?: number;
  prev_page?: number;
  next_offset?: number;
  prev_offset?: number;
}

export interface ConversationsResponse {
  conversations: ConversationListItemResponse[];
  pagination: ConversationPaginationResponse;
}

// Revisions
export interface CreateRevisionRequest {
  chatbot_id: string;
  original_message_id?: string;
  question: string;
  original_answer: string;
  revised_answer: string;
  revision_reason?: string | null;
}

export interface UpdateRevisionRequest {
  question?: string;
  revised_answer?: string;
  revision_reason?: string | null;
  is_active?: boolean;
}

export interface RevisionResponse {
  id: string;
  chatbot_id: string;
  original_message_id?: string | null;
  question: string;
  original_answer: string;
  revised_answer: string;
  revision_reason?: string | null;
  revised_by: string;
  created_at: string;
  updated_at: string;
  is_active: boolean;
  similarity?: number;
}

export interface RevisionsListResponse {
  revisions: RevisionResponse[];
}
