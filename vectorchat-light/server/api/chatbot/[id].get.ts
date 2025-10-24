import { createError, defineEventHandler, getRouterParam } from "h3";
import { getVectorchatBaseUrl } from "../../utils/vectorchat-auth";
import { createUserAuthHeaders } from "../../utils/ory-session";

export default defineEventHandler(async (event) => {
  const chatbotId = getRouterParam(event, "id");
  const config = useRuntimeConfig();

  const apiUrl = getVectorchatBaseUrl(config);

  if (!chatbotId) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing chatbot ID",
    });
  }

  try {
    // Fetch chatbot details from VectorChat API
    const authHeaders = await createUserAuthHeaders(event);
    const response = await fetch(`${apiUrl}/chat/chatbot/${chatbotId}`, {
      headers: {
        ...authHeaders,
      },
    });

    if (!response.ok) {
      if (response.status === 404) {
        throw createError({
          statusCode: 404,
          statusMessage: "Chatbot not found",
        });
      }
      if (response.status === 401) {
        throw createError({
          statusCode: 401,
          statusMessage: "Authentication required",
        });
      }
      throw createError({
        statusCode: response.status,
        statusMessage: "Failed to fetch chatbot data",
      });
    }

    const chatbotData = await response.json();
    return chatbotData;
  } catch (error: any) {
    console.error("Error fetching chatbot:", error);

    if (error?.statusCode) {
      throw error;
    }

    throw createError({
      statusCode: 500,
      statusMessage: "Internal server error",
    });
  }
});
