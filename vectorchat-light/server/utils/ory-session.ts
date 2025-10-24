import { H3Event, createError, getRequestHeader } from "h3";

/**
 * Ensures the incoming request carries a Kratos session cookie. VectorChat's Oathkeeper
 * proxy will perform the authoritative validation, but we can fail fast when the cookie
 * is missing.
 */
export function assertSessionCookie(event: H3Event): string {
  const cookieHeader = getRequestHeader(event, "cookie");

  if (!cookieHeader) {
    throw createError({
      statusCode: 401,
      statusMessage: "Authentication required",
    });
  }

  return cookieHeader;
}

/**
 * Builds the header map used to forward the authenticated user's context to VectorChat.
 */
export function createUserAuthHeaders(
  event: H3Event,
): Record<string, string> {
  const cookieHeader = assertSessionCookie(event);

  return {
    Cookie: cookieHeader,
  };
}
