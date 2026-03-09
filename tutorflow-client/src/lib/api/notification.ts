import { api } from "./client";

export interface Notification {
  id: string;
  title: string;
  message: string;
  read: boolean;
  created_at: string;
}

export const notificationApi = {
  list: async () => {
    return api.get<Notification[]>("/notifications");
  },

  markAsRead: async (id: string) => {
    return api.put(`/notifications/${id}/read`, {});
  },

  markAllAsRead: async () => {
    return api.post("/notifications/read-all", {});
  },

  delete: async (id: string) => {
    return api.delete(`/notifications/${id}`);
  },
};
