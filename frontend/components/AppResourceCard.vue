<template>
  <Card
    :class="[
      'group relative overflow-hidden transition-all duration-200 shadow-lg hover:shadow-md',
      cardClass,
    ]"
  >
    <div class="flex h-full flex-col">
      <CardHeader class="space-y-3 pb-0">
        <div class="flex items-start justify-between gap-3">
          <div class="flex min-w-0 items-center gap-3">
            <CardIconBadge
              v-if="hasIcon"
              :variant="iconVariant"
              :size="iconSize"
              class="flex-shrink-0"
            >
              <slot name="icon" />
            </CardIconBadge>
            <div class="min-w-0">
              <CardTitle class="truncate text-lg font-semibold">
                {{ title }}
              </CardTitle>
              <div
                v-if="hasSubtitle"
                class="mt-1 flex flex-wrap items-center gap-2"
              >
                <slot name="subtitle" />
              </div>
              <CardDescription
                v-if="showDescription"
                class="line-clamp-2 text-sm text-muted-foreground"
              >
                <slot name="description">
                  {{ description }}
                </slot>
              </CardDescription>
            </div>
          </div>
          <div v-if="hasMeta" class="flex items-center gap-2">
            <slot name="meta" />
          </div>
        </div>
        <slot name="header-extra" />
      </CardHeader>

      <CardContent v-if="$slots.content" class="pt-4">
        <slot name="content" />
      </CardContent>

      <CardFooter v-if="$slots.footer" :class="['mt-auto', footerClass]">
        <slot name="footer" />
      </CardFooter>
    </div>
    <NuxtLink
      v-if="to"
      :to="to"
      class="absolute inset-0"
      :aria-label="linkAriaLabel || `View ${title}`"
    />
  </Card>
</template>

<script setup lang="ts">
import { computed, useSlots } from "vue";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import CardIconBadge, {
  type CardIconSize,
  type CardIconVariant,
} from "@/components/CardIconBadge.vue";

const props = withDefaults(
  defineProps<{
    title: string;
    description?: string | null;
    to?: string;
    linkAriaLabel?: string;
    cardClass?: string;
    footerClass?: string;
    iconVariant?: CardIconVariant;
    iconSize?: CardIconSize;
  }>(),
  {
    description: null,
    to: undefined,
    linkAriaLabel: "",
    cardClass: "",
    footerClass:
      "flex-col items-start gap-2 border-t pt-4 text-sm text-muted-foreground",
    iconVariant: "slate",
    iconSize: "lg",
  },
);

const slots = useSlots();

const hasIcon = computed(() => Boolean(slots.icon));
const hasMeta = computed(() => Boolean(slots.meta));
const hasSubtitle = computed(() => Boolean(slots.subtitle));
const showDescription = computed(
  () => Boolean(slots.description) || Boolean(props.description),
);
</script>
