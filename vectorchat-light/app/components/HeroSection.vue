<script setup lang="ts">
import { nextTick, onMounted, ref } from "vue";
import { useRouter } from "#imports";
import { Button } from "@/components/ui/button";
import AuthRequiredModal from "@/components/AuthRequiredModal.vue";
import { useAuthPrompt } from "@/composables/useAuthPrompt";

const siteUrl = ref("");
const errorMessage = ref("");
const isGenerating = ref(false);
const generationStep = ref("");
const inputRef = ref<HTMLInputElement | null>(null);

type GenerateChatbotResponse = {
  chatbotId: string;
  siteUrl: string;
  previewUrl: string;
  message: string;
};

const router = useRouter();
const SITE_URL_STORAGE_KEY = "vc-light:pending-site-url";

const {
  ensureAuthenticated,
  loginHref,
  isCheckingSession,
  shouldShowPrompt,
  updateLoginHref,
} = useAuthPrompt({
  getReturnTo: () =>
    typeof window !== "undefined" ? window.location.href : undefined,
});

const urlPattern =
  /^https?:\/\/(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(?:[-a-zA-Z0-9@:%_+.~#?&/=]*)?$/;

function rememberPendingSiteUrl(url: string) {
  if (typeof window === "undefined") return;
  try {
    sessionStorage.setItem(SITE_URL_STORAGE_KEY, url);
  } catch (error) {
    console.warn("Failed to store pending site URL", error);
  }
}

function clearPendingSiteUrl() {
  if (typeof window === "undefined") return;
  try {
    sessionStorage.removeItem(SITE_URL_STORAGE_KEY);
  } catch (error) {
    console.warn("Failed to clear pending site URL", error);
  }
}

function restorePendingSiteUrl() {
  if (typeof window === "undefined") return;
  try {
    const pending = sessionStorage.getItem(SITE_URL_STORAGE_KEY);
    if (pending && !siteUrl.value) {
      siteUrl.value = pending;
    }
  } catch (error) {
    console.warn("Failed to restore pending site URL", error);
  }
}

async function handleSubmit() {
  if (isGenerating.value) return;

  const trimmed = siteUrl.value.trim();
  siteUrl.value = trimmed;
  errorMessage.value = "";
  if (!urlPattern.test(trimmed)) {
    errorMessage.value =
      "Enter a valid URL that starts with http:// or https://.";
    await nextTick();
    inputRef.value?.focus();
    return;
  }
  rememberPendingSiteUrl(trimmed);
  updateLoginHref();
  const session = await ensureAuthenticated();
  if (!session) {
    return;
  }
  clearPendingSiteUrl();

  isGenerating.value = true;
  generationStep.value = "Creating your AI chatbot...";

  try {
    generationStep.value = "Processing website content...";

    const response = await fetch("/api/chatbots", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ siteUrl: trimmed }),
    });

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      const error = new Error(
        errorData.message || `HTTP ${response.status}: ${response.statusText}`,
      );
      if (errorData && typeof errorData === "object" && "code" in errorData) {
        (error as { code?: string }).code = (
          errorData as { code?: string }
        ).code;
      }
      throw error;
    }

    const chatbotData: GenerateChatbotResponse = await response.json();

    generationStep.value = "Finalizing chatbot setup...";
    generationStep.value = "Success! Redirecting to your chatbot...";

    // Brief delay to show success message
    await new Promise((resolve) => setTimeout(resolve, 1000));

    await router.push({
      path: chatbotData.previewUrl,
      query: { siteUrl: chatbotData.siteUrl },
    });
  } catch (error: any) {
    console.error("Failed to generate chatbot", error);

    // Provide more specific error messages based on the error type
    const errorMessageText = error?.message || "";

    if (errorMessageText === "LIMIT_REACHED") {
      errorMessage.value =
        "You have already created a chatbot. You are allowed to create only one chatbot on the free plan.";
    } else if (errorMessageText.includes("HTTP 400")) {
      errorMessage.value =
        "Invalid website URL. Please check the URL and try again.";
    } else if (errorMessageText.includes("HTTP 401")) {
      errorMessage.value = "Please sign in to continue.";
    } else if (errorMessageText.includes("HTTP 500")) {
      if (errorMessageText.includes("website")) {
        errorMessage.value =
          "Unable to process your website. Please ensure the URL is accessible and try again.";
      } else if (errorMessageText.includes("chatbot")) {
        errorMessage.value = "Failed to create chatbot. Please try again.";
      } else {
        errorMessage.value = "Server error occurred. Please try again later.";
      }
    } else if (
      error?.name === "TypeError" ||
      errorMessageText.includes("fetch")
    ) {
      errorMessage.value =
        "Network error. Please check your connection and try again.";
    } else {
      errorMessage.value =
        "We couldn't generate your chatbot right now. Please try again.";
    }

    await nextTick();
    inputRef.value?.focus();
  } finally {
    isGenerating.value = false;
    generationStep.value = "";
  }
}

onMounted(() => {
  updateLoginHref();
  restorePendingSiteUrl();
});
</script>

