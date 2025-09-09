<template>
  <div class="min-h-screen bg-background">
    <div class="flex h-screen overflow-hidden">
      <!-- Sidebar -->
      <div class="hidden border-r bg-muted/40 lg:block lg:w-72">
        <div class="flex h-full flex-col gap-2">
          <div class="flex h-[60px] items-center border-b px-6">
            <NuxtLink to="/" class="flex items-center gap-2 font-semibold">
              <span>VectorChat</span>
            </NuxtLink>
          </div>
          <div class="flex-1 overflow-auto py-2">
            <nav class="grid items-start px-4 text-sm font-medium">
              <NuxtLink
                to="/chat"
                class="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
                active-class="bg-muted text-foreground"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="h-4 w-4"
                >
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
                </svg>
                Chats
              </NuxtLink>
              <NuxtLink
                to="/subscription"
                class="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
                active-class="bg-muted text-foreground"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="h-4 w-4"
                >
                  <rect width="20" height="14" x="2" y="5" rx="2" />
                  <line x1="2" x2="22" y1="10" y2="10" />
                </svg>
                Subscription
              </NuxtLink>
              <NuxtLink
                to="/settings"
                class="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
                active-class="bg-muted text-foreground"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  width="24"
                  height="24"
                  viewBox="0 0 24 24"
                  fill="none"
                  stroke="currentColor"
                  stroke-width="2"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                  class="h-4 w-4"
                >
                  <path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"></path>
                  <circle cx="12" cy="12" r="3"></circle>
                </svg>
                API Settings
              </NuxtLink>
            </nav>
          </div>
          <div class="mt-auto p-4">
            <div class="flex items-center gap-4 rounded-lg border p-4">
              <div class="flex h-9 w-9 items-center justify-center rounded-full bg-muted uppercase text-xs font-medium">
                {{ initials }}
              </div>
              <div class="flex flex-col min-w-0">
                <span class="text-sm font-medium truncate">{{ displayName }}</span>
                <span class="text-xs text-muted-foreground truncate">{{ email }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
      <!-- Main Content -->
      <div class="flex flex-1 flex-col overflow-hidden">
        <header class="sticky top-0 z-30 flex h-14 items-center gap-4 border-b bg-background px-4 sm:px-6">
          <button
            class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 lg:hidden"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="h-6 w-6"
            >
              <line x1="4" x2="20" y1="12" y2="12"></line>
              <line x1="4" x2="20" y1="6" y2="6"></line>
              <line x1="4" x2="20" y1="18" y2="18"></line>
            </svg>
            <span class="sr-only">Toggle Menu</span>
          </button>
        </header>
        <main class="flex-1 overflow-y-auto p-4 sm:p-6">
          <slot />
        </main>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const { get } = useSession();
const session = get();

const displayName = computed(() => session.value?.user?.name || "User");
const email = computed(() => session.value?.user?.email || "");
const initials = computed(() => {
  const n = displayName.value.trim();
  if (!n) return "";
  const parts = n.split(/\s+/).filter(Boolean);
  const chars = parts.slice(0, 2).map(p => p[0]).join("").toUpperCase();
  return chars || (email.value[0]?.toUpperCase() || "");
});
</script>
