import { ref } from 'vue'
import type { Ref } from 'vue'

export interface AnswerRevision {
  id: string
  chatbot_id: string
  original_message_id?: string
  question: string
  original_answer: string
  revised_answer: string
  revision_reason?: string
  revised_by: string
  created_at: string
  updated_at: string
  is_active: boolean
  similarity?: number
}

export interface CreateRevisionRequest {
  chatbot_id: string
  original_message_id?: string
  question: string
  original_answer: string
  revised_answer: string
  revision_reason?: string
}

export interface UpdateRevisionRequest {
  question?: string
  revised_answer?: string
  revision_reason?: string
  is_active?: boolean
}

export const useRevisions = () => {
  const revisions: Ref<AnswerRevision[]> = ref([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  // Fetch all revisions for a chatbot
  const fetchRevisions = async (chatbotId: string, includeInactive = false) => {
    loading.value = true
    error.value = null

    try {
      const { $fetch } = useNuxtApp()
      const params = new URLSearchParams()
      if (includeInactive) {
        params.append('includeInactive', 'true')
      }

      const response = await $fetch(`/api/admin/revisions/${chatbotId}?${params.toString()}`, {
        method: 'GET',
        credentials: 'include'
      })

      revisions.value = response.revisions || []
      return revisions.value
    } catch (err: any) {
      error.value = err.data?.error || 'Failed to fetch revisions'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Create a new revision
  const createRevision = async (revision: CreateRevisionRequest) => {
    loading.value = true
    error.value = null

    try {
      const { $fetch } = useNuxtApp()
      const response = await $fetch('/api/admin/revisions', {
        method: 'POST',
        body: revision,
        credentials: 'include'
      })

      // Add the new revision to the list
      if (response) {
        revisions.value.unshift(response)
      }

      return response
    } catch (err: any) {
      error.value = err.data?.error || 'Failed to create revision'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Update an existing revision
  const updateRevision = async (revisionId: string, updates: UpdateRevisionRequest) => {
    loading.value = true
    error.value = null

    try {
      const { $fetch } = useNuxtApp()
      await $fetch(`/api/admin/revisions/${revisionId}`, {
        method: 'PUT',
        body: updates,
        credentials: 'include'
      })

      // Update the revision in the local list
      const index = revisions.value.findIndex(r => r.id === revisionId)
      if (index !== -1) {
        revisions.value[index] = {
          ...revisions.value[index],
          ...updates,
          updated_at: new Date().toISOString()
        }
      }

      return true
    } catch (err: any) {
      error.value = err.data?.error || 'Failed to update revision'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Deactivate a revision
  const deactivateRevision = async (revisionId: string) => {
    loading.value = true
    error.value = null

    try {
      const { $fetch } = useNuxtApp()
      await $fetch(`/api/admin/revisions/${revisionId}`, {
        method: 'DELETE',
        credentials: 'include'
      })

      // Update the revision status in the local list
      const index = revisions.value.findIndex(r => r.id === revisionId)
      if (index !== -1) {
        revisions.value[index].is_active = false
      }

      return true
    } catch (err: any) {
      error.value = err.data?.error || 'Failed to deactivate revision'
      throw err
    } finally {
      loading.value = false
    }
  }

  // Clear all revisions
  const clearRevisions = () => {
    revisions.value = []
    error.value = null
  }

  return {
    revisions: readonly(revisions),
    loading: readonly(loading),
    error: readonly(error),
    fetchRevisions,
    createRevision,
    updateRevision,
    deactivateRevision,
    clearRevisions
  }
}
