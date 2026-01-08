// Server-side API utilities for SSR data fetching
const API_URL = process.env.API_URL || "http://localhost:8080/api/v1";

interface ApiResponse<T> {
  success: boolean;
  data: T;
  error?: {
    code: string;
    message: string;
  };
}

export async function serverFetch<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T | null> {
  try {
    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
      next: { revalidate: 60 }, // Cache for 60 seconds
    });

    if (!response.ok) {
      console.error(`API Error: ${response.status} ${response.statusText}`);
      return null;
    }

    const json: ApiResponse<T> = await response.json();
    return json.data;
  } catch (error) {
    console.error("Server fetch error:", error);
    return null;
  }
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
}

export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  course_count: number;
}

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
}

export const serverApi = {
  courses: {
    list: (params?: { page?: number; limit?: number; search?: string }) =>
      serverFetch<PaginatedResponse<Course>>(
        `/courses?page=${params?.page || 1}&limit=${params?.limit || 12}${
          params?.search ? `&search=${params.search}` : ""
        }`
      ),
    getBySlug: (slug: string) => serverFetch<Course>(`/courses/${slug}`),
  },
  categories: {
    list: () => serverFetch<Category[]>("/courses/categories"),
  },
};
