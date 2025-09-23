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
  const requestId = Math.random().toString(36).substring(2, 15);
  const body = await readBody<GenerateChatbotRequest | null>(event);
  const siteUrl = body?.siteUrl?.trim();
  const config = useRuntimeConfig();

  const apiKey = config.vectorchatApiKey as string;
  const apiUrl = config.vectorchatUrl as string;

  console.log(
    `[${requestId}] Starting chatbot creation process for site: ${siteUrl}`,
  );

  if (!siteUrl) {
    console.error(`[${requestId}] Missing required field: siteUrl`);
    throw createError({
      statusCode: 400,
      statusMessage: "Missing required field: siteUrl",
    });
  }

  try {
    // Step 1: Create a new chatbot with website assistant system prompt
    console.log(
      `[${requestId}] Step 1: Creating chatbot for hostname: ${new URL(siteUrl).hostname}`,
    );

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

    console.log(
      `[${requestId}] Chatbot creation response status: ${createChatbotResponse.status} ${createChatbotResponse.statusText}`,
    );

    if (!createChatbotResponse.ok) {
      const errorText = await createChatbotResponse.text();
      console.error(
        `[${requestId}] Failed to create chatbot. Status: ${createChatbotResponse.status}, Response: ${errorText}`,
      );
      throw new Error(
        `Failed to create chatbot: ${createChatbotResponse.statusText}`,
      );
    }

    const chatbotData = await createChatbotResponse.json();
    const actualChatbotId = chatbotData.id;

    console.log(
      `[${requestId}] Chatbot created successfully with ID: ${actualChatbotId}`,
    );

    // Step 2: Upload the website to index for this chatbot
    console.log(
      `[${requestId}] Step 2: Uploading website content for indexing`,
    );

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

    console.log(
      `[${requestId}] Website upload response status: ${websiteUploadResponse.status} ${websiteUploadResponse.statusText}`,
    );

    if (!websiteUploadResponse.ok) {
      const errorText = await websiteUploadResponse.text();
      console.error(
        `[${requestId}] Failed to upload website. Status: ${websiteUploadResponse.status}, Response: ${errorText}`,
      );

      // If website upload fails, we should clean up the created chatbot
      console.log(
        `[${requestId}] Attempting to cleanup chatbot ${actualChatbotId} due to website upload failure`,
      );
      try {
        const deleteResponse = await fetch(
          `${apiUrl}/chat/chatbot/${actualChatbotId}`,
          {
            method: "DELETE",
            headers: {
              "X-API-Key": apiKey,
            },
          },
        );
        console.log(
          `[${requestId}] Chatbot cleanup response status: ${deleteResponse.status} ${deleteResponse.statusText}`,
        );
      } catch (cleanupError) {
        console.error(
          `[${requestId}] Failed to cleanup chatbot after website upload failure:`,
          cleanupError,
        );
      }

      throw new Error(
        `Failed to upload website: ${websiteUploadResponse.statusText}`,
      );
    }

    console.log(
      `[${requestId}] Website content uploaded and indexed successfully`,
    );

    // Step 3: Return successful response
    console.log(`[${requestId}] Step 3: Returning successful response`);

    const response: GenerateChatbotResponse = {
      chatbotId: actualChatbotId,
      siteUrl,
      previewUrl: `/preview/${encodeURIComponent(actualChatbotId)}`,
      message: "Chatbot created successfully with website content indexed.",
    };

    console.log(
      `[${requestId}] Chatbot creation process completed successfully. ChatbotId: ${actualChatbotId}`,
    );

    return response;
  } catch (error) {
    console.error(`[${requestId}] Error creating chatbot:`, {
      error: error instanceof Error ? error.message : String(error),
      stack: error instanceof Error ? error.stack : undefined,
      siteUrl,
    });

    throw createError({
      statusCode: 500,
      statusMessage:
        error instanceof Error ? error.message : "Failed to create chatbot",
    });
  }
});
