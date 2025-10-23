import { createError } from "h3";

type CachedToken = {
  token: string;
  expiresAt: number;
};

let cachedToken: CachedToken | null = null;
let pendingRefresh: Promise<string> | null = null;

const CLOCK_SKEW_BUFFER_MS = 60 * 1000; // refresh 1 minute before expiry
const MIN_EXPIRY_MS = 30 * 1000; // fallback minimum lifetime

type VectorchatConfig = {
  vectorchatUrl?: string;
  vectorchatClientId?: string;
  vectorchatClientSecret?: string;
};

export async function getVectorchatAccessToken(
  config: VectorchatConfig,
): Promise<string> {
  const now = Date.now();
  if (cachedToken && cachedToken.expiresAt > now + 1000) {
    return cachedToken.token;
  }

  if (pendingRefresh) {
    return pendingRefresh;
  }

  const apiUrl = getVectorchatBaseUrl(config);
  const clientId = config.vectorchatClientId;
  const clientSecret = config.vectorchatClientSecret;

  if (!clientId || !clientSecret) {
    throw createError({
      statusCode: 500,
      statusMessage:
        "VectorChat client credentials are not configured. Please set NUXT_VECTORCHAT_CLIENT_ID and NUXT_VECTORCHAT_CLIENT_SECRET.",
    });
  }

  pendingRefresh = fetch(`${apiUrl}/public/oauth/token`, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({
      client_id: clientId,
      client_secret: clientSecret,
    }),
  })
    .then(async (response) => {
      if (!response.ok) {
        const errorBody = await safeParseJSON(response);
        console.error(
          "[vectorchat-auth] Failed to obtain access token",
          response.status,
          response.statusText,
          errorBody,
        );
        throw createError({
          statusCode: response.status,
          statusMessage:
            (errorBody && (errorBody.error || errorBody.error_description)) ||
            "Failed to obtain VectorChat access token",
        });
      }

      const token = (await response.json()) as {
        access_token: string;
        expires_in?: number;
        token_type?: string;
        scope?: string;
      };

      if (!token?.access_token) {
        throw createError({
          statusCode: 500,
          statusMessage: "VectorChat token response is missing access_token",
        });
      }

      const expiresInSeconds =
        typeof token.expires_in === "number" ? token.expires_in : 0;
      const lifetimeMs = Math.max(
        expiresInSeconds * 1000 - CLOCK_SKEW_BUFFER_MS,
        MIN_EXPIRY_MS,
      );
      cachedToken = {
        token: token.access_token,
        expiresAt: Date.now() + lifetimeMs,
      };

      return cachedToken.token;
    })
    .finally(() => {
      pendingRefresh = null;
    });

  return pendingRefresh;
}

async function safeParseJSON(response: Response) {
  try {
    return await response.json();
  } catch {
    return null;
  }
}

export function getVectorchatBaseUrl(config: VectorchatConfig): string {
  const candidate = config.vectorchatUrl;

  if (!candidate) {
    throw createError({
      statusCode: 500,
      statusMessage:
        "VectorChat base URL is not configured. Please set NUXT_VECTORCHAT_URL.",
    });
  }

  return candidate.replace(/\/$/, "");
}