<template>
  <section class="relative isolate overflow-hidden pb-32 pt-32 sm:pb-36">
    <div
      class="absolute inset-x-0 top-0 h-full -z-30 bg-gradient-to-br from-green-300 via-blue-100 to-white"
    ></div>
    <div class="pointer-events-none absolute"></div>
    <div
      class="pointer-events-none absolute left-1/2 top-[-22rem] -z-20 hidden aspect-square w-[70rem] -translate-x-1/2 rounded-full bg-[radial-gradient(circle,_rgba(255,255,255,0.18)_0%,rgba(12,31,76,0)_70%)] lg:block"
    ></div>
    <div
      class="pointer-events-none absolute left-1/2 top-[-8rem] -z-10 aspect-square w-[44rem] -translate-x-1/2 rounded-full border border-white/10 bg-[radial-gradient(circle,_rgba(255,255,255,0.32)_0%,rgba(255,255,255,0)_60%)] blur-sm"
    ></div>

    <div
      class="relative z-10 mx-auto flex w-full max-w-6xl flex-col gap-16 px-4 text-gray-900 lg:flex-row lg:items-center"
    >
      <div class="w-full text-center space-y-12 md:space-y-20">
        <span
          class="inline-flex items-center gap-2 rounded-full bg-gray-500/10 px-4 py-1 text-xs font-semibold uppercase tracking-[0.25em] text-black/70 ring-1 ring-white/20"
        >
          Free AI Chatbot in under 2 minutes
        </span>
        <div class="space-y-5">
          <h1
            class="mt-6 text-bold text-4xl font-semibold leading-tight sm:text-5xl md:text-6xl"
          >
            From Static Pages to Smart Conversations — Free, Fast, No Code.
          </h1>

          <p class="text-gray-700">
            Boost engagement, answer questions instantly, and convert more
            visitors — without writing a single line of code.
          </p>
        </div>
        <form
          class="mt-10 w-full flex flex-col items-center justify-between rounded-2xl px-4 py-3 dark:border-white/10 dark:bg-slate-900/60"
          novalidate
          @submit.prevent="handleSubmit"
        >
          <div class="flex flex-col gap-4 sm:flex-row w-full">
            <div class="flex-1 space-y-2">
              <label
                for="website-url"
                class="inline-flex items-center gap-2 rounded-full bg-gray-500/10 px-4 py-1 text-xs font-semibold uppercase tracking-[0.25em] text-black/70 ring-1 ring-white/20"
              >
                Enter your Website URL
              </label>
              <div
                class="relative mt-6 w-full md:w-2/3 lg:w-1/2 max-w-4xl mx-auto"
              >
                <div
                  class="absolute inset-0 rounded-xl bg-gradient-to-r from-blue-500 via-purple-500 to-pink-500 p-[4px] animate-pulse"
                >
                  <div class="h-full w-full rounded-xl bg-white/70"></div>
                </div>
                <input
                  id="website-url"
                  ref="inputRef"
                  v-model="siteUrl"
                  type="url"
                  inputmode="url"
                  autocomplete="url"
                  placeholder="https://your-website.com"
                  :disabled="isGenerating"
                  required
                  :aria-invalid="errorMessage.length > 0"
                  aria-describedby="url-feedback"
                  class="text-center relative w-full rounded-xl bg-white/70 px-4 py-3 text-base text-slate-900 shadow-sm transition-colors focus:outline-hidden focus-visible:ring-4 focus-visible:ring-[#4f46e5]/40 motion-reduce:transition-none"
                />
              </div>
            </div>
          </div>
          <p
            id="url-feedback"
            class="mt-3 min-h-[1.25rem] text-left text-sm text-rose-300"
            aria-live="polite"
          >
            {{ errorMessage }}
          </p>
          <Button
            type="submit"
            @click="handleSubmit"
            class="mt-2 h-12 shrink-0 rounded-xl bg-gradient-to-r px-6 text-base font-semibold text-white shadow-lg transition-transform hover:translate-y-[-1px] focus-visible:ring-4 focus-visible:ring-white/40 motion-reduce:transform-none motion-reduce:transition-none bg-black cursor-pointer"
            :disabled="isGenerating"
            :aria-busy="isGenerating"
          >
            <span v-if="!isGenerating">Build My Chatbot ✨</span>
            <span v-else class="flex items-center gap-2">
              <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
                <circle
                  class="opacity-25"
                  cx="12"
                  cy="12"
                  r="10"
                  stroke="currentColor"
                  stroke-width="4"
                ></circle>
                <path
                  class="opacity-75"
                  fill="currentColor"
                  d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                ></path>
              </svg>
              {{ generationStep || "Generating…" }}
            </span>
          </Button>
          <p class="mt-4 text-black/70 text-xs">
            Free and easy forever. No credit card required.
          </p>
        </form>
      </div>
    </div>
  </section>
  <AuthRequiredModal
    :open="shouldShowPrompt"
    :is-checking="isCheckingSession"
    :login-href="loginHref"
    title="Sign in to build your chatbot"
    description="You need to be signed in to create your chatbot. We'll bring you right back here after you log in."
  />
</template>
