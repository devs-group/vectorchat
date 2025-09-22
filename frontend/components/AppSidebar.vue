<template>
  <Sidebar collapsible="icon" class="bg-sidebar text-sidebar-foreground">
    <SidebarHeader class="border-b border-sidebar-border h-16">
      <NuxtLink
        to="/"
        class="flex items-center gap-3 rounded-lg px-2 py-1 text-base font-semibold leading-none text-sidebar-foreground transition-colors hover:text-sidebar-foreground/80"
      >
        <div
          class="flex h-9 w-9 items-center justify-center rounded-lg bg-sidebar-accent text-sm font-semibold uppercase text-sidebar-accent-foreground"
        >
          VC
        </div>
        <div class="flex flex-col gap-0.5 group-data-[collapsible=icon]:hidden">
          <span class="text-sm font-semibold tracking-tight">VectorChat</span>
          <span class="text-xs font-normal text-muted-foreground"
            >Control Center</span
          >
        </div>
      </NuxtLink>
    </SidebarHeader>

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
          <SidebarMenuButton
            class="hover:bg-transparent"
            :tooltip="displayName"
          >
            <div class="flex items-center gap-3">
              <div
                class="flex h-9 w-9 items-center justify-center rounded-full bg-sidebar-accent text-sm font-medium uppercase text-sidebar-accent-foreground"
              >
                {{ initials }}
              </div>
              <div
                class="flex flex-col text-left text-sm leading-tight group-data-[collapsible=icon]:hidden"
              >
                <span class="font-medium text-sidebar-foreground truncate">
                  {{ displayName }}
                </span>
                <span class="text-xs text-muted-foreground truncate">
                  {{ email }}
                </span>
              </div>
            </div>
          </SidebarMenuButton>
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
import IconCreditCard from "@/components/icons/IconCreditCard.vue";
import IconGrid from "@/components/icons/IconGrid.vue";
import IconMessageSquare from "@/components/icons/IconMessageSquare.vue";
import IconSettings from "@/components/icons/IconSettings.vue";

const navItems = [
  { title: "Chats", to: "/chat", icon: IconMessageSquare },
  { title: "Knowledge Bases", to: "/knowledge-bases", icon: IconGrid },
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
</script>
