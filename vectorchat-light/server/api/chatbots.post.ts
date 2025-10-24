import { createError, defineEventHandler, readBody } from "h3";
import { getVectorchatBaseUrl } from "../utils/vectorchat-auth";
import { createUserAuthHeaders } from "../utils/ory-session";

type GenerateChatbotRequest = {
  siteUrl?: string;
};

type GenerateChatbotResponse = {
  chatbotId: string;
  siteUrl: string;
  previewUrl: string;
  message: string;
};

export default defineEventHandler(async (event) => {
  const body = await readBody<GenerateChatbotRequest | null>(event);
  const siteUrl = body?.siteUrl?.trim();
  const config = useRuntimeConfig();

  const apiUrl = getVectorchatBaseUrl(config);

  if (!siteUrl) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing required field: siteUrl",
    });
  }

  try {
    const authHeaders = await createUserAuthHeaders(event);

    // Step 1: Create a new chatbot with website assistant system prompt
    const createChatbotResponse = await fetch(`${apiUrl}/chat/chatbot`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...authHeaders,
      },
      body: JSON.stringify({
        name: `VC Lite Assistant: ${new URL(siteUrl).hostname}`,
        description:
          "AI assistant trained on website content to help visitors navigate and find information",
        system_instructions: `You are a helpful website assistant. Answer questions using only information from this website. Be conversational and direct. If you don't find specific information on the site, let users know and suggest contacting the website directly for more details.`,
        model_name: "gpt-5-nano",
        max_tokens: 1000,
        temperature_param: 0.7,
        save_messages: true,
        is_enabled: true,
        shared_knowledge_base_ids: [],
      }),
    });

    if (!createChatbotResponse.ok) {
      const errorBody = await safeParseJSON(createChatbotResponse);
      throw createError({
        statusCode: errorBody.statusCode,
        message: errorBody.code,
      });
    }

    const chatbotData = await createChatbotResponse.json();
    const actualChatbotId = chatbotData.id;

    // Step 2: Upload the website to index for this chatbot
    const websiteUploadResponse = await fetch(
      `${apiUrl}/chat/${actualChatbotId}/website`,
      {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          ...authHeaders,
        },
        body: JSON.stringify({
          url: siteUrl,
        }),
      },
    );

    if (!websiteUploadResponse.ok) {
      // If website upload fails, we should clean up the created chatbot
      try {
        await fetch(`${apiUrl}/chat/chatbot/${actualChatbotId}`, {
          method: "DELETE",
          headers: {
            ...authHeaders,
          },
        });
      } catch (cleanupError) {
        console.error(
          "Failed to cleanup chatbot after website upload failure:",
          cleanupError,
        );
      }

      const errorBody = await safeParseJSON(websiteUploadResponse);
      throw createError({
        statusCode: errorBody.statusCode,
        message: errorBody.code,
      });
    }

    // Step 3: Return successful response
    const response: GenerateChatbotResponse = {
      chatbotId: actualChatbotId,
      siteUrl,
      previewUrl: `/preview/${encodeURIComponent(actualChatbotId)}`,
      message: "Chatbot created successfully with website content indexed.",
    };

    return response;
  } catch (error: any) {
    if (error) {
      throw createError({
        statusCode: error.statusCode,
        message: error.message,
      });
    }
  }

  async function safeParseJSON(response: Response) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }
});
