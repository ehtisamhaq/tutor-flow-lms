"use client";

import { useState, useCallback, useEffect } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import {
  Play,
  FileText,
  HelpCircle,
  Check,
  ChevronDown,
  ChevronRight,
  Menu,
  X,
  Bookmark,
  Clock,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { VideoPlayer } from "@/components/video/video-player";
import { useVideoProgressStore } from "@/store/video-progress-store";
import { cn } from "@/lib/utils";
import api from "@/lib/api";
import { toast } from "sonner";

interface Module {
  id: string;
  title: string;
  order: number;
  lessons: Lesson[];
}

interface Lesson {
  id: string;
  title: string;
  type: "video" | "text" | "quiz";
  duration_minutes: number;
  order: number;
  is_preview: boolean;
  video_url?: string;
  content?: string;
  captions?: { src: string; label: string; language: string }[];
}

interface LearningContentProps {
  courseSlug: string;
  courseId: string;
  modules: Module[];
  currentLessonId?: string;
  currentLesson?: Lesson;
}

export function LearningContent({
  courseSlug,
  courseId,
  modules,
  currentLessonId,
  currentLesson,
}: LearningContentProps) {
  const router = useRouter();
  const [sidebarOpen, setSidebarOpen] = useState(true);
  const [expandedModules, setExpandedModules] = useState<
    Record<string, boolean>
  >(() => {
    const expanded: Record<string, boolean> = {};
    for (const module of modules) {
      if (module.lessons.some((l) => l.id === currentLessonId)) {
        expanded[module.id] = true;
      }
    }
    return expanded;
  });

  const {
    updateProgress,
    getProgress,
    addBookmark,
    removeBookmark,
    getBookmarks,
  } = useVideoProgressStore();

  const lessonProgress = currentLessonId
    ? getProgress(currentLessonId)
    : undefined;
  const lessonBookmarks = currentLessonId ? getBookmarks(currentLessonId) : [];

  const toggleModule = (moduleId: string) => {
    setExpandedModules((prev) => ({
      ...prev,
      [moduleId]: !prev[moduleId],
    }));
  };

  const getLessonIcon = (type: string) => {
    switch (type) {
      case "video":
        return Play;
      case "quiz":
        return HelpCircle;
      default:
        return FileText;
    }
  };

  const handleProgress = useCallback(
    (data: { currentTime: number; duration: number; percent: number }) => {
      if (!currentLessonId) return;

      updateProgress(currentLessonId, {
        courseId,
        currentTime: data.currentTime,
        duration: data.duration,
        percent: data.percent,
      });

      // Sync with server every 30 seconds
      if (Math.floor(data.currentTime) % 30 === 0 && data.currentTime > 0) {
        api
          .post("/progress/update", {
            lesson_id: currentLessonId,
            current_time: data.currentTime,
            percent: data.percent,
          })
          .catch(() => {}); // Silent fail
      }
    },
    [currentLessonId, courseId, updateProgress]
  );

  const handleBookmarkAdd = (time: number) => {
    if (!currentLessonId || !currentLesson) return;

    const title = `Bookmark at ${formatTime(time)}`;
    addBookmark({
      courseId,
      lessonId: currentLessonId,
      time,
      title,
    });
    toast.success("Bookmark added");
  };

  const handleBookmarkRemove = (id: string) => {
    removeBookmark(id);
    toast.success("Bookmark removed");
  };

  const handleComplete = async () => {
    if (!currentLessonId) return;

    try {
      await api.post("/progress/complete", { lesson_id: currentLessonId });
      toast.success("Lesson completed!");
    } catch {
      // Silent fail
    }

    // Find next lesson
    let foundCurrent = false;
    for (const module of modules) {
      for (const lesson of module.lessons) {
        if (foundCurrent) {
          router.push(`/learn/${courseSlug}?lesson=${lesson.id}`);
          return;
        }
        if (lesson.id === currentLessonId) {
          foundCurrent = true;
        }
      }
    }
    // No next lesson - course complete
    toast.success("Course completed! 🎉");
    router.push("/dashboard/my-courses");
  };

  const formatTime = (time: number) => {
    const minutes = Math.floor(time / 60);
    const seconds = Math.floor(time % 60);
    return `${minutes}:${seconds.toString().padStart(2, "0")}`;
  };

  return (
    <div className="flex-1 flex">
      {/* Sidebar */}
      <aside
        className={cn(
          "fixed lg:static inset-y-0 left-0 z-40 w-80 border-r bg-card transform transition-transform lg:translate-x-0 overflow-y-auto",
          sidebarOpen ? "translate-x-0" : "-translate-x-full"
        )}
        style={{ top: "57px", height: "calc(100vh - 57px)" }}
      >
        <div className="flex items-center justify-between p-4 border-b lg:hidden">
          <span className="font-semibold">Course Content</span>
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setSidebarOpen(false)}
          >
            <X className="h-5 w-5" />
          </Button>
        </div>

        <div className="p-4">
          {modules.map((module) => (
            <div key={module.id} className="mb-2">
              <button
                className="w-full flex items-center justify-between p-2 rounded-lg hover:bg-muted text-left"
                onClick={() => toggleModule(module.id)}
              >
                <span className="font-medium text-sm">{module.title}</span>
                {expandedModules[module.id] ? (
                  <ChevronDown className="h-4 w-4" />
                ) : (
                  <ChevronRight className="h-4 w-4" />
                )}
              </button>

              {expandedModules[module.id] && (
                <div className="ml-2 mt-1 space-y-1">
                  {module.lessons.map((lesson) => {
                    const Icon = getLessonIcon(lesson.type);
                    const isActive = lesson.id === currentLessonId;
                    const progress = getProgress(lesson.id);

                    return (
                      <Link
                        key={lesson.id}
                        href={`/learn/${courseSlug}?lesson=${lesson.id}`}
                        className={cn(
                          "flex items-center gap-3 p-2 rounded-lg text-sm transition-colors",
                          isActive
                            ? "bg-primary/10 text-primary"
                            : "hover:bg-muted text-muted-foreground hover:text-foreground"
                        )}
                      >
                        <Icon className="h-4 w-4 shrink-0" />
                        <span className="flex-1 truncate">{lesson.title}</span>
                        <div className="flex items-center gap-1">
                          {progress && progress.percent >= 95 && (
                            <Check className="h-4 w-4 text-green-600" />
                          )}
                          <span className="text-xs">
                            {lesson.duration_minutes}m
                          </span>
                        </div>
                      </Link>
                    );
                  })}
                </div>
              )}
            </div>
          ))}
        </div>

        {/* Bookmarks Section */}
        {lessonBookmarks.length > 0 && (
          <div className="p-4 border-t">
            <h4 className="text-sm font-semibold flex items-center gap-2 mb-3">
              <Bookmark className="h-4 w-4" />
              Bookmarks
            </h4>
            <div className="space-y-2">
              {lessonBookmarks.map((bookmark) => (
                <button
                  key={bookmark.id}
                  className="w-full text-left text-sm p-2 rounded hover:bg-muted flex items-center gap-2"
                  onClick={() => {
                    // TODO: Seek to bookmark time via ref
                  }}
                >
                  <Clock className="h-3 w-3 text-muted-foreground" />
                  <span>{formatTime(bookmark.time)}</span>
                  <span className="text-muted-foreground truncate">
                    {bookmark.title}
                  </span>
                </button>
              ))}
            </div>
          </div>
        )}
      </aside>

      {/* Mobile sidebar toggle */}
      <Button
        variant="outline"
        size="icon"
        className="fixed bottom-4 left-4 z-50 lg:hidden shadow-lg"
        onClick={() => setSidebarOpen(!sidebarOpen)}
      >
        <Menu className="h-5 w-5" />
      </Button>

      {/* Content Area */}
      <main className="flex-1 flex flex-col">
        {currentLesson ? (
          <>
            {/* Video Player */}
            {currentLesson.type === "video" && currentLesson.video_url ? (
              <VideoPlayer
                src={currentLesson.video_url}
                title={currentLesson.title}
                captions={currentLesson.captions}
                bookmarks={lessonBookmarks}
                initialTime={lessonProgress?.currentTime || 0}
                onProgress={handleProgress}
                onBookmarkAdd={handleBookmarkAdd}
                onBookmarkRemove={handleBookmarkRemove}
                onComplete={handleComplete}
              />
            ) : currentLesson.type === "video" ? (
              <div className="aspect-video bg-black flex items-center justify-center">
                <div className="text-white text-center">
                  <Play className="h-16 w-16 mx-auto mb-4 opacity-50" />
                  <p className="text-lg opacity-75">Video Player</p>
                  <p className="text-sm opacity-50">{currentLesson.title}</p>
                </div>
              </div>
            ) : currentLesson.type === "quiz" ? (
              <div className="aspect-video bg-muted flex items-center justify-center">
                <div className="text-center">
                  <HelpCircle className="h-16 w-16 mx-auto mb-4 opacity-50" />
                  <p className="text-lg font-medium">Quiz</p>
                  <p className="text-sm text-muted-foreground">
                    {currentLesson.title}
                  </p>
                </div>
              </div>
            ) : (
              <div className="p-8 prose prose-slate dark:prose-invert max-w-none">
                <div
                  dangerouslySetInnerHTML={{
                    __html:
                      currentLesson.content || "<p>No content available</p>",
                  }}
                />
              </div>
            )}

            {/* Lesson Info */}
            <div className="p-6 border-t">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-xl font-semibold">{currentLesson.title}</h2>
                <Button onClick={handleComplete}>
                  <Check className="mr-2 h-4 w-4" />
                  Mark Complete & Continue
                </Button>
              </div>
              <p className="text-muted-foreground">
                {currentLesson.type === "video" && (
                  <span className="flex items-center gap-2">
                    <Clock className="h-4 w-4" />
                    {currentLesson.duration_minutes} minutes
                    {lessonProgress && (
                      <span className="text-primary">
                        • {Math.round(lessonProgress.percent)}% watched
                      </span>
                    )}
                  </span>
                )}
              </p>
            </div>
          </>
        ) : (
          <div className="flex-1 flex items-center justify-center text-muted-foreground">
            <p>Select a lesson to begin</p>
          </div>
        )}
      </main>
    </div>
  );
}
