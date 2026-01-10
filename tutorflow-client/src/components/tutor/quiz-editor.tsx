"use client";

import { useState, useEffect } from "react";
import { toast } from "sonner";
import {
  Plus,
  Trash2,
  Edit2,
  Check,
  X,
  GripVertical,
  AlertCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  getQuizByLessonAction,
  createQuizAction,
  updateQuizAction,
  addQuestionAction,
  updateQuestionAction,
  deleteQuestionAction,
} from "@/app/tutor/courses/actions";

interface QuizEditorProps {
  lessonId: string;
  lessonTitle: string;
}

interface Quiz {
  id: string;
  title: string;
  description?: string;
  questions: Question[];
  passing_score: number;
  time_limit?: number;
}

interface Question {
  id: string;
  question_text: string;
  question_type:
    | "single_choice"
    | "multiple_choice"
    | "true_false"
    | "short_answer"
    | "essay";
  points: number;
  options: Option[];
}

interface Option {
  id: string;
  option_text: string;
  is_correct: boolean;
}

export function QuizEditor({ lessonId, lessonTitle }: QuizEditorProps) {
  const [loading, setLoading] = useState(true);
  const [quiz, setQuiz] = useState<Quiz | null>(null);
  const [isCreating, setIsCreating] = useState(false);

  // Question Form State
  const [isQuestionModalOpen, setIsQuestionModalOpen] = useState(false);
  const [editingQuestion, setEditingQuestion] = useState<Question | null>(null);
  const [questionForm, setQuestionForm] = useState<{
    text: string;
    type: string;
    points: number;
    options: { text: string; isCorrect: boolean }[];
  }>({
    text: "",
    type: "single_choice",
    points: 1,
    options: [
      { text: "", isCorrect: false },
      { text: "", isCorrect: false },
    ],
  });

  useEffect(() => {
    loadQuiz();
  }, [lessonId]);

  const loadQuiz = async () => {
    setLoading(true);
    try {
      const res = await getQuizByLessonAction(lessonId);
      if (res.success && res.data) {
        setQuiz(res.data);
      } else {
        setQuiz(null);
      }
    } catch (error) {
      console.error(error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateQuiz = async () => {
    setIsCreating(true);
    try {
      const res = await createQuizAction({
        lesson_id: lessonId,
        title: `${lessonTitle} Quiz`,
        description: "Test your knowledge",
        passing_score: 60,
        max_attempts: 1,
      });
      if (res.success) {
        setQuiz({ ...res.data, questions: [] });
        toast.success("Quiz initialized");
      } else {
        toast.error("Failed to create quiz");
      }
    } catch (error) {
      toast.error("Error creating quiz");
    } finally {
      setIsCreating(false);
    }
  };

  const openAddQuestion = () => {
    setEditingQuestion(null);
    setQuestionForm({
      text: "",
      type: "single_choice",
      points: 1,
      options: [
        { text: "", isCorrect: false },
        { text: "", isCorrect: false },
      ],
    });
    setIsQuestionModalOpen(true);
  };

  const openEditQuestion = (q: Question) => {
    setEditingQuestion(q);
    setQuestionForm({
      text: q.question_text,
      type: q.question_type,
      points: q.points,
      options: q.options.map((o) => ({
        text: o.option_text,
        isCorrect: o.is_correct,
      })),
    });
    setIsQuestionModalOpen(true);
  };

  const handleSaveQuestion = async () => {
    if (!quiz) return;
    if (!questionForm.text) {
      toast.error("Question text is required");
      return;
    }

    const payload = {
      question_text: questionForm.text,
      question_type: questionForm.type,
      points: questionForm.points,
      options: questionForm.options.map((o) => ({
        option_text: o.text,
        is_correct: o.isCorrect,
      })),
    };

    try {
      let res;
      if (editingQuestion) {
        res = await updateQuestionAction(editingQuestion.id, payload);
      } else {
        res = await addQuestionAction(quiz.id, payload);
      }

      if (res.success) {
        toast.success(editingQuestion ? "Question updated" : "Question added");
        setIsQuestionModalOpen(false);
        loadQuiz(); // Reload to get fresh state
      } else {
        toast.error(res.error || "Failed to save question");
      }
    } catch (error) {
      toast.error("Error saving question");
    }
  };

  const handleDeleteQuestion = async (id: string) => {
    if (!confirm("Are you sure?")) return;
    try {
      const res = await deleteQuestionAction(id);
      if (res.success) {
        toast.success("Question deleted");
        loadQuiz();
      } else {
        toast.error("Failed to delete");
      }
    } catch (error) {
      toast.error("Error deleting question");
    }
  };

  // Option helpers
  const updateOption = (
    idx: number,
    field: "text" | "isCorrect",
    value: string | boolean
  ) => {
    const newOptions = [...questionForm.options];
    newOptions[idx] = { ...newOptions[idx], [field]: value } as {
      text: string;
      isCorrect: boolean;
    };
    setQuestionForm({ ...questionForm, options: newOptions });
  };

  const addOption = () => {
    setQuestionForm({
      ...questionForm,
      options: [...questionForm.options, { text: "", isCorrect: false }],
    });
  };

  const removeOption = (idx: number) => {
    const newOptions = questionForm.options.filter((_, i) => i !== idx);
    setQuestionForm({ ...questionForm, options: newOptions });
  };

  if (loading) return <div>Loading quiz...</div>;

  if (!quiz) {
    return (
      <div className="flex flex-col items-center justify-center p-8 border-2 border-dashed rounded-lg bg-muted/30">
        <AlertCircle className="h-10 w-10 text-muted-foreground mb-4" />
        <h3 className="text-lg font-medium mb-2">No Quiz Yet</h3>
        <p className="text-sm text-muted-foreground mb-6 text-center max-w-sm">
          Create a quiz to test your students' knowledge of this lesson.
        </p>
        <Button type="button" onClick={handleCreateQuiz} disabled={isCreating}>
          {isCreating ? "Creating..." : "Create Quiz"}
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h3 className="text-lg font-medium">{quiz.title}</h3>
          <p className="text-sm text-muted-foreground">
            {quiz.questions?.length || 0} Questions • {quiz.passing_score}%
            Passing Score
          </p>
        </div>
        <Button type="button" onClick={openAddQuestion} size="sm">
          <Plus className="h-4 w-4 mr-2" />
          Add Question
        </Button>
      </div>

      <div className="space-y-4">
        {quiz.questions?.map((q, idx) => (
          <Card key={q.id}>
            <CardHeader className="py-4">
              <div className="flex justify-between items-start">
                <div className="flex gap-3">
                  <span className="font-medium text-muted-foreground">
                    Q{idx + 1}.
                  </span>
                  <div>
                    <p className="font-medium">{q.question_text}</p>
                    <p className="text-xs text-muted-foreground mt-1 capitalize">
                      {q.question_type.replace("_", " ")} • {q.points} pts
                    </p>
                  </div>
                </div>
                <div className="flex gap-1">
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    onClick={() => openEditQuestion(q)}
                  >
                    <Edit2 className="h-4 w-4" />
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="text-destructive"
                    onClick={() => handleDeleteQuestion(q.id)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              </div>
            </CardHeader>
          </Card>
        ))}
      </div>

      <Dialog open={isQuestionModalOpen} onOpenChange={setIsQuestionModalOpen}>
        <DialogContent className="max-w-2xl">
          <DialogHeader>
            <DialogTitle>
              {editingQuestion ? "Edit Question" : "Add Question"}
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label>Question Text</Label>
              <Input
                value={questionForm.text}
                onChange={(e) =>
                  setQuestionForm({ ...questionForm, text: e.target.value })
                }
                placeholder="e.g., What is the capital of France?"
              />
            </div>

            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Type</Label>
                <Select
                  value={questionForm.type}
                  onValueChange={(v) =>
                    setQuestionForm({ ...questionForm, type: v })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="single_choice">Single Choice</SelectItem>
                    <SelectItem value="multiple_choice">
                      Multiple Choice
                    </SelectItem>
                    <SelectItem value="true_false">True/False</SelectItem>
                    <SelectItem value="short_answer">Short Answer</SelectItem>
                    <SelectItem value="essay">Essay</SelectItem>
                  </SelectContent>
                </Select>
              </div>
              <div className="space-y-2">
                <Label>Points</Label>
                <Input
                  type="number"
                  min="1"
                  value={questionForm.points}
                  onChange={(e) =>
                    setQuestionForm({
                      ...questionForm,
                      points: parseInt(e.target.value) || 0,
                    })
                  }
                />
              </div>
            </div>

            <div className="space-y-3">
              <Label>Answer Options</Label>
              {questionForm.options.map((opt, idx) => (
                <div key={idx} className="flex gap-2 items-center">
                  <Button
                    type="button"
                    variant={opt.isCorrect ? "default" : "outline"}
                    size="icon"
                    className={
                      opt.isCorrect ? "bg-green-600 hover:bg-green-700" : ""
                    }
                    onClick={() =>
                      updateOption(idx, "isCorrect", !opt.isCorrect)
                    }
                  >
                    <Check className="h-4 w-4" />
                  </Button>
                  <Input
                    value={opt.text}
                    onChange={(e) => updateOption(idx, "text", e.target.value)}
                    placeholder={`Option ${idx + 1}`}
                    className="flex-1"
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon"
                    className="text-destructive"
                    onClick={() => removeOption(idx)}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </div>
              ))}
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={addOption}
                className="mt-2"
              >
                <Plus className="h-3 w-3 mr-2" /> Add Option
              </Button>
            </div>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => setIsQuestionModalOpen(false)}
            >
              Cancel
            </Button>
            <Button type="button" onClick={handleSaveQuestion}>
              Save Question
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
