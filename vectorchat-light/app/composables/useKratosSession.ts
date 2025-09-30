export type KratosSession = {
  id: string;
  identity: {
    id: string;
    traits: Record<string, any>;
  };
};

export function useKratosSession() {
  const session = useState<KratosSession | null>(
    "vc-light-kratos-session",
    () => null,
  );
  const config = useRuntimeConfig();

  const loadSession = async () => {
    try {
      session.value = await $fetch<KratosSession>(
        `${config.public.kratosPublicUrl}/sessions/whoami`,
        { credentials: "include" },
      );
      return session.value;
    } catch (error: any) {
      if (error?.response?.status === 401) {
        session.value = null;
        return null;
      }
      console.warn("Failed to fetch Kratos session", error);
      session.value = null;
      return null;
    }
  };

  return {
    session,
    loadSession,
  };
}
