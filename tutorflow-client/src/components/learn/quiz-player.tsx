"use client";

import { useState, useEffect, useCallback } from "react";
import {
  ChevronLeft,
  ChevronRight,
  Send,
  Loader2,
  CheckCircle2,
  XCircle,
  Timer,
  AlertCircle,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import api from "@/lib/api";
import { toast } from "sonner";

interface QuizOption {
  id: string;
  option_text: string;
}

interface QuizQuestion {
  id: string;
  question_type: "single_choice" | "multiple_choice" | "true_false";
  question_text: string;
  points: number;
  options: QuizOption[];
}

interface Quiz {
  id: string;
  title: string;
  description?: string;
  time_limit?: number;
  passing_score: number;
  is_published: boolean;
  questions: QuizQuestion[];
}

interface QuizAttempt {
  id: string;
  score?: number;
  max_score?: number;
  percentage?: number;
  passed?: boolean;
  completed_at?: string;
}

interface QuizPlayerProps {
  lessonId: string;
  onComplete?: () => void;
}

export function QuizPlayer({ lessonId, onComplete }: QuizPlayerProps) {
  const [loading, setLoading] = useState(true);
  const [quiz, setQuiz] = useState<Quiz | null>(null);
  const [currentQuestionIndex, setCurrentQuestionIndex] = useState(0);
  const [answers, setAnswers] = useState<Record<string, string[]>>({});
  const [attempt, setAttempt] = useState<QuizAttempt | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchQuiz = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const response = await api.get<Quiz>(`/quizzes/lesson/${lessonId}`);
      if (response.data) {
        setQuiz(response.data);
      }
    } catch (err: any) {
      console.error("Failed to fetch quiz:", err);
      setError("Failed to load quiz data. Please try again.");
    } finally {
      setLoading(false);
    }
  }, [lessonId]);

  useEffect(() => {
    fetchQuiz();
    // Reset state when lesson changes
    setCurrentQuestionIndex(0);
    setAnswers({});
    setAttempt(null);
  }, [fetchQuiz]);

  const handleOptionSelect = (
    questionId: string,
    optionId: string,
    type: string,
  ) => {
    if (attempt) return; // Can't change after submission

    setAnswers((prev) => {
      const current = prev[questionId] || [];
      if (type === "single_choice" || type === "true_false") {
        return { ...prev, [questionId]: [optionId] };
      } else {
        const exists = current.includes(optionId);
        if (exists) {
          const newAnswers = { ...prev };
          newAnswers[questionId] = current.filter((id) => id !== optionId);
          return newAnswers;
        } else {
          const newAnswers = { ...prev };
          newAnswers[questionId] = [...current, optionId];
          return newAnswers;
        }
      }
    });
  };

  const handleSubmit = async () => {
    if (!quiz) return;

    try {
      setSubmitting(true);
      // Create attempt first
      const startRes = await api.post<{ id: string }>(
        `/quizzes/${quiz.id}/attempts`,
        {},
      );
      const attemptId = startRes.data.id;

      // Map answers to the format expected by backend (questions-options mapping)
      const submissionData = {
        answers: JSON.stringify(answers),
      };

      const submitRes = await api.post<QuizAttempt>(
        `/quizzes/attempts/${attemptId}/submit`,
        submissionData,
      );
      setAttempt(submitRes.data);

      if (submitRes.data.passed) {
        toast.success("Congratulations! You passed the quiz.");
        onComplete?.();
      } else {
        toast.error("You did not pass the quiz. Try again.");
      }
    } catch (err: any) {
      console.error("Failed to submit quiz:", err);
      const errorMessage =
        err.response?.data?.message ||
        err.message ||
        "Failed to submit your answers. Please try again.";
      toast.error(errorMessage);
    } finally {
      setSubmitting(false);
    }
  };

  if (loading) {
    return (
      <div className="aspect-video bg-muted flex items-center justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (error || !quiz) {
    return (
      <div className="aspect-video bg-muted flex items-center justify-center p-8">
        <div className="text-center max-w-md">
          <AlertCircle className="h-12 w-12 text-destructive mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">Error Loading Quiz</h3>
          <p className="text-muted-foreground mb-6">
            {error || "This quiz is not available."}
          </p>
          <Button onClick={fetchQuiz}>Try Again</Button>
        </div>
      </div>
    );
  }

  if (!quiz.is_published) {
    return (
      <div className="aspect-video bg-muted flex items-center justify-center p-8">
        <div className="text-center max-w-md">
          <AlertCircle className="h-12 w-12 text-yellow-500 mx-auto mb-4" />
          <h3 className="text-lg font-semibold mb-2">Quiz Not Available</h3>
          <p className="text-muted-foreground">
            This quiz is not yet published. Please check back later.
          </p>
        </div>
      </div>
    );
  }

  if (attempt) {
    return (
      <div className="aspect-video bg-muted overflow-y-auto p-8 flex items-center justify-center">
        <div className="max-w-md w-full text-center space-y-6">
          <div className="flex justify-center">
            {attempt.passed ? (
              <CheckCircle2 className="h-20 w-20 text-green-500" />
            ) : (
              <XCircle className="h-20 w-20 text-destructive" />
            )}
          </div>
          <div className="space-y-2">
            <h2 className="text-3xl font-bold">
              {attempt.passed ? "Quiz Passed!" : "Quiz Failed"}
            </h2>
            <p className="text-muted-foreground">
              You scored {Math.round(attempt.percentage || 0)}% in this attempt.
            </p>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="p-4 bg-background rounded-lg border">
              <p className="text-sm text-muted-foreground">Score</p>
              <p className="text-2xl font-bold">
                {attempt.score} / {attempt.max_score}
              </p>
            </div>
            <div className="p-4 bg-background rounded-lg border">
              <p className="text-sm text-muted-foreground">Passing Score</p>
              <p className="text-2xl font-bold">{quiz.passing_score}%</p>
            </div>
          </div>
          <div className="pt-4 flex gap-4 justify-center">
            <Button variant="outline" onClick={() => setAttempt(null)}>
              Try Again
            </Button>
            {attempt.passed && <Button onClick={onComplete}>Continue</Button>}
          </div>
        </div>
      </div>
    );
  }

  const currentQuestion = quiz.questions[currentQuestionIndex];
  const isFirstQuestion = currentQuestionIndex === 0;
  const isLastQuestion = currentQuestionIndex === quiz.questions.length - 1;

  if (!currentQuestion) {
    return (
      <div className="aspect-video bg-muted flex items-center justify-center p-8">
        <p>No questions found for this quiz.</p>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col bg-background border rounded-xl overflow-hidden shadow-sm">
      {/* Quiz Header */}
      <div className="p-4 border-b bg-muted/30 flex items-center justify-between">
        <div className="space-y-1">
          <h3 className="font-semibold">{quiz.title}</h3>
          <div className="flex items-center gap-3 text-xs text-muted-foreground">
            <span className="flex items-center gap-1">
              Question {currentQuestionIndex + 1} of {quiz.questions.length}
            </span>
            {quiz.time_limit && (
              <span className="flex items-center gap-1">
                <Timer className="h-3 w-3" />
                {quiz.time_limit}m limit
              </span>
            )}
          </div>
        </div>
        <div className="w-48 h-2 bg-muted rounded-full overflow-hidden">
          <div
            className="h-full bg-primary transition-all duration-300"
            style={{
              width: `${((currentQuestionIndex + 1) / quiz.questions.length) * 100}%`,
            }}
          />
        </div>
      </div>

      {/* Question Content */}
      <div className="flex-1 overflow-y-auto p-8 lg:p-12">
        <div className="max-w-2xl mx-auto space-y-8">
          <div className="space-y-4">
            <span className="text-primary text-sm font-medium tracking-wider uppercase">
              {currentQuestion.question_type.replace("_", " ")}
            </span>
            <h2 className="text-2xl font-bold leading-tight">
              {currentQuestion.question_text}
            </h2>
          </div>

          <div className="space-y-3">
            {currentQuestion.options.map((option) => {
              const isSelected = (answers[currentQuestion.id] || []).includes(
                option.id,
              );
              return (
                <button
                  key={option.id}
                  onClick={() =>
                    handleOptionSelect(
                      currentQuestion.id,
                      option.id,
                      currentQuestion.question_type,
                    )
                  }
                  className={cn(
                    "w-full flex items-center gap-4 p-4 rounded-xl border text-left transition-all",
                    isSelected
                      ? "border-primary bg-primary/5 ring-1 ring-primary"
                      : "hover:border-primary/50 hover:bg-muted/50",
                  )}
                >
                  <div
                    className={cn(
                      "h-5 w-5 rounded-full border flex items-center justify-center shrink-0",
                      isSelected
                        ? "border-primary bg-primary"
                        : "border-muted-foreground",
                    )}
                  >
                    {isSelected && (
                      <div className="h-2 w-2 rounded-full bg-white" />
                    )}
                  </div>
                  <span className="font-medium">{option.option_text}</span>
                </button>
              );
            })}
          </div>
        </div>
      </div>

      {/* Footer Controls */}
      <div className="p-4 border-t bg-muted/30 flex items-center justify-between">
        <Button
          variant="ghost"
          onClick={() => setCurrentQuestionIndex((v) => v - 1)}
          disabled={isFirstQuestion}
        >
          <ChevronLeft className="mr-2 h-4 w-4" />
          Previous
        </Button>

        {isLastQuestion ? (
          <Button
            onClick={handleSubmit}
            disabled={submitting || Object.keys(answers).length === 0}
            className="px-8 shadow-lg shadow-primary/20"
          >
            {submitting ? (
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            ) : (
              <Send className="mr-2 h-4 w-4" />
            )}
            Submit Quiz
          </Button>
        ) : (
          <Button
            onClick={() => setCurrentQuestionIndex((v) => v + 1)}
            className="px-8"
          >
            Next
            <ChevronRight className="ml-2 h-4 w-4" />
          </Button>
        )}
      </div>
    </div>
  );
}
