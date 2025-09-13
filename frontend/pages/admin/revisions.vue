<template>
  <div class="admin-revisions-page">
    <div class="page-header">
      <h1>Answer Revisions</h1>
      <div class="header-actions">
        <select v-model="selectedChatbotId" @change="handleChatbotChange" class="chatbot-selector">
          <option value="">Select a chatbot</option>
          <option v-for="chatbot in chatbots" :key="chatbot.id" :value="chatbot.id">
            {{ chatbot.name }}
          </option>
        </select>
        <button @click="refreshRevisions" :disabled="!selectedChatbotId" class="btn-refresh">
          <Icon name="mdi:refresh" /> Refresh
        </button>
        <button @click="openCreateModal" :disabled="!selectedChatbotId" class="btn-create">
          <Icon name="mdi:plus" /> New Revision
        </button>
      </div>
    </div>

    <div v-if="error" class="error-message">
      {{ error }}
    </div>

    <div v-if="loading" class="loading-spinner">
      <Icon name="mdi:loading" class="animate-spin" size="32" />
      Loading revisions...
    </div>

    <div v-else-if="revisions.length === 0 && selectedChatbotId" class="no-revisions">
      <Icon name="mdi:file-document-outline" size="48" />
      <p>No revisions found for this chatbot.</p>
      <button @click="openCreateModal" class="btn-create-first">
        Create First Revision
      </button>
    </div>

    <div v-else-if="revisions.length > 0" class="revisions-container">
      <div class="filters-bar">
        <input
          v-model="searchTerm"
          type="text"
          placeholder="Search revisions..."
          class="search-input"
        />
        <label class="checkbox-label">
          <input
            v-model="showInactive"
            type="checkbox"
            @change="handleShowInactiveChange"
          />
          Show inactive revisions
        </label>
      </div>

      <div class="revisions-table">
        <table>
          <thead>
            <tr>
              <th>Status</th>
              <th>Question</th>
              <th>Original Answer</th>
              <th>Revised Answer</th>
              <th>Reason</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="revision in filteredRevisions"
              :key="revision.id"
              :class="{ inactive: !revision.is_active }"
            >
              <td>
                <span
                  class="status-badge"
                  :class="revision.is_active ? 'active' : 'inactive'"
                >
                  {{ revision.is_active ? 'Active' : 'Inactive' }}
                </span>
              </td>
              <td class="question-cell">
                <div class="text-truncate" :title="revision.question">
                  {{ revision.question }}
                </div>
              </td>
              <td class="answer-cell">
                <div class="text-truncate" :title="revision.original_answer">
                  {{ revision.original_answer }}
                </div>
              </td>
              <td class="answer-cell">
                <div class="text-truncate" :title="revision.revised_answer">
                  {{ revision.revised_answer }}
                </div>
              </td>
              <td class="reason-cell">
                {{ revision.revision_reason || '-' }}
              </td>
              <td class="date-cell">
                {{ formatDate(revision.created_at) }}
              </td>
              <td class="actions-cell">
                <button
                  @click="openViewModal(revision)"
                  class="btn-action"
                  title="View details"
                >
                  <Icon name="mdi:eye" />
                </button>
                <button
                  @click="openEditModal(revision)"
                  class="btn-action"
                  title="Edit revision"
                >
                  <Icon name="mdi:pencil" />
                </button>
                <button
                  v-if="revision.is_active"
                  @click="deactivateRevision(revision.id)"
                  class="btn-action btn-danger"
                  title="Deactivate"
                >
                  <Icon name="mdi:close-circle" />
                </button>
                <button
                  v-else
                  @click="reactivateRevision(revision.id)"
                  class="btn-action btn-success"
                  title="Reactivate"
                >
                  <Icon name="mdi:check-circle" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- View Modal -->
    <div v-if="showViewModal" class="modal-overlay" @click.self="closeViewModal">
      <div class="modal-content large">
        <div class="modal-header">
          <h2>Revision Details</h2>
          <button @click="closeViewModal" class="btn-close">
            <Icon name="mdi:close" />
          </button>
        </div>

        <div class="modal-body">
          <div class="detail-grid">
            <div class="detail-item">
              <label>Status:</label>
              <span
                class="status-badge"
                :class="currentRevision.is_active ? 'active' : 'inactive'"
              >
                {{ currentRevision.is_active ? 'Active' : 'Inactive' }}
              </span>
            </div>
            <div class="detail-item">
              <label>Created:</label>
              <span>{{ formatFullDate(currentRevision.created_at) }}</span>
            </div>
            <div class="detail-item">
              <label>Updated:</label>
              <span>{{ formatFullDate(currentRevision.updated_at) }}</span>
            </div>
            <div class="detail-item">
              <label>Revised By:</label>
              <span>{{ currentRevision.revised_by }}</span>
            </div>
          </div>

          <div class="detail-section">
            <h3>Question</h3>
            <div class="detail-content">{{ currentRevision.question }}</div>
          </div>

          <div class="detail-section">
            <h3>Original Answer</h3>
            <div class="detail-content">{{ currentRevision.original_answer }}</div>
          </div>

          <div class="detail-section">
            <h3>Revised Answer</h3>
            <div class="detail-content revised">{{ currentRevision.revised_answer }}</div>
          </div>

          <div v-if="currentRevision.revision_reason" class="detail-section">
            <h3>Revision Reason</h3>
            <div class="detail-content">{{ currentRevision.revision_reason }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Edit/Create Modal -->
    <div v-if="showEditModal" class="modal-overlay" @click.self="closeEditModal">
      <div class="modal-content">
        <div class="modal-header">
          <h2>{{ editMode === 'create' ? 'Create New Revision' : 'Edit Revision' }}</h2>
          <button @click="closeEditModal" class="btn-close">
            <Icon name="mdi:close" />
          </button>
        </div>

        <div class="modal-body">
          <div class="form-group">
            <label for="edit-question">Question:</label>
            <textarea
              id="edit-question"
              v-model="editForm.question"
              rows="3"
              class="form-textarea"
              placeholder="Enter the question..."
              :disabled="editMode === 'edit'"
            ></textarea>
          </div>

          <div v-if="editMode === 'create'" class="form-group">
            <label for="edit-original">Original Answer:</label>
            <textarea
              id="edit-original"
              v-model="editForm.original_answer"
              rows="4"
              class="form-textarea"
              placeholder="Enter the original answer..."
            ></textarea>
          </div>

          <div class="form-group">
            <label for="edit-revised">Revised Answer:</label>
            <textarea
              id="edit-revised"
              v-model="editForm.revised_answer"
              rows="4"
              class="form-textarea"
              placeholder="Enter the revised answer..."
            ></textarea>
          </div>

          <div class="form-group">
            <label for="edit-reason">Revision Reason (optional):</label>
            <input
              id="edit-reason"
              v-model="editForm.revision_reason"
              type="text"
              class="form-input"
              placeholder="e.g., Incorrect information, outdated content..."
            />
          </div>

          <div v-if="editMode === 'edit'" class="form-group">
            <label class="checkbox-label">
              <input v-model="editForm.is_active" type="checkbox" />
              Active
            </label>
          </div>
        </div>

        <div class="modal-footer">
          <button @click="closeEditModal" class="btn-cancel">Cancel</button>
          <button @click="submitEdit" :disabled="!isFormValid || submitting" class="btn-submit">
            {{ submitting ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, watch } from 'vue'
import { useRevisions } from '~/composables/useRevisions'

// Composables
const {
  revisions,
  loading,
  error,
  fetchRevisions,
  createRevision,
  updateRevision,
  deactivateRevision: deactivateRevisionApi
} = useRevisions()

// State
const selectedChatbotId = ref('')
const chatbots = ref([]) // This should be populated from your chatbot API
const searchTerm = ref('')
const showInactive = ref(false)
const showViewModal = ref(false)
const showEditModal = ref(false)
const editMode = ref<'create' | 'edit'>('create')
const currentRevision = ref<any>(null)
const submitting = ref(false)

// Edit form
const editForm = reactive({
  id: '',
  question: '',
  original_answer: '',
  revised_answer: '',
  revision_reason: '',
  is_active: true
})

// Computed
const filteredRevisions = computed(() => {
  let filtered = revisions.value

  if (!showInactive.value) {
    filtered = filtered.filter(r => r.is_active)
  }

  if (searchTerm.value) {
    const term = searchTerm.value.toLowerCase()
    filtered = filtered.filter(r =>
      r.question.toLowerCase().includes(term) ||
      r.revised_answer.toLowerCase().includes(term) ||
      (r.revision_reason && r.revision_reason.toLowerCase().includes(term))
    )
  }

  return filtered
})

const isFormValid = computed(() => {
  if (editMode.value === 'create') {
    return editForm.question && editForm.original_answer && editForm.revised_answer
  }
  return editForm.revised_answer
})

// Methods
const handleChatbotChange = async () => {
  if (selectedChatbotId.value) {
    await fetchRevisions(selectedChatbotId.value, showInactive.value)
  }
}

const handleShowInactiveChange = async () => {
  if (selectedChatbotId.value) {
    await fetchRevisions(selectedChatbotId.value, showInactive.value)
  }
}

const refreshRevisions = async () => {
  if (selectedChatbotId.value) {
    await fetchRevisions(selectedChatbotId.value, showInactive.value)
  }
}

const openViewModal = (revision: any) => {
  currentRevision.value = revision
  showViewModal.value = true
}

const closeViewModal = () => {
  showViewModal.value = false
  currentRevision.value = null
}

const openCreateModal = () => {
  editMode.value = 'create'
  editForm.id = ''
  editForm.question = ''
  editForm.original_answer = ''
  editForm.revised_answer = ''
  editForm.revision_reason = ''
  editForm.is_active = true
  showEditModal.value = true
}

const openEditModal = (revision: any) => {
  editMode.value = 'edit'
  editForm.id = revision.id
  editForm.question = revision.question
  editForm.original_answer = revision.original_answer
  editForm.revised_answer = revision.revised_answer
  editForm.revision_reason = revision.revision_reason || ''
  editForm.is_active = revision.is_active
  showEditModal.value = true
}

const closeEditModal = () => {
  showEditModal.value = false
}

const submitEdit = async () => {
  if (!isFormValid.value || submitting.value) return

  submitting.value = true
  try {
    if (editMode.value === 'create') {
      await createRevision({
        chatbot_id: selectedChatbotId.value,
        question: editForm.question,
        original_answer: editForm.original_answer,
        revised_answer: editForm.revised_answer,
        revision_reason: editForm.revision_reason || undefined
      })
    } else {
      const updates: any = {
        revised_answer: editForm.revised_answer,
        revision_reason: editForm.revision_reason || undefined,
        is_active: editForm.is_active
      }
      await updateRevision(editForm.id, updates)
    }

    closeEditModal()
    await refreshRevisions()
    alert(editMode.value === 'create' ? 'Revision created successfully!' : 'Revision updated successfully!')
  } catch (err) {
    console.error('Failed to save revision:', err)
    alert('Failed to save revision. Please try again.')
  } finally {
    submitting.value = false
  }
}

const deactivateRevision = async (revisionId: string) => {
  if (confirm('Are you sure you want to deactivate this revision?')) {
    try {
      await deactivateRevisionApi(revisionId)
      await refreshRevisions()
    } catch (err) {
      console.error('Failed to deactivate revision:', err)
      alert('Failed to deactivate revision. Please try again.')
    }
  }
}

const reactivateRevision = async (revisionId: string) => {
  try {
    await updateRevision(revisionId, { is_active: true })
    await refreshRevisions()
  } catch (err) {
    console.error('Failed to reactivate revision:', err)
    alert('Failed to reactivate revision. Please try again.')
  }
}

const formatDate = (dateString: string) => {
  return new Date(dateString).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric'
  })
}

