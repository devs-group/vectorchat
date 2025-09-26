import { createError, defineEventHandler, getRouterParam, setHeader } from "h3";

export default defineEventHandler(async (event) => {
  const chatID = getRouterParam(event, "chatID");
  const widget = getRouterParam(event, "widget");
  const config = useRuntimeConfig();

  const apiKey = config.vectorchatApiKey as string;
  const apiUrl = config.vectorchatUrl as string;

  if (!chatID) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing chatbot ID",
    });
  }

  if (!widget) {
    throw createError({
      statusCode: 400,
      statusMessage: "Missing widget name",
    });
  }

  if (!widget.endsWith(".js")) {
    throw createError({
      statusCode: 400,
      statusMessage: "Invalid widget file format",
    });
  }

  try {
    // Fetch widget JavaScript from VectorChat API
    const response = await fetch(
      `${apiUrl || "http://localhost:8080"}/widgets/chats/${chatID}/${widget}`,
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
          statusMessage: "Widget not found",
        });
      }
      throw createError({
        statusCode: response.status,
        statusMessage: "Failed to fetch widget",
      });
    }

    const jsContent = await response.text();

    // Set appropriate headers for JavaScript content
    setHeader(event, "Content-Type", "application/javascript");
    setHeader(event, "Cache-Control", "public, max-age=3600");

    return jsContent;
  } catch (error: any) {
    console.error("Error fetching widget:", error);

    if (error?.statusCode) {
      throw error;
    }

    throw createError({
      statusCode: 500,
      statusMessage: "Internal server error",
    });
  }
});
