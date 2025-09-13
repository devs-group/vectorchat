import { toast } from "vue-sonner";

interface BackendErrorResponse {
  error?: string;
  message?: string;
  data?: any;
}

/**
 * Composable for consistent error handling across the application
 * Extracts error messages from various error formats and shows them in toast notifications
 */
export function useErrorHandler() {
  /**
   * Extracts error message from various error response formats
   */
  const extractErrorMessage = (error: any): string => {
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

    // Check for response._data (from $fetch errors)
    if (error.response?._data) {
      const responseData = error.response._data;
      if (typeof responseData === "object") {
        if (responseData.error) {
          return responseData.error;
        }
        if (responseData.message) {
          return responseData.message;
        }
      }
      if (typeof responseData === "string") {
        return responseData;
      }
    }

    // Check for statusMessage
    if (error.statusMessage) {
      return error.statusMessage;
    }

    // Fallback
    return "An unexpected error occurred";
  };

  /**
   * Shows an error toast with the extracted error message
   */
  const showError = (error: any, fallbackMessage?: string) => {
    const message = extractErrorMessage(error);
    toast.error("Error", {
      description: fallbackMessage || message,
    });
  };

  /**
   * Shows a success toast
   */
  const showSuccess = (message: string, description?: string) => {
    toast.success(message, {
      description,
    });
  };

  /**
   * Shows an info toast
   */
  const showInfo = (message: string, description?: string) => {
    toast.info(message, {
      description,
    });
  };

  /**
   * Shows a warning toast
   */
  const showWarning = (message: string, description?: string) => {
    toast.warning(message, {
      description,
    });
  };

  /**
   * Handles API errors with optional custom error handling
   */
  const handleApiError = async (
    apiCall: () => Promise<any>,
    options?: {
      onError?: (error: any) => void;
      showToast?: boolean;
      fallbackMessage?: string;
    }
  ) => {
    try {
      return await apiCall();
    } catch (error) {
      if (options?.onError) {
        options.onError(error);
      }
      if (options?.showToast !== false) {
        showError(error, options?.fallbackMessage);
      }
      throw error;
    }
  };

  return {
    extractErrorMessage,
    showError,
    showSuccess,
    showInfo,
    showWarning,
    handleApiError,
  };
}
