import { redirect } from "next/navigation";

// Server-side API utilities for SSR data fetching
const BACKEND_URL =
  process.env.BACKEND_API_URL || "http://localhost:8080/api/v1";

// Internal fetch that hits the backend directly (used by Proxy and can be used by SSR)
async function internalFetch<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T | null> {
  let responseStatus = 0;
  try {
    const response = await fetch(`${BACKEND_URL}${endpoint}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    responseStatus = response.status;

    if (!response.ok) {
      if (response.status === 401) {
        // Fall through to 401 check outside try/catch
      } else {
        return null;
      }
    } else {
      const json = await response.json();

      // If it's a paginated response, wrap it in our PaginatedResponse structure
      if (json.meta) {
        return {
          items: json.data,
          total: json.meta.total,
          page: json.meta.page,
          limit: json.meta.per_page,
          total_pages: json.meta.total_pages,
        } as any;
      }

      return json.data;
    }
  } catch (error) {
    // If it's already a redirect error or other Next.js specific error, let it bubble
    if (error && typeof error === "object" && "digest" in error) {
      throw error;
    }
    // Only return null for non-401 non-redirect errors
    if (responseStatus !== 401) {
      return null;
    }
  }

  if (responseStatus === 401) {
    redirect("/login");
  }

  return null;
}

export async function serverFetch<T>(
  endpoint: string,
  options: RequestInit = {},
): Promise<T | null> {
  return internalFetch<T>(endpoint, options);
}

// Authenticated server fetch - gets token from cookies and hits backend directly
export async function authServerFetch<T>(
  endpoint: string,
  options: RequestInit = {},
  revalidate: number | false = 60,
): Promise<T | null> {
  const { cookies } = await import("next/headers");
  const cookieStore = await cookies();
  const token = cookieStore.get("accessToken")?.value;

  const headers: Record<string, string> = {
    ...((options.headers as Record<string, string>) || {}),
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  // Add session ID if it exists
  const sessionId = cookieStore.get("session_id")?.value;
  if (sessionId) {
    headers["X-Session-ID"] = sessionId;
  }

  return internalFetch<T>(endpoint, {
    ...options,
    headers,
    next: revalidate === false ? { revalidate: 0 } : { revalidate },
    cache: revalidate === false ? "no-store" : undefined,
  });
}

// Typed API endpoints
export interface Course {
  id: string;
  title: string;
  slug: string;
  description: string;
  thumbnail_url?: string;
  price: number;
  discount_price?: number;
  level: "beginner" | "intermediate" | "advanced";
  duration_hours: number;
  rating: number;
  total_students: number;
  instructor: {
    id: string;
    first_name: string;
    last_name: string;
    avatar_url?: string;
  };
  modules?: Module[];
}

export interface Module {
  id: string;
  title: string;
  sort_order: number;
  lessons: Lesson[];
}

export interface Lesson {
  id: string;
  title: string;
  type: "video" | "text" | "quiz";
  sort_order: number;
  duration_minutes: number;
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  course_count: number;
}

export interface LearningPath {
  id: string;
  title: string;
  slug: string;
  description: string;
  thumbnail_url?: string;
  level: "beginner" | "intermediate" | "advanced";
  duration_hours: number;
  course_count: number;
  courses: Course[];
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages?: number;
}

// Mock data for when the API is unavailable

export const serverApi = {
  courses: {
    list: async (params?: {
      page?: number;
      limit?: number;
      search?: string;
    }) => {
      const result = await serverFetch<PaginatedResponse<Course>>(
        `/courses?page=${params?.page || 1}&limit=${params?.limit || 12}${
          params?.search ? `&search=${params.search}` : ""
        }`,
      );

      return (
        result || {
          items: [],
          total: 0,
          page: params?.page || 1,
          limit: params?.limit || 12,
        }
      );
    },

    getBySlug: async (slug: string) => {
      return serverFetch<Course>(`/courses/${slug}`);
    },
  },

  categories: {
    list: async () => {
      const result = await serverFetch<Category[]>("/courses/categories");
      return result || [];
    },
  },

  learningPaths: {
    list: async () => {
      const result =
        await serverFetch<PaginatedResponse<LearningPath>>("/learning-paths");
      return result?.items || [];
    },
  },
};
