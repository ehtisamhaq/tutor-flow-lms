"use client";

import { useState, useEffect, useRef } from "react";
import { useRouter } from "next/navigation";
import { useFormContext } from "react-hook-form";
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
  X,
  BookOpen,
  Loader2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  FormWrapper,
  InputField,
  TextareaField,
  SelectField,
} from "@/components/forms/rhf";

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
} from "@/app/(dashboard)/tutor/courses/lib/actions";
import { QuizEditor } from "./quiz-editor";
import { VideoUploader } from "./video-uploader";
import React from "react";

// ---------------------------------------------------------------------------
// Zod Schema
// ---------------------------------------------------------------------------
const courseSchema = z.object({
  title: z.string().min(5, "Title must be at least 5 characters"),
  description: z.string().min(50, "Description must be at least 50 characters"),
  price: z.coerce
    .number({ message: "Price is required" })
    .min(0, "Price must be 0 or greater"),
  discount_price: z.coerce
    .number()
    .min(0, "Discount price must be 0 or greater")
    .optional()
    .nullable(),
  level: z.enum(["beginner", "intermediate", "advanced"]),
  category_id: z.string().optional(),
});

type CourseFormData = z.infer<typeof courseSchema>;

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------
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
  basePath?: string;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------
const LEVEL_OPTIONS = [
  { value: "beginner", label: "Beginner" },
  { value: "intermediate", label: "Intermediate" },
  { value: "advanced", label: "Advanced" },
];

// ---------------------------------------------------------------------------
// Status badge
// ---------------------------------------------------------------------------
function StatusBadge({ status }: { status?: string }) {
  if (!status) return null;
  const styles: Record<string, string> = {
    published:
      "bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400",
    archived: "bg-gray-100 text-gray-600 dark:bg-gray-800 dark:text-gray-400",
    draft:
      "bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400",
  };
  return (
    <span
      className={`px-2 py-0.5 rounded text-xs font-semibold uppercase tracking-wide ${
        styles[status] ?? styles.draft
      }`}
    >
      {status}
    </span>
  );
}

function getLessonIcon(type: string) {
  switch (type) {
    case "video":
      return Video;
    case "quiz":
      return HelpCircle;
    default:
      return FileText;
  }
}

// ---------------------------------------------------------------------------
// Inner form — must be inside FormProvider (provided by FormWrapper)
// ---------------------------------------------------------------------------
function CourseFormFields({
  initialData,
  thumbnailPreview,
  thumbnailFile,
  onThumbnailChange,
  onThumbnailRemove,
}: {
  initialData: CourseFormProps["initialData"];
  thumbnailPreview: string | null;
  thumbnailFile: File | null;
  onThumbnailChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onThumbnailRemove: () => void;
}) {
  const { watch } = useFormContext<CourseFormData>();
  const descriptionLength = watch("description")?.length ?? 0;

  return (
    <div className="space-y-6">
      {/* Grid: left fields + thumbnail */}
      <div className="grid md:grid-cols-2 gap-6 items-start">
        {/* Left */}
        <div className="space-y-5">
          <InputField<CourseFormData>
            name="title"
            label="Course Title *"
            placeholder="e.g., Complete Web Development Bootcamp"
          />

          <SelectField<CourseFormData>
            name="level"
            label="Level *"
            options={LEVEL_OPTIONS}
            placeholder="Select level"
          />

          <div className="grid grid-cols-2 gap-4">
            <InputField<CourseFormData>
              name="price"
              label="Price ($) *"
              type="number"
              placeholder="0.00"
            />
            <InputField<CourseFormData>
              name="discount_price"
              label="Discount Price ($)"
              type="number"
              placeholder="0.00"
              description="Must be less than price"
            />
          </div>
        </div>

        {/* Thumbnail */}
        <div className="space-y-2">
          <label className="text-sm font-medium leading-none">
            Course Thumbnail
          </label>
          <div className="relative aspect-video border-2 border-dashed rounded-lg overflow-hidden hover:border-primary transition-colors group">
            {thumbnailPreview ? (
              <>
                <img
                  src={thumbnailPreview}
                  alt="Thumbnail preview"
                  className="w-full h-full object-cover"
                />
                <button
                  type="button"
                  onClick={onThumbnailRemove}
                  className="absolute top-2 right-2 bg-black/60 hover:bg-black/80 text-white rounded-full p-1 transition-colors"
                  aria-label="Remove thumbnail"
                >
                  <X className="h-4 w-4" />
                </button>
              </>
            ) : (
              <div className="absolute inset-0 flex flex-col items-center justify-center text-muted-foreground group-hover:text-primary transition-colors">
                <Upload className="h-8 w-8 mb-2" />
                <p className="text-sm font-medium">Click to upload thumbnail</p>
                <p className="text-xs mt-1 opacity-70">
                  PNG, JPG, WebP · Max 5 MB · 1280×720 recommended
                </p>
              </div>
            )}
            <input
              type="file"
              accept="image/*"
              className="absolute inset-0 opacity-0 cursor-pointer"
              onChange={onThumbnailChange}
            />
          </div>
          {thumbnailFile && (
            <p className="text-xs text-muted-foreground truncate">
              {thumbnailFile.name} ({(thumbnailFile.size / 1024).toFixed(0)} KB)
            </p>
          )}
        </div>
      </div>

      {/* Description */}
      <TextareaField<CourseFormData>
        name="description"
        label="Description *"
        placeholder="Describe what students will learn in this course… (minimum 50 characters)"
        rows={5}
        description={`${descriptionLength} characters (minimum 50)`}
      />
    </div>
  );
}

