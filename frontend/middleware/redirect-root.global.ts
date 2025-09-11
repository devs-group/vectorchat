export default defineNuxtRouteMiddleware((to) => {
  // Redirect root path to /chat
  if (to.path === '/') {
    return navigateTo('/chat', { replace: true })
  }
})
