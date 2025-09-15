export interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp?: string;
}

// Frontend conversation item for list view
export interface Conversation {
  session_id: string;
  // For list preview (from backend conversations list)
  first_message_content?: string;
  first_message_at?: string;
  last_message_at?: string;
  // For detail view, messages loaded lazily
  messages: Message[];
  // Created at convenience for UI (mapped from first_message_at)
  created_at: string;
}
