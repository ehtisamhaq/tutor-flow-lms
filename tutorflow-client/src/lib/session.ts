import { cookies } from "next/headers";
import { cache } from "react";

export async function createSession(accessToken: string, refreshToken: string) {
  const cookieStore = await cookies();

  cookieStore.set("accessToken", accessToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 60 * 60,
  });

  cookieStore.set("refreshToken", refreshToken, {
    httpOnly: true,
    secure: process.env.NODE_ENV === "production",
    sameSite: "lax",
    path: "/",
    maxAge: 7 * 24 * 60 * 60,
  });
}

export async function deleteSession() {
  const cookieStore = await cookies();
  cookieStore.delete("accessToken");
  cookieStore.delete("refreshToken");
}

import { authApi } from "./api";

export const getSession = cache(async () => {
  const cookieStore = await cookies();
  const accessToken = cookieStore.get("accessToken")?.value;
  const refreshToken = cookieStore.get("refreshToken")?.value;

  if (!accessToken) return null;

  try {
    const { data: user } = await authApi.getMe();

    if (!user) return null;

    return {
      accessToken,
      refreshToken,
      ...user,
    };
  } catch (error) {
    // If it's a 401, authApi might have already handled it or we handle it here
    console.error("Session verification failed:", error);
    return null;
  }
});
