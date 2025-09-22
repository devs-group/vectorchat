<script setup lang="ts">
import { computed, useSlots } from "vue";
import CardIconBadge, {
  type CardIconVariant,
} from "@/components/CardIconBadge.vue";

const props = withDefaults(
  defineProps<{
    title: string;
    subtitle?: string;
    color?: CardIconVariant;
    padded?: boolean;
  }>(),
  {
    color: "indigo",
    padded: true,
  },
);

const slots = useSlots();

const hasIcon = computed(() => Boolean(slots.icon));
</script>

<template>
  <div class="rounded-2xl border border-border bg-card shadow-sm">
    <div class="px-6 py-5 border-b border-border/70">
      <div class="flex items-start gap-3">
        <CardIconBadge v-if="hasIcon" :variant="color" class="mt-0.5">
          <slot name="icon" />
        </CardIconBadge>
        <div>
          <h2 class="text-lg font-medium">{{ title }}</h2>
          <p v-if="subtitle" class="text-sm text-muted-foreground">{{ subtitle }}</p>
        </div>
      </div>
    </div>

    <div v-if="padded" class="p-6 md:p-8">
      <slot />
    </div>
    <template v-else>
      <slot />
    </template>
  </div>
</template>
