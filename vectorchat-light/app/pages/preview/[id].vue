<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import { useRoute, useRouter } from "#imports";
import { Button } from "@/components/ui/button";
import AuthRequiredModal from "@/components/AuthRequiredModal.vue";
import { useAuthPrompt } from "@/composables/useAuthPrompt";
import { onClickOutside } from "@vueuse/core";
import { Check, ChevronDown } from "lucide-vue-next";

definePageMeta({
  layout: "landing",
});

interface WidgetOption {
  name: string;
  description: string;
  theme?: string;
}

function escapeAttribute(value: string) {
  return String(value)
    .replace(/&/g, "&amp;")
    .replace(/"/g, "&quot;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;");
}

const route = useRoute();
const router = useRouter();
const {
  loginHref,
  isCheckingSession,
  shouldShowPrompt,
  refreshSession,
  updateLoginHref,
} = useAuthPrompt({
  getReturnTo: () =>
    typeof window !== "undefined" ? window.location.href : undefined,
});

const isInteractionDisabled = computed(
  () => shouldShowPrompt.value || isCheckingSession.value,
);

const chatbotId = route.params.id as string;
const siteUrl = ref((route.query.siteUrl as string) || "");
const isLoading = ref(true);
const error = ref("");
const widgetError = ref("");
const chatbotData = ref<any>(null);

const widgetOptions = ref<WidgetOption[]>([]);
const selectedWidget = ref("");
const widgetIframeKey = ref(0);
const isWidgetDropdownOpen = ref(false);
const widgetDropdownRef = ref<HTMLElement | null>(null);

const widgetScriptUrl = computed(() => {
  if (!selectedWidget.value) return "";
  return `/api/widgets/${chatbotId}/${selectedWidget.value}.js?v=${widgetIframeKey.value}`;
});

const widgetSrcdoc = computed(() => {
  if (!widgetScriptUrl.value) return "";
  const scriptSrc = escapeAttribute(widgetScriptUrl.value);
  const attrs = [
    `src="${scriptSrc}"`,
    `data-chat-id="${escapeAttribute(chatbotId)}"`,
  ];
  const runtimeOrigin =
    typeof window !== "undefined" ? window.location.origin : "";
  if (runtimeOrigin) {
    attrs.push(`data-api-base="${escapeAttribute(runtimeOrigin)}"`);
  }
  return `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>VectorChat Widget Preview</title>
    <style>
      :root, body {
        margin: 0;
        padding: 0;
        font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
        background: linear-gradient(135deg, #f8fafc, #e2e8f0);
      }
    </style>
  </head>
  <body>
    <script ${attrs.join(" ")}><\/script>
  </body>
</html>`.trim();
});

function formatWidgetDisplayName(widgetName: string) {
  return widgetName
    .replace(/^vectorchat[-_]?/i, "")
    .replace(/-widget$/i, "")
    .split("-")
    .map((segment) => segment.charAt(0).toUpperCase() + segment.slice(1))
    .join(" ");
}

function handleWidgetSelection(widgetName: string) {
  if (isInteractionDisabled.value) return;
  if (selectedWidget.value === widgetName) {
    widgetIframeKey.value += 1;
    return;
  }
  selectedWidget.value = widgetName;
}

const selectedWidgetLabel = computed(() =>
  selectedWidget.value
    ? formatWidgetDisplayName(selectedWidget.value)
    : "Select widget",
);

function toggleWidgetDropdown() {
  if (isInteractionDisabled.value) return;
  isWidgetDropdownOpen.value = !isWidgetDropdownOpen.value;
}

function selectWidget(widgetName: string) {
  handleWidgetSelection(widgetName);
  isWidgetDropdownOpen.value = false;
}

function goBack() {
  router.push("/");
}

const handleFocus = async () => {
  await refreshSession({ showModalOnFailure: true });
};

const handleVisibilityChange = async () => {
  if (document.visibilityState === "visible") {
    await refreshSession({ showModalOnFailure: true });
  }
};

onMounted(async () => {
  if (!chatbotId) {
    router.push("/");
    return;
  }

  if (typeof window !== "undefined") {
    updateLoginHref();
    await refreshSession({ showModalOnFailure: true });
    window.addEventListener("focus", handleFocus);
    document.addEventListener("visibilitychange", handleVisibilityChange);
  } else {
    await refreshSession({ showModalOnFailure: true });
  }

  try {
    const chatbotResponse = await fetch(`/api/chatbot/${chatbotId}`);
    if (!chatbotResponse.ok) {
      throw new Error("Failed to load chatbot");
    }
    chatbotData.value = await chatbotResponse.json();
  } catch (err) {
    console.error("Error loading chatbot:", err);
    error.value = "Failed to load chatbot. It may still be processing.";
  }

  try {
    const widgetResponse = await fetch("/api/widgets");
    if (!widgetResponse.ok) {
      throw new Error("Failed to load widgets");
    }
    const payload = await widgetResponse.json();
    widgetOptions.value = Array.isArray(payload.widgets) ? payload.widgets : [];
    if (widgetOptions.value.length > 0) {
      const defaultWidget =
        widgetOptions.value.find(
          (option) => option.name === "vectorchat-glass-widget",
        ) || widgetOptions.value[0];
      selectedWidget.value = defaultWidget.name;
    }
  } catch (err) {
    console.error("Error loading widgets:", err);
    widgetError.value =
      "Unable to load widget themes right now. Please try again later.";
  } finally {
    isLoading.value = false;
  }
});

watch(selectedWidget, (newValue) => {
  if (!newValue) return;
  widgetIframeKey.value += 1;
});

onClickOutside(widgetDropdownRef, () => {
  isWidgetDropdownOpen.value = false;
});

onBeforeUnmount(() => {
  if (typeof window === "undefined") return;
  window.removeEventListener("focus", handleFocus);
  document.removeEventListener("visibilitychange", handleVisibilityChange);
});
</script>

<template>
  <AuthRequiredModal
    :open="shouldShowPrompt"
    :is-checking="isCheckingSession"
    :login-href="loginHref"
  />

  <!-- Loading State -->
  <div v-if="isLoading" class="flex items-center justify-center min-h-[60vh]">
    <div class="text-center">
      <div
        class="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"
      ></div>
      <p class="text-gray-600">Loading your chatbot...</p>
    </div>
  </div>

  <!-- Error State -->
  <div v-else-if="error" class="flex items-center justify-center min-h-[60vh]">
    <div class="text-center max-w-md">
      <div class="text-red-500 mb-4">
        <svg
          class="w-16 h-16 mx-auto"
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.964-.833-2.732 0L3.732 16.5c-.77.833.192 2.5 1.732 2.5z"
          />
        </svg>
      </div>
      <h2 class="text-xl font-semibold text-gray-900 mb-2">
        Something went wrong
      </h2>
      <p class="text-gray-600 mb-6">{{ error }}</p>
      <Button @click="goBack">Go Back</Button>
    </div>
  </div>

  <!-- Widget Preview -->
  <div v-else class="max-w-6xl mx-auto p-4 space-y-6 pt-30">
    <div class="bg-white rounded-xl shadow-lg overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-200">
        <!-- Preview Badge -->
        <span
          class="inline-flex items-center px-3 py-1 rounded-full text-sm font-medium bg-green-100 text-green-800"
        >
          ✨ Live Preview
        </span>
        <p class="text-sm text-gray-600 mt-2">
          <span v-if="selectedWidget">
            Showing
            {{ formatWidgetDisplayName(selectedWidget) }} for
            {{ chatbotData?.name || "your chatbot" }}.
          </span>
          <span v-else>Select a widget theme to start the preview.</span>
        </p>
      </div>

      <div class="bg-slate-100/70 p-4 md:p-6">
        <div class="preview-shell">
          <div ref="widgetDropdownRef" class="widget-switcher">
            <Button
              variant="outline"
              size="sm"
              class="widget-switcher__trigger"
              :disabled="!widgetOptions.length || isInteractionDisabled"
              @click="toggleWidgetDropdown"
            >
              <span class="widget-switcher__label">
                {{ selectedWidgetLabel }}
              </span>
              <ChevronDown
                class="widget-switcher__icon"
                :class="{ 'rotate-180': isWidgetDropdownOpen }"
              />
            </Button>
            <Transition
              enter-active-class="transition duration-150 ease-out"
              enter-from-class="opacity-0 scale-95 -translate-y-1"
              enter-to-class="opacity-100 scale-100 translate-y-0"
              leave-active-class="transition duration-100 ease-in"
              leave-from-class="opacity-100 scale-100 translate-y-0"
              leave-to-class="opacity-0 scale-95 -translate-y-1"
            >
              <div v-if="isWidgetDropdownOpen" class="widget-switcher__menu">
                <p v-if="widgetError" class="widget-switcher__error">
                  {{ widgetError }}
                </p>
                <p
                  v-else-if="!widgetOptions.length"
                  class="widget-switcher__empty"
                >
                  No widgets available yet.
                </p>
                <div v-else class="widget-switcher__list">
                  <button
                    v-for="widget in widgetOptions"
                    :key="widget.name"
                    type="button"
                    class="widget-switcher__item"
                    :data-active="selectedWidget === widget.name"
                    @click="selectWidget(widget.name)"
                  >
                    <span>{{ formatWidgetDisplayName(widget.name) }}</span>
                    <Check
                      v-if="selectedWidget === widget.name"
                      class="widget-switcher__check"
                    />
                  </button>
                </div>
              </div>
            </Transition>
          </div>
          <iframe
            v-if="selectedWidget"
            :key="widgetIframeKey"
            class="preview-frame"
            :srcdoc="widgetSrcdoc"
            title="Chat widget preview"
          ></iframe>
          <div
            v-else
            class="h-[520px] w-full flex items-center justify-center text-gray-500 text-sm"
          >
            Select a widget to load the preview.
          </div>
        </div>
      </div>
    </div>

    <div class="bg-white rounded-xl shadow p-6 space-y-4">
      <h3 class="text-lg font-semibold text-gray-900">Embed Instructions</h3>
      <p class="text-sm text-gray-600">
        Copy the script tag below to embed this chatbot on your site.
      </p>
      <div class="bg-slate-50 border border-slate-200 rounded-lg p-4">
        <code class="block text-xs text-slate-700 break-words">
          &lt;script src="https://vectorchat.com/widgets/chats/{{
            chatbotId
          }}/{{ selectedWidget || "[widget-name]" }}.js" data-chat-id="{{
            chatbotId
          }}" data-api-base="https://vectorchat.com"&gt;&lt;/script&gt;
        </code>
      </div>
      <div
        class="bg-blue-50 border border-blue-200 rounded-lg p-3 text-sm text-blue-800"
      >
        Need a different style? Custom widget options will be available soon.
      </div>
    </div>
  </div>
