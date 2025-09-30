export default defineNuxtRouteMiddleware(async (to, from) => {
  // Skip middleware for login page
  console.log(to);
  if (to.path === "/login") {
    return;
  }

  // Skip middleware for error page
  if (to.path === "/error") {
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
