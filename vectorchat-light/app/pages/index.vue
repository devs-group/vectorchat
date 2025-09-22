<script setup lang="ts">
import { computed, nextTick, ref } from "vue";

import { useHead } from "#imports";
import LandingMenu from "@/components/LandingMenu.vue";
import { Button } from "@/components/ui/button";

useHead({
  title: "VectorChat Light | Free Chatbot in Under 2 Minutes",
  meta: [
    {
      name: "description",
      content:
        "Type in your website address and launch an AI chatbot in minutes. VectorChat Light keeps visitors engaged without extra engineering.",
    },
    {
      property: "og:title",
      content: "VectorChat Light | Free Chatbot in Under 2 Minutes",
    },
    {
      property: "og:description",
      content:
        "Most sites are static walls of text. Let visitors interact, get instant answers, and stay engaged with VectorChat Light.",
    },
  ],
});

const siteUrl = ref("");
const errorMessage = ref("");
const isGenerating = ref(false);
const showResult = ref(false);
const copyStatus = ref<"idle" | "copied" | "error">("idle");

const inputRef = ref<HTMLInputElement | null>(null);
const resultRef = ref<HTMLElement | null>(null);

const snippet = computed(() => {
  if (!showResult.value) return "";

  try {
    const parsed = new URL(siteUrl.value);
    return `<script src="https://cdn.vectorchat.light/embed.js" data-site="${parsed.hostname}" async />`;
  } catch (error) {
    console.debug("Failed to derive hostname from URL", error);
    return `<script src="https://cdn.vectorchat.light/embed.js" async />`;
  }
});

