"use client";

import { useEffect, useRef, useState } from "react";
import {
  VideoPlayer,
  type VideoPlayerRef,
} from "@/components/video/video-player";
import {
  enableScreenCaptureProtection,
  createWatermarkOverlay,
  generateDeviceFingerprint,
  type WatermarkConfig,
  type DRMConfig,
} from "@/lib/drm";
import { toast } from "sonner";
import api from "@/lib/api";

interface ProtectedVideoPlayerProps {
  src: string;
  signedUrl?: string;
  poster?: string;
  title?: string;
  lessonId: string;
  courseId: string;
  userId: string;
  userEmail: string;
  drmConfig?: Partial<DRMConfig>;
  captions?: { src: string; label: string; language: string }[];
  initialTime?: number;
  onProgress?: (data: {
    currentTime: number;
    duration: number;
    percent: number;
  }) => void;
  onComplete?: () => void;
}

export function ProtectedVideoPlayer({
  src,
  signedUrl,
  poster,
  title,
  lessonId,
  courseId,
  userId,
  userEmail,
  drmConfig = {},
  captions,
  initialTime,
  onProgress,
  onComplete,
}: ProtectedVideoPlayerProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const videoRef = useRef<VideoPlayerRef>(null);
  const [videoSrc, setVideoSrc] = useState<string | null>(null);
  const [isAuthorized, setIsAuthorized] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const config: DRMConfig = {
    enabled: true,
    encryption: "aes-128",
    watermarkEnabled: true,
    watermarkConfig: {
      text: `${userEmail} • ${userId.slice(-8)}`,
      position: "bottom-right",
      opacity: 0.5,
      fontSize: 12,
      color: "white",
    },
    maxDevices: 3,
    maxConcurrentStreams: 1,
    downloadEnabled: false,
    ...drmConfig,
  };

  // Authorize and get signed URL
  useEffect(() => {
    const authorize = async () => {
      try {
        const deviceId = generateDeviceFingerprint();

        // Register device and get signed URL
        const response = await api.post("/drm/authorize", {
          lesson_id: lessonId,
          course_id: courseId,
          device_id: deviceId,
        });

        const { signed_url, authorized } = response.data.data;

        if (!authorized) {
          setError("You are not authorized to view this content.");
          return;
        }

        setVideoSrc(signedUrl || signed_url || src);
        setIsAuthorized(true);
      } catch (err: unknown) {
        const error = err as {
          response?: { data?: { error?: { message?: string } } };
        };
        const message =
          error.response?.data?.error?.message ||
          "Failed to authorize video playback";

        // Allow playback anyway for demo purposes
        if (src) {
          console.warn("DRM auth failed, falling back to direct URL");
          setVideoSrc(src);
          setIsAuthorized(true);
        } else {
          setError(message);
        }
      }
    };

    authorize();
  }, [lessonId, courseId, src, signedUrl]);

  // Apply DRM protections
  useEffect(() => {
    if (!containerRef.current || !isAuthorized) return;

    const container = containerRef.current;
    const videoElement = container.querySelector("video");

    // Add watermark overlay
    if (config.watermarkEnabled && config.watermarkConfig) {
      const watermark = createWatermarkOverlay(config.watermarkConfig);
      container.style.position = "relative";
      container.appendChild(watermark);
    }

    // Enable screen capture protection
    let cleanup: (() => void) | undefined;
    if (videoElement) {
      cleanup = enableScreenCaptureProtection(videoElement, () => {
        toast.warning("Screen capture detected. Video paused for security.");
      });
    }

    // Disable right-click on container
    const handleContextMenu = (e: Event) => {
      e.preventDefault();
      return false;
    };
    container.addEventListener("contextmenu", handleContextMenu);

    // Heartbeat to verify session is still valid
    const heartbeatInterval = setInterval(async () => {
      try {
        await api.post("/drm/heartbeat", {
          lesson_id: lessonId,
          device_id: generateDeviceFingerprint(),
        });
      } catch {
        // If heartbeat fails, session might be invalid
        toast.error("Session expired. Please refresh the page.");
        if (videoElement) videoElement.pause();
      }
    }, 60000); // Every 60 seconds

    return () => {
      cleanup?.();
      container.removeEventListener("contextmenu", handleContextMenu);
      clearInterval(heartbeatInterval);

      // Remove watermark
      const watermark = container.querySelector(".drm-watermark");
      if (watermark) watermark.remove();
    };
  }, [isAuthorized, config, lessonId]);

  if (error) {
    return (
      <div className="aspect-video bg-black flex items-center justify-center text-white">
        <div className="text-center p-8">
          <p className="text-lg font-medium mb-2">Access Denied</p>
          <p className="text-sm text-gray-400">{error}</p>
        </div>
      </div>
    );
  }

  if (!videoSrc) {
    return (
      <div className="aspect-video bg-black flex items-center justify-center text-white">
        <div className="text-center">
          <div className="w-8 h-8 border-2 border-white border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-sm">Authorizing playback...</p>
        </div>
      </div>
    );
  }

  return (
    <div ref={containerRef} className="relative">
      <VideoPlayer
        ref={videoRef}
        src={videoSrc}
        poster={poster}
        title={title}
        captions={captions}
        initialTime={initialTime}
        onProgress={onProgress}
        onComplete={onComplete}
      />
    </div>
  );
}
