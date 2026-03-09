import { api } from "./client";
import { User } from "@/store/auth-store";

export const userApi = {
  getProfile: async () => {
    return api.get<User>("/users/profile");
  },

  updateProfile: async (data: Partial<User>) => {
    return api.put<User>("/users/profile", data);
  },

  listUsers: async (params?: {
    page?: number;
    limit?: number;
    search?: string;
  }) => {
    return api.get<any>("/users", {
      // params are handled via URL in the standard api.get but we can helper here
    });
  },
};
