import { create } from "zustand";
import { persist } from "zustand/middleware";

interface VideoProgress {
  courseId: string;
  lessonId: string;
  currentTime: number;
  duration: number;
  percent: number;
  lastWatched: string;
}

interface VideoBookmark {
  id: string;
  courseId: string;
  lessonId: string;
  time: number;
  title: string;
  created_at: string;
}

interface VideoProgressState {
  progress: Record<string, VideoProgress>; // key: lessonId
  bookmarks: VideoBookmark[];

  // Actions
  updateProgress: (lessonId: string, data: Partial<VideoProgress>) => void;
  getProgress: (lessonId: string) => VideoProgress | undefined;
  addBookmark: (bookmark: Omit<VideoBookmark, "id" | "created_at">) => void;
  removeBookmark: (id: string) => void;
  getBookmarks: (lessonId: string) => VideoBookmark[];
}

export const useVideoProgressStore = create<VideoProgressState>()(
  persist(
    (set, get) => ({
      progress: {},
      bookmarks: [],

      updateProgress: (lessonId, data) => {
        set((state) => ({
          progress: {
            ...state.progress,
            [lessonId]: {
              ...state.progress[lessonId],
              ...data,
              lessonId,
              lastWatched: new Date().toISOString(),
            },
          },
        }));
      },

      getProgress: (lessonId) => {
        return get().progress[lessonId];
      },

      addBookmark: (bookmark) => {
        const id = `bookmark_${Date.now()}`;
        const newBookmark: VideoBookmark = {
          ...bookmark,
          id,
          created_at: new Date().toISOString(),
        };
        set((state) => ({
          bookmarks: [...state.bookmarks, newBookmark],
        }));
      },

      removeBookmark: (id) => {
        set((state) => ({
          bookmarks: state.bookmarks.filter((b) => b.id !== id),
        }));
      },

      getBookmarks: (lessonId) => {
        return get().bookmarks.filter((b) => b.lessonId === lessonId);
      },
    }),
    {
      name: "video-progress",
      partialize: (state) => ({
        progress: state.progress,
        bookmarks: state.bookmarks,
      }),
    }
  )
);