const formatFullDate = (dateString: string) => {
  return new Date(dateString).toLocaleString('en-US', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
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
.admin-revisions-page {
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

.btn-refresh,
.btn-create,
.btn-create-first {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 1rem;
  border: none;
  border-radius: 0.375rem;
  cursor: pointer;
  transition: background 0.2s;
  font-weight: 500;
}

.btn-refresh {
  background: #4299e1;
  color: white;
}

.btn-refresh:hover:not(:disabled) {
  background: #3182ce;
}

.btn-create,
.btn-create-first {
  background: #48bb78;
  color: white;
}

.btn-create:hover:not(:disabled),
.btn-create-first:hover {
  background: #38a169;
}

.btn-refresh:disabled,
.btn-create:disabled {
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

.no-revisions {
  text-align: center;
  padding: 4rem;
  color: #718096;
}

.no-revisions p {
  margin: 1rem 0 2rem;
}

.revisions-container {
  background: white;
  border-radius: 0.5rem;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.filters-bar {
  display: flex;
  gap: 2rem;
  align-items: center;
  padding: 1rem;
  border-bottom: 1px solid #e2e8f0;
}

.search-input {
  flex: 1;
  max-width: 400px;
  padding: 0.5rem 1rem;
  border: 1px solid #e2e8f0;
  border-radius: 0.375rem;
  font-size: 1rem;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  cursor: pointer;
  color: #4a5568;
}

.checkbox-label input {
  cursor: pointer;
}

.revisions-table {
  overflow-x: auto;
}

table {
  width: 100%;
  border-collapse: collapse;
}

thead {
  background: #f7fafc;
}

th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-weight: 600;
  color: #4a5568;
  border-bottom: 2px solid #e2e8f0;
}

tbody tr {
  border-bottom: 1px solid #e2e8f0;
  transition: background 0.2s;
}

tbody tr:hover {
  background: #f7fafc;
}

tbody tr.inactive {
  opacity: 0.6;
}

td {
  padding: 0.75rem 1rem;
  color: #2d3748;
}

.status-badge {
  display: inline-block;
  padding: 0.25rem 0.75rem;
  border-radius: 9999px;
  font-size: 0.875rem;
  font-weight: 500;
}

.status-badge.active {
  background: #c6f6d5;
  color: #22543d;
}

.status-badge.inactive {
  background: #fed7d7;
  color: #742a2a;
}

.text-truncate {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.question-cell,
.answer-cell {
  max-width: 250px;
}

.reason-cell {
  max-width: 150px;
  font-size: 0.875rem;
  color: #718096;
}

.date-cell {
  font-size: 0.875rem;
  color: #718096;
  white-space: nowrap;
}

.actions-cell {
  display: flex;
  gap: 0.5rem;
}

.btn-action {
  padding: 0.375rem;
  background: white;
  border: 1px solid #e2e8f0;
  border-radius: 0.25rem;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
}

.btn-action:hover {
  background: #f7fafc;
  border-color: #cbd5e0;
}

.btn-action.btn-danger:hover {
  background: #fed7d7;
  border-color: #fc8181;
  color: #c53030;
}

.btn-action.btn-success:hover {
  background: #c6f6d5;
  border-color: #9ae6b4;
  color: #22543d;
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

.modal-content.large {
  max-width: 800px;
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

.detail-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
  margin-bottom: 2rem;
}

.detail-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.detail-item label {
  font-weight: 500;
  color: #4a5568;
}

.detail-section {
  margin-bottom: 1.5rem;
}

.detail-section h3 {
  font-size: 1.125rem;
  font-weight: 600;
  color: #2d3748;
  margin-bottom: 0.5rem;
}

.detail-content {
  padding: 1rem;
  background: #f7fafc;
  border-radius: 0.375rem;
  color: #2d3748;
  line-height: 1.6;
  white-space: pre-wrap;
}

.detail-content.revised {
  background: #f0fff4;
  border-left: 3px solid #48bb78;
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

.form-textarea:disabled {
  background: #f7fafc;
  cursor: not-allowed;
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
  background: #48bb78;
  color: white;
  border: none;
}

.btn-submit:hover:not(:disabled) {
  background: #38a169;
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
