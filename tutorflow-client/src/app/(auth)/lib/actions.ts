"use server";

import { z } from "zod";
import { createSession, deleteSession } from "@/lib/session";
import { redirect } from "next/navigation";
import { User } from "@/store/auth-store";
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
  const BACKEND_URL =
    process.env.BACKEND_API_URL || "http://localhost:8080/api/v1";

  let redirectPath = null;

  try {
    const response = await fetch(`${BACKEND_URL}/auth/login`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password }),
    });

    const data = await response.json();
    console.log(data);

    if (!response.ok) {
      toast.error(data.error?.message || "Invalid credentials");
      return {
        errors: {
          form: [data.error?.message || "Invalid credentials"],
        },
      };
    }

    const { user, tokens } = data.data;

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

export async function logout() {
  await deleteSession();
  redirect("/login");
}
