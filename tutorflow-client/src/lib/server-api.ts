// Server-side API utilities for SSR data fetching
const BACKEND_URL =
  process.env.BACKEND_API_URL || "http://localhost:8080/api/v1";

// Internal fetch that hits the backend directly (used by Proxy and can be used by SSR)
async function internalFetch<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T | null> {
  try {
    const response = await fetch(`${BACKEND_URL}${endpoint}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...options.headers,
      },
    });

    if (!response.ok) {
      return null;
    }

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
  } catch (error) {
    return null;
  }
}

export async function serverFetch<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<T | null> {
  return internalFetch<T>(endpoint, options);
}

// Authenticated server fetch - gets token from cookies and hits backend directly
export async function authServerFetch<T>(
  endpoint: string,
  options: RequestInit = {},
  revalidate: number | false = 60
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
  const sessionId = cookieStore.get("sessionId")?.value;
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

export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  limit: number;
  total_pages?: number;
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
