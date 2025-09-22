<script setup lang="ts">
import { computed } from "vue";

const variantClasses = {
  indigo: "bg-gradient-to-br from-blue-500 to-indigo-500",
  purple: "bg-gradient-to-br from-fuchsia-500 to-violet-500",
  green: "bg-gradient-to-br from-emerald-500 to-teal-500",
  orange: "bg-gradient-to-br from-amber-500 to-orange-500",
  pink: "bg-gradient-to-br from-rose-500 to-pink-500",
  slate: "bg-gradient-to-br from-slate-600 to-slate-800",
} as const;

const sizeClasses = {
  sm: "h-8 w-8 rounded-lg",
  md: "h-9 w-9 rounded-xl",
  lg: "h-10 w-10 rounded-xl",
} as const;

export type CardIconVariant = keyof typeof variantClasses;
export type CardIconSize = keyof typeof sizeClasses;

const props = withDefaults(
  defineProps<{
    variant?: CardIconVariant;
    size?: CardIconSize;
  }>(),
  {
    variant: "indigo",
    size: "md",
  },
);

const classes = computed(
  () =>
    `inline-flex items-center justify-center text-white shadow-sm ${variantClasses[props.variant]} ${sizeClasses[props.size]}`,
);
</script>

<template>
  <div :class="classes">
    <slot />
  </div>
</template>
