import { api } from "./client";

export interface Announcement {
  id: string;
  title: string;
  content: string;
  created_at: string;
  updated_at: string;
}

export const announcementApi = {
  list: async () => {
    return api.get<Announcement[]>("/announcements");
  },

  get: async (id: string) => {
    return api.get<Announcement>(`/announcements/${id}`);
  },

  create: async (data: any) => {
    return api.post<Announcement>("/announcements", data);
  },

  update: async (id: string, data: any) => {
    return api.put<Announcement>(`/announcements/${id}`, data);
  },

  delete: async (id: string) => {
    return api.delete(`/announcements/${id}`);
  },
};
