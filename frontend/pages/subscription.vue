<template>
  <div class="flex flex-col gap-6">
    <div class="flex items-center justify-between">
      <h1 class="text-3xl font-bold tracking-tight">Subscription</h1>
    </div>

    <!-- Skeleton loader for subscription section -->
    <div class="rounded-lg border p-6" v-if="isLoadingSub">
      <div class="flex items-center justify-between">
        <Skeleton class="h-6 w-40" />
        <Skeleton class="h-9 w-32" />
      </div>
      <div class="mt-4 space-y-2">
        <Skeleton class="h-4 w-64" />
      </div>
    </div>

    <div class="rounded-lg border p-6" v-else>
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold mb-2">Your subscription</h2>
        <Button
          v-if="currentSub"
          size="sm"
          variant="secondary"
          :disabled="isOpeningPortal"
          @click="openPortal"
        >
          {{ isOpeningPortal ? "Opening…" : "Manage Billing" }}
        </Button>
      </div>
      <div v-if="currentSub">
        <div class="flex flex-wrap items-center gap-3 text-sm">
          <span>
            Status:
            <span
              :class="statusClass(currentSub.status)"
              class="px-2 py-0.5 rounded-full"
              >{{ prettyStatus(currentSub.status) }}</span
            >
          </span>
          <span
            v-if="willCancelAtPeriodEnd"
            class="px-2 py-0.5 rounded-full bg-gray-100 text-gray-700"
            >Will not renew</span
          >
          <span v-if="showNextDate"
            >{{ nextDateLabel }}: {{ nextDateFormatted }}</span
          >
          <span v-if="isSubActive && currentPlan && currentPlan.display_name"
            >Plan: <strong>{{ currentPlan.display_name }}</strong></span
          >
        </div>
      </div>
      <div v-else class="text-sm text-muted-foreground">
        No active subscription.
      </div>
    </div>

    <div class="rounded-lg border">
      <div class="p-6">
        <h2 class="text-lg font-semibold">Plans</h2>
        <p class="text-sm text-muted-foreground">
          Choose a plan and subscribe using Stripe Checkout.
        </p>
      </div>
      <div class="p-6 grid gap-4 sm:grid-cols-2 lg:grid-cols-3 items-stretch">
        <div
          v-for="plan in plans || []"
          :key="plan.id"
          class="rounded-lg h-full"
        >
          <!-- Gradient border wrapper when current plan -->
          <div
            class="h-full rounded-lg"
            :class="[
              isSubActive && currentPlanKey === plan.key
                ? 'bg-gradient-to-br from-emerald-500 to-teal-500 p-[2px]'
                : '',
            ]"
          >
            <!-- Inner card -->
            <div
              class="rounded-lg p-4 flex flex-col relative border bg-card shadow-sm border-border h-full"
              :class="{
                'border-transparent':
                  isSubActive && currentPlanKey === plan.key,
              }"
            >
              <div
                v-if="isSubActive && currentPlanKey === plan.key"
                class="absolute -top-3 left-3"
              >
                <span
                  class="text-xs font-medium bg-green-100 text-green-700 px-2 py-0.5 rounded-full border border-green-200 shadow-sm"
                  >Current</span
                >
              </div>
              <div>
                <div class="flex items-center justify-between mb-2">
                  <h3 class="text-base font-semibold">
                    {{ plan.display_name }}
                  </h3>
                  <span
                    v-if="plan.plan_definition?.tags?.length"
                    class="text-xs text-muted-foreground"
                  >
                    {{ plan.plan_definition.tags.join(", ") }}
                  </span>
                </div>
                <div class="text-2xl font-bold">
                  {{ formatPrice(plan.amount_cents, plan.currency) }}
                  <span class="text-sm font-normal text-muted-foreground"
                    >/ {{ plan.billing_interval }}</span
                  >
                </div>
                <ul
                  v-if="plan.plan_definition?.features"
                  class="mt-3 text-sm text-muted-foreground space-y-1"
                >
                  <li
                    v-for="(val, key) in plan.plan_definition.features"
                    :key="key"
                  >
                    {{ key }}: {{ formatFeature(val) }}
                  </li>
                </ul>
              </div>
              <div class="mt-auto">
                <Button
                  :disabled="isCreatingCheckout || isBlockingSub"
                  @click="subscribe(plan)"
                >
                  {{
                    isCreatingCheckout
                      ? "Redirecting…"
                      : currentPlanKey === plan.key
                        ? "Subscribed"
                        : isBlockingSub
                          ? "Manage in Billing"
                          : "Subscribe"
                  }}
                </Button>
              </div>
            </div>
          </div>
        </div>

        <div
          v-if="!isLoadingPlans && (!plans || plans.length === 0)"
          class="col-span-full text-center text-sm text-muted-foreground py-6"
        >
          No plans available.
        </div>
        <div
          v-if="isLoadingPlans"
          class="col-span-full flex justify-center py-6"
        >
          <svg
            class="animate-spin h-5 w-5 text-muted-foreground"
            xmlns="http://www.w3.org/2000/svg"
            fill="none"
            viewBox="0 0 24 24"
          >
            <circle
              class="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              stroke-width="4"
            />
            <path
              class="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from "vue";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import type { Plan, Subscription } from "~/types/api";

