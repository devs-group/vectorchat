import type {
  ChatbotResponse,
  GenerateAPIKeyRequest,
  Plan,
  Subscription,
} from "~/types/api";

/**
 * Composable for the VectorChat API service
 * Provides methods for interacting with all endpoints
 */
export function useApiService() {
  // Auth endpoints
  const getSession = () => {
    return useApi(
      async () => {
        return await useApiFetch("/auth/session");
      },
      {
        errorMessage: "Failed to get session information",
      },
    );
  };

  const generateApiKey = <T>() => {
    return useApi(
      async (req: GenerateAPIKeyRequest) => {
        return await useApiFetch<T>("/auth/apikey", {
          method: "POST",
          body: req,
        });
      },
      {
        showSuccessToast: true,
        successMessage: "API key generated successfully",
      },
    );
  };

  const listApiKeys = <T>() => {
    return useApi(
      async (page: number = 1, limit: number = 10) => {
        const params = new URLSearchParams({
          page: page.toString(),
          limit: limit.toString(),
        });
        return await useApiFetch<T>(`/auth/apikey?${params}`);
      },
      {
        errorMessage: "Failed to fetch API keys",
      },
    );
  };

  const revokeApiKey = () => {
    return useApi(
      async (id: string) => {
        return await useApiFetch(`/auth/apikey/${id}`, {
          method: "DELETE",
        });
      },
      {
        showSuccessToast: true,
        successMessage: "API key revoked successfully",
      },
    );
  };

  const loginWithGithub = () => {
    // Open GitHub auth in browser window
    window.location.href = `${useRuntimeConfig().public.apiBase}/auth/github`;
  };

  const logout = () => {
    return useApi(
      async () => {
        return await useApiFetch("/auth/logout", {
          method: "POST",
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Logged out successfully",
      },
    );
  };

  const githubAuthCallback = (queryParams: string) => {
    return useApi(
      async () => {
        return await useApiFetch("/auth/github/callback?" + queryParams, {
          method: "GET",
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Logged in successfully",
      },
    );
  };

  // Chat endpoints
  const createChatbot = () => {
    return useApi(
      async (data: {
        name: string;
        description: string;
        model_name: string;
        system_instructions: string;
        max_tokens: number;
        temperature_param: number;
      }) => {
        return await useApiFetch("/chat/chatbot", {
          method: "POST",
          body: data,
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Chatbot created successfully",
      },
    );
  };

  // Conversations endpoints
  const listConversations = () => {
    return useApi(
      async (data: { chatbotId: string; limit: number; offset: number }) => {
        const params = new URLSearchParams({
          limit: String(data.limit),
          offset: String(data.offset),
        });
        return await useApiFetch<import("~/types/api").ConversationsResponse>(
          `/conversation/conversations/${data.chatbotId}?${params.toString()}`,
          { method: "GET" },
        );
      },
      { errorMessage: "Failed to fetch conversations" },
    );
  };

  const getConversationMessages = () => {
    return useApi(
      async (data: { chatbotId: string; sessionId: string }) => {
        return await useApiFetch<{
          messages: import("~/types/api").MessageDetails[];
        }>(`/conversation/conversations/${data.chatbotId}/${data.sessionId}`, {
          method: "GET",
        });
      },
      { errorMessage: "Failed to fetch conversation messages" },
    );
  };

  // Revisions endpoints
  const getRevisions = () => {
    return useApi(
      async (data: { chatbotId: string; includeInactive?: boolean }) => {
        const params = new URLSearchParams();
        if (data.includeInactive) params.set("includeInactive", "true");
        return await useApiFetch<import("~/types/api").RevisionsListResponse>(
          `/conversation/revisions/${data.chatbotId}${params.toString() ? `?${params.toString()}` : ""}`,
          { method: "GET" },
        );
      },
      { errorMessage: "Failed to fetch revisions" },
    );
  };

  const createRevision = () => {
    return useApi(
      async (req: import("~/types/api").CreateRevisionRequest) => {
        return await useApiFetch<import("~/types/api").RevisionResponse>(
          "/conversation/revisions",
          {
            method: "POST",
            body: req,
          },
        );
      },
      { showSuccessToast: true, successMessage: "Revision saved" },
    );
  };

  const updateRevision = () => {
    return useApi(
      async (data: {
        revisionId: string;
        body: import("~/types/api").UpdateRevisionRequest;
      }) => {
        return await useApiFetch(`/conversation/revisions/${data.revisionId}`, {
          method: "PUT",
          body: data.body,
        });
      },
      { showSuccessToast: true, successMessage: "Revision updated" },
    );
  };

  const deleteRevision = () => {
    return useApi(
      async (revisionId: string) => {
        return await useApiFetch(`/conversation/revisions/${revisionId}`, {
          method: "DELETE",
        });
      },
      { showSuccessToast: true, successMessage: "Revision canceled" },
    );
  };

  const listChatbots = () => {
    return useApi(
      async () => {
        return await useApiFetch<{ chatbots: ChatbotResponse[] }>(
          "/chat/chatbots",
          {
            method: "GET",
          },
        );
      },
      {
        showSuccessToast: false,
      },
    );
  };

  const getChatbot = () => {
    return useApi(
      async (chatbotId: string) => {
        return await useApiFetch<{ chatbot: ChatbotResponse }>(
          `/chat/chatbot/${chatbotId}`,
          {
            method: "GET",
          },
        );
      },
      {
        errorMessage: "Failed to fetch chatbot details",
      },
    );
  };

  const updateChatbot = () => {
    return useApi(
      async (chatbotData: {
        id: string;
        name?: string;
        description?: string;
        system_instructions?: string;
        model_name?: string;
        temperature_param?: number;
        max_tokens?: number;
      }) => {
        return await useApiFetch<{ chatbot: ChatbotResponse }>(
          `/chat/chatbot/${chatbotData.id}`,
          {
            method: "PUT",
            body: chatbotData,
          },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Chatbot updated",
        errorMessage: "Failed to update chatbot",
      },
    );
  };

  const toggleChatbot = () => {
    return useApi(
      async (data: { chatbotId: string; isEnabled: boolean }) => {
        return await useApiFetch(`/chat/chatbot/${data.chatbotId}/toggle`, {
          method: "PATCH",
          body: {
            is_enabled: data.isEnabled,
          },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Chatbot updated",
        errorMessage: "Failed to toggle chatbot state",
      },
    );
  };

  const deleteChatbot = () => {
    return useApi(
      async (chatbotId: string) => {
        return await useApiFetch(`/chat/chatbot/${chatbotId}`, {
          method: "DELETE",
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Chatbot deleted successfully",
      },
    );
  };

  const sendChatMessage = (
    chatID: string,
    query: string,
    sessionId?: string | null,
  ) => {
    return useApi(async () => {
      const body: { query: string; session_id?: string } = { query };
      if (sessionId) {
        body.session_id = sessionId;
      }
      return await useApiFetch(`/chat/${chatID}/message`, {
        method: "POST",
        body,
      });
    });
  };

  const uploadFile = async (chatID: string, file: File) => {
    const config = useRuntimeConfig();
    const formData = new FormData();
    formData.append("file", file);
    try {
      return await $fetch(`/chat/${chatID}/upload`, {
        baseURL: config.public.apiBase as string,
        method: "POST",
        body: formData,
        credentials: "include",
      });
    } catch (error: any) {
      // If the error has a response with data, throw that instead
      // This ensures the backend error message is preserved
      if (error.data) {
        throw error.data;
      }
      throw error;
    }
  };

  const uploadText = (chatID: string, text: string) => {
    return useApi(
      async () => {
        return await useApiFetch(`/chat/${chatID}/text`, {
          method: "POST",
          body: { text },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Text added successfully",
      },
    );
  };

  const uploadWebsite = (chatID: string, url: string) => {
    return useApi(
      async () => {
        return await useApiFetch(`/chat/${chatID}/website`, {
          method: "POST",
          body: { url },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Website indexing started",
      },
    );
  };

  const updateFile = (chatID: string, filename: string, file: File) => {
    return useApi(
      async () => {
        const formData = new FormData();
        formData.append("file", file);

        return await useApiFetch(`/chat/${chatID}/files/${filename}`, {
          method: "PUT",
          body: formData,
        });
      },
      {
        showSuccessToast: true,
        successMessage: "File updated successfully",
      },
    );
  };

  const deleteFile = (chatID: string, filename: string) => {
    return useApi(
      async () => {
        return await useApiFetch(`/chat/${chatID}/files/${filename}`, {
          method: "DELETE",
        });
      },
      {
        showSuccessToast: true,
        successMessage: "File deleted successfully",
      },
    );
  };

  const listChatFiles = (chatID: string) => {
    return useApi(
      async () => {
        return await useApiFetch(`/chat/${chatID}/files`);
      },
      {
        errorMessage: "Failed to fetch chat files",
        cacheKey: `chatFiles-${chatID}`,
      },
    );
  };

  const listTextSources = (chatID: string) => {
    return useApi(
      async () => {
        return await useApiFetch(`/chat/${chatID}/text`);
      },
      {
        errorMessage: "Failed to fetch text sources",
        cacheKey: `textSources-${chatID}`,
      },
    );
  };

  const deleteTextSource = (chatID: string, id: string) => {
    return useApi(
      async () => {
        return await useApiFetch(`/chat/${chatID}/text/${id}`, {
          method: "DELETE",
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Text source deleted successfully",
      },
    );
  };

  // Health check
  const healthCheck = () => {
    return useApi(async () => {
      return await useApiFetch("/health");
    });
  };

  // Billing
  const listPlans = () => {
    return useApi(
      async () => {
        return await useApiFetch<Plan[]>("/billing/plans", { method: "GET" });
      },
      { errorMessage: "Failed to fetch plans" },
    );
  };

  const createCheckoutSession = () => {
    return useApi(
      async (body: {
        customer_id: string;
        plan_key: string;
        success_url: string;
        cancel_url: string;
        allow_promotion_codes?: boolean;
        idempotency_key?: string;
        metadata?: Record<string, string>;
      }) => {
        return await useApiFetch<{ session_id: string; url: string }>(
          "/billing/checkout-session",
          {
            method: "POST",
            body,
          },
        );
      },
      { errorMessage: "Failed to create checkout session" },
    );
  };

  const getSubscription = () => {
    return useApi(
      async () => {
        return await useApiFetch<{ subscription: Subscription | null }>(
          "/billing/subscription",
          { method: "GET" },
        );
      },
      { errorMessage: "Failed to fetch subscription" },
    );
  };

  const createPortalSession = () => {
    return useApi(
      async (body: { return_url: string }) => {
        return await useApiFetch<{ url: string }>("/billing/portal-session", {
          method: "POST",
          body,
        });
      },
      { errorMessage: "Failed to open billing portal" },
    );
  };

  return {
    // Auth
    getSession,
    generateApiKey,
    listApiKeys,
    revokeApiKey,
    loginWithGithub,
    logout,
    githubAuthCallback,

    // Chat
    createChatbot,
    listChatbots,
    getChatbot,
    updateChatbot,
    toggleChatbot,
    deleteChatbot,
    sendChatMessage,
    uploadFile,
    uploadText,
    uploadWebsite,
    updateFile,
    deleteFile,
    listChatFiles,
    listTextSources,
    deleteTextSource,

    // Conversations
    listConversations,
    getConversationMessages,
    getRevisions,
    createRevision,
    updateRevision,
    deleteRevision,

    // Health
    healthCheck,

    // Billing
    listPlans,
    createCheckoutSession,
    getSubscription,
    createPortalSession,
  };
}
