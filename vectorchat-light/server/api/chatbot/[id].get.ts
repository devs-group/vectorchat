import { createError, defineEventHandler, getRouterParam } from "h3";
import {
  getVectorchatAccessToken,
  getVectorchatBaseUrl,
} from "../../utils/vectorchat-auth";

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
    const accessToken = await getVectorchatAccessToken(config);
    const response = await fetch(
      `${apiUrl}/chat/chatbot/${chatbotId}`,
      {
        headers: {
          Authorization: `Bearer ${accessToken}`,
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
