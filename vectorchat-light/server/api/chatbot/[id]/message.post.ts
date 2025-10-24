import { createError, defineEventHandler, getRouterParam, readBody } from "h3";
import { getVectorchatBaseUrl } from "../../../utils/vectorchat-auth";
import { createUserAuthHeaders } from "../../../utils/ory-session";

type ChatMessageRequest = {
  query: string;
  session_id?: string;
};

export default defineEventHandler(async (event) => {
  const chatbotId = getRouterParam(event, "id");
  const body = await readBody<ChatMessageRequest>(event);
  const config = useRuntimeConfig();

  const apiUrl = getVectorchatBaseUrl(config);

  if (!chatbotId) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing chatbot ID",
    });
  }

  if (!body?.query) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing required field: query",
    });
  }

  try {
    const authHeaders = await createUserAuthHeaders(event);
    // Send message to VectorChat API
    const response = await fetch(`${apiUrl}/chat/${chatbotId}/message`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...authHeaders,
      },
      body: JSON.stringify({
        query: body.query,
        session_id: body.session_id,
      }),
    });

    if (!response.ok) {
      const errorBody = await safeParseJSON(response);
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
      console.error("[vectorchat-light] Chat message failed", {
        status: response.status,
        statusText: response.statusText,
        errorBody,
      });
      throw createError({
        statusCode: response.status,
        statusMessage: "Failed to send message",
      });
    }

    const responseData = await response.json();
    return responseData;
  } catch (error: any) {
    console.error("Error sending message:", error);

    if (error?.statusCode) {
      throw error;
    }

    throw createError({
      statusCode: 500,
      statusMessage: "Internal server error",
    });
  }

  async function safeParseJSON(response: Response) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }
});
