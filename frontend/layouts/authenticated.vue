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
                <IconMessageSquare class="h-4 w-4" />
                Chats
              </NuxtLink>
              <NuxtLink
                to="/subscription"
                class="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
                active-class="bg-muted text-foreground"
              >
                <IconCreditCard class="h-4 w-4" />
                Subscription
              </NuxtLink>
              <NuxtLink
                to="/settings"
                class="flex items-center gap-3 rounded-lg px-3 py-2 text-muted-foreground transition-all hover:text-foreground"
                active-class="bg-muted text-foreground"
              >
                <IconSettings class="h-4 w-4" />
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
            <IconMenu class="h-6 w-6" />
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
import IconCreditCard from "@/components/icons/IconCreditCard.vue";
import IconMenu from "@/components/icons/IconMenu.vue";
import IconMessageSquare from "@/components/icons/IconMessageSquare.vue";
import IconSettings from "@/components/icons/IconSettings.vue";

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
