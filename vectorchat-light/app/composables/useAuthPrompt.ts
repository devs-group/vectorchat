import { onMounted, ref } from "vue";
import { useRuntimeConfig } from "#imports";
import { useKratosSession } from "@/composables/useKratosSession";

type SyncOptions = {
  showModalOnFailure?: boolean;
};

type AuthPromptOptions = {
  getReturnTo?: () => string | undefined;
};

export function useAuthPrompt(options: AuthPromptOptions = {}) {
  const config = useRuntimeConfig();
  const { session, loadSession } = useKratosSession();

  const loginHref = ref(config.public.frontendLoginUrl || "#");
  const isCheckingSession = ref(false);
  const shouldShowPrompt = ref(false);

  const buildLoginHref = (returnTo?: string) => {
    const base = config.public.frontendLoginUrl;
    if (!base) {
      return "#";
    }

    try {
      const url = new URL(base);
      if (returnTo) {
        url.searchParams.set("return_to", returnTo);
      } else if (typeof window !== "undefined") {
        url.searchParams.set("return_to", window.location.href);
      }
      return url.toString();
    } catch (error) {
      console.warn("Failed to construct login URL", error);
      return base;
    }
  };

  const updateLoginHref = () => {
    const returnTo = options.getReturnTo?.();
    loginHref.value = buildLoginHref(returnTo);
  };

  onMounted(() => {
    updateLoginHref();
  });

  const syncSession = async (
    { showModalOnFailure = false }: SyncOptions = {},
  ) => {
    updateLoginHref();
    if (session.value) {
      shouldShowPrompt.value = false;
      return session.value;
    }

    isCheckingSession.value = true;
    try {
      const currentSession = await loadSession();
      if (currentSession) {
        shouldShowPrompt.value = false;
        return currentSession;
      }

      if (showModalOnFailure) {
        shouldShowPrompt.value = true;
      }
      return null;
    } finally {
      isCheckingSession.value = false;
    }
  };

  const ensureAuthenticated = () =>
    syncSession({ showModalOnFailure: true });

  const refreshSession = (options?: SyncOptions) =>
    syncSession(options);

  return {
    session,
    loginHref,
    isCheckingSession,
    shouldShowPrompt,
    ensureAuthenticated,
    refreshSession,
    updateLoginHref,
  };
}
