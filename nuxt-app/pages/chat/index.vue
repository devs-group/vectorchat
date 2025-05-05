<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold tracking-tight">Chats</h1>
      <Button @click="createNewChat">
        <svg
          xmlns="http://www.w3.org/2000/svg"
          width="24"
          height="24"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          class="mr-2 h-4 w-4"
        >
          <path d="M5 12h14"></path>
          <path d="M12 5v14"></path>
        </svg>
        New Chat
      </Button>
    </div>

    <div class="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="chat in chats"
        :key="chat.id"
        class="group relative rounded-lg border p-4 hover:border-accent"
      >
        <div class="flex flex-col gap-2">
          <div class="flex items-center justify-between">
            <h3 class="font-semibold">{{ chat.name }}</h3>
            <div class="flex items-center gap-2">
              <span class="text-xs text-muted-foreground">
                {{ formatDate(chat.created_at) }}
              </span>
              <Button
                variant="ghost"
                size="icon"
                class="h-8 w-8"
                @click="deleteChat(chat.id)"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="h-4 w-4"
                >
                  <path d="M3 6h18"></path>
                  <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6"></path>
                  <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path>
                </svg>
              </Button>
            </div>
          </div>
          <p class="text-sm text-muted-foreground">{{ chat.description }}</p>
          <div class="flex items-center gap-2 text-sm text-muted-foreground">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4"
            >
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
            </svg>
            <span>{{ chat.message_count }} messages</span>
          </div>
          <div class="flex items-center gap-2 text-sm text-muted-foreground">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-4 w-4"
            >
              <path d="M14.5 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V7.5L14.5 2z"></path>
              <polyline points="14 2 14 8 20 8"></polyline>
            </svg>
            <span>{{ chat.file_count }} files</span>
          </div>
        </div>
        <NuxtLink
          :to="`/chat/${chat.id}`"
          class="absolute inset-0"
          aria-label="View chat"
        ></NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
definePageMeta({
  layout: 'authenticated'
})

// Types based on Swagger API
interface Chat {
  id: string
  name: string
  description: string
  created_at: string
  updated_at: string
  user_id: string
  message_count: number
  file_count: number
}

// Mock data for development
const chats = ref<Chat[]>([
  {
    id: '1',
    name: 'Project Documentation',
    description: 'Chat about project documentation and setup',
    created_at: '2024-03-20T10:00:00Z',
    updated_at: '2024-03-20T15:30:00Z',
    user_id: '1',
    message_count: 12,
    file_count: 3,
  },
  {
    id: '2',
    name: 'API Integration',
    description: 'Discussion about API endpoints and integration',
    created_at: '2024-03-19T14:20:00Z',
    updated_at: '2024-03-20T09:15:00Z',
    user_id: '1',
    message_count: 8,
    file_count: 2,
  },
  {
    id: '3',
    name: 'Database Schema',
    description: 'Chat about database structure and relationships',
    created_at: '2024-03-18T09:45:00Z',
    updated_at: '2024-03-19T16:20:00Z',
    user_id: '1',
    message_count: 15,
    file_count: 4,
  },
])

// Format date for display
const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}

// Create new chat
const createNewChat = async () => {
  // TODO: Implement chat creation using the API
  // POST /chat/chatbot
  console.log('Create new chat')
}

// Delete chat
const deleteChat = async (chatId: string) => {
  // TODO: Implement chat deletion using the API
  console.log('Delete chat:', chatId)
}

// TODO: Implement API integration
// const fetchChats = async () => {
//   try {
//     const response = await fetch('/api/chats')
//     const data = await response.json()
//     chats.value = data
//   } catch (error) {
//     console.error('Error fetching chats:', error)
//   }
// }

// onMounted(() => {
//   fetchChats()
// })
</script> 