</template>

<style scoped>
.preview-shell {
  position: relative;
  border-radius: 1rem;
  border: 1px solid rgba(148, 163, 184, 0.35);
  background: linear-gradient(
    180deg,
    rgba(255, 255, 255, 0.95),
    rgba(226, 232, 240, 0.7)
  );
  padding: 2.5rem 2.5rem 4rem;
  min-height: 760px;
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

.preview-frame {
  width: 100%;
  height: 760px;
  border: none;
  border-radius: 0.75rem;
  background-color: transparent;
}

.widget-switcher {
  position: absolute;
  top: 1.5rem;
  left: 1.5rem;
  z-index: 20;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.widget-switcher__trigger {
  min-width: 11rem;
  justify-content: space-between;
  font-weight: 500;
  border-color: rgba(148, 163, 184, 0.6);
}

.widget-switcher__icon {
  height: 1rem;
  width: 1rem;
  transition: transform 150ms ease;
}

.widget-switcher__menu {
  min-width: 14rem;
  background: #fff;
  border: 1px solid rgba(148, 163, 184, 0.35);
  border-radius: 0.75rem;
  box-shadow: 0 20px 45px rgba(15, 23, 42, 0.16);
  padding: 0.75rem;
}

.widget-switcher__error {
  font-size: 0.8125rem;
  color: #dc2626;
}

.widget-switcher__empty {
  font-size: 0.8125rem;
  color: #475569;
}

.widget-switcher__list {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.widget-switcher__item {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
  padding: 0.6rem 0.75rem;
  border-radius: 0.6rem;
  font-size: 0.875rem;
  color: #0f172a;
  transition:
    background-color 120ms ease,
    color 120ms ease;
}

.widget-switcher__item:hover {
  background: rgba(191, 219, 254, 0.3);
}

.widget-switcher__item[data-active="true"] {
  background: rgba(59, 130, 246, 0.12);
  color: #1d4ed8;
  font-weight: 600;
}

.widget-switcher__check {
  height: 1rem;
  width: 1rem;
}

.widget-switcher__label {
  font-size: 0.875rem;
}
</style>
