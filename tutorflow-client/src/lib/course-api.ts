import { api } from "./api";
import { Module, Lesson } from "./server-api"; // Import types

export const courseApi = {
  // --- Modules ---
  createModule: async (
    courseId: string,
    data: { title: string; order: number }
  ) => {
    return api.post(`/courses/${courseId}/modules`, data);
  },

  updateModule: async (
    courseId: string,
    moduleId: string,
    data: { title?: string; order?: number; is_published?: boolean }
  ) => {
    return api.put(`/courses/${courseId}/modules/${moduleId}`, data);
  },

  deleteModule: async (courseId: string, moduleId: string) => {
    return api.delete(`/courses/${courseId}/modules/${moduleId}`);
  },

  reorderModules: async (courseId: string, moduleIds: string[]) => {
    return api.patch(`/courses/${courseId}/modules/reorder`, {
      module_ids: moduleIds,
    });
  },

  // --- Lessons ---
  createLesson: async (
    moduleId: string,
    data: {
      title: string;
      type: "video" | "text" | "quiz";
      order: number;
      description?: string;
      content?: string;
      video_url?: string;
      duration_minutes?: number;
    }
  ) => {
    const { type, order, ...rest } = data;
    return api.post(`/courses/modules/${moduleId}/lessons`, {
      ...rest,
      lesson_type: type, // Map frontend 'type' to backend 'lesson_type'
      access_type: "enrolled",
    });
  },

  updateLesson: async (
    moduleId: string,
    lessonId: string,
    data: {
      title?: string;
      description?: string;
      content?: string;
      video_url?: string;
      duration_minutes?: number;
      order?: number;
      is_published?: boolean;
      is_preview?: boolean;
    }
  ) => {
    return api.put(`/courses/modules/${moduleId}/lessons/${lessonId}`, data);
  },

  deleteLesson: async (moduleId: string, lessonId: string) => {
    return api.delete(`/courses/modules/${moduleId}/lessons/${lessonId}`);
  },

  // NOTE: ReorderLessons backend endpoint might be missing.
  // We will assume it exists or implement it later.
  // For now, this might fail if the endpoint isn't there.
  reorderLessons: async (moduleId: string, lessonIds: string[]) => {
    // Placeholder: If backend doesn't have a reorder endpoint for lessons,
    // we might need to rely on individual updates or add the endpoint.
    // return api.patch(`/courses/modules/${moduleId}/lessons/reorder`, { lesson_ids: lessonIds });
    console.warn("reorderLessons API not fully implemented on backend yet.");
    return Promise.resolve();
  },
};
