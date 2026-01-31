import { useState, useCallback } from 'react';
import { Platform } from 'react-native';

// Base URL for API - adjust based on platform
const getBaseUrl = (): string => {
  if (Platform.OS === 'web') {
    // In web, use relative URLs (same origin)
    return '';
  }
  // For native, use the actual server URL
  // In development, this would be your local IP
  return process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080';
};

interface ApiState<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
}

interface UseApiResult<T> extends ApiState<T> {
  execute: () => Promise<T | null>;
  reset: () => void;
}

interface UseApiOptions {
  immediate?: boolean;
}

export function useApi<T>(
  endpoint: string,
  options: RequestInit = {},
  apiOptions: UseApiOptions = {}
): UseApiResult<T> {
  const [state, setState] = useState<ApiState<T>>({
    data: null,
    loading: apiOptions.immediate ?? false,
    error: null,
  });

  const execute = useCallback(async (): Promise<T | null> => {
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const baseUrl = getBaseUrl();
      const url = `${baseUrl}${endpoint}`;

      const response = await fetch(url, {
        ...options,
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          ...options.headers,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Request failed: ${response.status}`);
      }

      const data = await response.json();
      setState({ data, loading: false, error: null });
      return data;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setState({ data: null, loading: false, error: errorMessage });
      return null;
    }
  }, [endpoint, options]);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return { ...state, execute, reset };
}

// Convenience hooks for common HTTP methods

export function useGet<T>(endpoint: string, options: UseApiOptions = {}) {
  return useApi<T>(endpoint, { method: 'GET' }, options);
}

export function usePost<T, B = unknown>(endpoint: string) {
  const [state, setState] = useState<ApiState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(async (body: B): Promise<T | null> => {
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const baseUrl = getBaseUrl();
      const url = `${baseUrl}${endpoint}`;

      const response = await fetch(url, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Request failed: ${response.status}`);
      }

      const data = await response.json();
      setState({ data, loading: false, error: null });
      return data;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setState({ data: null, loading: false, error: errorMessage });
      return null;
    }
  }, [endpoint]);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return { ...state, execute, reset };
}

export function usePut<T, B = unknown>(endpoint: string) {
  const [state, setState] = useState<ApiState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(async (body: B): Promise<T | null> => {
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const baseUrl = getBaseUrl();
      const url = `${baseUrl}${endpoint}`;

      const response = await fetch(url, {
        method: 'PUT',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(body),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Request failed: ${response.status}`);
      }

      const data = await response.json();
      setState({ data, loading: false, error: null });
      return data;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setState({ data: null, loading: false, error: errorMessage });
      return null;
    }
  }, [endpoint]);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return { ...state, execute, reset };
}

export function useDelete<T>(endpoint: string) {
  const [state, setState] = useState<ApiState<T>>({
    data: null,
    loading: false,
    error: null,
  });

  const execute = useCallback(async (): Promise<T | null> => {
    setState(prev => ({ ...prev, loading: true, error: null }));

    try {
      const baseUrl = getBaseUrl();
      const url = `${baseUrl}${endpoint}`;

      const response = await fetch(url, {
        method: 'DELETE',
        credentials: 'include',
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `Request failed: ${response.status}`);
      }

      const data = await response.json();
      setState({ data, loading: false, error: null });
      return data;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Unknown error';
      setState({ data: null, loading: false, error: errorMessage });
      return null;
    }
  }, [endpoint]);

  const reset = useCallback(() => {
    setState({ data: null, loading: false, error: null });
  }, []);

  return { ...state, execute, reset };
}
