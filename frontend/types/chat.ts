export interface Message {
  role: 'user' | 'assistant';
  content: string;
  timestamp?: string;
}

export interface Conversation {
  session_id: string;
  messages: Message[];
  created_at: string;
}