// ---------------------------------------------------------------------------
// Main Component
// ---------------------------------------------------------------------------
export function CourseForm({
  initialData,
  initialModules = [],
  basePath = "/tutor",
}: CourseFormProps) {
  const router = useRouter();
  const [modules, setModules] = useState<Module[]>(initialModules);
  const [thumbnailPreview, setThumbnailPreview] = useState<string | null>(null);
  const [thumbnailFile, setThumbnailFile] = useState<File | null>(null);

  // Module / lesson loading states
  const [isAddingModule, setIsAddingModule] = useState(false);
  const [deletingLessonId, setDeletingLessonId] = useState<string | null>(null);
  const [addingLessonModuleId, setAddingLessonModuleId] = useState<
    string | null
  >(null);

  // Confirm dialogs
  const [deleteModuleId, setDeleteModuleId] = useState<string | null>(null);
  const [isDeletingModule, setIsDeletingModule] = useState(false);
  const [showPublishConfirm, setShowPublishConfirm] = useState(false);
  const [isPublishing, setIsPublishing] = useState(false);

  // Stable module sync ref
  const prevModuleIdsRef = useRef<string>("");

  useEffect(() => {
    const currentIds = initialModules.map((m) => m.id).join(",");
    if (currentIds !== prevModuleIdsRef.current) {
      prevModuleIdsRef.current = currentIds;
      setModules((prevModules) => {
        const expandedState = new Map(
          prevModules.map((m) => [m.id, m.isExpanded]),
        );
        return initialModules.map((m) => ({
          ...m,
          isExpanded: expandedState.get(m.id) ?? false,
        }));
      });
    }
  }, [initialModules]);

  // ---------------------------------------------------------------------------
  // Thumbnail helpers
  // ---------------------------------------------------------------------------
  const handleThumbnailChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    if (!file.type.startsWith("image/")) {
      toast.error("Please select an image file");
      return;
    }
    if (file.size > 5 * 1024 * 1024) {
      toast.error("Thumbnail must be smaller than 5 MB");
      return;
    }
    setThumbnailFile(file);
    const reader = new FileReader();
    reader.onload = () => setThumbnailPreview(reader.result as string);
    reader.readAsDataURL(file);
  };

  const removeThumbnail = () => {
    setThumbnailFile(null);
    setThumbnailPreview(null);
  };

  // ---------------------------------------------------------------------------
  // Form submit — receives (data, form) via FormWrapper render prop
  // ---------------------------------------------------------------------------
  const handleSubmit = async (
    data: CourseFormData,
    form: { setError: Function },
  ) => {
    const setError = form.setError;
    const formData = new FormData();
    formData.append("title", data.title);
    formData.append("description", data.description);
    formData.append("price", String(data.price));
    if (data.discount_price != null) {
      formData.append("discount_price", String(data.discount_price));
    }
    formData.append("level", data.level);
    if (data.category_id) formData.append("category_id", data.category_id);
    if (thumbnailFile) formData.append("thumbnail", thumbnailFile);

    console.log(formData, initialData);

    const result = initialData?.id
      ? await updateCourseAction(initialData.id, formData)
      : await createCourseAction(formData);

    if (result.success) {
      toast.success(
        initialData?.id
          ? "Course updated successfully"
          : "Course created! Now add your modules and lessons.",
      );
      router.push(`${basePath}/courses`);
    } else if (result.error?.code === "VALIDATION_ERROR") {
      const details = result.error.details as {
        field: string;
        message: string;
      }[];
      details.forEach((err) =>
        setError(err.field, { type: "server", message: err.message }),
      );
      toast.error("Please fix the validation errors and try again");
    } else {
      toast.error(result.error?.message ?? "Failed to save course");
    }
  };

  // ---------------------------------------------------------------------------
  // Publish
  // ---------------------------------------------------------------------------
  const handlePublish = async () => {
    if (!initialData?.id) return;
    setIsPublishing(true);
    try {
      const result = await publishCourseAction(initialData.id);
      if (result.success) {
        toast.success("Course published successfully!");
        setShowPublishConfirm(false);
        router.refresh();
      } else {
        toast.error(result.error?.message ?? "Failed to publish course");
      }
    } catch {
      toast.error("An error occurred while publishing");
    } finally {
      setIsPublishing(false);
    }
  };

  // ---------------------------------------------------------------------------
  // Module management
  // ---------------------------------------------------------------------------
  const addModule = async () => {
    if (!initialData?.id) {
      toast.error("Please save the course details first.");
      return;
    }
    setIsAddingModule(true);
    try {
      const result = await createModuleAction(initialData.id, {
        title: `Module ${modules.length + 1}`,
        order: modules.length + 1,
      });
      if (!result.success) {
        toast.error(
          typeof result.error === "string"
            ? result.error
            : "Failed to create module",
        );
      } else {
        toast.success("Module created");
      }
    } catch {
      toast.error("Failed to create module");
    } finally {
      setIsAddingModule(false);
    }
  };

  const handleDeleteModule = async () => {
    if (!initialData?.id || !deleteModuleId) return;
    setIsDeletingModule(true);
    try {
      const result = await deleteModuleAction(initialData.id, deleteModuleId);
      if (result.success) {
        toast.success("Module deleted");
        setDeleteModuleId(null);
      } else {
        toast.error(
          typeof result.error === "string"
            ? result.error
            : "Failed to delete module",
        );
      }
    } catch {
      toast.error("Failed to delete module");
    } finally {
      setIsDeletingModule(false);
    }
  };

  const updateModuleTitleLocal = (moduleId: string, title: string) =>
    setModules((prev) =>
      prev.map((m) => (m.id === moduleId ? { ...m, title } : m)),
    );

  const saveModuleTitle = async (moduleId: string, title: string) => {
    if (!initialData?.id) return;
    const result = await updateModuleAction(initialData.id, moduleId, {
      title,
    }).catch(() => null);
    if (result && !result.success) {
      toast.error(
        typeof result.error === "string"
          ? result.error
          : "Failed to update module",
      );
    }
  };

  const toggleModule = (moduleId: string) =>
    setModules((prev) =>
      prev.map((m) =>
        m.id === moduleId ? { ...m, isExpanded: !m.isExpanded } : m,
      ),
    );

  // ---------------------------------------------------------------------------
  // Lesson management
  // ---------------------------------------------------------------------------
  const addLesson = async (
    moduleId: string,
    type: "video" | "text" | "quiz",
  ) => {
    if (!initialData?.id) return;
    setAddingLessonModuleId(moduleId);
    try {
      const module = modules.find((m) => m.id === moduleId);
      const result = await createLessonAction(moduleId, {
        title: `New ${type.charAt(0).toUpperCase() + type.slice(1)}`,
        type,
        order: (module?.lessons.length ?? 0) + 1,
      });
      if (result.success) {
        toast.success("Lesson created");
      } else {
        toast.error(
          typeof result.error === "string"
            ? result.error
            : "Failed to create lesson",
        );
      }
    } catch {
      toast.error("Failed to create lesson");
    } finally {
      setAddingLessonModuleId(null);
    }
  };

  const removeLesson = async (moduleId: string, lessonId: string) => {
    if (!initialData?.id) return;
    setDeletingLessonId(lessonId);
    try {
      const result = await deleteLessonAction(moduleId, lessonId);
      if (result.success) {
        toast.success("Lesson deleted");
      } else {
        toast.error(
          typeof result.error === "string"
            ? result.error
            : "Failed to delete lesson",
        );
      }
    } catch {
      toast.error("Failed to delete lesson");
    } finally {
      setDeletingLessonId(null);
    }
  };

  const saveLesson = async (
    moduleId: string,
    lessonId: string,
    data: Partial<Lesson>,
  ) => {
    const result = await updateLessonAction(moduleId, lessonId, data).catch(
      () => null,
    );
    if (result && !result.success) {
      toast.error(
        typeof result.error === "string"
          ? result.error
          : "Failed to update lesson",
      );
    }
  };

  const updateLessonValue = (
    moduleId: string,
    lessonId: string,
    field: keyof Lesson,
    value: any,
  ) =>
    setModules((prev) =>
      prev.map((m) =>
        m.id !== moduleId
          ? m
          : {
              ...m,
              lessons: m.lessons.map((l) =>
                l.id === lessonId ? { ...l, [field]: value } : l,
              ),
            },
      ),
    );

  // ---------------------------------------------------------------------------
  // Render
  // ---------------------------------------------------------------------------
  return (
    <FormWrapper<CourseFormData>
      schema={courseSchema}
      mode="onBlur"
      isEditMode={!!initialData?.id}
      submitLabel={initialData?.id ? "Update Course" : "Create Course"}
      defaultValues={{
        title: initialData?.title ?? "",
        description: initialData?.description ?? "",
        price: initialData?.price ?? 0,
        discount_price: initialData?.discount_price ?? null,
        level: (initialData?.level as CourseFormData["level"]) ?? "beginner",
        category_id: initialData?.category_id ?? "",
      }}
      onSubmit={(data, form) => handleSubmit(data, form)}
      onCancel={() => router.push(`${basePath}/courses`)}
      showActions={false}
      className="space-y-8"
    >
      {(form) => (
        <>
          {/* Course info fields */}
          <CourseFormFields
            initialData={initialData}
            thumbnailPreview={thumbnailPreview}
            thumbnailFile={thumbnailFile}
            onThumbnailChange={handleThumbnailChange}
            onThumbnailRemove={removeThumbnail}
          />

          {/* ── Curriculum ── */}
          <div className="space-y-4">
            {!initialData?.id ? (
              <Card className="border-dashed bg-muted/30">
                <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground space-y-3">
                  <BookOpen className="h-10 w-10 opacity-40" />
                  <h3 className="font-semibold text-foreground">
                    Curriculum Locked
                  </h3>
                  <p className="text-sm text-center max-w-xs">
                    Save the basic course details above to unlock the curriculum
                    builder and start adding modules and lessons.
                  </p>
                </CardContent>
              </Card>
            ) : (
              <>
                {/* Header */}
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3 flex-wrap">
                    <h3 className="text-lg font-semibold">Course Curriculum</h3>
                    <StatusBadge status={initialData?.status} />
                    <span className="text-sm text-muted-foreground">
                      {modules.length} module{modules.length !== 1 ? "s" : ""}
                      {" · "}
                      {modules.reduce(
                        (acc, m) => acc + m.lessons.length,
                        0,
                      )}{" "}
                      lessons
                    </span>
                  </div>
                  <Button
                    type="button"
                    onClick={addModule}
                    disabled={isAddingModule}
                    size="sm"
                  >
                    {isAddingModule ? (
                      <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    ) : (
                      <Plus className="mr-2 h-4 w-4" />
                    )}
                    Add Module
                  </Button>
                </div>

                {/* Empty state */}
                {modules.length === 0 ? (
                  <Card className="border-dashed">
                    <CardContent className="flex flex-col items-center justify-center py-12 text-muted-foreground space-y-3">
                      <BookOpen className="h-8 w-8 opacity-40" />
                      <p className="text-sm font-medium text-foreground">
                        No modules yet
                      </p>
                      <p className="text-xs text-center max-w-xs">
                        Add your first module to start organising your course
                        content into sections.
                      </p>
                      <Button
                        type="button"
                        variant="outline"
                        size="sm"
                        onClick={addModule}
                        disabled={isAddingModule}
                      >
                        {isAddingModule ? (
                          <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                        ) : (
                          <Plus className="mr-2 h-4 w-4" />
                        )}
                        Add First Module
                      </Button>
                    </CardContent>
                  </Card>
                ) : (
                  <div className="space-y-3">
                    {modules.map((module, index) => (
                      <Card key={module.id} className="overflow-hidden">
                        {/* Module header row */}
                        <div className="flex items-center gap-2 px-4 py-3 border-b bg-muted/20">
                          <GripVertical className="h-5 w-5 text-muted-foreground/50 shrink-0 cursor-move" />
                          <span className="text-sm font-mono text-muted-foreground w-6 shrink-0">
                            {index + 1}.
                          </span>
                          <Input
                            value={module.title}
                            onChange={(e) =>
                              updateModuleTitleLocal(module.id, e.target.value)
                            }
                            onBlur={() =>
                              saveModuleTitle(module.id, module.title)
                            }
                            className="flex-1 h-8 text-sm"
                            placeholder="Module title"
                          />
                          <Button
                            type="button"
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8 shrink-0"
                            onClick={() => toggleModule(module.id)}
                            aria-label={
                              module.isExpanded
                                ? "Collapse module"
                                : "Expand module"
                            }
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
                            className="h-8 w-8 shrink-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                            onClick={() => setDeleteModuleId(module.id)}
                            aria-label="Delete module"
                          >
                            <Trash2 className="h-4 w-4" />
                          </Button>
                        </div>

                        {/* Module body */}
                        {module.isExpanded && (
                          <div className="p-4 space-y-2">
                            {module.lessons.length === 0 && (
                              <p className="text-xs text-muted-foreground text-center py-3">
                                No lessons yet. Add one below.
                              </p>
                            )}

                            {module.lessons.map((lesson, lessonIndex) => {
                              const Icon = getLessonIcon(lesson.type);
                              const isDeleting = deletingLessonId === lesson.id;
                              return (
                                <div
                                  key={lesson.id}
                                  className="p-2 bg-muted/40 rounded-lg space-y-2 border"
                                >
                                  <div className="flex items-center gap-2">
                                    <GripVertical className="h-4 w-4 text-muted-foreground/50 shrink-0 cursor-move" />
                                    <span className="text-xs text-muted-foreground w-5 shrink-0">
                                      {lessonIndex + 1}.
                                    </span>
                                    <Icon className="h-4 w-4 text-muted-foreground shrink-0" />
                                    <Input
                                      value={lesson.title}
                                      onChange={(e) =>
                                        updateLessonValue(
                                          module.id,
                                          lesson.id,
                                          "title",
                                          e.target.value,
                                        )
                                      }
                                      onBlur={() =>
                                        saveLesson(module.id, lesson.id, {
                                          title: lesson.title,
                                        })
                                      }
                                      className="flex-1 h-8 text-sm"
                                      placeholder="Lesson title"
                                    />
                                    <Input
                                      type="number"
                                      min={0}
                                      value={lesson.duration_minutes}
                                      onChange={(e) =>
                                        updateLessonValue(
                                          module.id,
                                          lesson.id,
                                          "duration_minutes",
                                          parseInt(e.target.value) || 0,
                                        )
                                      }
                                      onBlur={() =>
                                        saveLesson(module.id, lesson.id, {
                                          duration_minutes:
                                            lesson.duration_minutes,
                                        })
                                      }
                                      className="w-16 h-8 text-xs shrink-0"
                                      placeholder="min"
                                      aria-label="Duration in minutes"
                                    />
                                    <span className="text-xs text-muted-foreground capitalize shrink-0 hidden sm:inline">
                                      {lesson.type}
                                    </span>
                                    <Button
                                      type="button"
                                      variant="ghost"
                                      size="icon"
                                      className="h-8 w-8 shrink-0 text-destructive hover:text-destructive hover:bg-destructive/10"
                                      onClick={() =>
                                        removeLesson(module.id, lesson.id)
                                      }
                                      disabled={isDeleting}
                                      aria-label="Delete lesson"
                                    >
                                      {isDeleting ? (
                                        <Loader2 className="h-3 w-3 animate-spin" />
                                      ) : (
                                        <Trash2 className="h-3 w-3" />
                                      )}
                                    </Button>
                                  </div>

                                  {/* Content by type */}
                                  <div className="pl-9">
                                    {lesson.type === "video" && (
                                      <div className="space-y-2">
                                        <VideoUploader
                                          lessonId={lesson.id}
                                          initialVideoUrl={lesson.video_url}
                                          onUploadComplete={(url) =>
                                            updateLessonValue(
                                              module.id,
                                              lesson.id,
                                              "video_url",
                                              url,
                                            )
                                          }
                                          onDelete={() =>
                                            updateLessonValue(
                                              module.id,
                                              lesson.id,
                                              "video_url",
                                              "",
                                            )
                                          }
                                        />
                                        <Input
                                          placeholder="Or paste external URL (YouTube, Vimeo…)"
                                          value={lesson.video_url || ""}
                                          onChange={(e) =>
                                            updateLessonValue(
                                              module.id,
                                              lesson.id,
                                              "video_url",
                                              e.target.value,
                                            )
                                          }
                                          onBlur={() =>
                                            saveLesson(module.id, lesson.id, {
                                              video_url: lesson.video_url,
                                            })
                                          }
                                          className="h-8 text-xs bg-background"
                                        />
                                      </div>
                                    )}
                                    {lesson.type === "text" && (
                                      <textarea
                                        placeholder="Article content…"
                                        value={lesson.content || ""}
                                        onChange={(e) =>
                                          updateLessonValue(
                                            module.id,
                                            lesson.id,
                                            "content",
                                            e.target.value,
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

                            {/* Add lesson buttons */}
                            <div className="flex gap-2 pt-1">
                              {(["video", "text", "quiz"] as const).map(
                                (type) => {
                                  const icons = {
                                    video: Video,
                                    text: FileText,
                                    quiz: HelpCircle,
                                  };
                                  const labels = {
                                    video: "Video",
                                    text: "Article",
                                    quiz: "Quiz",
                                  };
                                  const Icon = icons[type];
                                  const isAdding =
                                    addingLessonModuleId === module.id;
                                  return (
                                    <Button
                                      key={type}
                                      type="button"
                                      variant="outline"
                                      size="sm"
                                      onClick={() => addLesson(module.id, type)}
                                      disabled={isAdding}
                                    >
                                      {isAdding ? (
                                        <Loader2 className="mr-1.5 h-3.5 w-3.5 animate-spin" />
                                      ) : (
                                        <Icon className="mr-1.5 h-3.5 w-3.5" />
                                      )}
                                      {labels[type]}
                                    </Button>
                                  );
                                },
                              )}
                            </div>
                          </div>
                        )}
                      </Card>
                    ))}
                  </div>
                )}
              </>
            )}
          </div>

          {/* ── Footer actions ── */}
          <div className="flex justify-end gap-3 pt-6 border-t">
            <Button
              type="button"
              variant="outline"
              onClick={() => router.push(`${basePath}/courses`)}
              disabled={form.formState.isSubmitting}
            >
              Cancel
            </Button>

            {initialData?.id && initialData.status === "draft" && (
              <Button
                type="button"
                variant="secondary"
                onClick={() => setShowPublishConfirm(true)}
                disabled={form.formState.isSubmitting || isPublishing}
              >
                Publish Course
              </Button>
            )}

            <Button type="submit" disabled={form.formState.isSubmitting}>
              {form.formState.isSubmitting ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Saving…
                </>
              ) : initialData?.id ? (
                "Update Course"
              ) : (
                "Create Course"
              )}
            </Button>
          </div>

          {/* ── Delete Module Dialog ── */}
          <Dialog
            open={!!deleteModuleId}
            onOpenChange={(open) => !open && setDeleteModuleId(null)}
          >
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Delete Module</DialogTitle>
                <DialogDescription>
                  Are you sure you want to delete this module? All lessons
                  within it will be permanently removed. This action cannot be
                  undone.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setDeleteModuleId(null)}
                  disabled={isDeletingModule}
                >
                  Cancel
                </Button>
                <Button
                  variant="destructive"
                  onClick={handleDeleteModule}
                  disabled={isDeletingModule}
                >
                  {isDeletingModule && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  Delete Module
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>

          {/* ── Publish Confirm Dialog ── */}
          <Dialog
            open={showPublishConfirm}
            onOpenChange={(open) => !open && setShowPublishConfirm(false)}
          >
            <DialogContent>
              <DialogHeader>
                <DialogTitle>Publish Course</DialogTitle>
                <DialogDescription>
                  Publishing makes this course visible to all students. Make
                  sure all modules and lessons are complete. You can archive it
                  later if needed.
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <Button
                  variant="outline"
                  onClick={() => setShowPublishConfirm(false)}
                  disabled={isPublishing}
                >
                  Cancel
                </Button>
                <Button onClick={handlePublish} disabled={isPublishing}>
                  {isPublishing && (
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  )}
                  Publish Course
                </Button>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        </>
      )}
    </FormWrapper>
  );
}
