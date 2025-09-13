<template>
  <div class="admin-conversations-page">
    <div class="page-header">
      <h1>Conversation History</h1>
      <div class="header-actions">
        <select v-model="selectedChatbotId" @change="handleChatbotChange" class="chatbot-selector">
          <option value="">Select a chatbot</option>
          <option v-for="chatbot in chatbots" :key="chatbot.id" :value="chatbot.id">
            {{ chatbot.name }}
          </option>
        </select>
        <button @click="refreshConversations" :disabled="!selectedChatbotId" class="btn-refresh">
          <Icon name="mdi:refresh" /> Refresh
        </button>
      </div>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="loading" class="loading-spinner">
      <Icon name="mdi:loading" class="animate-spin" size="32" />
      Loading conversations...
    </div>

    <div v-else-if="conversations.length === 0 && selectedChatbotId" class="no-conversations">
      <Icon name="mdi:chat-outline" size="48" />
      <p>No conversations found for this chatbot.</p>
    </div>

    <div v-else-if="conversations.length > 0" class="conversations-container">
      <div class="search-bar">
        <input
          v-model="searchTerm"
          type="text"
          placeholder="Search conversations..."
          class="search-input"
        />
      </div>

      <div class="conversations-list">
        <div
          v-for="conversation in filteredConversations"
          :key="conversation.session_id"
          class="conversation-card"
          :class="{ expanded: expandedConversations.has(conversation.session_id) }"
        >
          <div class="conversation-header" @click="toggleConversation(conversation.session_id)">
            <div class="conversation-info">
              <span class="session-id">Session: {{ conversation.session_id.slice(0, 8) }}...</span>
              <span class="timestamp">{{ formatDate(conversation.created_at) }}</span>
              <span class="message-count">{{ conversation.messages.length }} messages</span>
            </div>
            <Icon
              :name="expandedConversations.has(conversation.session_id) ? 'mdi:chevron-up' : 'mdi:chevron-down'"
              size="24"
            />
          </div>

          <div v-if="expandedConversations.has(conversation.session_id)" class="conversation-messages">
            <div v-for="(message, index) in conversation.messages" :key="message.id" class="message">
              <div class="message-header">
                <span class="message-role" :class="`role-${message.role}`">
                  <Icon :name="message.role === 'user' ? 'mdi:account' : 'mdi:robot'" />
                  {{ message.role === 'user' ? 'User' : 'Assistant' }}
                </span>
                <span class="message-time">{{ formatTime(message.created_at) }}</span>
              </div>
              <div class="message-content">{{ message.content }}</div>

              <!-- Revision button for assistant messages that follow user messages -->
              <div
                v-if="message.role === 'assistant' && index > 0 && conversation.messages[index - 1].role === 'user'"
                class="message-actions"
              >
                <button
                  @click="openRevisionModal(conversation.messages[index - 1], message)"
                  class="btn-revise"
                >
                  <Icon name="mdi:pencil" /> Revise Answer
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="hasMore" class="load-more">
        <button @click="loadMore" :disabled="loadingMore" class="btn-load-more">
          {{ loadingMore ? 'Loading...' : 'Load More Conversations' }}
        </button>
      </div>
    </div>

    <!-- Revision Modal -->
    <div v-if="showRevisionModal" class="modal-overlay" @click.self="closeRevisionModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>Revise Answer</h2>
          <button @click="closeRevisionModal" class="btn-close">
            <Icon name="mdi:close" />
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label>Original Question:</label>
            <div class="readonly-text">{{ revisionForm.question }}</div>
          </div>

          <div class="form-group">
            <label>Original Answer:</label>
            <div class="readonly-text">{{ revisionForm.original_answer }}</div>
          </div>

          <div class="form-group">
            <label for="revised-answer">Revised Answer:</label>
            <textarea
              id="revised-answer"
              v-model="revisionForm.revised_answer"
              rows="6"
              class="form-textarea"
              placeholder="Enter the corrected answer..."
            ></textarea>
          </div>

          <div class="form-group">
            <label for="revision-reason">Revision Reason (optional):</label>
            <input
              id="revision-reason"
              v-model="revisionForm.revision_reason"
              type="text"
              class="form-input"
              placeholder="e.g., Incorrect information, outdated content..."
            />
          </div>
        </div>

        <div class="modal-footer">
          <button @click="closeRevisionModal" class="btn-cancel">Cancel</button>
          <button @click="submitRevision" :disabled="!revisionForm.revised_answer || submittingRevision" class="btn-submit">
            {{ submittingRevision ? 'Saving...' : 'Save Revision' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive } from 'vue'
import { useAdminConversations } from '~/composables/useAdminConversations'
import { useRevisions } from '~/composables/useRevisions'

// Composables
const { conversations, loading, error, fetchConversations, loadMoreConversations } = useAdminConversations()
const { createRevision } = useRevisions()

// State
const selectedChatbotId = ref('')
const chatbots = ref([]) // This should be populated from your chatbot API
const searchTerm = ref('')
const expandedConversations = ref(new Set<string>())
const showRevisionModal = ref(false)
const submittingRevision = ref(false)
const loadingMore = ref(false)
const hasMore = ref(true)

// Revision form
const revisionForm = reactive({
  chatbot_id: '',
  original_message_id: '',
  question: '',
  original_answer: '',
  revised_answer: '',
  revision_reason: ''
})

// Computed
const filteredConversations = computed(() => {
  if (!searchTerm.value) return conversations.value

  const term = searchTerm.value.toLowerCase()
  return conversations.value.filter(conversation =>
    conversation.messages.some(message =>
      message.content.toLowerCase().includes(term)
    )
  )
})

// Methods
const handleChatbotChange = async () => {
  if (selectedChatbotId.value) {
    expandedConversations.value.clear()
    await fetchConversations(selectedChatbotId.value)
    hasMore.value = conversations.value.length >= 20
  }
}

const refreshConversations = async () => {
  if (selectedChatbotId.value) {
    await fetchConversations(selectedChatbotId.value)
    hasMore.value = conversations.value.length >= 20
  }
}

const loadMore = async () => {
  if (selectedChatbotId.value && !loadingMore.value) {
    loadingMore.value = true
    try {
      const response = await loadMoreConversations(selectedChatbotId.value)
      hasMore.value = response.conversations.length >= 20
    } finally {
      loadingMore.value = false
    }
  }
}

const toggleConversation = (sessionId: string) => {
  if (expandedConversations.value.has(sessionId)) {
    expandedConversations.value.delete(sessionId)
  } else {
    expandedConversations.value.add(sessionId)
  }
}

const openRevisionModal = (userMessage: any, assistantMessage: any) => {
  revisionForm.chatbot_id = selectedChatbotId.value
  revisionForm.original_message_id = assistantMessage.id
  revisionForm.question = userMessage.content
  revisionForm.original_answer = assistantMessage.content
  revisionForm.revised_answer = assistantMessage.content // Pre-fill with original
  revisionForm.revision_reason = ''
  showRevisionModal.value = true
}

const closeRevisionModal = () => {
  showRevisionModal.value = false
  // Reset form
  Object.keys(revisionForm).forEach(key => {
    revisionForm[key] = ''
  })
}

const submitRevision = async () => {
  if (!revisionForm.revised_answer || submittingRevision.value) return

  submittingRevision.value = true
  try {
    await createRevision({
      chatbot_id: revisionForm.chatbot_id,
      original_message_id: revisionForm.original_message_id,
      question: revisionForm.question,
      original_answer: revisionForm.original_answer,
      revised_answer: revisionForm.revised_answer,
      revision_reason: revisionForm.revision_reason || undefined
    })

    closeRevisionModal()
    // Show success notification
    alert('Revision saved successfully!')
  } catch (err) {
    console.error('Failed to save revision:', err)
    alert('Failed to save revision. Please try again.')
  } finally {
    submittingRevision.value = false
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatTime = (dateString: string) => {
  return new Date(dateString).toLocaleTimeString('en-US', {
    hour: '2-digit',
    minute: '2-digit'
  })
}

// Load chatbots on mount
onMounted(async () => {
  // TODO: Fetch chatbots from API
  // const response = await $fetch('/api/chatbots')
  // chatbots.value = response.chatbots
})
</script>

<style scoped>
.admin-conversations-page {
  max-width: 1400px;
  margin: 0 auto;
  padding: 2rem;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;
}

.page-header h1 {
  font-size: 2rem;
  font-weight: 600;
  color: #1a202c;
}

.header-actions {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.chatbot-selector {
  padding: 0.5rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  background: white;
  min-width: 200px;
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  background: #4299e1;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-refresh:hover:not(:disabled) {
  background: #3182ce;
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.error-message {
  background: #fed7d7;
  color: #c53030;
  padding: 1rem;
  border-radius: 0.375rem;
  margin-bottom: 1rem;
}

.loading-spinner {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 4rem;
  color: #718096;
}

.no-conversations {
  text-align: center;
  padding: 4rem;
  color: #718096;
}

.conversations-container {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.search-bar {
  padding: 1rem;
  border-bottom: 1px solid #e2e8f0;
}

.search-input {
  width: 100%;
  padding: 0.5rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  font-size: 1rem;
}

.conversations-list {
  max-height: 70vh;
  overflow-y: auto;
}

.conversation-card {
  border-bottom: 1px solid #e2e8f0;
  transition: background 0.2s;
}

.conversation-card:hover {
  background: #f7fafc;
}

.conversation-card.expanded {
  background: #f7fafc;
}

.conversation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem;
  cursor: pointer;
}

.conversation-info {
  display: flex;
  gap: 1rem;
  align-items: center;
}

.session-id {
  font-family: monospace;
  font-size: 0.875rem;
  color: #4a5568;
}

.timestamp {
  color: #718096;
  font-size: 0.875rem;
}

.message-count {
  background: #edf2f7;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  color: #4a5568;
}

.conversation-messages {
  padding: 0 1rem 1rem;
}

.message {
  margin-bottom: 1rem;
  padding: 1rem;
  background: #f7fafc;
  border-radius: 0.375rem;
}

.message-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.message-role {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  font-weight: 500;
}

.role-user {
  color: #2b6cb0;
}

.role-assistant {
  color: #38a169;
}

.message-time {
  color: #718096;
  font-size: 0.875rem;
}

.message-content {
  color: #2d3748;
  line-height: 1.5;
  white-space: pre-wrap;
}

.message-actions {
  margin-top: 0.75rem;
}

.btn-revise {
  display: flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.375rem 0.75rem;
  background: #805ad5;
  color: white;
  border: none;
  border-radius: 0.25rem;
  font-size: 0.875rem;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-revise:hover {
  background: #6b46c1;
}

.load-more {
  padding: 1rem;
  text-align: center;
}

.btn-load-more {
  padding: 0.5rem 1.5rem;
  background: #4299e1;
  color: white;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-load-more:hover:not(:disabled) {
  background: #3182ce;
}

.btn-load-more:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Modal Styles */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: white;
  border-radius: 0.5rem;
  width: 90%;
  max-width: 600px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1.5rem;
  border-bottom: 1px solid #e2e8f0;
}

.modal-header h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: #1a202c;
}

.btn-close {
  background: none;
  border: none;
  font-size: 1.5rem;
  color: #718096;
  cursor: pointer;
  padding: 0;
  display: flex;
  align-items: center;
}

.btn-close:hover {
  color: #4a5568;
}

.modal-body {
  padding: 1.5rem;
}

.form-group {
  margin-bottom: 1.5rem;
}

.form-group label {
  display: block;
  margin-bottom: 0.5rem;
  font-weight: 500;
  color: #4a5568;
}

.readonly-text {
  padding: 0.75rem;
  background: #f7fafc;
  border-radius: 0.375rem;
  color: #2d3748;
  line-height: 1.5;
  white-space: pre-wrap;
}

.form-textarea,
.form-input {
  width: 100%;
  padding: 0.75rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-textarea:focus,
.form-input:focus {
  outline: none;
  border-color: #4299e1;
  box-shadow: 0 0 0 3px rgba(66, 153, 225, 0.1);
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 1rem;
  padding: 1.5rem;
  border-top: 1px solid #e2e8f0;
}

.btn-cancel,
.btn-submit {
  padding: 0.5rem 1.5rem;
  border-radius: 0.375rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-cancel {
  background: white;
  color: #4a5568;
  border: 1px solid #e2e8f0;
}

.btn-cancel:hover {
  background: #f7fafc;
}

.btn-submit {
  background: #805ad5;
  color: white;
  border: none;
}

.btn-submit:hover:not(:disabled) {
  background: #6b46c1;
}

.btn-submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.animate-spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
