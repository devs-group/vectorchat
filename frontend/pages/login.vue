<template>
  <div class="grid min-h-screen lg:grid-cols-2">
    <div
      class="relative hidden bg-primary/5 lg:flex lg:flex-col lg:items-center lg:justify-center"
    >
      <div
        class="absolute inset-0 bg-gradient-to-br from-primary/10 via-primary/5 to-transparent"
      />
      <div class="relative z-10 flex flex-col items-center px-8 text-center">
        <div
          class="mb-6 flex h-20 w-20 items-center justify-center rounded-2xl bg-primary/10"
        >
          <IconMessageSquare class="h-10 w-10 text-primary" />
        </div>
        <h2 class="mb-3 text-3xl font-bold tracking-tight">VectorChat</h2>
        <p class="max-w-sm text-lg text-muted-foreground">
          Chat with your documents using AI. Upload, index, and get instant
          answers from your knowledge base.
        </p>
      </div>
      <div
        class="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-border to-transparent"
      />
    </div>

    <div class="flex flex-col">
      <div class="flex flex-1 items-center justify-center px-4 py-12 sm:px-6">
        <div class="w-full max-w-sm space-y-6">
          <div class="space-y-2 text-center lg:text-left">
            <div class="mb-4 flex items-center justify-center gap-2 lg:hidden">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10"
              >
                <IconMessageSquare class="h-5 w-5 text-primary" />
              </div>
              <span class="text-xl font-bold">VectorChat</span>
            </div>
            <h1 class="text-2xl font-semibold tracking-tight">Welcome back</h1>
            <p class="text-sm text-muted-foreground">
              Sign in to your account to continue
            </p>
          </div>

          <div
            v-if="errorMessage"
            class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive"
          >
            {{ errorMessage }}
          </div>

          <div class="space-y-4">
            <Button
              variant="outline"
              class="h-11 w-full gap-2 text-base font-medium"
              :disabled="!githubNode || isLoading"
              @click="handleGithubLogin"
            >
              <IconGithub class="h-5 w-5" />
              <span v-if="isLoading">Loading...</span>
              <span v-else>Continue with GitHub</span>
            </Button>
          </div>

          <p class="text-center text-xs text-muted-foreground lg:text-left">
            By continuing, you agree to our terms of service
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import IconGithub from "@/components/icons/IconGithub.vue";
import IconMessageSquare from "@/components/icons/IconMessageSquare.vue";
import { useKratosSession } from "~/composables/useKratosSession";

type KratosNodeAttributes = {
  name?: string;
  type?: string;
  value?: unknown;
  disabled?: boolean;
};

type KratosNode = {
  group?: string;
  attributes?: KratosNodeAttributes;
};

type ContinueWith = {
  action?: string;
  redirect_browser_to?: string;
  flow?: {
    id?: string;
    type?: string;
    continue_with?: ContinueWith[];
  };
};

type FlowSession = {
  id?: string;
};

type LoginFlowResponse = {
  id: string;
  state?: string;
  return_to?: string;
  continue_with?: ContinueWith[];
  session?: FlowSession | null;
  session_token?: string | null;
  ui?: {
    action?: string;
    method?: string;
    nodes?: KratosNode[];
  };
};

const config = useRuntimeConfig();
const route = useRoute();
const isLoading = ref(true);
const errorMessage = ref<string | null>(null);
const loginFlow = ref<LoginFlowResponse | null>(null);
const { loadSession } = useKratosSession();

const githubNode = computed(() => {
  const nodes = loginFlow.value?.ui?.nodes || [];
  return nodes.find(
    (node) => node.group === "oidc" && node.attributes.value === "github",
  );
});

const redirectToKratos = () => {
  const kratosUrl = new URL(
    `${config.public.kratosPublicUrl}/self-service/login/browser`,
  );

  const returnToParam = route.query.return_to;
  if (typeof returnToParam === "string" && returnToParam) {
    kratosUrl.searchParams.set("return_to", returnToParam);
  } else {
    kratosUrl.searchParams.set(
      "return_to",
      `${window.location.origin}${route.path}`,
    );
  }

  window.location.replace(kratosUrl.toString());
};

