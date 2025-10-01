<template>
  <div class="flex min-h-[calc(100vh-3.5rem)] items-center justify-center">
    <div class="mx-auto flex w-full max-w-sm flex-col justify-center space-y-6">
      <div class="flex flex-col space-y-2 text-center">
        <h1 class="text-2xl font-semibold tracking-tight">VectorChat</h1>
        <p class="text-sm text-muted-foreground">
          Sign in to your account to continue
        </p>
      </div>

      <div
        v-if="errorMessage"
        class="rounded border border-red-200 bg-red-50 p-4 text-sm text-red-700"
      >
        {{ errorMessage }}
      </div>

      <div class="grid gap-4">
        <Button
          variant="outline"
          class="w-full"
          :disabled="!githubNode || isLoading"
          @click="handleGithubLogin"
        >
          <IconGithub
            class="mr-2 h-4 w-4"
            aria-hidden="true"
            focusable="false"
            data-prefix="fab"
            data-icon="github"
            role="img"
          />
          <span v-if="isLoading">Loading providers…</span>
          <span v-else>{{
            githubNode?.meta?.label?.text || "Sign in with GitHub"
          }}</span>
        </Button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import IconGithub from "@/components/icons/IconGithub.vue";
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