definePageMeta({
  layout: "authenticated",
});

const apiService = useApiService();
const { get: getSession } = useSession();
const sessionRef = getSession();

const {
  data: plans,
  execute: fetchPlans,
  isLoading: isLoadingPlans,
} = apiService.listPlans();
const {
  data: subResp,
  execute: fetchSubscription,
  isLoading: isLoadingSub,
} = apiService.getSubscription();
const {
  execute: createPortal,
  isLoading: isOpeningPortal,
  data: portalResp,
} = apiService.createPortalSession();
const {
  execute: createCheckout,
  isLoading: isCreatingCheckout,
  data: checkoutResp,
} = apiService.createCheckoutSession();

const currentSub = computed(
  () => (subResp.value as any)?.subscription as Subscription | null | undefined,
);
const currentPlanKey = computed(() => {
  const meta = currentSub.value?.metadata || null;
  if (!meta) return undefined;
  return (meta["plan_key"] as string) || undefined;
});
const currentPlan = computed(() =>
  (plans.value || []).find((p: any) => p.key === currentPlanKey.value),
);
const isSubActive = computed(() => {
  const s = (currentSub.value?.status || "").toLowerCase();
  return s === "active" || s === "trialing" || s === "past_due";
});
const isBlockingSub = computed(
  () => isSubActive.value && !willCancelAtPeriodEnd.value,
);

const formatPrice = (amountCents: number, currency: string) => {
  const amount = (amountCents || 0) / 100;
  try {
    return new Intl.NumberFormat(undefined, {
      style: "currency",
      currency: currency?.toUpperCase() || "USD",
    }).format(amount);
  } catch {
    return `$${amount.toFixed(2)}`;
  }
};

const formatFeature = (v: any) => {
  if (typeof v === "boolean") return v ? "Yes" : "No";
  if (typeof v === "number") return v.toString();
  return String(v);
};

const formatDate = (iso: string) => {
  try {
    const d = new Date(iso);
    return new Intl.DateTimeFormat(undefined, { dateStyle: "medium" }).format(
      d,
    );
  } catch {
    return iso;
  }
};

const prettyStatus = (s: string) =>
  s.replace(/_/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
const statusClass = (s: string) => {
  const ok = ["active", "trialing"];
  const warn = ["past_due", "incomplete", "incomplete_expired"];
  const bad = ["canceled", "unpaid"];
  if (ok.includes(s)) return "bg-green-100 text-green-700";
  if (warn.includes(s)) return "bg-yellow-100 text-yellow-700";
  if (bad.includes(s)) return "bg-red-100 text-red-700";
  return "bg-gray-100 text-gray-700";
};

const isCanceled = computed(
  () => (currentSub.value?.status || "").toLowerCase() === "canceled",
);
const willCancelAtPeriodEnd = computed(
  () => !!currentSub.value?.cancel_at_period_end && !isCanceled.value,
);
const showNextDate = computed(
  () => !!currentSub.value?.current_period_end && !isCanceled.value,
);
const nextDateLabel = computed(() =>
  willCancelAtPeriodEnd.value ? "Ends on" : "Renews on",
);
const nextDateFormatted = computed(() =>
  formatDate(currentSub.value?.current_period_end as string),
);

const subscribe = async (plan: Plan) => {
  const userId = sessionRef.value?.user?.id;
  if (!userId) {
    console.error("No user session; cannot subscribe");
    return;
  }
  const origin = window.location.origin;
  const success_url = `${origin}/subscription?status=success`;
  const cancel_url = `${origin}/subscription?status=cancelled`;
  await createCheckout({
    customer_id: userId,
    plan_key: plan.key,
    success_url,
    cancel_url,
    allow_promotion_codes: true,
  });

  const resp = (checkoutResp.value as any) || null;
  if (!resp) return;

  const { session_id, url } = resp;
  const pk = useRuntimeConfig().public.stripePk as string | undefined;
  if (pk) {
    try {
      const { loadStripe } = await import("@stripe/stripe-js");
      const stripe = await loadStripe(pk);
      if (stripe && session_id) {
        await stripe.redirectToCheckout({ sessionId: session_id });
        return;
      }
    } catch (e) {
      console.warn("Stripe JS failed, falling back to URL redirect", e);
    }
  }
  if (url) window.location.href = url;
};

const openPortal = async () => {
  const origin = window.location.origin;
  const return_url = `${origin}/subscription`;
  await createPortal({ return_url });
  const resp = (portalResp.value as any) || null;
  const url = resp?.url as string | undefined;
  if (url) window.location.href = url;
};

onMounted(async () => {
  await Promise.all([fetchPlans(), fetchSubscription()]);
});
</script>