const resolveReturnTo = () => {
  const explicitReturnTo =
    typeof route.query.return_to === "string" && route.query.return_to
      ? route.query.return_to
      : null;

  if (explicitReturnTo) {
    return explicitReturnTo;
  }

  const flowReturnTo = loginFlow.value?.return_to;
  if (typeof flowReturnTo === "string" && flowReturnTo) {
    return flowReturnTo;
  }

  return window.location.origin;
};

const redirectIfAuthenticated = async () => {
  try {
    const session = await loadSession();
    if (session) {
      window.location.href = resolveReturnTo();
      return true;
    }
  } catch (error) {
    console.warn("Failed to check existing session", error);
  }
  return false;
};

const findRedirectInInstructions = (
  instructions?: ContinueWith[] | null,
): string | null => {
  if (!Array.isArray(instructions)) return null;

  for (const instruction of instructions) {
    if (
      instruction &&
      typeof instruction.redirect_browser_to === "string" &&
      instruction.redirect_browser_to
    ) {
      return instruction.redirect_browser_to;
    }

    const nested = findRedirectInInstructions(instruction?.flow?.continue_with);
    if (nested) return nested;
  }

  return null;
};

const followPostLoginRedirect = (flow: LoginFlowResponse | null) => {
  if (process.server || !flow) return false;

  const redirectTarget = findRedirectInInstructions(flow.continue_with);
  if (redirectTarget) {
    window.location.href = redirectTarget;
    return true;
  }

  const successStates = new Set([
    "success",
    "passed_challenge",
    "success_logged_in",
  ]);
  const hasSession = Boolean(flow.session?.id || flow.session_token);

  if (successStates.has(flow.state || "") || hasSession) {
    window.location.href = resolveReturnTo();
    return true;
  }

  return false;
};

const fetchLoginFlow = async (flowId: string) => {
  try {
    isLoading.value = true;
    loginFlow.value = await $fetch<LoginFlowResponse>(
      `${config.public.kratosPublicUrl}/self-service/login/flows`,
      {
        params: { id: flowId },
        credentials: "include",
      },
    );
    if (followPostLoginRedirect(loginFlow.value)) {
      return;
    }
    errorMessage.value = null;
  } catch (error: any) {
    const status = error?.response?.status;
    console.error("Failed to load Kratos login flow", error, status);
    if ([400, 403, 404, 410].includes(status)) {
      redirectToKratos();
      return;
    }
    errorMessage.value =
      "Unable to load login flow. Please refresh and try again.";
  } finally {
    isLoading.value = false;
  }
};

const ensureFlowId = () => {
  const flowId = route.query.flow;
  return typeof flowId === "string" ? flowId : null;
};

const ensureFlow = () => {
  if (process.server) return;

  const flowId = ensureFlowId();
  if (!flowId) {
    redirectToKratos();
    return;
  }

  void fetchLoginFlow(flowId);
};

const handleGithubLogin = () => {
  const node = githubNode.value;
  const action = loginFlow.value?.ui?.action;
  const method = loginFlow.value?.ui?.method || "POST";

  if (!node || !action) {
    errorMessage.value = "GitHub login is currently unavailable.";
    return;
  }

  errorMessage.value = null;

  const form = document.createElement("form");
  form.method = method.toUpperCase();
  form.action = action;
  form.style.display = "none";

  const appendField = (name: string, value: string) => {
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = name;
    input.value = value;
    form.appendChild(input);
  };

  const nodes = loginFlow.value?.ui?.nodes || [];
  nodes.forEach((currentNode) => {
    const attributes = currentNode.attributes;
    if (!attributes || attributes.disabled) return;

    const name = attributes.name;
    const value = attributes.value;
    if (!name || typeof value === "undefined") return;

    const isTargetProvider = currentNode === node;
    const isHiddenField = attributes.type === "hidden";

    if (isHiddenField || isTargetProvider) {
      appendField(name, String(value));
    }
  });

  document.body.appendChild(form);
  form.submit();
};

onMounted(async () => {
  if (await redirectIfAuthenticated()) return;
  ensureFlow();
});

watch(
  () => route.query.flow,
  (newFlow, oldFlow) => {
    if (newFlow === oldFlow) return;
    const flowId = ensureFlowId();
    if (!flowId) {
      redirectToKratos();
      return;
    }
    void fetchLoginFlow(flowId);
  },
);
</script>
