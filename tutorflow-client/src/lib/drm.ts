// DRM (Digital Rights Management) utilities for content protection

/**
 * Generate a signed URL for protected video content
 * The URL expires after a specified time and is tied to a specific user/session
 */
export interface SignedUrlOptions {
  videoId: string;
  userId: string;
  sessionId: string;
  expiresIn?: number; // seconds, default 4 hours
}

export interface SignedUrlResponse {
  url: string;
  expires_at: string;
  token: string;
}

/**
 * Watermark configuration for video overlay
 */
export interface WatermarkConfig {
  text: string;
  position:
    | "top-left"
    | "top-right"
    | "bottom-left"
    | "bottom-right"
    | "center";
  opacity: number;
  fontSize: number;
  color: string;
}

/**
 * Device tracking for concurrent stream limits
 */
export interface DeviceInfo {
  id: string;
  name: string;
  type: "desktop" | "mobile" | "tablet";
  browser: string;
  os: string;
  lastActive: string;
  isCurrent: boolean;
}

/**
 * DRM configuration per course/video
 */
export interface DRMConfig {
  enabled: boolean;
  encryption: "aes-128" | "widevine" | "fairplay" | "none";
  watermarkEnabled: boolean;
  watermarkConfig?: WatermarkConfig;
  maxDevices: number;
  maxConcurrentStreams: number;
  downloadEnabled: boolean;
  expiresAfterDays?: number; // Content access expiry
}

/**
 * Generate device fingerprint for tracking
 */
export function generateDeviceFingerprint(): string {
  if (typeof window === "undefined") return "";

  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");
  if (ctx) {
    ctx.textBaseline = "top";
    ctx.font = "14px Arial";
    ctx.fillText("TutorFlow", 2, 2);
  }

  const components = [
    navigator.userAgent,
    navigator.language,
    screen.width + "x" + screen.height,
    screen.colorDepth,
    new Date().getTimezoneOffset(),
    navigator.hardwareConcurrency || 0,
    canvas.toDataURL(),
  ];

  return hashCode(components.join("|"));
}

function hashCode(str: string): string {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash;
  }
  return Math.abs(hash).toString(16);
}

/**
 * Get device type from user agent
 */
export function getDeviceType(): "desktop" | "mobile" | "tablet" {
  if (typeof window === "undefined") return "desktop";

  const ua = navigator.userAgent.toLowerCase();
  if (/tablet|ipad|playbook|silk/.test(ua)) return "tablet";
  if (/mobile|iphone|ipod|android|blackberry|opera mini|iemobile/.test(ua))
    return "mobile";
  return "desktop";
}

/**
 * Get browser name
 */
export function getBrowserName(): string {
  if (typeof window === "undefined") return "unknown";

  const ua = navigator.userAgent;
  if (ua.includes("Firefox")) return "Firefox";
  if (ua.includes("Edg")) return "Edge";
  if (ua.includes("Chrome")) return "Chrome";
  if (ua.includes("Safari")) return "Safari";
  if (ua.includes("Opera") || ua.includes("OPR")) return "Opera";
  return "Unknown";
}

/**
 * Get OS name
 */
export function getOSName(): string {
  if (typeof window === "undefined") return "unknown";

  const ua = navigator.userAgent;
  if (ua.includes("Win")) return "Windows";
  if (ua.includes("Mac")) return "macOS";
  if (ua.includes("Linux")) return "Linux";
  if (ua.includes("Android")) return "Android";
  if (ua.includes("iOS") || ua.includes("iPhone") || ua.includes("iPad"))
    return "iOS";
  return "Unknown";
}

/**
 * Anti-screen capture detection (limited effectiveness)
 * This is mainly for deterrence, not a foolproof solution
 */
export function enableScreenCaptureProtection(
  videoElement: HTMLVideoElement,
  onCapture?: () => void
): () => void {
  if (typeof window === "undefined") return () => {};

  // Disable context menu on video
  const handleContextMenu = (e: Event) => {
    e.preventDefault();
    return false;
  };

  // Detect visibility change (might indicate screen recording)
  const handleVisibilityChange = () => {
    if (document.hidden) {
      videoElement.pause();
    }
  };

  // Detect keyboard shortcuts for screenshots
  const handleKeyDown = (e: KeyboardEvent) => {
    // PrintScreen, Cmd+Shift+3/4 (Mac), etc.
    if (
      e.key === "PrintScreen" ||
      (e.metaKey &&
        e.shiftKey &&
        (e.key === "3" || e.key === "4" || e.key === "5"))
    ) {
      e.preventDefault();
      videoElement.pause();
      onCapture?.();
    }
  };

  videoElement.addEventListener("contextmenu", handleContextMenu);
  document.addEventListener("visibilitychange", handleVisibilityChange);
  document.addEventListener("keydown", handleKeyDown);

  return () => {
    videoElement.removeEventListener("contextmenu", handleContextMenu);
    document.removeEventListener("visibilitychange", handleVisibilityChange);
    document.removeEventListener("keydown", handleKeyDown);
  };
}

/**
 * Create watermark overlay element
 */
export function createWatermarkOverlay(
  config: WatermarkConfig
): HTMLDivElement {
  const overlay = document.createElement("div");
  overlay.className = "drm-watermark";
  overlay.textContent = config.text;
  overlay.style.cssText = `
    position: absolute;
    color: ${config.color};
    opacity: ${config.opacity};
    font-size: ${config.fontSize}px;
    font-family: sans-serif;
    pointer-events: none;
    user-select: none;
    z-index: 100;
    ${getPositionStyles(config.position)}
  `;
  return overlay;
}

function getPositionStyles(position: WatermarkConfig["position"]): string {
  switch (position) {
    case "top-left":
      return "top: 10px; left: 10px;";
    case "top-right":
      return "top: 10px; right: 10px;";
    case "bottom-left":
      return "bottom: 60px; left: 10px;";
    case "bottom-right":
      return "bottom: 60px; right: 10px;";
    case "center":
      return "top: 50%; left: 50%; transform: translate(-50%, -50%);";
    default:
      return "top: 10px; left: 10px;";
  }
}

/**
 * DRM status check
 */
export interface DRMStatus {
  isSupported: boolean;
  widevine: boolean;
  fairplay: boolean;
  playready: boolean;
}

export async function checkDRMSupport(): Promise<DRMStatus> {
  if (typeof window === "undefined" || !navigator.requestMediaKeySystemAccess) {
    return {
      isSupported: false,
      widevine: false,
      fairplay: false,
      playready: false,
    };
  }

  const config = [
    {
      initDataTypes: ["cenc"],
      videoCapabilities: [{ contentType: 'video/mp4; codecs="avc1.42E01E"' }],
    },
  ];

  const checkKeySystem = async (keySystem: string): Promise<boolean> => {
    try {
      await navigator.requestMediaKeySystemAccess(keySystem, config);
      return true;
    } catch {
      return false;
    }
  };

  const [widevine, fairplay, playready] = await Promise.all([
    checkKeySystem("com.widevine.alpha"),
    checkKeySystem("com.apple.fps.1_0"),
    checkKeySystem("com.microsoft.playready"),
  ]);

  return {
    isSupported: widevine || fairplay || playready,
    widevine,
    fairplay,
    playready,
  };
}
