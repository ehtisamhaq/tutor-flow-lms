"use server";

import { z } from "zod";
import { createSession, deleteSession } from "@/lib/session";
import { redirect } from "next/navigation";
import { User } from "@/store/auth-store";
import { authApi } from "@/lib/api";
import { toast } from "sonner";

const loginSchema = z.object({
  email: z.string().email("Please enter a valid email"),
  password: z.string().min(1, "Password is required"),
});

export type LoginState =
  | {
      errors?: {
        email?: string[];
        password?: string[];
        form?: string[];
      };
      message?: string;
    }
  | undefined;

export async function login(
  prevState: LoginState,
  formData: FormData,
): Promise<LoginState> {
  const validatedFields = loginSchema.safeParse({
    email: formData.get("email"),
    password: formData.get("password"),
  });

  if (!validatedFields.success) {
    return {
      errors: validatedFields.error.flatten().fieldErrors,
    };
  }

  const { email, password } = validatedFields.data;

  let redirectPath = null;

  try {
    const { data } = await authApi.login({ email, password });
    const { user, tokens } = data;

    await createSession(tokens.access_token, tokens.refresh_token);

    // We can't redirect with user data, so the client will need to fetch it
    // or we assume the client store is updated separately (which is tricky with server actions).
    // For now, we'll just redirect and let the client fetch "me" or handling it via middleware/layout.

    switch (user.role) {
      case "student":
        redirectPath = "/dashboard";
        break;
      case "tutor":
        redirectPath = "/tutor";
        break;
      case "manager":
        redirectPath = "/manager";
        break;
      case "admin":
        redirectPath = "/admin";
        break;
    }
  } catch (error) {
    return {
      errors: {
        form: ["Failed to connect to the server"],
      },
    };
  }

  if (redirectPath) {
    redirect(redirectPath);
  }
}

export async function getAccessToken() {
  const { cookies } = await import("next/headers");
  const cookieStore = await cookies();
  return cookieStore.get("accessToken")?.value;
}

export async function logout() {
  await deleteSession();
  redirect("/login");
}

export async function getDemoCredentials(role: string) {
  if (process.env.NODE_ENV !== "development") {
    return null;
  }

  switch (role) {
    case "admin":
      return {
        email: process.env.DEMO_ADMIN_EMAIL || "admin@tutorflow.com",
        password: process.env.DEMO_ADMIN_PASSWORD || "password123",
      };
    case "manager":
      return {
        email: process.env.DEMO_MANAGER_EMAIL || "manager@tutorflow.com",
        password: process.env.DEMO_MANAGER_PASSWORD || "password123",
      };
    case "tutor":
      return {
        email: process.env.DEMO_TUTOR_EMAIL || "tutor@tutorflow.com",
        password: process.env.DEMO_TUTOR_PASSWORD || "password123",
      };
    case "student":
      return {
        email: process.env.DEMO_STUDENT_EMAIL || "student@tutorflow.com",
        password: process.env.DEMO_STUDENT_PASSWORD || "password123",
      };
    default:
      return null;
  }
}
