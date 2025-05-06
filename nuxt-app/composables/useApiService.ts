import type { GenerateAPIKeyRequest } from "~/types/api";

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
          body: req
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
      async () => {
        return await useApiFetch<T>("/auth/apikey");
      },
      {
        errorMessage: "Failed to fetch API keys",
        cacheKey: "apiKeys",
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

  const sendChatMessage = (chatID: string, query: string) => {
    return useApi(async () => {
      return await useApiFetch(`/chat/${chatID}/message`, {
        method: "POST",
        body: { query },
      });
    });
  };

  const uploadFile = (chatID: string, file: File) => {
    return useApi(
      async () => {
        const formData = new FormData();
        formData.append("file", file);

        return await useApiFetch(`/chat/${chatID}/upload`, {
          method: "POST",
          body: formData,
        });
      },
      {
        showSuccessToast: true,
        successMessage: "File uploaded successfully",
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

  // User endpoints
  const getUserInfo = () => {
    return useApi(
      async () => {
        return await useApiFetch("/");
      },
      {
        cacheKey: "userInfo",
      },
    );
  };

  // Health check
  const healthCheck = () => {
    return useApi(async () => {
      return await useApiFetch("/health");
    });
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
    sendChatMessage,
    uploadFile,
    updateFile,
    deleteFile,
    listChatFiles,

    // User
    getUserInfo,

    // Health
    healthCheck,
  };
}
