import { createError, defineEventHandler, readBody } from "h3";
import {
  getVectorchatAccessToken,
  getVectorchatBaseUrl,
} from "../utils/vectorchat-auth";

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

  console.log(config);

  try {
    const accessToken = await getVectorchatAccessToken(config);
    const authHeader = {
      Authorization: `Bearer ${accessToken}`,
    };

    // Step 1: Create a new chatbot with website assistant system prompt
    const createChatbotResponse = await fetch(`${apiUrl}/chat/chatbot`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        ...authHeader,
      },
      body: JSON.stringify({
        name: `VC Lite Assistant: ${new URL(siteUrl).hostname}`,
        description:
          "AI assistant trained on website content to help visitors navigate and find information",
        system_instructions: `**You are the friendly website assistant for ${siteUrl}!**

        **Your Goal:** To be a helpful and welcoming guide for every visitor. Your job is to make their experience on the site easy and to help them find exactly what they're looking for.

        **Your Most Important Rule:** You must **only** use information found on ${siteUrl}. Think of the website as your entire world. If it's not on the site, you don't know it.
        * **NO** outside knowledge.
        * **NO** guessing.
        * **NO** making up facts, links, or contact info.

        **How to Behave:**
        * **Tone:** Be conversational, patient, and positive. Imagine you're a helpful customer service representative.
        * **Clarity:** Give clear and simple answers.
        * **Efficiency:** Help users get their answers quickly.

        **How to Handle Questions:**
        1.  **Answer from the Source:** When you have the answer, state it clearly and mention it's from the website.
        2.  **Guide Them:** If a user is lost, point them to the right page or section.
        3.  **If You Can't Find It:** Don't just say "I don't know." Instead, try one of these:
            * "I'm not finding that exact detail on the site, but I *can* tell you about [Related Topic]. Is that helpful?"
            * "For that specific question, the best people to ask would be the team at ${siteUrl}. You can reach them through the 'Contact Us' page."
        4.  **Off-Topic Questions:** If someone asks about the weather, your opinions, or another website, gently guide them back.
            * **Example:** "That's an interesting question! However, my purpose is to help you with ${siteUrl}. Do you have any questions about our products or services I can help with?"`,
        model_name: "gpt-5-nano",
        max_tokens: 1000,
        temperature_param: 0.7,
        save_messages: true,
        is_enabled: false,
        shared_knowledge_base_ids: [],
      }),
    });

    if (!createChatbotResponse.ok) {
      const errorBody = await safeParseJSON(createChatbotResponse);
      console.error("[vectorchat-light] Create chatbot failed", {
        status: createChatbotResponse.status,
        statusText: createChatbotResponse.statusText,
        errorBody,
      });
      throw new Error(
        `Failed to create chatbot: ${createChatbotResponse.status} ${createChatbotResponse.statusText}`,
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
          ...authHeader,
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
            ...authHeader,
          },
        });
      } catch (cleanupError) {
        console.error(
          "Failed to cleanup chatbot after website upload failure:",
          cleanupError,
        );
      }

      const errorBody = await safeParseJSON(websiteUploadResponse);
      console.error("[vectorchat-light] Website upload failed", {
        status: websiteUploadResponse.status,
        statusText: websiteUploadResponse.statusText,
        errorBody,
      });
      throw new Error(
        `Failed to upload website: ${websiteUploadResponse.status} ${websiteUploadResponse.statusText}`,
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

  async function safeParseJSON(response: Response) {
    try {
      return await response.json();
    } catch {
      return null;
    }
  }
});
