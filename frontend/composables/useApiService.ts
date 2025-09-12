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
  const createChatbot = (chatbotData: {
    name: string;
    description: string;
    model_name: string;
    system_instructions: string;
    max_tokens: number;
    temperature_param: number;
  }) => {
    return useApi(
      async () => {
        return await useApiFetch("/chat/chatbot", {
          method: "POST",
          body: chatbotData,
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Chatbot created successfully",
      },
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
        showSuccessToast: true,
        successMessage: "Chatbots retrieved successfully",
      },
    );
  };

  const getChatbot = (chatbotId: string) => {
    return useApi(
      async () => {
        return await useApiFetch<{ chatbot: ChatbotResponse }>(
          `/chat/chatbot/${chatbotId}`,
          {
            method: "GET",
          },
        );
      },
      {
        errorMessage: "Failed to fetch chatbot details",
        cacheKey: `chatbot-${chatbotId}`,
      },
    );
  };

  const updateChatbot = (
    chatbotId: string,
    chatbotData: {
      name?: string;
      description?: string;
      system_instructions?: string;
      model_name?: string;
      temperature_param?: number;
      max_tokens?: number;
    },
  ) => {
    return useApi(
      async () => {
        return await useApiFetch<{ chatbot: ChatbotResponse }>(
          `/chat/chatbot/${chatbotId}`,
          {
            method: "PUT",
            body: chatbotData,
          },
        );
      },
      {
        errorMessage: "Failed to update chatbot",
      },
    );
  };

  const toggleChatbot = (chatbotId: string, isEnabled: boolean) => {
    return useApi(
      async () => {
        return await useApiFetch(`/chat/chatbot/${chatbotId}/toggle`, {
          method: "PATCH",
          body: {
            is_enabled: isEnabled,
          },
        });
      },
      {
        showSuccessToast: true,
        successMessage: isEnabled ? "Chatbot enabled" : "Chatbot disabled",
        errorMessage: "Failed to toggle chatbot state",
      },
    );
  };

  const deleteChatbot = (chatbotId: string) => {
    return useApi(
      async () => {
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

  const sendChatMessage = (chatID: string, query: string) => {
    return useApi(async () => {
      return await useApiFetch(`/chat/${chatID}/message`, {
        method: "POST",
        body: { query },
      });
    });
  };

  const uploadFile = async (chatID: string, file: File) => {
    const config = useRuntimeConfig();
    const formData = new FormData();
    formData.append("file", file);
    return await $fetch(`/chat/${chatID}/upload`, {
      baseURL: config.public.apiBase as string,
      method: "POST",
      body: formData,
      credentials: "include",
    });
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

    // Health
    healthCheck,

    // Billing
    listPlans,
    createCheckoutSession,
    getSubscription,
    createPortalSession,
  };
}
