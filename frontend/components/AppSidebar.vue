<template>
  <Sidebar collapsible="icon" class="bg-sidebar text-sidebar-foreground">
    <SidebarHeader class="border-b border-sidebar-border h-16">
      <NuxtLink
        to="/"
        class="flex items-center gap-3 rounded-lg px-2 py-1 text-base font-semibold leading-none text-sidebar-foreground transition-colors hover:text-sidebar-foreground/80"
      >
        <img
          src="/vc.svg"
          alt="VectorChat logo"
          class="h-9 w-9 rounded-lg"
        />
        <div class="flex flex-col gap-0.5 group-data-[collapsible=icon]:hidden">
          <span class="text-sm font-semibold tracking-tight">VectorChat</span>
        </div>
      </NuxtLink>
    </SidebarHeader>

    <div class="px-3 py-2">
      <OrganizationSwitcher />
    </div>

    <SidebarContent>
      <SidebarGroup>
        <SidebarGroupContent>
          <SidebarMenu>
            <SidebarMenuItem v-for="item in navItems" :key="item.to">
              <SidebarMenuButton
                as-child
                :tooltip="item.title"
                :is-active="isActive(item.to)"
              >
                <NuxtLink :to="item.to">
                  <component :is="item.icon" />
                  <span>{{ item.title }}</span>
                </NuxtLink>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroupContent>
      </SidebarGroup>
    </SidebarContent>

    <SidebarFooter class="border-t border-sidebar-border">
      <SidebarMenu>
        <SidebarMenuItem>
          <DropdownMenu>
            <DropdownMenuTrigger as-child>
              <SidebarMenuButton
                size="lg"
                class="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                :tooltip="displayName"
              >
                <div
                  class="flex h-8 w-8 items-center justify-center rounded-full bg-sidebar-accent text-sm font-medium uppercase text-sidebar-accent-foreground"
                >
                  {{ initials }}
                </div>
                <div
                  class="grid flex-1 text-left text-sm leading-tight group-data-[collapsible=icon]:hidden"
                >
                  <span class="truncate font-medium">{{ displayName }}</span>
                  <span class="truncate text-xs text-muted-foreground">
                    {{ email }}
                  </span>
                </div>
                <IconChevronsUpDown
                  class="ml-auto h-4 w-4 group-data-[collapsible=icon]:hidden"
                />
              </SidebarMenuButton>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              class="w-[--reka-dropdown-menu-trigger-width] min-w-56 rounded-lg"
              side="bottom"
              align="end"
              :side-offset="4"
            >
              <DropdownMenuLabel class="p-0 font-normal">
                <div class="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                  <div
                    class="flex h-8 w-8 items-center justify-center rounded-full bg-muted text-sm font-medium uppercase"
                  >
                    {{ initials }}
                  </div>
                  <div class="grid flex-1 text-left text-sm leading-tight">
                    <span class="truncate font-medium">{{ displayName }}</span>
                    <span class="truncate text-xs text-muted-foreground">
                      {{ email }}
                    </span>
                  </div>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem @click="handleLogout">
                <IconLogOut class="mr-2 h-4 w-4" />
                Log out
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    </SidebarFooter>

    <SidebarRail />
  </Sidebar>
</template>

<script setup lang="ts">
import { computed } from "vue";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import IconChevronsUpDown from "@/components/icons/IconChevronsUpDown.vue";
import IconCreditCard from "@/components/icons/IconCreditCard.vue";
import IconGrid from "@/components/icons/IconGrid.vue";
import IconLogOut from "@/components/icons/IconLogOut.vue";
import IconMessageSquare from "@/components/icons/IconMessageSquare.vue";
import IconSettings from "@/components/icons/IconSettings.vue";
import OrganizationSwitcher from "@/components/OrganizationSwitcher.vue";

const navItems = [
  { title: "Chats", to: "/chat", icon: IconMessageSquare },
  { title: "Knowledge Bases", to: "/knowledge-bases", icon: IconGrid },
  { title: "Organizations", to: "/organizations", icon: IconGrid },
  { title: "Subscription", to: "/subscription", icon: IconCreditCard },
  { title: "API Settings", to: "/settings", icon: IconSettings },
] as const;

const route = useRoute();
const { get } = useSession();
const session = get();

const displayName = computed(() => session.value?.user?.name || "User");
const email = computed(() => session.value?.user?.email || "");
const initials = computed(() => {
  const name = displayName.value.trim();
  if (name) {
    const parts = name.split(/\s+/).filter(Boolean);
    const chars = parts
      .slice(0, 2)
      .map((part) => part[0] ?? "")
      .join("")
      .toUpperCase();
    if (chars) return chars;
  }
  return email.value[0]?.toUpperCase() || "";
});

function isActive(path: string) {
  if (!path) return false;
  const current = route.path;
  if (current === path) return true;
  return current.startsWith(`${path}/`);
}

const config = useRuntimeConfig();

async function handleLogout() {
  try {
    const logoutFlow = await $fetch<{ logout_url: string }>(
      `${config.public.kratosPublicUrl}/self-service/logout/browser`,
      { credentials: "include" },
    );
    if (logoutFlow?.logout_url) {
      window.location.href = logoutFlow.logout_url;
    }
  } catch (error) {
    console.error("Failed to initiate logout", error);
  }
}
</script>
