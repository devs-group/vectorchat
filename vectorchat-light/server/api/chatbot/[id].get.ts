import { createError, defineEventHandler, getRouterParam } from "h3";

export default defineEventHandler(async (event) => {
  const chatbotId = getRouterParam(event, "id");
  const config = useRuntimeConfig();

  const apiKey = config.vectorchatApiKey as string;
  const apiUrl = config.vectorchatUrl as string;

  if (!chatbotId) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing chatbot ID",
    });
  }

  try {
    // Fetch chatbot details from VectorChat API
    const response = await fetch(
      `${apiUrl || "http://localhost:8080"}/chat/chatbot/${chatbotId}`,
      {
        headers: {
          "X-API-Key": apiKey || "",
        },
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
        statusMessage: "Failed to fetch chatbot data",
      });
    }

    const chatbotData = await response.json();
    return chatbotData;
  } catch (error) {
    console.error("Error fetching chatbot:", error);

    if (error.statusCode) {
      throw error;
    }

    throw createError({
      statusCode: 500,
      statusMessage: "Internal server error",
    });
  }
});
