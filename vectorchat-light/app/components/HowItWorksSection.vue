<script setup lang="ts">
const howItWorksCards = [
  {
    gradientFrom: "from-sky-200/60",
    gradientTo: "to-blue-200/60",
    title: "1. Enter your website address",
    description:
      "VectorChat Light will index your website pages and automatically create a chatbot, powered by AI.",
    containerClass: "bg-transparent",
    borderClass: "",
    img: "1.png",
  },
  {
    gradientFrom: "from-indigo-200/60",
    gradientTo: "to-purple-200/60",
    title: "2. Embed one line of code",
    description:
      "Copy the script tag and paste it into your existing site. No complex setup.",
    containerClass:
      "border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-900",
    borderClass: "",
    img: "2.png",
  },
  {
    gradientFrom: "from-emerald-200/60",
    gradientTo: "to-teal-200/60",
    title: "3. Engage & track",
    description:
      "Let people chat with your content and track engagement in real time.",
    containerClass:
      "border border-slate-200 bg-white dark:border-slate-700 dark:bg-slate-900",
    borderClass: "",
    img: "3.png",
  },
];
</script>

<template>
  <section class="bg-gray-100 py-25">
    <div class="relative mx-auto w-full max-w-6xl px-4 pb-24">
      <header class="mx-auto max-w-2xl text-center">
        <h2 class="text-3xl font-semibold text-slate-900 dark:text-white">
          How it works
        </h2>
        <p class="mt-3 text-base text-slate-600 dark:text-slate-300">
          Launch a helpful chatbot in three simple steps—no engineering hours
          required.
        </p>
      </header>

      <div class="mt-14 grid gap-8 lg:grid-cols-3 lg:items-start">
        <article
          v-for="(card, index) in howItWorksCards"
          :key="index"
          :class="`h-[320px] group relative overflow-hidden rounded-[2rem] p-8 shadow-[0_30px_80px_-45px_rgba(15,23,42,0.45)] transition-transform motion-reduce:transform-none ${card.containerClass} bg-gradient-to-br from-gray-100 bg-green-200 to-white`"
        >
          <div
            :class="`flex justify-center items-center *:h-28 rounded-2xl bg-gradient-to-br ${card.gradientFrom} via-white ${card.gradientTo}`"
          >
            <img
              class="rounded-lg h-48"
              :src="'/' + card.img"
              :alt="card.title"
            />
          </div>

          <h3 class="mt-6 text-xl font-semibold text-slate-900 dark:text-white">
            {{ card.title }}
          </h3>
          <p class="mt-3 text-sm text-slate-600 dark:text-slate-300">
            {{ card.description }}
          </p>
        </article>
      </div>
    </div>
  </section>
</template>
