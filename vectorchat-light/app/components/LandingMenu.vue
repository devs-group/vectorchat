<script setup lang="ts">
import { computed, onMounted, ref } from "vue";

import { Menu } from "lucide-vue-next";

import { Button } from "@/components/ui/button";
import {
  NavigationMenu,
  NavigationMenuItem,
  NavigationMenuLink,
  NavigationMenuList,
} from "@/components/ui/navigation-menu";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useKratosSession } from "@/composables/useKratosSession";

const config = useRuntimeConfig();
const navItems = [
  { label: "How it works", href: "#how-it-works" },
  { label: "Integration", href: "#integration" },
];

const loginHref = ref(config.public.frontendLoginUrl);
const { session, loadSession } = useKratosSession();
const isAuthenticated = computed(() => Boolean(session.value));
const dashboardHref = computed(() => {
  if (config.public.vectorchatUrl) {
    return config.public.vectorchatUrl as string;
  }
  return loginHref.value;
});

onMounted(() => {
  if (typeof window === "undefined") return;
  try {
    const url = new URL(config.public.frontendLoginUrl);
    url.searchParams.set("return_to", window.location.origin);
    loginHref.value = url.toString();
  } catch (error) {
    console.warn("Failed to construct login URL", error);
  }
  loadSession();
});
</script>

<template>
  <header class="fixed inset-x-0 top-0 z-50">
    <div class="mx-auto max-w-6xl px-4 sm:px-6 lg:px-8">
      <div
        class="mt-6 flex items-center justify-between gap-4 rounded-2xl border border-white/30 bg-white/70 px-4 py-3 shadow-[0_18px_40px_-30px_rgba(15,23,42,0.45)] backdrop-blur-2xl transition supports-[backdrop-filter]:bg-white/55 dark:border-white/10 dark:bg-slate-900/60"
      >
        <NuxtLink
          to="/"
          class="w-1/5 inline-flex items-center gap-2 rounded-xl px-2 py-1 text-sm font-semibold text-slate-900/90 transition-colors hover:text-slate-900 dark:text-white"
        >
          <span class="text-black text-lg">VC</span>
          <span
            class="text-xs uppercase tracking-wide text-black bg-white/40 py-1 px-2 rounded-full"
          >
            Light
          </span>
        </NuxtLink>

        <NavigationMenu :viewport="false" class="w-1/5 hidden lg:flex">
          <NavigationMenuList
            class="gap-1 rounded-full bg-white/40 p-1 text-sm font-medium text-slate-600 dark:bg-slate-800/50 dark:text-slate-200"
          >
            <NavigationMenuItem v-for="item in navItems" :key="item.label">
              <NavigationMenuLink
                :href="item.href"
                class="rounded-full px-4 py-2 text-slate-700/85 transition-colors duration-200 hover:bg-white/70 hover:text-slate-900 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary dark:text-slate-200/85 dark:hover:bg-slate-700/50 dark:hover:text-white"
              >
                {{ item.label }}
              </NavigationMenuLink>
            </NavigationMenuItem>
          </NavigationMenuList>
        </NavigationMenu>

        <div class="hidden items-center gap-2 lg:flex">
          <a
            v-if="!isAuthenticated"
            :href="loginHref"
            class="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900 dark:text-slate-300 dark:hover:text-white"
          >
            Sign in
          </a>
          <a
            v-else
            :href="dashboardHref"
            class="text-sm font-medium text-slate-600 transition-colors hover:text-slate-900 dark:text-slate-300 dark:hover:text-white"
          >
            Open app
          </a>
          <Button class="px-5 text-sm font-semibold">Start for free</Button>
        </div>

        <Sheet>
          <SheetTrigger as-child>
            <Button
              variant="ghost"
              size="icon"
              class="lg:hidden text-slate-700/90 transition hover:bg-white/60 hover:text-slate-900 dark:text-slate-100 dark:hover:bg-slate-800/80"
              aria-label="Open navigation"
            >
              <Menu class="size-5" />
            </Button>
          </SheetTrigger>
          <SheetContent
            side="left"
            class="w-full max-w-xs bg-background/95 px-5 py-6 shadow-xl backdrop-blur-xl supports-[backdrop-filter]:bg-background/80"
          >
            <div class="flex h-full flex-col gap-6">
              <div class="flex items-center justify-between">
                <div class="flex flex-col">
                  <span class="text-base font-semibold text-foreground"
                    >VectorChat</span
                  >
                  <span class="text-xs text-muted-foreground"
                    >Conversational AI for teams</span
                  >
                </div>
              </div>

              <NavigationMenu
                :viewport="false"
                class="flex w-full justify-start max-w-none"
              >
                <NavigationMenuList
                  class="w-full flex-col items-start justify-start gap-1 text-base text-foreground"
                >
                  <NavigationMenuItem
                    v-for="item in navItems"
                    :key="item.label"
                    class="w-full"
                  >
                    <SheetClose as-child>
                      <NavigationMenuLink
                        :href="item.href"
                        class="w-full rounded-xl px-3 py-2 text-foreground/90 transition-colors bg-muted/80 hover:bg-muted/80 hover:text-foreground focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary"
                      >
                        {{ item.label }}
                      </NavigationMenuLink>
                    </SheetClose>
                  </NavigationMenuItem>
                </NavigationMenuList>
              </NavigationMenu>

              <div class="mt-auto flex flex-col gap-2">
                <SheetClose as-child>
                  <a
                    v-if="!isAuthenticated"
                    :href="loginHref"
                    class="rounded-lg px-3 py-2 text-center text-sm font-medium text-muted-foreground transition hover:bg-muted hover:text-foreground"
                  >
                    Sign in
                  </a>
                  <a
                    v-else
                    :href="dashboardHref"
                    class="rounded-lg px-3 py-2 text-center text-sm font-medium text-muted-foreground transition hover:bg-muted hover:text-foreground"
                  >
                    Open app
                  </a>
                </SheetClose>
                <SheetClose as-child>
                  <Button class="h-11 w-full text-sm font-semibold">
                    Start free
                  </Button>
                </SheetClose>
              </div>
            </div>
          </SheetContent>
        </Sheet>
      </div>
    </div>
  </header>
</template>
