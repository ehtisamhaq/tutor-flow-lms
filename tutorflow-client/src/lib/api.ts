import Cookies from "js-cookie";
import { redirect } from "next/navigation";

const BASE_URL =
  typeof window === "undefined"
    ? process.env.BACKEND_API_URL || "http://localhost:8080/api/v1"
    : "/api";

async function getAuthToken() {
  if (typeof window === "undefined") {
    try {
      const { cookies } = await import("next/headers");
      return (await cookies()).get("accessToken")?.value;
    } catch (e) {
      // Not in a request context (e.g. build time or static generation without headers)
      return null;
    }
  }
  return Cookies.get("accessToken");
}

interface RawApiResponse {
  data?: unknown;
  meta?: {
    total: number;
    page: number;
    per_page: number;
    total_pages: number;
  };
}

export interface ApiResponse<T> {
  data: T;
}

async function handleResponse<T>(response: Response): Promise<ApiResponse<T>> {
  const isJson = response.headers
    .get("content-type")
    ?.includes("application/json");
  const data = isJson ? await response.json() : await response.text();

  if (response.status === 401) {
    if (typeof window === "undefined") {
      redirect("/login");
    } else {
      window.location.href = "/login";
      return new Promise(() => {});
    }
  }

  if (!response.ok) {
    return Promise.reject({
      response: {
        data: isJson ? data : { error: { message: data } },
        status: response.status,
      },
    });
  }

  if (isJson && data && typeof data === "object") {
    const rawData = data as RawApiResponse;
    if (rawData.meta) {
      return {
        data: {
          items: rawData.data,
          total: rawData.meta.total,
          page: rawData.meta.page,
          limit: rawData.meta.per_page,
          total_pages: rawData.meta.total_pages,
        } as unknown as T,
      };
    }
    if (rawData.data !== undefined) {
      return { data: rawData.data as T };
    }
  }

  return { data: data as T };
}

export const api = {
  get: async <T>(url: string, options: RequestInit = {}) => {
    const token = await getAuthToken();
    const response = await fetch(`${BASE_URL}${url}`, {
      ...options,
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });
    return handleResponse<T>(response);
  },

  post: async <T>(url: string, body: unknown, options: RequestInit = {}) => {
    const token = await getAuthToken();
    const isFormData = body instanceof FormData;

    const response = await fetch(`${BASE_URL}${url}`, {
      ...options,
      method: "POST",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(isFormData ? {} : { "Content-Type": "application/json" }),
        ...options.headers,
      },
      body: isFormData ? body : JSON.stringify(body),
    });
    return handleResponse<T>(response);
  },

  put: async <T>(url: string, body: unknown, options: RequestInit = {}) => {
    const token = await getAuthToken();
    const isFormData = body instanceof FormData;

    const response = await fetch(`${BASE_URL}${url}`, {
      ...options,
      method: "PUT",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(isFormData ? {} : { "Content-Type": "application/json" }),
        ...options.headers,
      },
      body: isFormData ? body : JSON.stringify(body),
    });
    return handleResponse<T>(response);
  },

  patch: async <T>(url: string, body: unknown, options: RequestInit = {}) => {
    const token = await getAuthToken();
    const isFormData = body instanceof FormData;

    const response = await fetch(`${BASE_URL}${url}`, {
      ...options,
      method: "PATCH",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(isFormData ? {} : { "Content-Type": "application/json" }),
        ...options.headers,
      },
      body: isFormData ? body : JSON.stringify(body),
    });
    return handleResponse<T>(response);
  },

  delete: async <T>(url: string, options: RequestInit = {}) => {
    const token = await getAuthToken();
    const response = await fetch(`${BASE_URL}${url}`, {
      ...options,
      method: "DELETE",
      headers: {
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });
    return handleResponse<T>(response);
  },
};

export default api;
