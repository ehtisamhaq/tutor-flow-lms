"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import {
  Upload,
  Plus,
  Trash2,
  GripVertical,
  Video,
  FileText,
  HelpCircle,
  ChevronDown,
  ChevronUp,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import api from "@/lib/api";
import { cn } from "@/lib/utils";
import {
  createCourseAction,
  updateCourseAction,
  createModuleAction,
  updateModuleAction,
  deleteModuleAction,
  createLessonAction,
  updateLessonAction,
  deleteLessonAction,
  publishCourseAction,
} from "@/app/tutor/courses/actions";
import { QuizEditor } from "./quiz-editor";
// import { courseApi } from "@/lib/course-api"; // Migrated to Server Actions

// Validation schema
const courseSchema = z.object({
  title: z.string().min(5, "Title must be at least 5 characters"),
  description: z.string().min(50, "Description must be at least 50 characters"),
  price: z.number().min(0, "Price must be positive"),
  discount_price: z.number().optional(),
  level: z.enum(["beginner", "intermediate", "advanced"]),
  category_id: z.string().optional(),
  thumbnail: z.any().optional(),
});

type CourseFormData = z.infer<typeof courseSchema>;

interface Module {
  id: string;
  title: string;
  order: number;
  lessons: Lesson[];
  isExpanded: boolean;
}

interface Lesson {
  id: string;
  title: string;
  type: "video" | "text" | "quiz";
  order: number;
  duration_minutes: number;
  video_url?: string;
  content?: string;
}

interface CourseFormProps {
  initialData?: Partial<CourseFormData> & {
    id?: string;
    status?: "draft" | "published" | "archived";
  };
  initialModules?: Module[];
}

