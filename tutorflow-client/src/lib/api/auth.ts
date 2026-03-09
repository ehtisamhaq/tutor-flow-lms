import { api } from "./client";
import { User } from "@/store/auth-store";

export const authApi = {
  getMe: async () => {
    return api.get<User>("/auth/me");
  },

  login: async (credentials: any) => {
    return api.post<{
      user: User;
      tokens: { access_token: string; refresh_token: string };
    }>("/auth/login", credentials);
  },

  refresh: async (refreshToken: string) => {
    return api.post<{
      access_token: string;
      refresh_token: string;
    }>("/auth/refresh", { refresh_token: refreshToken });
  },

  logout: async (refreshToken: string) => {
    return api.post("/auth/logout", { refresh_token: refreshToken });
  },
};
