import type { ChatbotCreateRequest, ChatbotResponse, ChatResponse, ChatMessageRequest, FileUploadResponse, ChatFilesResponse } from '~/types/api'

export const useChat = () => {
  const config = useRuntimeConfig()
  const apiBase = config.public.apiBase || ''

  // Create a new chatbot
  const createChatbot = async (data: ChatbotCreateRequest): Promise<ChatbotResponse> => {
    const response = await fetch(`${apiBase}/chat/chatbot`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // TODO: Add API key or session token
      },
      body: JSON.stringify(data),
    })

    if (!response.ok) {
      throw new Error('Failed to create chatbot')
    }

    return response.json()
  }

  // Send a message to a chat
  const sendMessage = async (chatId: string, message: ChatMessageRequest): Promise<ChatResponse> => {
    const response = await fetch(`${apiBase}/chat/${chatId}/message`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        // TODO: Add API key or session token
      },
      body: JSON.stringify(message),
    })

    if (!response.ok) {
      throw new Error('Failed to send message')
    }

    return response.json()
  }

  // Upload a file to a chat
  const uploadFile = async (chatId: string, file: File): Promise<FileUploadResponse> => {
    const formData = new FormData()
    formData.append('file', file)

    const response = await fetch(`${apiBase}/chat/${chatId}/upload`, {
      method: 'POST',
      headers: {
        // TODO: Add API key or session token
      },
      body: formData,
    })

    if (!response.ok) {
      throw new Error('Failed to upload file')
    }

    return response.json()
  }

  // List files in a chat
  const listFiles = async (chatId: string): Promise<ChatFilesResponse> => {
    const response = await fetch(`${apiBase}/chat/${chatId}/files`, {
      headers: {
        // TODO: Add API key or session token
      },
    })

    if (!response.ok) {
      throw new Error('Failed to list files')
    }

    return response.json()
  }

  // Delete a file from a chat
  const deleteFile = async (chatId: string, filename: string): Promise<void> => {
    const response = await fetch(`${apiBase}/chat/${chatId}/files/${filename}`, {
      method: 'DELETE',
      headers: {
        // TODO: Add API key or session token
      },
    })

    if (!response.ok) {
      throw new Error('Failed to delete file')
    }
  }

  return {
    createChatbot,
    sendMessage,
    uploadFile,
    listFiles,
    deleteFile,
  }
} 