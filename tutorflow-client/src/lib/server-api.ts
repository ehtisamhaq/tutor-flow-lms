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
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000); // 5 second timeout

    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      signal: controller.signal,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
      next: { revalidate: 60 }, // Cache for 60 seconds
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      console.error(`API Error: ${response.status} ${response.statusText}`);
      return null;
    }

    const json: ApiResponse<T> = await response.json();
    return json.data;
  } catch (error) {
    // Don't log in production to avoid noise when backend is down
    if (process.env.NODE_ENV === "development") {
      console.warn(
        `Server fetch failed for ${endpoint}:`,
        error instanceof Error ? error.message : "Unknown error"
      );
    }
    return null;
  }
}

// Authenticated server fetch - gets token from cookies
export async function authServerFetch<T>(
  endpoint: string,
  options: RequestInit = {},
  revalidate: number | false = 60
): Promise<T | null> {
  // Dynamic import to avoid bundling issues
  const { cookies } = await import("next/headers");
  const cookieStore = await cookies();
  const token = cookieStore.get("accessToken");

  if (!token) {
    return null;
  }

  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);

    const response = await fetch(`${API_URL}${endpoint}`, {
      ...options,
      signal: controller.signal,
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${token.value}`,
        ...options.headers,
      },
      next: revalidate === false ? { revalidate: 0 } : { revalidate },
      cache: revalidate === false ? "no-store" : undefined,
    });

    clearTimeout(timeoutId);

    if (!response.ok) {
      if (process.env.NODE_ENV === "development") {
        console.error(
          `Auth API Error: ${response.status} ${response.statusText}`
        );
      }
      return null;
    }

    const json: ApiResponse<T> = await response.json();
    return json.data;
  } catch (error) {
    if (process.env.NODE_ENV === "development") {
      console.warn(
        `Auth fetch failed for ${endpoint}:`,
        error instanceof Error ? error.message : "Unknown error"
      );
    }
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

// Mock data for when the API is unavailable
const MOCK_COURSES: Course[] = [
  {
    id: "1",
    title: "Complete Web Development Bootcamp",
    slug: "complete-web-development-bootcamp",
    description:
      "Learn web development from scratch with HTML, CSS, JavaScript, React, Node.js, and more.",
    thumbnail_url: undefined,
    price: 99.99,
    discount_price: 49.99,
    level: "beginner",
    duration_hours: 48,
    rating: 4.8,
    total_students: 12500,
    instructor: {
      id: "1",
      first_name: "John",
      last_name: "Doe",
    },
  },
  {
    id: "2",
    title: "Advanced React Patterns",
    slug: "advanced-react-patterns",
    description:
      "Master advanced React patterns including hooks, context, and performance optimization.",
    thumbnail_url: undefined,
    price: 79.99,
    level: "advanced",
    duration_hours: 24,
    rating: 4.9,
    total_students: 5200,
    instructor: {
      id: "2",
      first_name: "Jane",
      last_name: "Smith",
    },
  },
  {
    id: "3",
    title: "Python for Data Science",
    slug: "python-for-data-science",
    description:
      "Learn Python programming for data science, machine learning, and analytics.",
    thumbnail_url: undefined,
    price: 89.99,
    discount_price: 59.99,
    level: "intermediate",
    duration_hours: 36,
    rating: 4.7,
    total_students: 8900,
    instructor: {
      id: "3",
      first_name: "Alex",
      last_name: "Johnson",
    },
  },
];

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
        }`
      );

      // Return mock data if API is unavailable
      if (!result) {
        return {
          items: MOCK_COURSES,
          total: MOCK_COURSES.length,
          page: params?.page || 1,
          limit: params?.limit || 12,
        };
      }

      return result;
    },

    getBySlug: async (slug: string) => {
      const result = await serverFetch<Course>(`/courses/${slug}`);

      // Return mock course if API is unavailable
      if (!result) {
        const mockCourse = MOCK_COURSES.find((c) => c.slug === slug);
        return mockCourse || MOCK_COURSES[0];
      }

      return result;
    },
  },

  categories: {
    list: async () => {
      const result = await serverFetch<Category[]>("/courses/categories");
      return result || [];
    },
  },
};
