"use client";

import { useState } from "react";
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
import api from "@/lib/api";
import { cn } from "@/lib/utils";

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
}

interface CourseFormProps {
  initialData?: Partial<CourseFormData> & { id?: string };
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
  const addModule = () => {
    const newModule: Module = {
      id: `temp_${Date.now()}`,
      title: `Module ${modules.length + 1}`,
      order: modules.length + 1,
      lessons: [],
      isExpanded: true,
    };
    setModules([...modules, newModule]);
  };

  const removeModule = (moduleId: string) => {
    setModules(modules.filter((m) => m.id !== moduleId));
  };

  const updateModuleTitle = (moduleId: string, title: string) => {
    setModules(modules.map((m) => (m.id === moduleId ? { ...m, title } : m)));
  };

  const toggleModule = (moduleId: string) => {
    setModules(
      modules.map((m) =>
        m.id === moduleId ? { ...m, isExpanded: !m.isExpanded } : m
      )
    );
  };

  // Lesson management
  const addLesson = (moduleId: string, type: "video" | "text" | "quiz") => {
    setModules(
      modules.map((m) => {
        if (m.id === moduleId) {
          const newLesson: Lesson = {
            id: `temp_lesson_${Date.now()}`,
            title: `New ${type.charAt(0).toUpperCase() + type.slice(1)}`,
            type,
            order: m.lessons.length + 1,
            duration_minutes: 0,
          };
          return { ...m, lessons: [...m.lessons, newLesson] };
        }
        return m;
      })
    );
  };

  const removeLesson = (moduleId: string, lessonId: string) => {
    setModules(
      modules.map((m) => {
        if (m.id === moduleId) {
          return {
            ...m,
            lessons: m.lessons.filter((l) => l.id !== lessonId),
          };
        }
        return m;
      })
    );
  };

  const updateLessonTitle = (
    moduleId: string,
    lessonId: string,
    title: string
  ) => {
    setModules(
      modules.map((m) => {
        if (m.id === moduleId) {
          return {
            ...m,
            lessons: m.lessons.map((l) =>
              l.id === lessonId ? { ...l, title } : l
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
      formData.append("modules", JSON.stringify(modules));

      if (initialData?.id) {
        await api.put(`/courses/${initialData.id}`, formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        toast.success("Course updated successfully");
      } else {
        await api.post("/courses", formData, {
          headers: { "Content-Type": "multipart/form-data" },
        });
        toast.success("Course created successfully");
      }

      router.push("/tutor/courses");
      router.refresh();
    } catch (error: unknown) {
      const err = error as {
        response?: { data?: { error?: { message?: string } } };
      };
      toast.error(
        err.response?.data?.error?.message || "Failed to save course"
      );
    } finally {
      setIsLoading(false);
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
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold">Course Content</h3>
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
                        updateModuleTitle(module.id, e.target.value)
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
                      onClick={() => removeModule(module.id)}
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
                            className="flex items-center gap-3 p-2 bg-muted/50 rounded-lg"
                          >
                            <GripVertical className="h-4 w-4 text-muted-foreground cursor-move" />
                            <span className="text-sm text-muted-foreground w-8">
                              {lessonIndex + 1}.
                            </span>
                            <Icon className="h-4 w-4 text-muted-foreground" />
                            <Input
                              value={lesson.title}
                              onChange={(e) =>
                                updateLessonTitle(
                                  module.id,
                                  lesson.id,
                                  e.target.value
                                )
                              }
                              className="flex-1 h-8"
                            />
                            <span className="text-xs text-muted-foreground capitalize">
                              {lesson.type}
                            </span>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8 text-destructive"
                              onClick={() => removeLesson(module.id, lesson.id)}
                            >
                              <Trash2 className="h-3 w-3" />
                            </Button>
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
      </div>

      {/* Submit */}
      <div className="flex justify-end gap-4">
        <Button type="button" variant="outline" asChild>
          <a href="/tutor/courses">Cancel</a>
        </Button>
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
