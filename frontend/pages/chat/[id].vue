<template>
  <div class="min-h-screen bg-background">
    <div class="container mx-auto p-4 md:p-6">
      <!-- Desktop: Two column layout -->
      <div
        class="hidden md:flex md:gap-6 min-h-[calc(100vh-3rem)]"
      >
        <!-- Left column with tabs -->
        <div
          class="border rounded-lg bg-card overflow-hidden w-full md:max-w-3xl md:flex-shrink-0"
        >
          <div class="h-full flex flex-col">
            <!-- Tab navigation -->
            <div class="p-4">
              <PillTabs v-model="activeDesktopTab">
                <PillTab value="details" @click="navigateToDetails">
                  Details
                </PillTab>
                <PillTab value="history" @click="navigateToHistory">
                  Chat History
                </PillTab>
              </PillTabs>
            </div>

            <!-- Tab content -->
            <div class="flex-1 overflow-y-auto p-6">
              <NuxtPage />
            </div>
          </div>
        </div>

        <!-- Right column with test panel -->
        <div class="border rounded-lg bg-card p-6 overflow-hidden flex-1">
          <TestPanel />
        </div>
      </div>

      <!-- Mobile: Single column with tabs -->
      <div class="md:hidden">
        <div class="w-full">
          <!-- Tab navigation -->
          <div class="border rounded-t-lg bg-card p-4">
            <PillTabs v-model="activeMobileTab">
              <PillTab value="details" @click="handleMobileTabClick('details')">
                Details
              </PillTab>
              <PillTab value="history" @click="handleMobileTabClick('history')">
                History
              </PillTab>
              <PillTab value="test" @click="handleMobileTabClick('test')">
                Test
              </PillTab>
            </PillTabs>
          </div>

          <!-- Tab content -->
          <div
            v-if="!showMobileTest"
            class="border-x border-b rounded-b-lg bg-card p-4"
          >
            <NuxtPage />
          </div>

          <!-- Mobile Test Panel -->
          <div v-else class="border-x border-b rounded-b-lg bg-card p-4">
            <div class="mb-4">
              <button
                @click="showMobileTest = false"
                class="text-sm text-muted-foreground hover:text-foreground"
              >
                ← Back to tabs
              </button>
            </div>
            <TestPanel />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from "vue";
import TestPanel from "./components/TestPanel.vue";
import { PillTabs, PillTab } from "@/components/ui/pill-tabs";

definePageMeta({
  layout: "authenticated",
});

// Route
const route = useRoute();
const chatId = computed(() => route.params.id as string);

// State
const showMobileTest = ref(false);

// Computed
const isDetailsActive = computed(() => {
  return (
    route.path === `/chat/${chatId.value}/details` ||
    route.path === `/chat/${chatId.value}`
  );
});

const isHistoryActive = computed(() => {
  return route.path === `/chat/${chatId.value}/history`;
});

// Active tab state for pill tabs
const activeDesktopTab = ref(
  isDetailsActive.value
    ? "details"
    : isHistoryActive.value
      ? "history"
      : "details",
);
const activeMobileTab = ref(
  showMobileTest.value
    ? "test"
    : isDetailsActive.value
      ? "details"
      : isHistoryActive.value
        ? "history"
        : "details",
);

// Navigation functions
const router = useRouter();

const navigateToDetails = () => {
  router.push(`/chat/${chatId.value}/details`);
};

const navigateToHistory = () => {
  router.push(`/chat/${chatId.value}/history`);
};

const handleMobileTabClick = (tab: string) => {
  if (tab === "test") {
    showMobileTest.value = true;
  } else {
    showMobileTest.value = false;
    if (tab === "details") {
      navigateToDetails();
    } else if (tab === "history") {
      navigateToHistory();
    }
  }
};

// Watch route changes to update active tabs
watch(
  () => route.path,
  () => {
    if (isDetailsActive.value) {
      activeDesktopTab.value = "details";
      if (!showMobileTest.value) {
        activeMobileTab.value = "details";
      }
    } else if (isHistoryActive.value) {
      activeDesktopTab.value = "history";
      if (!showMobileTest.value) {
        activeMobileTab.value = "history";
      }
    }
  },
);

// Watch showMobileTest to update active mobile tab
watch(showMobileTest, (newValue) => {
  if (newValue) {
    activeMobileTab.value = "test";
  } else {
    activeMobileTab.value = isDetailsActive.value
      ? "details"
      : isHistoryActive.value
        ? "history"
        : "details";
  }
});
</script>

<style scoped>
/* Custom scrollbar */
:deep(.overflow-y-auto) {
  scrollbar-width: thin;
  scrollbar-color: rgba(156, 163, 175, 0.5) transparent;
}

:deep(.overflow-y-auto::-webkit-scrollbar) {
  width: 6px;
}

:deep(.overflow-y-auto::-webkit-scrollbar-track) {
  background: transparent;
}

:deep(.overflow-y-auto::-webkit-scrollbar-thumb) {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}

:deep(.overflow-y-auto::-webkit-scrollbar-thumb:hover) {
  background-color: rgba(156, 163, 175, 0.7);
}
</style>
