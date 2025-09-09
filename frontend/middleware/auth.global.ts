export default defineNuxtRouteMiddleware(async (to, from) => {
  // Skip middleware for login and callback pages
  if (to.path === "/login" || to.path === "/auth/github/callback") {
    return;
  }

  // Use the session composable so session is cached globally
  const { call } = useSession();
  const res = await call();
  if (!res.ok) {
    return navigateTo("/login");
  }
  // If session exists, continue to the requested page
});