const urlPattern =
  /^https?:\/\/(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}(?:[-a-zA-Z0-9@:%_+.~#?&/=]*)?$/;

async function handleSubmit() {
  if (isGenerating.value) return;

  const trimmed = siteUrl.value.trim();
  siteUrl.value = trimmed;
  errorMessage.value = "";
  copyStatus.value = "idle";
  showResult.value = false;

  if (!urlPattern.test(trimmed)) {
    errorMessage.value =
      "Enter a valid URL that starts with http:// or https://.";
    await nextTick();
    inputRef.value?.focus();
    return;
  }

  isGenerating.value = true;

  await new Promise((resolve) => setTimeout(resolve, 600));

  isGenerating.value = false;
  showResult.value = true;

  await nextTick();
  resultRef.value?.focus();
}
</script>

<template>
  <div class="relative min-h-screen bg-background text-foreground">
    <LandingMenu />

    <main class="flex flex-col">
      <section class="relative isolate overflow-hidden pb-32 pt-32 sm:pb-36">
        <div
          class="absolute inset-x-0 top-0 h-[52rem] -z-30 bg-gradient-to-br from-[black] via-green-900 to-[black]"
        ></div>
        <div
          class="pointer-events-none absolute inset-x-0 bottom-0 h-48 -z-20 bg-gradient-to-b from-transparent via-[#0c1f4c]/40 to-white dark:to-slate-950"
        ></div>
        <div
          class="pointer-events-none absolute left-1/2 top-[-22rem] -z-20 hidden aspect-square w-[70rem] -translate-x-1/2 rounded-full bg-[radial-gradient(circle,_rgba(255,255,255,0.18)_0%,rgba(12,31,76,0)_70%)] lg:block"
        ></div>
        <div
          class="pointer-events-none absolute left-1/2 top-[-8rem] -z-10 aspect-square w-[44rem] -translate-x-1/2 rounded-full border border-white/10 bg-[radial-gradient(circle,_rgba(255,255,255,0.32)_0%,rgba(255,255,255,0)_60%)] blur-sm"
        ></div>

        <div
          class="relative z-10 mx-auto flex w-full max-w-6xl flex-col gap-16 px-4 text-white lg:flex-row lg:items-center"
        >
          <div class="w-full text-center space-y-10">
            <span
              class="inline-flex items-center gap-2 rounded-full bg-white/10 px-4 py-1 text-xs font-semibold uppercase tracking-[0.25em] text-white/70 ring-1 ring-white/20"
            >
              Free AI Chatbot in under 2 minutes
            </span>
            <h1
              class="mt-6 text-balance text-4xl font-semibold leading-tight sm:text-5xl md:text-6xl"
            >
              Enter your Website Address and create an AI Chatbot.
            </h1>
            <p class="mt-4 text-lg text-white">
              Most sites are static walls of text. <br />
              Visitors leave without finding what they need. With VectorChat
              Light, let users interact, get instant answers, and stay engaged
              <br />
              without endless scrolling...
            </p>

            <form
              class="mt-10 w-full rounded-[2rem] border border-white/20 bg-white/90 p-6 text-slate-900"
              novalidate
              @submit.prevent="handleSubmit"
            >
              <div class="flex flex-col gap-4 sm:flex-row">
                <div class="flex-1 space-y-2">
                  <label
                    class="block text-center text-sm font-semibold text-slate-700"
                    for="website-url"
                  >
                    Enter your Website URL
                  </label>
                  <input
                    id="website-url"
                    ref="inputRef"
                    v-model="siteUrl"
                    type="url"
                    inputmode="url"
                    autocomplete="url"
                    placeholder="https://your-website.com"
                    required
                    :aria-invalid="errorMessage.length > 0"
                    aria-describedby="url-feedback"
                    class="w-full rounded-xl border border-slate-200 bg-white/70 px-4 py-3 text-base text-slate-900 shadow-sm transition-colors focus:border-transparent focus:outline-hidden focus-visible:ring-4 focus-visible:ring-[#4f46e5]/40 motion-reduce:transition-none"
                  />
                </div>
              </div>
              <p
                id="url-feedback"
                class="mt-3 min-h-[1.25rem] text-left text-sm text-rose-500"
                aria-live="polite"
              >
                {{ errorMessage }}
              </p>
              <Button
                type="submit"
                class="h-12 shrink-0 rounded-xl bg-gradient-to-r px-6 text-base font-semibold text-white shadow-lg transition-transform hover:translate-y-[-1px] focus-visible:ring-4 focus-visible:ring-white/40 motion-reduce:transform-none motion-reduce:transition-none bg-black cursor-pointer"
                :disabled="isGenerating"
                :aria-busy="isGenerating"
              >
                <span v-if="!isGenerating">Generate Free Chatbot</span>
                <span v-else>Generating…</span>
              </Button>
              <p class="mt-4 text-gray-900 text-xs">
                Free, easy and no credit card required.
              </p>
            </form>
          </div>
        </div>
      </section>

      <section class="relative mx-auto w-full max-w-6xl px-4 pb-24">
        <header class="mx-auto max-w-2xl text-center">
          <h2 class="text-3xl font-semibold text-slate-900 dark:text-white">
            How it works
          </h2>
          <p class="mt-3 text-base text-slate-600 dark:text-slate-300">
            Launch a helpful chatbot in three simple steps—no engineering hours
            required.
          </p>
        </header>

        <div class="mt-14 grid gap-8 lg:grid-cols-3">
          <article
            class="group relative overflow-hidden rounded-[2rem] border border-slate-200 bg-white p-8 shadow-[0_30px_80px_-45px_rgba(15,23,42,0.45)] transition-transform hover:-translate-y-2 motion-reduce:transform-none dark:border-slate-700 dark:bg-slate-900"
          >
            <div
              class="h-28 rounded-2xl bg-gradient-to-br from-sky-200/60 via-white to-blue-200/60"
            ></div>
            <h3
              class="mt-6 text-xl font-semibold text-slate-900 dark:text-white"
            >
              Enter your website address
            </h3>
            <p class="mt-3 text-sm text-slate-600 dark:text-slate-300">
              VectorChat Light will index your website pages and automatically
              create a chatbot, powered by AI.
            </p>
          </article>

          <article
            class="group relative overflow-hidden rounded-[2rem] border border-slate-200 bg-white p-8 shadow-[0_30px_80px_-45px_rgba(15,23,42,0.45)] transition-transform hover:-translate-y-2 motion-reduce:transform-none lg:translate-y-8 dark:border-slate-700 dark:bg-slate-900"
          >
            <div
              class="h-28 rounded-2xl bg-gradient-to-br from-indigo-200/60 via-white to-purple-200/60"
            ></div>
            <h3
              class="mt-6 text-xl font-semibold text-slate-900 dark:text-white"
            >
              Embed one line of code
            </h3>
            <p class="mt-3 text-sm text-slate-600 dark:text-slate-300">
              Copy the script tag and paste it into your existing site. No
              complex setup.
            </p>
          </article>

          <article
            class="group relative overflow-hidden rounded-[2rem] border border-slate-200 bg-white p-8 shadow-[0_30px_80px_-45px_rgba(15,23,42,0.45)] transition-transform hover:-translate-y-2 motion-reduce:transform-none lg:-translate-y-6 dark:border-slate-700 dark:bg-slate-900"
          >
            <div
              class="h-28 rounded-2xl bg-gradient-to-br from-emerald-200/60 via-white to-teal-200/60"
            ></div>
            <h3
              class="mt-6 text-xl font-semibold text-slate-900 dark:text-white"
            >
              Engage &amp; track
            </h3>
            <p class="mt-3 text-sm text-slate-600 dark:text-slate-300">
              Let people interact with your content and track engagement in real
              time.
            </p>
          </article>
        </div>
      </section>
    </main>
  </div>
</template>
