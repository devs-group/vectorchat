import { ref } from 'vue'
import type { Ref } from 'vue'

export interface Message {
  id: string
  chatbot_id: string
  role: 'user' | 'assistant'
  content: string
  created_at: string
}

export interface Conversation {
  session_id: string
  messages: Message[]
  created_at: string
}

export interface ConversationsResponse {
  conversations: Conversation[]
  total_count: number
  limit: number
  offset: number
}

export const useAdminConversations = () => {
  const conversations: Ref<Conversation[]> = ref([])
  const totalCount = ref(0)
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Fetch conversations for a chatbot
  const fetchConversations = async (
    chatbotId: string,
    limit = 20,
    offset = 0
  ): Promise<ConversationsResponse> => {
    loading.value = true
    error.value = null

    try {
      const { $fetch } = useNuxtApp()
      const params = new URLSearchParams({
        limit: limit.toString(),
        offset: offset.toString()
      })

      const response = await $fetch<ConversationsResponse>(
        `/api/admin/conversations/${chatbotId}?${params.toString()}`,
        {
          method: 'GET',
          credentials: 'include'
        }
      )

      conversations.value = response.conversations || []
      totalCount.value = response.total_count || 0

      return response
    } catch (err: any) {
      error.value = err.data?.error || 'Failed to fetch conversations'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Load more conversations (pagination)
  const loadMoreConversations = async (
    chatbotId: string,
    limit = 20
  ): Promise<ConversationsResponse> => {
    const offset = conversations.value.length
    const response = await fetchConversations(chatbotId, limit, offset)

    // Append new conversations to existing ones
    if (response.conversations) {
      conversations.value.push(...response.conversations)
    }

    return response
  }

  // Find a specific conversation by session ID
  const findConversation = (sessionId: string): Conversation | undefined => {
    return conversations.value.find(c => c.session_id === sessionId)
  }

  // Get messages from a conversation in a Q&A format
  const getQAPairs = (conversation: Conversation): Array<{ question: string; answer: string; messageId: string }> => {
    const pairs: Array<{ question: string; answer: string; messageId: string }> = []

    for (let i = 0; i < conversation.messages.length - 1; i++) {
      const currentMsg = conversation.messages[i]
      const nextMsg = conversation.messages[i + 1]

      if (currentMsg.role === 'user' && nextMsg.role === 'assistant') {
        pairs.push({
          question: currentMsg.content,
          answer: nextMsg.content,
          messageId: nextMsg.id
        })
      }
    }

    return pairs
  }

  // Clear all conversations
  const clearConversations = () => {
    conversations.value = []
    totalCount.value = 0
    error.value = null
  }

  // Search conversations for specific text
  const searchConversations = (searchTerm: string): Conversation[] => {
    const term = searchTerm.toLowerCase()
    return conversations.value.filter(conversation =>
      conversation.messages.some(message =>
        message.content.toLowerCase().includes(term)
      )
    )
  }

  return {
    conversations: readonly(conversations),
    totalCount: readonly(totalCount),
    loading: readonly(loading),
    error: readonly(error),
    fetchConversations,
    loadMoreConversations,
    findConversation,
    getQAPairs,
    clearConversations,
    searchConversations
  }
}
