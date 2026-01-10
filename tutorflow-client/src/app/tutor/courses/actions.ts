"use server";

import { revalidatePath, revalidateTag } from "next/cache";
import api from "@/lib/api";

// Helper to check for Next.js redirect errors
const isRedirectError = (error: any) => {
  return (
    error &&
    typeof error === "object" &&
    error.digest &&
    error.digest.startsWith("NEXT_REDIRECT")
  );
};

export async function createCourseAction(formData: FormData) {
  console.log("createCourseAction received keys:", Array.from(formData.keys()));
  try {
    const result: any = await api.post("/courses", formData);
    console.log("api.post result:", result);

    revalidatePath("/tutor/courses");
    return {
      success: true,
      data: result.data,
    };
  } catch (error: any) {
    // If it's a redirect error from the api utility (e.g. 401), re-throw it
    if (isRedirectError(error)) {
      throw error;
    }
    // For other failures (like validation), return the error object
    return {
      success: false,
      error: error.response?.data?.error || {
        message: error.message || "Failed to create course",
      },
    };
  }
}

export async function updateCourseAction(id: string, formData: FormData) {
  try {
    const result: any = await api.put(`/courses/${id}`, formData);

    revalidatePath("/tutor/courses");
    revalidatePath(`/tutor/courses/${id}`);
    revalidateTag(`course-${id}`, "max");
    return {
      success: true,
      data: result.data,
    };
  } catch (error: any) {
    if (isRedirectError(error)) {
      throw error;
    }
    return {
      success: false,
      error: error.response?.data?.error || {
        message: error.message || "Failed to update course",
      },
    };
  }
}

// --- Module Actions ---

export async function createModuleAction(
  courseId: string,
  data: { title: string; order: number }
) {
  try {
    const result: any = await api.post(`/courses/${courseId}/modules`, data);
    revalidatePath(`/tutor/courses/${courseId}`);
    revalidateTag(`course-${courseId}`, "max");
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error: error.message || "Failed to create module",
    };
  }
}

export async function updateModuleAction(
  courseId: string,
  moduleId: string,
  data: { title?: string; order?: number; is_published?: boolean }
) {
  try {
    const result: any = await api.put(
      `/courses/${courseId}/modules/${moduleId}`,
      data
    );
    revalidatePath(`/tutor/courses/${courseId}`);
    revalidateTag(`course-${courseId}`, "max");
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error: error.message || "Failed to update module",
    };
  }
}

export async function deleteModuleAction(courseId: string, moduleId: string) {
  try {
    await api.delete(`/courses/${courseId}/modules/${moduleId}`);
    revalidatePath(`/tutor/courses/${courseId}`);
    revalidateTag(`course-${courseId}`, "max");
    return { success: true };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error: error.message || "Failed to delete module",
    };
  }
}

// --- Lesson Actions ---

export async function createLessonAction(
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
) {
  try {
    const { type, order, ...rest } = data;
    const payload = {
      ...rest,
      lesson_type: type, // Map frontend 'type' to backend 'lesson_type'
      access_type: "enrolled",
    };
    const result: any = await api.post(
      `/courses/modules/${moduleId}/lessons`,
      payload
    );
    // Note: We need the courseId to revalidate the page.
    // Since we don't have it easily here, we might need to rely on the client to refresh
    // or pass courseId to this action.
    // For now, let's just return success and let client handle refresh or revalidate generic path.
    revalidatePath("/tutor/courses/[id]", "page");
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error: error.message || "Failed to create lesson",
    };
  }
}

export async function updateLessonAction(
  moduleId: string,
  lessonId: string,
  data: any
) {
  try {
    const result: any = await api.put(
      `/courses/modules/${moduleId}/lessons/${lessonId}`,
      data
    );
    revalidatePath("/tutor/courses/[id]", "page");
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error: error.message || "Failed to update lesson",
    };
  }
}

export async function deleteLessonAction(moduleId: string, lessonId: string) {
  try {
    await api.delete(`/courses/modules/${moduleId}/lessons/${lessonId}`);
    revalidatePath("/tutor/courses/[id]", "page");
    return { success: true };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error: error.message || "Failed to delete lesson",
    };
  }
}

// --- Quiz Actions ---

export async function getQuizByLessonAction(lessonId: string) {
  try {
    const result: any = await api.get(`/quizzes/lesson/${lessonId}`, {
      next: { tags: [`quiz-lesson-${lessonId}`] },
    });
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    // 404 is expected if quiz doesn't exist yet
    if (error.response?.status === 404) {
      return { success: true, data: null };
    }
    return {
      success: false,
      error:
        error.response?.data?.error?.message ||
        error.message ||
        "Failed to fetch quiz",
    };
  }
}

export async function createQuizAction(data: {
  lesson_id: string;
  title: string;
  description?: string;
  passing_score?: number;
  max_attempts?: number;
}) {
  try {
    const result: any = await api.post("/quizzes", data);
    revalidateTag(`quiz-lesson-${data.lesson_id}`, "max");
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error:
        error.response?.data?.error?.message ||
        error.message ||
        "Action failed",
    };
  }
}

export async function updateQuizAction(quizId: string, data: any) {
  try {
    const result: any = await api.put(`/quizzes/${quizId}`, data);
    if (result.data?.lesson_id) {
      revalidateTag(`quiz-lesson-${result.data.lesson_id}`, "max");
    }
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error:
        error.response?.data?.error?.message ||
        error.message ||
        "Action failed",
    };
  }
}

export async function addQuestionAction(quizId: string, data: any) {
  try {
    const result: any = await api.post(`/quizzes/${quizId}/questions`, data);
    // Since we don't know lessonId efficiently here without fetching,
    // we might need a better strategy or pass lessonId from client.
    // For now, let's assume the question response might have it or we can't easily revalidate
    // the *lesson* tag specifically without it.
    // HOWEVER, the `updateQuizAction` above has it.
    // Let's rely on client reloading for now, OR return data to client.
    // Actually, `QuizEditor` calls `loadQuiz()` manually on success, so client-side refetch works.
    // But for caching correctness:
    if (result.data?.quiz?.lesson_id) {
      revalidateTag(`quiz-lesson-${result.data.quiz.lesson_id}`, "max");
    }
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error:
        error.response?.data?.error?.message ||
        error.message ||
        "Action failed",
    };
  }
}

export async function updateQuestionAction(questionId: string, data: any) {
  try {
    const result: any = await api.put(`/quizzes/questions/${questionId}`, data);
    if (result.data?.quiz?.lesson_id) {
      revalidateTag(`quiz-lesson-${result.data.quiz.lesson_id}`, "max");
    }
    return { success: true, data: result.data };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error:
        error.response?.data?.error?.message ||
        error.message ||
        "Action failed",
    };
  }
}

export async function deleteQuestionAction(questionId: string) {
  try {
    await api.delete(`/quizzes/questions/${questionId}`);
    return { success: true };
  } catch (error: any) {
    if (isRedirectError(error)) throw error;
    return {
      success: false,
      error:
        error.response?.data?.error?.message ||
        error.message ||
        "Action failed",
    };
  }
}
export async function publishCourseAction(id: string) {
  try {
    const result: any = await api.patch(`/courses/${id}/publish`, {});

    revalidatePath("/tutor/courses");
    revalidatePath(`/tutor/courses/${id}`);
    revalidateTag(`course-${id}`, "max");
    return {
      success: true,
      data: result.data,
    };
  } catch (error: any) {
    if (isRedirectError(error)) {
      throw error;
    }
    return {
      success: false,
      error: error.response?.data?.error || {
        message: error.message || "Failed to publish course",
      },
    };
  }
}
