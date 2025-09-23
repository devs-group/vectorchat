import { createError, defineEventHandler, readBody } from "h3";

type GenerateChatbotRequest = {
  siteUrl?: string;
};

type GenerateChatbotResponse = {
  chatbotId: string;
  siteUrl: string;
  previewUrl: string;
  message: string;
};

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms));

export default defineEventHandler(async (event) => {
  const body = await readBody<GenerateChatbotRequest | null>(event);
  const siteUrl = body?.siteUrl?.trim();
  const config = useRuntimeConfig();

  const apiKey = config.vectorchatApiKey as string;
  const apiUrl = config.vectorchatUrl as string;

  if (!siteUrl) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing required field: siteUrl",
    });
  }

  try {
    // Step 1: Create a new chatbot with website assistant system prompt
    const createChatbotResponse = await fetch(`${apiUrl}/chat/chatbot`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "X-API-Key": apiKey,
      },
      body: JSON.stringify({
        name: `VC Lite Assistant: ${new URL(siteUrl).hostname}`,
        description:
          "AI assistant trained on website content to help visitors navigate and find information",
        system_instructions: `You are a helpful website assistant for ${siteUrl}. Your role is to help visitors navigate the website, find information, and answer questions based on the content you have been trained on.

Key guidelines:
- Be friendly, professional, and helpful
- Provide accurate information based on the website content
- If you don't know something, direct users to contact the site owner or suggest relevant pages
- Help users find what they're looking for quickly
- Use a conversational tone while remaining informative
- Always stay focused on helping with website-related queries`,
        model_name: "gpt-3.5-turbo",
        max_tokens: 1000,
        temperature_param: 0.7,
        save_messages: true,
        shared_knowledge_base_ids: [],
      }),
    });

    if (!createChatbotResponse.ok) {
      throw new Error(
        `Failed to create chatbot: ${createChatbotResponse.statusText}`,
      );
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
          "X-API-Key": apiKey,
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
            "X-API-Key": apiKey,
          },
        });
      } catch (cleanupError) {
        console.error(
          "Failed to cleanup chatbot after website upload failure:",
          cleanupError,
        );
      }

      throw new Error(
        `Failed to upload website: ${websiteUploadResponse.statusText}`,
      );
    }

    // Step 3: Return successful response
    const response: GenerateChatbotResponse = {
      chatbotId: actualChatbotId,
      siteUrl,
      previewUrl: `/preview/${encodeURIComponent(actualChatbotId)}`,
      message: "Chatbot created successfully with website content indexed.",
    };

    return response;
  } catch (error) {
    console.error("Error creating chatbot:", error);

    throw createError({
      statusCode: 500,
      statusMessage:
        error instanceof Error ? error.message : "Failed to create chatbot",
    });
  }
});
