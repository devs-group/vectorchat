export default defineNuxtRouteMiddleware(async (to, from) => {
  // Skip middleware for login and callback pages
  if (to.path === "/login" || to.path === "/auth/github/callback") {
    return;
  }

  const apiService = useApiService();
  const { execute, error, data } = apiService.getSession();
  await execute();

  // Only redirect to login if there's an error or no session data
  if (error.value || !data.value) {
    return navigateTo("/login");
  }
  // If we have a valid session, just continue to the requested page
  // No need to navigate again as we're already on the correct path
});