export function CourseForm({
  initialData,
  initialModules = [],
}: CourseFormProps) {
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [modules, setModules] = useState<Module[]>(initialModules);
  const [thumbnailPreview, setThumbnailPreview] = useState<string | null>(null);

  // Delete Module State
  const [deleteModuleId, setDeleteModuleId] = useState<string | null>(null);

  useEffect(() => {
    // Preserve isExpanded state when server data updates
    setModules((prevModules) => {
      const expandedState = new Map(
        prevModules.map((m) => [m.id, m.isExpanded])
      );

      return initialModules.map((m) => ({
        ...m,
        isExpanded: expandedState.get(m.id) || false,
      }));
    });
  }, [initialModules]);

  const {
    register,
    handleSubmit,
    formState: { errors },
    watch,
    setValue,
  } = useForm<CourseFormData>({
    resolver: zodResolver(courseSchema),
    defaultValues: {
      title: initialData?.title || "",
      description: initialData?.description || "",
      price: initialData?.price || 0,
      discount_price: initialData?.discount_price,
      level: initialData?.level || "beginner",
      category_id: initialData?.category_id,
    },
  });

  const price = watch("price");

  const handleThumbnailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setValue("thumbnail", file);
      const reader = new FileReader();
      reader.onload = () => setThumbnailPreview(reader.result as string);
      reader.readAsDataURL(file);
    }
  };

  // Module management
  const addModule = async () => {
    if (!initialData?.id) {
      toast.error("Please save the course details first.");
      return;
    }

    try {
      const newOrder = modules.length + 1;
      await createModuleAction(initialData.id, {
        title: `Module ${newOrder}`,
        order: newOrder,
      });

      // router.refresh(); // Handled by Server Action revalidatePath
      toast.success("Module created");
    } catch (error) {
      toast.error("Failed to create module");
    }
  };

  const confirmDeleteModule = (moduleId: string) => {
    setDeleteModuleId(moduleId);
  };

  const handleDeleteModule = async () => {
    if (!initialData?.id || !deleteModuleId) return;
    try {
      await deleteModuleAction(initialData.id, deleteModuleId);
      toast.success("Module deleted");
      setDeleteModuleId(null);
    } catch (error) {
      toast.error("Failed to delete module");
    }
  };

  const updateModuleTitleLocal = (moduleId: string, title: string) => {
    setModules(modules.map((m) => (m.id === moduleId ? { ...m, title } : m)));
  };

  const saveModuleTitle = async (moduleId: string, title: string) => {
    if (!initialData?.id) return;
    try {
      await updateModuleAction(initialData.id, moduleId, { title });
      // router.refresh();
    } catch (error) {
      toast.error("Failed to update module title");
    }
  };

  const toggleModule = (moduleId: string) => {
    setModules(
      modules.map((m) =>
        m.id === moduleId ? { ...m, isExpanded: !m.isExpanded } : m
      )
    );
  };

  // Lesson management
  const addLesson = async (
    moduleId: string,
    type: "video" | "text" | "quiz"
  ) => {
    if (!initialData?.id) return;
    try {
      const module = modules.find((m) => m.id === moduleId);
      const newOrder = (module?.lessons.length || 0) + 1;

      await createLessonAction(moduleId, {
        title: `New ${type.charAt(0).toUpperCase() + type.slice(1)}`,
        type,
        order: newOrder,
      });

      // router.refresh();
      toast.success("Lesson created");
    } catch (error) {
      toast.error("Failed to create lesson");
    }
  };

  const removeLesson = async (moduleId: string, lessonId: string) => {
    if (!initialData?.id) return;
    try {
      await deleteLessonAction(moduleId, lessonId);
      // router.refresh();
      toast.success("Lesson deleted");
    } catch (error) {
      toast.error("Failed to delete lesson");
    }
  };

  const saveLesson = async (
    moduleId: string,
    lessonId: string,
    data: Partial<Lesson>
  ) => {
    try {
      await updateLessonAction(moduleId, lessonId, data);
      // router.refresh();
    } catch (error) {
      toast.error("Failed to update lesson");
    }
  };

  const updateLessonValue = (
    moduleId: string,
    lessonId: string,
    field: keyof Lesson,
    value: any
  ) => {
    setModules(
      modules.map((m) => {
        if (m.id === moduleId) {
          return {
            ...m,
            lessons: m.lessons.map((l) =>
              l.id === lessonId ? { ...l, [field]: value } : l
            ),
          };
        }
        return m;
      })
    );
  };

  const getLessonIcon = (type: string) => {
    switch (type) {
      case "video":
        return Video;
      case "quiz":
        return HelpCircle;
      default:
        return FileText;
    }
  };

  const onSubmit = async (data: CourseFormData) => {
    setIsLoading(true);
    try {
      const formData = new FormData();
      formData.append("title", data.title);
      formData.append("description", data.description);
      formData.append("price", data.price.toString());
      if (data.discount_price) {
        formData.append("discount_price", data.discount_price.toString());
      }
      formData.append("level", data.level);
      if (data.category_id) {
        formData.append("category_id", data.category_id);
      }
      if (data.thumbnail) {
        formData.append("thumbnail", data.thumbnail);
      }
      // Modules are now handled granularly via API

      let result;
      if (initialData?.id) {
        result = await updateCourseAction(initialData.id, formData);
      } else {
        result = await createCourseAction(formData);
      }

      if (result.success) {
        toast.success(
          initialData?.id
            ? "Course updated successfully"
            : "Course created successfully"
        );
        router.push("/tutor/courses");
      } else {
        // Robust error handling for Server Action result
        if (result.error?.code === "VALIDATION_ERROR") {
          const details = result.error.details;
          details.forEach((err: { field: string; message: string }) => {
            // @ts-ignore
            setError(err.field, { message: err.message });
          });
          toast.error("Please fix validation errors");
        } else {
          toast.error(result.error?.message || "Failed to save course");
        }
      }
    } catch (error: any) {
      toast.error("An unexpected error occurred");
    } finally {
      setIsLoading(false);
    }
  };

  const handlePublish = async () => {
    if (!initialData?.id) return;

    const confirm = window.confirm(
      "Are you sure you want to publish this course? It will be visible to all students."
    );
    if (!confirm) return;

    setIsLoading(true);
    try {
      const result = await publishCourseAction(initialData.id);
      if (result.success) {
        toast.success("Course published successfully!");
        router.refresh();
      } else {
        toast.error(result.error?.message || "Failed to publish course");
      }
    } catch (error) {
      toast.error("An error occurred while publishing");
    } finally {
      setIsLoading(false);
    }
  };

  const getStatusBadge = (status?: string) => {
    switch (status) {
      case "published":
        return (
          <span className="bg-green-100 text-green-700 px-2 py-1 rounded text-xs font-semibold uppercase">
            Published
          </span>
        );
      case "archived":
        return (
          <span className="bg-gray-100 text-gray-700 px-2 py-1 rounded text-xs font-semibold uppercase">
            Archived
          </span>
        );
      case "draft":
      default:
        return (
          <span className="bg-yellow-100 text-yellow-700 px-2 py-1 rounded text-xs font-semibold uppercase">
            Draft
          </span>
        );
    }
  };

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="space-y-8">
      {/* Basic Info */}
      <div className="grid md:grid-cols-2 gap-6">
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="title">Course Title *</Label>
            <Input
              id="title"
              placeholder="e.g., Complete Web Development Bootcamp"
              {...register("title")}
            />
            {errors.title && (
              <p className="text-sm text-destructive">{errors.title.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="level">Level *</Label>
            <select
              id="level"
              className="w-full h-10 px-3 border rounded-md bg-background"
              {...register("level")}
            >
              <option value="beginner">Beginner</option>
              <option value="intermediate">Intermediate</option>
              <option value="advanced">Advanced</option>
            </select>
          </div>

          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="price">Price ($) *</Label>
              <Input
                id="price"
                type="number"
                step="0.01"
                min="0"
                {...register("price", { valueAsNumber: true })}
              />
              {errors.price && (
                <p className="text-sm text-destructive">
                  {errors.price.message}
                </p>
              )}
            </div>

            <div className="space-y-2">
              <Label htmlFor="discount_price">Discount Price ($)</Label>
              <Input
                id="discount_price"
                type="number"
                step="0.01"
                min="0"
                max={price}
                {...register("discount_price", { valueAsNumber: true })}
              />
            </div>
          </div>
        </div>

        {/* Thumbnail Upload */}
        <div className="space-y-2">
          <Label>Course Thumbnail</Label>
          <div className="relative aspect-video border-2 border-dashed rounded-lg overflow-hidden hover:border-primary transition-colors">
            {thumbnailPreview ? (
              <img
                src={thumbnailPreview}
                alt="Thumbnail preview"
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground">
                <Upload className="h-8 w-8 mb-2" />
                <p className="text-sm">Click to upload thumbnail</p>
                <p className="text-xs">Recommended: 1280x720px</p>
              </div>
            )}
            <input
              type="file"
              accept="image/*"
              className="absolute inset-0 opacity-0 cursor-pointer"
              onChange={handleThumbnailChange}
            />
          </div>
        </div>
      </div>

      {/* Description */}
      <div className="space-y-2">
        <Label htmlFor="description">Description *</Label>
        <textarea
          id="description"
          rows={4}
          className="w-full px-3 py-2 border rounded-md bg-background resize-none"
          placeholder="Describe what students will learn in this course..."
          {...register("description")}
        />
        {errors.description && (
          <p className="text-sm text-destructive">
            {errors.description.message}
          </p>
        )}
      </div>

      {/* Modules & Lessons */}
      <div className="space-y-4">
        {!initialData?.id ? (
          <Card className="border-dashed bg-muted/30">
            <CardContent className="flex flex-col items-center justify-center py-10 text-muted-foreground space-y-2">
              <h3 className="font-semibold">Curriculum Locked</h3>
              <p className="text-sm">
                Save the course details to start adding modules and lessons.
              </p>
            </CardContent>
          </Card>
        ) : (
          <>
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <h3 className="text-lg font-semibold">Course Content</h3>
                {getStatusBadge(initialData?.status)}
                <p className="text-sm text-muted-foreground">
                  Organize your course into modules and lessons
                </p>
              </div>
              <Button type="button" onClick={addModule}>
                <Plus className="mr-2 h-4 w-4" />
                Add Module
              </Button>
            </div>

            {modules.length === 0 ? (
              <Card className="text-center py-8">
                <CardContent>
                  <p className="text-muted-foreground mb-4">
                    No modules yet. Add your first module to get started.
                  </p>
                  <Button type="button" variant="outline" onClick={addModule}>
                    <Plus className="mr-2 h-4 w-4" />
                    Add Module
                  </Button>
                </CardContent>
              </Card>
            ) : (
              <div className="space-y-4">
                {modules.map((module, index) => (
                  <Card key={module.id}>
                    <CardHeader className="py-3">
                      <div className="flex items-center gap-3">
                        <GripVertical className="h-5 w-5 text-muted-foreground cursor-move" />
                        <span className="text-sm font-medium text-muted-foreground">
                          {index + 1}
                        </span>
                        <Input
                          value={module.title}
                          onChange={(e) =>
                            updateModuleTitleLocal(module.id, e.target.value)
                          }
                          onBlur={() =>
                            saveModuleTitle(module.id, module.title)
                          }
                          className="flex-1"
                        />
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          onClick={() => toggleModule(module.id)}
                        >
                          {module.isExpanded ? (
                            <ChevronUp className="h-4 w-4" />
                          ) : (
                            <ChevronDown className="h-4 w-4" />
                          )}
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="text-destructive"
                          onClick={() => confirmDeleteModule(module.id)}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </div>
                    </CardHeader>

                    {module.isExpanded && (
                      <CardContent className="pt-0">
                        <div className="space-y-2 mb-4">
                          {module.lessons.map((lesson, lessonIndex) => {
                            const Icon = getLessonIcon(lesson.type);
                            return (
                              <div
                                key={lesson.id}
                                className="p-2 bg-muted/50 rounded-lg space-y-2"
                              >
                                <div className="flex items-center gap-3">
                                  <GripVertical className="h-4 w-4 text-muted-foreground cursor-move" />
                                  <span className="text-sm text-muted-foreground w-8">
                                    {lessonIndex + 1}.
                                  </span>
                                  <Icon className="h-4 w-4 text-muted-foreground" />
                                  <Input
                                    value={lesson.title}
                                    onChange={(e) =>
                                      updateLessonValue(
                                        module.id,
                                        lesson.id,
                                        "title",
                                        e.target.value
                                      )
                                    }
                                    onBlur={() =>
                                      saveLesson(module.id, lesson.id, {
                                        title: lesson.title,
                                      })
                                    }
                                    className="flex-1 h-8"
                                  />
                                  <Input
                                    type="number"
                                    value={lesson.duration_minutes}
                                    onChange={(e) =>
                                      updateLessonValue(
                                        module.id,
                                        lesson.id,
                                        "duration_minutes",
                                        parseInt(e.target.value) || 0
                                      )
                                    }
                                    onBlur={() =>
                                      saveLesson(module.id, lesson.id, {
                                        duration_minutes:
                                          lesson.duration_minutes,
                                      })
                                    }
                                    className="w-16 h-8 text-xs"
                                    placeholder="Min"
                                  />
                                  <span className="text-xs text-muted-foreground capitalize">
                                    {lesson.type}
                                  </span>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="h-8 w-8 text-destructive"
                                    onClick={() =>
                                      removeLesson(module.id, lesson.id)
                                    }
                                  >
                                    <Trash2 className="h-3 w-3" />
                                  </Button>
                                </div>

                                {/* Content Fields */}
                                <div className="pl-8">
                                  {lesson.type === "video" && (
                                    <Input
                                      placeholder="Video URL (e.g., YouTube, Vimeo)"
                                      value={lesson.video_url || ""}
                                      onChange={(e) =>
                                        updateLessonValue(
                                          module.id,
                                          lesson.id,
                                          "video_url",
                                          e.target.value
                                        )
                                      }
                                      onBlur={() =>
                                        saveLesson(module.id, lesson.id, {
                                          video_url: lesson.video_url,
                                        })
                                      }
                                      className="h-8 text-xs bg-background"
                                    />
                                  )}
                                  {lesson.type === "text" && (
                                    <textarea
                                      placeholder="Article content..."
                                      value={lesson.content || ""}
                                      onChange={(e) =>
                                        updateLessonValue(
                                          module.id,
                                          lesson.id,
                                          "content",
                                          e.target.value
                                        )
                                      }
                                      onBlur={() =>
                                        saveLesson(module.id, lesson.id, {
                                          content: lesson.content,
                                        })
                                      }
                                      className="flex w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 min-h-[100px]"
                                    />
                                  )}
                                  {lesson.type === "quiz" && (
                                    <div className="pt-2">
                                      <QuizEditor
                                        lessonId={lesson.id}
                                        lessonTitle={lesson.title}
                                      />
                                    </div>
                                  )}
                                </div>
                              </div>
                            );
                          })}
                        </div>

                        <div className="flex gap-2">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => addLesson(module.id, "video")}
                          >
                            <Video className="mr-2 h-4 w-4" />
                            Video
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => addLesson(module.id, "text")}
                          >
                            <FileText className="mr-2 h-4 w-4" />
                            Article
                          </Button>
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            onClick={() => addLesson(module.id, "quiz")}
                          >
                            <HelpCircle className="mr-2 h-4 w-4" />
                            Quiz
                          </Button>
                        </div>
                      </CardContent>
                    )}
                  </Card>
                ))}
              </div>
            )}
          </>
        )}
      </div>

      <Dialog
        open={!!deleteModuleId}
        onOpenChange={(open) => !open && setDeleteModuleId(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Module</DialogTitle>
            <DialogDescription>
              Are you sure you want to delete this module? All lessons within it
              will be permanently deleted. This action cannot be undone.
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteModuleId(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDeleteModule}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Submit */}
      <div className="flex justify-end gap-4 mt-8 pt-6 border-t">
        <Button type="button" variant="outline" asChild>
          <a href="/tutor/courses">Cancel</a>
        </Button>

        {initialData?.id && initialData.status === "draft" && (
          <Button
            type="button"
            variant="secondary"
            onClick={handlePublish}
            disabled={isLoading}
          >
            Publish Course
          </Button>
        )}

        <Button type="submit" disabled={isLoading}>
          {isLoading
            ? "Saving..."
            : initialData?.id
            ? "Update Course"
            : "Create Course"}
        </Button>
      </div>
    </form>
  );
}
