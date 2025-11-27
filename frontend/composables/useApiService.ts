import type {
  ChatbotResponse,
  GenerateAPIKeyRequest,
  Plan,
  Subscription,
  CreateRevisionRequest,
  ConversationsResponse,
  MessageDetails,
  RevisionsListResponse,
  RevisionResponse,
  UpdateRevisionRequest,
  SharedKnowledgeBase,
  SharedKnowledgeBaseCreateRequest,
  SharedKnowledgeBaseUpdateRequest,
  SharedKnowledgeBaseListResponse,
  SharedKnowledgeBaseFileUploadResponse,
  SharedKnowledgeBaseFilesResponse,
  SharedKnowledgeBaseTextSourcesResponse,
  CrawlSchedule,
  CrawlScheduleListResponse,
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
        save_messages: boolean;
        use_max_tokens?: boolean;
        is_enabled?: boolean;
        shared_knowledge_base_ids?: string[];
      }) => {
        return await useApiFetch<{ chatbot: ChatbotResponse }>(
          "/chat/chatbot",
          {
            method: "POST",
            body: data,
          },
        );
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
      async (data: {
        chatbotId: string;
        page?: number;
        limit?: number;
        offset?: number;
      }) => {
        const params = new URLSearchParams();
        if (data.limit !== undefined) params.set("limit", String(data.limit));
        if (data.page !== undefined) {
          params.set("page", String(data.page));
        } else if (data.offset !== undefined) {
          params.set("offset", String(data.offset));
        }

        const query = params.toString();
        return await useApiFetch<ConversationsResponse>(
          `/conversation/conversations/${data.chatbotId}${
            query ? `?${query}` : ""
          }`,
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
          messages: MessageDetails[];
        }>(`/conversation/conversations/${data.chatbotId}/${data.sessionId}`, {
          method: "GET",
        });
      },
      { errorMessage: "Failed to fetch conversation messages" },
    );
  };

  const deleteConversation = () => {
    return useApi(
      async (data: { chatbotId: string; sessionId: string }) => {
        return await useApiFetch(
          `/conversation/conversations/${data.chatbotId}/${data.sessionId}`,
          { method: "DELETE" },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Conversation deleted",
        errorMessage: "Failed to delete conversation",
      },
    );
  };

  // Revisions endpoints
  const getRevisions = () => {
    return useApi(
      async (data: { chatbotId: string; includeInactive?: boolean }) => {
        const params = new URLSearchParams();
        if (data.includeInactive) params.set("includeInactive", "true");
        return await useApiFetch<RevisionsListResponse>(
          `/conversation/revisions/${data.chatbotId}${params.toString() ? `?${params.toString()}` : ""}`,
          { method: "GET" },
        );
      },
      { errorMessage: "Failed to fetch revisions" },
    );
  };

  const createRevision = () => {
    return useApi(
      async (req: CreateRevisionRequest) => {
        return await useApiFetch<RevisionResponse>("/conversation/revisions", {
          method: "POST",
          body: req,
        });
      },
      { showSuccessToast: true, successMessage: "Revision saved" },
    );
  };

  const updateRevision = () => {
    return useApi(
      async (data: { revisionId: string; body: UpdateRevisionRequest }) => {
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
        save_messages?: boolean;
        use_max_tokens?: boolean;
        shared_knowledge_base_ids?: string[];
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

  // Shared knowledge bases
  const listSharedKnowledgeBases = () => {
    return useApi(
      async () => {
        return await useApiFetch<SharedKnowledgeBaseListResponse>(
          "/knowledge-bases",
          {
            method: "GET",
          },
        );
      },
      { showSuccessToast: false },
    );
  };

  const createSharedKnowledgeBase = () => {
    return useApi(
      async (body: SharedKnowledgeBaseCreateRequest) => {
        return await useApiFetch<SharedKnowledgeBase>("/knowledge-bases", {
          method: "POST",
          body,
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Knowledge base created",
        errorMessage: "Failed to create knowledge base",
      },
    );
  };

  const getSharedKnowledgeBase = () => {
    return useApi(
      async (kbId: string) => {
        return await useApiFetch<SharedKnowledgeBase>(
          `/knowledge-bases/${kbId}`,
          {
            method: "GET",
          },
        );
      },
      { errorMessage: "Failed to fetch knowledge base" },
    );
  };

  const updateSharedKnowledgeBase = () => {
    return useApi(
      async (data: { id: string; body: SharedKnowledgeBaseUpdateRequest }) => {
        return await useApiFetch<SharedKnowledgeBase>(
          `/knowledge-bases/${data.id}`,
          {
            method: "PUT",
            body: data.body,
          },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Knowledge base updated",
        errorMessage: "Failed to update knowledge base",
      },
    );
  };

  const deleteSharedKnowledgeBase = () => {
    return useApi(
      async (kbId: string) => {
        return await useApiFetch(`/knowledge-bases/${kbId}`, {
          method: "DELETE",
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Knowledge base deleted",
        errorMessage: "Failed to delete knowledge base",
      },
    );
  };

  const uploadSharedKnowledgeBaseFile = async (
    kbId: string,
    file: File,
  ): Promise<SharedKnowledgeBaseFileUploadResponse> => {
    const config = useRuntimeConfig();
    const formData = new FormData();
    formData.append("file", file);

    return await $fetch(`/knowledge-bases/${kbId}/upload`, {
      baseURL: config.public.apiBase as string,
      method: "POST",
      body: formData,
      credentials: "include",
    });
  };

  const uploadSharedKnowledgeBaseText = () => {
    return useApi(
      async (data: { kbId: string; text: string }) => {
        return await useApiFetch(`/knowledge-bases/${data.kbId}/text`, {
          method: "POST",
          body: { text: data.text },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Text added successfully",
      },
    );
  };

  const uploadSharedKnowledgeBaseWebsite = () => {
    return useApi(
      async (data: { kbId: string; url: string }) => {
        return await useApiFetch(`/knowledge-bases/${data.kbId}/website`, {
          method: "POST",
          body: { url: data.url },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Website indexing started",
      },
    );
  };

  const listSharedKnowledgeBaseFiles = (kbId: string) => {
    return useApi(
      async () => {
        return await useApiFetch<SharedKnowledgeBaseFilesResponse>(
          `/knowledge-bases/${kbId}/files`,
          {
            method: "GET",
          },
        );
      },
      { errorMessage: "Failed to fetch files" },
    );
  };

  const deleteSharedKnowledgeBaseFile = () => {
    return useApi(
      async (data: { kbId: string; filename: string }) => {
        return await useApiFetch(
          `/knowledge-bases/${data.kbId}/files/${data.filename}`,
          {
            method: "DELETE",
          },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "File deleted successfully",
      },
    );
  };

  const listSharedKnowledgeBaseTextSources = (kbId: string) => {
    return useApi(
      async () => {
        return await useApiFetch<SharedKnowledgeBaseTextSourcesResponse>(
          `/knowledge-bases/${kbId}/text`,
          {
            method: "GET",
          },
        );
      },
      { errorMessage: "Failed to fetch text sources" },
    );
  };

  const deleteSharedKnowledgeBaseTextSource = () => {
    return useApi(
      async (data: { kbId: string; id: string }) => {
        return await useApiFetch(
          `/knowledge-bases/${data.kbId}/text/${data.id}`,
          {
            method: "DELETE",
          },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Text source deleted successfully",
      },
    );
  };

  // Crawl schedules (chat)
  const listChatCrawlSchedules = () => {
    return useApi(
      async (chatId: string) => {
        return await useApiFetch<CrawlScheduleListResponse>(
          `/chat/${chatId}/crawl-schedules`,
          { method: "GET" },
        );
      },
      { errorMessage: "Failed to load crawl schedules" },
    );
  };

  const upsertChatCrawlSchedule = () => {
    return useApi(
      async (data: { chatId: string; body: Partial<CrawlSchedule> }) => {
        return await useApiFetch<CrawlSchedule>(
          `/chat/${data.chatId}/crawl-schedules`,
          { method: "PUT", body: data.body },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Schedule saved",
        errorMessage: "Failed to save schedule",
      },
    );
  };

  const deleteChatCrawlSchedule = () => {
    return useApi(
      async (data: { chatId: string; scheduleId: string }) => {
        return await useApiFetch(
          `/chat/${data.chatId}/crawl-schedules/${data.scheduleId}`,
          { method: "DELETE" },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Schedule deleted",
        errorMessage: "Failed to delete schedule",
      },
    );
  };

  // Crawl schedules (shared KB)
  const listSharedCrawlSchedules = () => {
    return useApi(
      async (kbId: string) => {
        return await useApiFetch<CrawlScheduleListResponse>(
          `/knowledge-bases/${kbId}/crawl-schedules`,
          { method: "GET" },
        );
      },
      { errorMessage: "Failed to load crawl schedules" },
    );
  };

  const upsertSharedCrawlSchedule = () => {
    return useApi(
      async (data: { kbId: string; body: Partial<CrawlSchedule> }) => {
        return await useApiFetch<CrawlSchedule>(
          `/knowledge-bases/${data.kbId}/crawl-schedules`,
          { method: "PUT", body: data.body },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Schedule saved",
        errorMessage: "Failed to save schedule",
      },
    );
  };

  const deleteSharedCrawlSchedule = () => {
    return useApi(
      async (data: { kbId: string; scheduleId: string }) => {
        return await useApiFetch(
          `/knowledge-bases/${data.kbId}/crawl-schedules/${data.scheduleId}`,
          { method: "DELETE" },
        );
      },
      {
        showSuccessToast: true,
        successMessage: "Schedule deleted",
        errorMessage: "Failed to delete schedule",
      },
    );
  };

  const crawlSharedOnce = () => {
    return useApi(
      async (data: { kbId: string; url: string }) => {
        return await useApiFetch(`/knowledge-bases/${data.kbId}/crawl-now`, {
          method: "POST",
          body: { url: data.url },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Crawl enqueued",
        errorMessage: "Failed to enqueue crawl",
      },
    );
  };

  const crawlChatOnce = () => {
    return useApi(
      async (data: { chatId: string; url: string }) => {
        return await useApiFetch(`/chat/${data.chatId}/crawl-now`, {
          method: "POST",
          body: { url: data.url },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Crawl enqueued",
        errorMessage: "Failed to enqueue crawl",
      },
    );
  };

  const getCrawlQueueMetrics = () => {
    return useApi(
      async () => {
        return await useApiFetch(`/queue/crawl/metrics`, { method: "GET" });
      },
      {
        errorMessage: "Failed to load crawl queue metrics",
      },
    );
  };

  const sendChatMessage = () => {
    return useApi(
      async (data: {
        chatID: string;
        query: string;
        sessionId?: string | null;
      }) => {
        const body = {
          query: data.query,
          ...(data.sessionId && { session_id: data.sessionId }),
        };
        return await useApiFetch(`/chat/${data.chatID}/message`, {
          method: "POST",
          body,
        });
      },
    );
  };

  const streamChatMessage = (
    data: {
      chatID: string;
      query: string;
      sessionId?: string | null;
    },
    handlers: {
      onChunk?: (chunk: string) => void | Promise<void>;
      onDone?: (payload: {
        content: string;
        sessionId?: string;
      }) => void | Promise<void>;
      onError?: (payload: { message: string }) => void | Promise<void>;
    },
  ) => {
    const controller = new AbortController();
    const config = useRuntimeConfig();

    const body = {
      query: data.query,
      ...(data.sessionId && { session_id: data.sessionId }),
    };

    const processEvent = async (rawEvent: string) => {
      const dataLines = rawEvent
        .split("\n")
        .filter((line) => line.startsWith("data:"))
        .map((line) => line.slice(5).trim());

      if (dataLines.length === 0) {
        return;
      }

      const payloadRaw = dataLines.join("");

      try {
        const payload = JSON.parse(payloadRaw) as {
          type: string;
          content?: string;
          session_id?: string;
          error?: string;
        };

        if (payload.type === "chunk" && payload.content && handlers.onChunk) {
          await handlers.onChunk(payload.content);
        } else if (payload.type === "done" && handlers.onDone) {
          await handlers.onDone({
            content: payload.content ?? "",
            sessionId: payload.session_id,
          });
        } else if (payload.type === "error" && handlers.onError) {
          await handlers.onError({
            message: payload.error ?? "Stream error",
          });
        }
      } catch (error) {
        console.error("Failed to parse stream payload", error, rawEvent);
      }
    };

    (async () => {
      try {
        const response = await fetch(
          `${config.public.apiBase as string}/chat/${data.chatID}/stream-message`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify(body),
            credentials: "include",
            signal: controller.signal,
          },
        );

        if (!response.ok) {
          let message = `Stream request failed (${response.status})`;
          try {
            const errorBody = await response.json();
            message = errorBody?.error ?? message;
          } catch (parseError) {
            // ignore, use default message
          }
          if (handlers.onError) {
            await handlers.onError({ message });
          }
          return;
        }

        if (!response.body) {
          if (handlers.onError) {
            await handlers.onError({
              message: "Stream response body is empty",
            });
          }
          return;
        }

        const reader = response.body.getReader();
        const decoder = new TextDecoder();
        let buffer = "";

        while (true) {
          const { value, done } = await reader.read();
          if (done) {
            buffer += decoder.decode();
            break;
          }
          buffer += decoder.decode(value, { stream: true });

          let separatorIndex = buffer.indexOf("\n\n");
          while (separatorIndex !== -1) {
            const rawEvent = buffer.slice(0, separatorIndex);
            buffer = buffer.slice(separatorIndex + 2);
            await processEvent(rawEvent.trim());
            separatorIndex = buffer.indexOf("\n\n");
          }
        }

        if (buffer.trim() !== "") {
          await processEvent(buffer.trim());
        }
      } catch (error: any) {
        if (controller.signal.aborted) {
          return;
        }
        console.error("Stream error", error);
        if (handlers.onError) {
          await handlers.onError({
            message: error?.message ?? "Stream connection failed",
          });
        }
      }
    })();

    return () => controller.abort();
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

  const uploadText = () => {
    return useApi(
      async (data: { chatID: string; text: string }) => {
        return await useApiFetch(`/chat/${data.chatID}/text`, {
          method: "POST",
          body: { text: data.text },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Text added successfully",
      },
    );
  };

  const uploadWebsite = () => {
    return useApi(
      async (data: { chatID: string; url: string }) => {
        return await useApiFetch(`/chat/${data.chatID}/website`, {
          method: "POST",
          body: { url: data.url },
        });
      },
      {
        showSuccessToast: true,
        successMessage: "Website indexing started",
      },
    );
  };

  const deleteFile = () => {
    return useApi(
      async (data: { chatID: string; filename: string }) => {
        return await useApiFetch(
          `/chat/${data.chatID}/files/${data.filename}`,
          {
            method: "DELETE",
          },
        );
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

  const deleteTextSource = () => {
    return useApi(
      async (data: { chatID: string; id: string }) => {
        return await useApiFetch(`/chat/${data.chatID}/text/${data.id}`, {
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
        return await useApiFetch<Plan[]>("/public/billing/plans", {
          method: "GET",
        });
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
        return await useApiFetch<{
          subscription: Subscription | null;
          plan: Plan | null;
        }>("/billing/subscription", { method: "GET" });
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
    logout,

    // Chat
    createChatbot,
    listChatbots,
    getChatbot,
    updateChatbot,
    toggleChatbot,
    deleteChatbot,
    sendChatMessage,
    streamChatMessage,
    uploadFile,
    uploadText,
    uploadWebsite,
    deleteFile,
    listChatFiles,
    listTextSources,
    deleteTextSource,

    // Shared knowledge bases
    listSharedKnowledgeBases,
    createSharedKnowledgeBase,
    getSharedKnowledgeBase,
    updateSharedKnowledgeBase,
    deleteSharedKnowledgeBase,
    uploadSharedKnowledgeBaseFile,
    uploadSharedKnowledgeBaseText,
    uploadSharedKnowledgeBaseWebsite,
    listSharedKnowledgeBaseFiles,
    deleteSharedKnowledgeBaseFile,
    listSharedKnowledgeBaseTextSources,
    deleteSharedKnowledgeBaseTextSource,
    listChatCrawlSchedules,
    upsertChatCrawlSchedule,
    deleteChatCrawlSchedule,
    listSharedCrawlSchedules,
    upsertSharedCrawlSchedule,
    deleteSharedCrawlSchedule,
    crawlSharedOnce,
    crawlChatOnce,
    getCrawlQueueMetrics,

    // Conversations
    listConversations,
    getConversationMessages,
    deleteConversation,
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
