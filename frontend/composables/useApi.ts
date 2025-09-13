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

interface BackendErrorResponse {
  error?: string;
  message?: string;
  data?: any;
}

const globalState = reactive<Record<string, unknown>>({});

/**
 * Extracts error message from various error response formats
 */
function extractErrorMessage(error: any): string {
  // If error is null or undefined
  if (!error) {
    return "An unknown error occurred";
  }

  // If error has a data property (from fetch response)
  if (error.data) {
    // Check if data has the backend error structure
    if (typeof error.data === "object") {
      const backendError = error.data as BackendErrorResponse;
      // Priority: error field > message field
      if (backendError.error) {
        return backendError.error;
      }
      if (backendError.message) {
        return backendError.message;
      }
    }
    // If data is a string
    if (typeof error.data === "string") {
      return error.data;
    }
  }

  // If error itself has the backend structure
  if (typeof error === "object") {
    const backendError = error as BackendErrorResponse;
    if (backendError.error) {
      return backendError.error;
    }
    if (backendError.message) {
      return backendError.message;
    }
  }

  // If error has a message property (standard Error object)
  if (error.message) {
    return error.message;
  }

  // If error is a string
  if (typeof error === "string") {
    return error;
  }

  // Fallback
  return "An unexpected error occurred";
}

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
    error.value = null; // Reset error state

    try {
      const {
        data: apiData,
        error: apiError,
        pending,
      } = await apiCall(...args);

      isLoading.value = pending.value;

      if (apiError.value) {
        error.value = apiError.value;

        // Extract the actual error message from the backend response
        const errorMessage = extractErrorMessage(apiError.value);

        // Show error toast with the backend error message
        toast.error("Error", {
          description: errorMessage,
        });
      } else if (apiData.value) {
        // Check if the response itself contains an error (some endpoints might return 200 with error in body)
        const responseData = apiData.value as any;
        if (responseData?.error) {
          error.value = responseData.error;
          toast.error("Error", {
            description: responseData.error,
          });
          return;
        }

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
    } catch (err) {
      // Handle any unexpected errors
      error.value = err;
      isLoading.value = false;

      const errorMessage = extractErrorMessage(err);
      toast.error("Error", {
        description: errorMessage,
      });
    }
  };

  return {
    data,
    error,
    isLoading,
    execute,
  };
}

/**
 * Custom fetch wrapper with proper error handling
 */
export const useApiFetch = <T>(request: any, opts?: any) => {
  const config = useRuntimeConfig();

  return useFetch<T>(request, {
    baseURL: config.public.apiBase as string,
    credentials: "include",
    onResponseError({ response }) {
      // Ensure error responses are properly formatted
      if (response._data) {
        // The backend returns error in the format: { error: "message" }
        // Make sure this is preserved in the error object
        response._data = response._data;
      }
    },
    ...opts,
  });
};
