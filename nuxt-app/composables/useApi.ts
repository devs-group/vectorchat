import type { UseFetchOptions } from "#app";
import { toast } from "vue-sonner";

interface UseApiOptions {
  showSuccessToast?: boolean;
  successMessage?: string;
  errorMessage?: string;
  cacheKey?: string;
}

interface UseApiReturn<T, P> {
  data: Ref<T | null>;
  error: Ref<unknown>;
  isLoading: Ref<boolean>;
  execute: (...args: P[]) => Promise<void>;
}

const globalState = reactive<Record<string, unknown>>({});

/**
 * A composable function for making API calls with built-in error handling, loading state, and caching.
 *
 * @param apiCall - A function that returns a Promise with the API call result.
 * @param options - Configuration options for the API call.
 * @returns An object containing reactive references for data, error, loading state, and an execute function.
 *
 *
 * Usage:
 * const { data, error, isLoading, execute } = useApi(() => APIService.getSomeData(), {
 *   showSuccessToast: true,
 *   successMessage: 'Data fetched successfully',
 *   errorMessage: 'Failed to fetch data',
 *   cacheKey: 'someDataCacheKey'
 * });
 *
 * // Call execute() to perform the API call
 * execute();
 */
export function useApi<T, P>(
  apiCall: (...args: P[]) => Promise<{
    data: Ref<T>;
    error: Ref<unknown>;
    pending: Ref<boolean>;
  }>,
  options: UseApiOptions = {},
): UseApiReturn<T, P> {
  const data = ref<T | null>(null) as Ref<T | null>;
  const error = ref<unknown>(null);
  const isLoading = ref(false);

  const execute = async (...args: P[]) => {
    isLoading.value = true;
    const { data: apiData, error: apiError, pending } = await apiCall(...args);

    isLoading.value = pending.value;

    if (apiError.value) {
      error.value = apiError.value;
      toast.error("Error", {
        description: options.errorMessage || "Something went wrong",
      });
    } else {
      data.value = apiData.value;

      if (options.cacheKey) {
        globalState[options.cacheKey] = apiData.value;
      }

      if (options.showSuccessToast) {
        toast.success("Success", {
          description:
            options.successMessage || "Operation completed successfully",
        });
      }
    }
  };

  return {
    data,
    error,
    isLoading,
    execute,
  };
}

export const useApiFetch = <T>(request: any, opts?: any) => {
  const config = useRuntimeConfig();
  return useFetch<T>(request, {
    baseURL: config.public.apiBase as string,
    credentials: "include",
    ...opts,
  });
};
