<template>
  <SidebarProvider class="bg-background">
    <AppSidebar />

    <SidebarInset>
      <header
        class="sticky top-0 z-30 flex h-16 items-center gap-4 border-b border-border bg-background px-2"
      >
        <SidebarTrigger class="" />
        <slot name="header" />
      </header>
      <div class="flex-1 overflow-y-auto p-7">
        <slot />
      </div>
    </SidebarInset>
  </SidebarProvider>
</template>

<script setup lang="ts">
import AppSidebar from "@/components/AppSidebar.vue";
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from "@/components/ui/sidebar";
</script>
