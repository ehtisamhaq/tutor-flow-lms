"use client";

import { useState, useEffect, useRef } from "react";
import {
  Upload,
  X,
  CheckCircle,
  Loader2,
  AlertCircle,
  Play,
  Trash2,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";
import api from "@/lib/api";
import { Progress } from "../ui/progress";

interface VideoUploaderProps {
  lessonId: string;
  initialVideoUrl?: string;
  onUploadComplete: (videoUrl: string) => void;
  onDelete: () => void;
}

interface VideoStatus {
  id: string;
  status: "pending" | "processing" | "completed" | "failed";
  processing_error?: string;
}

export function VideoUploader({
  lessonId,
  initialVideoUrl,
  onUploadComplete,
  onDelete,
}: VideoUploaderProps) {
  const [file, setFile] = useState<File | null>(null);
  const [uploadProgress, setUploadProgress] = useState<number>(0);
  const [isUploading, setIsUploading] = useState(false);
  const [isProcessing, setIsProcessing] = useState(false);
  const [status, setStatus] = useState<VideoStatus | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);
  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Poll for status if processing
  useEffect(() => {
    if (status?.status === "processing" || status?.status === "pending") {
      setIsProcessing(true);
      startPolling();
    } else {
      setIsProcessing(false);
      stopPolling();
    }
    return () => stopPolling();
  }, [status]);

  // Check initial status if we have a lessonId
  useEffect(() => {
    if (lessonId) {
      checkStatus();
    }
  }, [lessonId]);

  const checkStatus = async () => {
    try {
      const response = await api.get<VideoStatus>(
        `/videos/lessons/${lessonId}/status`,
      );
      if (response.data) {
        setStatus(response.data);
      }
    } catch (error) {
      // It's fine if no video exists yet
    }
  };

  const startPolling = () => {
    if (pollIntervalRef.current) return;
    pollIntervalRef.current = setInterval(async () => {
      try {
        const response = await api.get<VideoStatus>(
          `/videos/lessons/${lessonId}/status`,
        );
        if (response.data) {
          const newStatus = response.data;
          setStatus(newStatus);
          if (newStatus.status === "completed") {
            toast.success("Video processing completed!");
            onUploadComplete(`/videos/stream/${newStatus.id}/playlist.m3u8`);
            stopPolling();
          } else if (newStatus.status === "failed") {
            toast.error("Video processing failed");
            stopPolling();
          }
        }
      } catch (error) {
        stopPolling();
      }
    }, 5000);
  };

  const stopPolling = () => {
    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current);
      pollIntervalRef.current = null;
    }
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const selectedFile = e.target.files?.[0];
    if (selectedFile) {
      if (selectedFile.size > 500 * 1024 * 1024) {
        toast.error("File is too large (max 500MB)");
        return;
      }
      setFile(selectedFile);
    }
  };

  const handleUpload = async () => {
    if (!file) return;

    setIsUploading(true);
    setUploadProgress(0);

    const formData = new FormData();
    formData.append("video", file);

    try {
      // We use XHR here instead of fetch/api to get upload progress
      const xhr = new XMLHttpRequest();
      const token = document.cookie
        .split("; ")
        .find((row) => row.startsWith("accessToken="))
        ?.split("=")[1];

      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) {
          const progress = Math.round((event.loaded * 100) / event.total);
          setUploadProgress(progress);
        }
      };

      const result = await new Promise<any>((resolve, reject) => {
        xhr.open("POST", `/api/videos/lessons/${lessonId}/upload`);
        if (token) {
          xhr.setRequestHeader("Authorization", `Bearer ${token}`);
        }

        xhr.onload = () => {
          if (xhr.status >= 200 && xhr.status < 300) {
            try {
              resolve(JSON.parse(xhr.response));
            } catch (e) {
              resolve(xhr.response);
            }
          } else {
            reject(new Error(xhr.statusText || "Upload failed"));
          }
        };
        xhr.onerror = () => reject(new Error("Network error"));
        xhr.send(formData);
      });

      console.log({ result });

      if (result && result.success) {
        toast.success("Upload successful! Processing started...");
        setStatus(result.data);
        setFile(null);
      } else {
        toast.error(result?.error?.message || "Upload failed");
      }
    } catch (error: any) {
      console.log({ error });
      toast.error(error.message || "Upload failed");
    } finally {
      setIsUploading(false);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Are you sure you want to delete this video?")) return;

    try {
      await api.delete(`/videos/lessons/${lessonId}`);
      toast.success("Video deleted");
      setStatus(null);
      onDelete();
    } catch (error) {
      toast.error("Failed to delete video");
    }
  };

  if (status?.status === "completed") {
    return (
      <div className="flex items-center justify-between p-3 bg-green-50 dark:bg-green-900/10 border border-green-200 dark:border-green-800 rounded-lg">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 bg-green-100 dark:bg-green-900/30 rounded flex items-center justify-center">
            <CheckCircle className="h-5 w-5 text-green-600" />
          </div>
          <div>
            <p className="text-sm font-medium text-green-800 dark:text-green-200">
              Video Ready
            </p>
            <p className="text-xs text-green-700 dark:text-green-300">
              Secure HLS streaming enabled
            </p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="icon"
            className="h-8 w-8 text-destructive"
            onClick={handleDelete}
          >
            <Trash2 className="h-4 w-4" />
          </Button>
        </div>
      </div>
    );
  }

  if (isProcessing || status?.status === "processing") {
    return (
      <div className="space-y-3 p-4 border rounded-lg bg-muted/30">
        <div className="flex items-center gap-3">
          <Loader2 className="h-5 w-5 animate-spin text-primary" />
          <div className="flex-1">
            <p className="text-sm font-medium">Processing Video...</p>
            <p className="text-xs text-muted-foreground">
              Generating HLS variants and enabling encryption.
            </p>
          </div>
        </div>
        <Progress value={undefined} className="h-2" />
      </div>
    );
  }

  if (isUploading) {
    return (
      <div className="space-y-3 p-4 border rounded-lg bg-muted/30">
        <div className="flex items-center justify-between text-sm">
          <span className="font-medium">Uploading video...</span>
          <span>{uploadProgress}%</span>
        </div>
        <Progress value={uploadProgress} className="h-2" />
        <Button
          variant="ghost"
          size="sm"
          onClick={() => setIsUploading(false)}
          className="w-full text-xs"
        >
          Cancel
        </Button>
      </div>
    );
  }

  if (status?.status === "failed") {
    return (
      <div className="space-y-3 p-4 border border-destructive/50 rounded-lg bg-destructive/5">
        <div className="flex items-center gap-3">
          <AlertCircle className="h-5 w-5 text-destructive" />
          <div className="flex-1">
            <p className="text-sm font-medium text-destructive">
              Processing Failed
            </p>
            <p className="text-xs text-destructive/80">
              {status.processing_error ||
                "Unknown error occurred during transcoding."}
            </p>
          </div>
        </div>
        <Button
          variant="outline"
          size="sm"
          className="w-full"
          onClick={() => setStatus(null)}
        >
          Try Again
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-3">
      {!file ? (
        <div
          onClick={() => fileInputRef.current?.click()}
          className="border-2 border-dashed rounded-lg p-6 flex flex-col items-center justify-center cursor-pointer hover:border-primary/50 hover:bg-muted/50 transition-all"
        >
          <Upload className="h-8 w-8 text-muted-foreground mb-2" />
          <p className="text-sm font-medium">Click to upload video</p>
          <p className="text-xs text-muted-foreground">
            MP4, MOV, MKV up to 500MB
          </p>
          <input
            type="file"
            ref={fileInputRef}
            className="hidden"
            accept="video/*"
            onChange={handleFileChange}
          />
        </div>
      ) : (
        <div className="flex items-center justify-between p-3 border rounded-lg bg-muted/50">
          <div className="flex items-center gap-3 overflow-hidden">
            <Play className="h-5 w-5 text-primary" />
            <div className="flex-1 overflow-hidden">
              <p className="text-sm font-medium truncate">{file.name}</p>
              <p className="text-xs text-muted-foreground">
                {(file.size / (1024 * 1024)).toFixed(2)} MB
              </p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="ghost" size="icon" onClick={() => setFile(null)}>
              <X className="h-4 w-4" />
            </Button>
            <Button size="sm" onClick={handleUpload}>
              Upload
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
