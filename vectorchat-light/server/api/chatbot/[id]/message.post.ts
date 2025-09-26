import { createError, defineEventHandler, getRouterParam, readBody } from "h3";

type ChatMessageRequest = {
  query: string;
  session_id?: string;
};

export default defineEventHandler(async (event) => {
  const chatbotId = getRouterParam(event, "id");
  const body = await readBody<ChatMessageRequest>(event);
  const config = useRuntimeConfig();

  const apiKey = config.vectorchatApiKey as string;
  const apiUrl = config.vectorchatUrl as string;

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
    // Send message to VectorChat API
    const response = await fetch(
      `${apiUrl || "http://localhost:8080"}/chat/${chatbotId}/message`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-API-Key": apiKey || "",
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
