<script setup lang="ts">
import { computed, nextTick, ref } from "vue";

import { useHead, useRoute } from "#imports";
import LandingMenu from "@/components/LandingMenu.vue";
import { Button } from "@/components/ui/button";

const route = useRoute();

const createId = () => Math.random().toString(36).slice(2, 10);

type Message = {
  id: string;
  role: "user" | "assistant";
  text: string;
};

const chatbotId = computed(() => String(route.params.chatbotId ?? ""));

const siteUrlQuery = computed(() => {
  const raw = route.query.siteUrl;
  if (Array.isArray(raw)) {
    return raw[0] ?? "";
  }
  return raw ?? "";
});

const fallbackSiteUrl = "https://example.com";
const siteUrl = computed(() => siteUrlQuery.value || fallbackSiteUrl);

const siteHostname = computed(() => {
  try {
    return new URL(siteUrl.value).hostname;
  } catch (error) {
    return siteUrl.value.replace(/^https?:\/\//, "") || "your-site.com";
  }
});

useHead({
  title: "Chatbot Preview • VectorChat Light",
  meta: [
    {
      name: "description",
      content: `Interact with the VectorChat Light preview chatbot for ${siteHostname.value}.`,
    },
  ],
});

const messages = ref<Message[]>([
  {
    id: createId(),
    role: "assistant",
    text: `Hi! I'm the VectorChat assistant for ${siteHostname.value}. Ask anything about the site and I'll do my best to help.`,
  },
]);

const userInput = ref("");
const isResponding = ref(false);
const messagesContainer = ref<HTMLElement | null>(null);

const embedSnippet = computed(
  () =>
    `<script src="https://cdn.vectorchat.light/embed.js" data-site="${siteHostname.value}" async />`,
);

const scrollToLatest = () => {
  const el = messagesContainer.value;
  if (!el) return;

  el.scrollTo({
    top: el.scrollHeight,
    behavior: "smooth",
  });
};

const handleSend = async () => {
  const text = userInput.value.trim();
  if (!text || isResponding.value) return;

  messages.value.push({
    id: createId(),
    role: "user",
    text,
  });
  userInput.value = "";

  await nextTick();
  scrollToLatest();

  isResponding.value = true;

  await new Promise((resolve) => setTimeout(resolve, 800));

  messages.value.push({
    id: createId(),
    role: "assistant",
    text: `Here's a mock answer about "${text}". Imagine this pulling from ${siteHostname.value} using your VectorChat knowledge base.`,
  });

  isResponding.value = false;

  await nextTick();
  scrollToLatest();
};
</script>

<template>
  <div class="relative min-h-screen bg-[#050816] text-white">
    <LandingMenu />

    <main
      class="mx-auto flex w-full max-w-6xl flex-col gap-10 px-4 pb-16 pt-28"
    >
      <div class="flex flex-col gap-4">
        <NuxtLink
          to="/"
          class="inline-flex w-max items-center gap-2 text-sm font-semibold text-white/70 transition hover:text-white"
        >
          <span aria-hidden="true">←</span>
          Back to landing
        </NuxtLink>
        <div class="space-y-2">
          <span
            class="inline-flex items-center gap-2 rounded-full bg-white/10 px-4 py-1 text-xs font-semibold uppercase tracking-[0.3em] text-white/60"
          >
            Chatbot Preview
          </span>
          <h1 class="text-4xl font-semibold leading-tight sm:text-5xl">
            Your VectorChat assistant is ready.
          </h1>
          <p class="text-base text-white/70 sm:text-lg">
            Preview the experience visitors will have on
            <span class="font-semibold text-white">{{ siteHostname }}</span
            >. Ask a question below to try it out.
          </p>
        </div>
      </div>

      <div class="grid gap-8 lg:grid-cols-[minmax(0,1.8fr)_minmax(0,1.1fr)]">
        <section
          class="flex min-h-[32rem] flex-col rounded-[2rem] border border-white/10 bg-white/5 backdrop-blur-xl"
        >
          <header
            class="flex items-center justify-between border-b border-white/10 px-6 py-4 text-sm text-white/70"
          >
            <div class="flex items-center gap-2">
              <span
                class="inline-block size-2 rounded-full bg-emerald-400 ring-2 ring-emerald-400/30"
              ></span>
              Live preview
            </div>
            <span
              class="rounded-full border border-white/10 px-3 py-1 text-xs font-medium text-white/60"
            >
              ID: {{ chatbotId }}
            </span>
          </header>

          <div
            ref="messagesContainer"
            class="flex-1 space-y-4 overflow-y-auto px-6 py-6"
          >
            <div
              v-for="message in messages"
              :key="message.id"
              class="flex"
              :class="
                message.role === 'assistant' ? 'justify-start' : 'justify-end'
              "
            >
              <div
                :class="[
                  'max-w-[80%] rounded-2xl px-4 py-3 text-sm leading-relaxed',
                  message.role === 'assistant'
                    ? 'bg-white/10 text-white/90'
                    : 'bg-emerald-400 text-emerald-950',
                ]"
              >
                {{ message.text }}
              </div>
            </div>
            <div v-if="isResponding" class="flex justify-start">
              <div
                class="inline-flex items-center gap-2 rounded-2xl bg-white/10 px-4 py-3 text-sm text-white/70"
              >
                <span
                  class="size-2 animate-pulse rounded-full bg-white/70"
                ></span>
                Typing a response...
              </div>
            </div>
          </div>

          <form
            class="flex flex-col gap-3 border-t border-white/10 px-6 py-4 sm:flex-row"
            @submit.prevent="handleSend"
          >
            <label class="sr-only" for="chat-input">Ask something</label>
            <input
              id="chat-input"
              v-model="userInput"
              type="text"
              autocomplete="off"
              placeholder="Ask about pricing, features, or anything on the site..."
              class="flex-1 rounded-xl border border-white/15 bg-white/10 px-4 py-3 text-sm text-white placeholder:text-white/60 focus:border-white/30 focus:outline-hidden focus-visible:ring-4 focus-visible:ring-white/30"
            />
            <Button
              type="submit"
              class="h-12 rounded-xl bg-emerald-400 px-6 text-sm font-semibold text-emerald-950 transition hover:bg-emerald-300"
              :disabled="isResponding || !userInput.trim()"
            >
              {{ isResponding ? "Waiting..." : "Send message" }}
            </Button>
          </form>
        </section>

        <aside
          class="flex flex-col gap-6 rounded-[2rem] border border-white/10 bg-white/5 p-6 text-white/80 backdrop-blur-xl"
        >
          <div class="space-y-2">
            <h2 class="text-xl font-semibold text-white">Next steps</h2>
            <p class="text-sm text-white/70">
              Copy this snippet into your site to embed the chatbot once you're
              ready to go live.
            </p>
          </div>
          <pre
            class="overflow-x-auto rounded-xl bg-black/40 p-4 text-xs leading-relaxed text-white/80"
          >
<code>{{ embedSnippet }}</code>
          </pre>
          <div class="space-y-2 text-sm text-white/70">
            <p>
              The assistant is connected to
              <span class="font-semibold text-white">{{ siteHostname }}</span
              >.
            </p>
            <p>
              This preview uses mocked answers - connect your knowledge base to
              provide real responses.
            </p>
          </div>
          <NuxtLink
            to="/"
            class="inline-flex items-center justify-center rounded-xl border border-white/20 px-4 py-3 text-sm font-semibold text-white/80 transition hover:text-white"
          >
            Back to landing
          </NuxtLink>
        </aside>
      </div>
    </main>
  </div>
</template>
