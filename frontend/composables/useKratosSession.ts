export type KratosIdentity = {
  id: string;
  traits: Record<string, any>;
  metadata_public?: Record<string, any> | null;
  created_at?: string;
  updated_at?: string;
};

export type KratosSession = {
  id: string;
  identity: KratosIdentity;
  expires_at?: string;
  authenticated_at?: string;
};

export function useKratosSession() {
  const config = useRuntimeConfig();
  const state = useState<KratosSession | null>("kratos-session", () => null);

  const loadSession = async () => {
    try {
      state.value = await $fetch<KratosSession>(
        `${config.public.kratosPublicUrl}/sessions/whoami`,
        { credentials: "include" },
      );
      return state.value;
    } catch (error: any) {
      if (error?.response?.status === 401) {
        state.value = null;
        return null;
      }
      console.warn("Failed to load Kratos session", error);
      state.value = null;
      return null;
    }
  };

  return {
    session: state,
    loadSession,
  };
}
