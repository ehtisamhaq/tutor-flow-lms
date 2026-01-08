"use client";

import {
  useState,
  useRef,
  useEffect,
  useCallback,
  forwardRef,
  useImperativeHandle,
} from "react";
import {
  Play,
  Pause,
  Volume2,
  VolumeX,
  Maximize,
  Minimize,
  Settings,
  Bookmark,
  BookmarkCheck,
  SkipBack,
  SkipForward,
  Subtitles,
  Check,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

interface Caption {
  src: string;
  label: string;
  language: string;
}

interface VideoBookmark {
  id: string;
  time: number;
  title: string;
  created_at: string;
}

interface VideoPlayerProps {
  src: string;
  poster?: string;
  title?: string;
  captions?: Caption[];
  bookmarks?: VideoBookmark[];
  initialTime?: number;
  onProgress?: (progress: {
    currentTime: number;
    duration: number;
    percent: number;
  }) => void;
  onBookmarkAdd?: (time: number) => void;
  onBookmarkRemove?: (id: string) => void;
  onComplete?: () => void;
}

export interface VideoPlayerRef {
  play: () => void;
  pause: () => void;
  seek: (time: number) => void;
  getCurrentTime: () => number;
}

const PLAYBACK_SPEEDS = [0.5, 0.75, 1, 1.25, 1.5, 1.75, 2];

export const VideoPlayer = forwardRef<VideoPlayerRef, VideoPlayerProps>(
  (
    {
      src,
      poster,
      title,
      captions = [],
      bookmarks = [],
      initialTime = 0,
      onProgress,
      onBookmarkAdd,
      onBookmarkRemove,
      onComplete,
    },
    ref
  ) => {
    const videoRef = useRef<HTMLVideoElement>(null);
    const containerRef = useRef<HTMLDivElement>(null);
    const progressRef = useRef<HTMLDivElement>(null);

    const [isPlaying, setIsPlaying] = useState(false);
    const [isMuted, setIsMuted] = useState(false);
    const [volume, setVolume] = useState(1);
    const [currentTime, setCurrentTime] = useState(0);
    const [duration, setDuration] = useState(0);
    const [isFullscreen, setIsFullscreen] = useState(false);
    const [showControls, setShowControls] = useState(true);
    const [playbackSpeed, setPlaybackSpeed] = useState(1);
    const [showSpeedMenu, setShowSpeedMenu] = useState(false);
    const [showCaptionMenu, setShowCaptionMenu] = useState(false);
    const [activeCaption, setActiveCaption] = useState<string | null>(null);
    const [buffered, setBuffered] = useState(0);

    const controlsTimeoutRef = useRef<NodeJS.Timeout | null>(null);

    // Expose methods to parent
    useImperativeHandle(ref, () => ({
      play: () => videoRef.current?.play(),
      pause: () => videoRef.current?.pause(),
      seek: (time: number) => {
        if (videoRef.current) videoRef.current.currentTime = time;
      },
      getCurrentTime: () => videoRef.current?.currentTime || 0,
    }));

    // Auto-hide controls
    const showControlsTemporarily = useCallback(() => {
      setShowControls(true);
      if (controlsTimeoutRef.current) {
        clearTimeout(controlsTimeoutRef.current);
      }
      if (isPlaying) {
        controlsTimeoutRef.current = setTimeout(() => {
          setShowControls(false);
        }, 3000);
      }
    }, [isPlaying]);

    // Initialize video
    useEffect(() => {
      const video = videoRef.current;
      if (!video) return;

      const handleLoadedMetadata = () => {
        setDuration(video.duration);
        if (initialTime > 0) {
          video.currentTime = initialTime;
        }
      };

      const handleTimeUpdate = () => {
        setCurrentTime(video.currentTime);
        onProgress?.({
          currentTime: video.currentTime,
          duration: video.duration,
          percent: (video.currentTime / video.duration) * 100,
        });
      };

      const handleProgress = () => {
        if (video.buffered.length > 0) {
          const bufferedEnd = video.buffered.end(video.buffered.length - 1);
          setBuffered((bufferedEnd / video.duration) * 100);
        }
      };

      const handleEnded = () => {
        setIsPlaying(false);
        onComplete?.();
      };

      const handlePlay = () => setIsPlaying(true);
      const handlePause = () => setIsPlaying(false);

      video.addEventListener("loadedmetadata", handleLoadedMetadata);
      video.addEventListener("timeupdate", handleTimeUpdate);
      video.addEventListener("progress", handleProgress);
      video.addEventListener("ended", handleEnded);
      video.addEventListener("play", handlePlay);
      video.addEventListener("pause", handlePause);

      return () => {
        video.removeEventListener("loadedmetadata", handleLoadedMetadata);
        video.removeEventListener("timeupdate", handleTimeUpdate);
        video.removeEventListener("progress", handleProgress);
        video.removeEventListener("ended", handleEnded);
        video.removeEventListener("play", handlePlay);
        video.removeEventListener("pause", handlePause);
      };
    }, [initialTime, onProgress, onComplete]);

    // Fullscreen change handler
    useEffect(() => {
      const handleFullscreenChange = () => {
        setIsFullscreen(!!document.fullscreenElement);
      };

      document.addEventListener("fullscreenchange", handleFullscreenChange);
      return () =>
        document.removeEventListener(
          "fullscreenchange",
          handleFullscreenChange
        );
    }, []);

    // Keyboard shortcuts
    useEffect(() => {
      const handleKeyDown = (e: KeyboardEvent) => {
        if (!containerRef.current?.contains(document.activeElement)) return;

        switch (e.key) {
          case " ":
          case "k":
            e.preventDefault();
            togglePlay();
            break;
          case "ArrowLeft":
            e.preventDefault();
            skip(-10);
            break;
          case "ArrowRight":
            e.preventDefault();
            skip(10);
            break;
          case "ArrowUp":
            e.preventDefault();
            changeVolume(0.1);
            break;
          case "ArrowDown":
            e.preventDefault();
            changeVolume(-0.1);
            break;
          case "m":
            e.preventDefault();
            toggleMute();
            break;
          case "f":
            e.preventDefault();
            toggleFullscreen();
            break;
        }
      };

      window.addEventListener("keydown", handleKeyDown);
      return () => window.removeEventListener("keydown", handleKeyDown);
    }, []);

    const togglePlay = () => {
      if (videoRef.current) {
        if (isPlaying) {
          videoRef.current.pause();
        } else {
          videoRef.current.play();
        }
      }
    };

    const toggleMute = () => {
      if (videoRef.current) {
        videoRef.current.muted = !isMuted;
        setIsMuted(!isMuted);
      }
    };

    const changeVolume = (delta: number) => {
      if (videoRef.current) {
        const newVolume = Math.max(0, Math.min(1, volume + delta));
        videoRef.current.volume = newVolume;
        setVolume(newVolume);
        if (newVolume === 0) {
          setIsMuted(true);
        } else if (isMuted) {
          setIsMuted(false);
        }
      }
    };

    const skip = (seconds: number) => {
      if (videoRef.current) {
        videoRef.current.currentTime = Math.max(
          0,
          Math.min(duration, currentTime + seconds)
        );
      }
    };

    const handleProgressClick = (e: React.MouseEvent<HTMLDivElement>) => {
      if (!progressRef.current || !videoRef.current) return;
      const rect = progressRef.current.getBoundingClientRect();
      const percent = (e.clientX - rect.left) / rect.width;
      videoRef.current.currentTime = percent * duration;
    };

    const toggleFullscreen = () => {
      if (!containerRef.current) return;

      if (!isFullscreen) {
        containerRef.current.requestFullscreen();
      } else {
        document.exitFullscreen();
      }
    };

    const setSpeed = (speed: number) => {
      if (videoRef.current) {
        videoRef.current.playbackRate = speed;
        setPlaybackSpeed(speed);
        setShowSpeedMenu(false);
      }
    };

    const toggleCaption = (language: string | null) => {
      if (videoRef.current) {
        const tracks = videoRef.current.textTracks;
        for (let i = 0; i < tracks.length; i++) {
          tracks[i].mode =
            tracks[i].language === language ? "showing" : "hidden";
        }
        setActiveCaption(language);
        setShowCaptionMenu(false);
      }
    };

    const formatTime = (time: number) => {
      const hours = Math.floor(time / 3600);
      const minutes = Math.floor((time % 3600) / 60);
      const seconds = Math.floor(time % 60);

      if (hours > 0) {
        return `${hours}:${minutes.toString().padStart(2, "0")}:${seconds
          .toString()
          .padStart(2, "0")}`;
      }
      return `${minutes}:${seconds.toString().padStart(2, "0")}`;
    };

    const isBookmarked = bookmarks.some(
      (b) => Math.abs(b.time - currentTime) < 3
    );

    const handleBookmark = () => {
      if (isBookmarked) {
        const bookmark = bookmarks.find(
          (b) => Math.abs(b.time - currentTime) < 3
        );
        if (bookmark) onBookmarkRemove?.(bookmark.id);
      } else {
        onBookmarkAdd?.(currentTime);
      }
    };

    return (
      <div
        ref={containerRef}
        className="relative bg-black aspect-video group"
        onMouseMove={showControlsTemporarily}
        onMouseLeave={() => isPlaying && setShowControls(false)}
        tabIndex={0}
      >
        {/* Video Element */}
        <video
          ref={videoRef}
          src={src}
          poster={poster}
          className="w-full h-full"
          onClick={togglePlay}
          playsInline
        >
          {captions.map((caption) => (
            <track
              key={caption.language}
              kind="subtitles"
              src={caption.src}
              srcLang={caption.language}
              label={caption.label}
            />
          ))}
        </video>

        {/* Bookmark markers on progress bar */}
        {bookmarks.length > 0 && (
          <div className="absolute bottom-12 left-0 right-0 h-1 pointer-events-none">
            {bookmarks.map((bookmark) => (
              <div
                key={bookmark.id}
                className="absolute w-2 h-2 bg-yellow-400 rounded-full transform -translate-x-1/2 -translate-y-1/2"
                style={{
                  left: `${(bookmark.time / duration) * 100}%`,
                  top: "50%",
                }}
                title={bookmark.title}
              />
            ))}
          </div>
        )}

        {/* Controls Overlay */}
        <div
          className={cn(
            "absolute inset-0 bg-gradient-to-t from-black/80 via-transparent to-black/40 transition-opacity",
            showControls ? "opacity-100" : "opacity-0 pointer-events-none"
          )}
        >
          {/* Top Bar */}
          <div className="absolute top-0 left-0 right-0 p-4 flex items-center justify-between">
            {title && (
              <h3 className="text-white font-medium truncate">{title}</h3>
            )}
          </div>

          {/* Center Play Button */}
          <button
            onClick={togglePlay}
            className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-16 h-16 bg-white/20 backdrop-blur-sm rounded-full flex items-center justify-center hover:bg-white/30 transition-colors"
          >
            {isPlaying ? (
              <Pause className="h-8 w-8 text-white" />
            ) : (
              <Play className="h-8 w-8 text-white ml-1" />
            )}
          </button>

          {/* Bottom Controls */}
          <div className="absolute bottom-0 left-0 right-0 p-4 space-y-2">
            {/* Progress Bar */}
            <div
              ref={progressRef}
              className="relative h-1 bg-white/30 rounded-full cursor-pointer group/progress"
              onClick={handleProgressClick}
            >
              {/* Buffered */}
              <div
                className="absolute h-full bg-white/50 rounded-full"
                style={{ width: `${buffered}%` }}
              />
              {/* Progress */}
              <div
                className="absolute h-full bg-primary rounded-full"
                style={{ width: `${(currentTime / duration) * 100}%` }}
              />
              {/* Thumb */}
              <div
                className="absolute w-3 h-3 bg-primary rounded-full transform -translate-y-1/3 -translate-x-1/2 opacity-0 group-hover/progress:opacity-100 transition-opacity"
                style={{ left: `${(currentTime / duration) * 100}%` }}
              />
            </div>

            {/* Control Buttons */}
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {/* Play/Pause */}
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-white hover:bg-white/20"
                  onClick={togglePlay}
                >
                  {isPlaying ? (
                    <Pause className="h-5 w-5" />
                  ) : (
                    <Play className="h-5 w-5" />
                  )}
                </Button>

                {/* Skip Back */}
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-white hover:bg-white/20"
                  onClick={() => skip(-10)}
                >
                  <SkipBack className="h-5 w-5" />
                </Button>

                {/* Skip Forward */}
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-white hover:bg-white/20"
                  onClick={() => skip(10)}
                >
                  <SkipForward className="h-5 w-5" />
                </Button>

                {/* Volume */}
                <div className="flex items-center gap-1">
                  <Button
                    variant="ghost"
                    size="icon"
                    className="text-white hover:bg-white/20"
                    onClick={toggleMute}
                  >
                    {isMuted || volume === 0 ? (
                      <VolumeX className="h-5 w-5" />
                    ) : (
                      <Volume2 className="h-5 w-5" />
                    )}
                  </Button>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.1"
                    value={isMuted ? 0 : volume}
                    onChange={(e) => {
                      const newVolume = parseFloat(e.target.value);
                      if (videoRef.current) {
                        videoRef.current.volume = newVolume;
                        setVolume(newVolume);
                        setIsMuted(newVolume === 0);
                      }
                    }}
                    className="w-20 accent-primary"
                  />
                </div>

                {/* Time */}
                <span className="text-white text-sm">
                  {formatTime(currentTime)} / {formatTime(duration)}
                </span>
              </div>

              <div className="flex items-center gap-2">
                {/* Bookmark */}
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-white hover:bg-white/20"
                  onClick={handleBookmark}
                  title="Add bookmark"
                >
                  {isBookmarked ? (
                    <BookmarkCheck className="h-5 w-5 text-yellow-400" />
                  ) : (
                    <Bookmark className="h-5 w-5" />
                  )}
                </Button>

                {/* Speed */}
                <div className="relative">
                  <Button
                    variant="ghost"
                    size="sm"
                    className="text-white hover:bg-white/20 text-sm"
                    onClick={() => {
                      setShowSpeedMenu(!showSpeedMenu);
                      setShowCaptionMenu(false);
                    }}
                  >
                    {playbackSpeed}x
                  </Button>
                  {showSpeedMenu && (
                    <div className="absolute bottom-full right-0 mb-2 bg-black/90 rounded-lg p-2 min-w-[100px]">
                      {PLAYBACK_SPEEDS.map((speed) => (
                        <button
                          key={speed}
                          className={cn(
                            "w-full text-left px-3 py-1.5 text-sm rounded hover:bg-white/20 flex items-center justify-between",
                            playbackSpeed === speed
                              ? "text-primary"
                              : "text-white"
                          )}
                          onClick={() => setSpeed(speed)}
                        >
                          {speed}x
                          {playbackSpeed === speed && (
                            <Check className="h-4 w-4" />
                          )}
                        </button>
                      ))}
                    </div>
                  )}
                </div>

                {/* Captions */}
                {captions.length > 0 && (
                  <div className="relative">
                    <Button
                      variant="ghost"
                      size="icon"
                      className={cn(
                        "text-white hover:bg-white/20",
                        activeCaption && "text-primary"
                      )}
                      onClick={() => {
                        setShowCaptionMenu(!showCaptionMenu);
                        setShowSpeedMenu(false);
                      }}
                    >
                      <Subtitles className="h-5 w-5" />
                    </Button>
                    {showCaptionMenu && (
                      <div className="absolute bottom-full right-0 mb-2 bg-black/90 rounded-lg p-2 min-w-[120px]">
                        <button
                          className={cn(
                            "w-full text-left px-3 py-1.5 text-sm rounded hover:bg-white/20 flex items-center justify-between",
                            !activeCaption ? "text-primary" : "text-white"
                          )}
                          onClick={() => toggleCaption(null)}
                        >
                          Off
                          {!activeCaption && <Check className="h-4 w-4" />}
                        </button>
                        {captions.map((caption) => (
                          <button
                            key={caption.language}
                            className={cn(
                              "w-full text-left px-3 py-1.5 text-sm rounded hover:bg-white/20 flex items-center justify-between",
                              activeCaption === caption.language
                                ? "text-primary"
                                : "text-white"
                            )}
                            onClick={() => toggleCaption(caption.language)}
                          >
                            {caption.label}
                            {activeCaption === caption.language && (
                              <Check className="h-4 w-4" />
                            )}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Fullscreen */}
                <Button
                  variant="ghost"
                  size="icon"
                  className="text-white hover:bg-white/20"
                  onClick={toggleFullscreen}
                >
                  {isFullscreen ? (
                    <Minimize className="h-5 w-5" />
                  ) : (
                    <Maximize className="h-5 w-5" />
                  )}
                </Button>
              </div>
            </div>
          </div>
        </div>
      </div>
    );
  }
);

VideoPlayer.displayName = "VideoPlayer";
