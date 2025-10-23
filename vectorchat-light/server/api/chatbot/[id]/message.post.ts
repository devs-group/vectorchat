import {
  createError,
  defineEventHandler,
  getRouterParam,
  readBody,
} from "h3";
import {
  getVectorchatAccessToken,
  getVectorchatBaseUrl,
} from "../../../utils/vectorchat-auth";

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
    const accessToken = await getVectorchatAccessToken(config);
    // Send message to VectorChat API
    const response = await fetch(
      `${apiUrl}/chat/${chatbotId}/message`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${accessToken}`,
        },
        body: JSON.stringify({
          query: body.query,
          session_id: body.session_id || `preview-${Date.now()}`,
        }),
      },
    );

    if (!response.ok) {
      if (response.status === 404) {
        throw createError({
          statusCode: 404,
          statusMessage: "Chatbot not found",
        });
      }
      throw createError({
        statusCode: response.status,
        statusMessage: "Failed to send message",
      });
    }

    const responseData = await response.json();
    return responseData;
  } catch (error) {
    console.error("Error sending message:", error);

    if (error.statusCode) {
      throw error;
    }

    throw createError({
      statusCode: 500,
      statusMessage: "Internal server error",
    });
  }
});
