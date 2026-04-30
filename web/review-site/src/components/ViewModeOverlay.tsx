import { useState, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import type { RootState, AppDispatch } from "../store";
import {
  setOverlayOpacity,
  setOverlayBlend,
} from "../store/slices/viewSlice";
import { Zap } from "lucide-react";

export function ViewModeOverlay({
  leftUrl,
  rightUrl,
}: {
  leftUrl: string;
  rightUrl: string;
}) {
  const dispatch = useDispatch<AppDispatch>();
  const opacity = useSelector((s: RootState) => s.view.overlayOpacity);
  const blend = useSelector((s: RootState) => s.view.overlayBlend);
  const [swapPressed, setSwapPressed] = useState(false);

  useEffect(() => {
    const down = (e: KeyboardEvent) => {
      if (
        e.key === "f" &&
        !e.repeat &&
        !(e.target instanceof HTMLTextAreaElement) &&
        !(e.target instanceof HTMLInputElement) &&
        !(e.target instanceof HTMLSelectElement)
      ) {
        e.preventDefault();
        setSwapPressed(true);
      }
    };
    const up = (e: KeyboardEvent) => {
      if (e.key === "f") setSwapPressed(false);
    };
    window.addEventListener("keydown", down);
    window.addEventListener("keyup", up);
    return () => {
      window.removeEventListener("keydown", down);
      window.removeEventListener("keyup", up);
    };
  }, []);

  const effectiveOpacity = swapPressed
    ? opacity > 50
      ? 0
      : 100
    : opacity;

  return (
    <div className="space-y-3">
      {/* Controls */}
      <div className="flex items-center gap-3 px-3 py-2 bg-cream-light border border-border rounded-[2px]">
        <span className="text-xs text-text-muted uppercase tracking-wider">
          A
        </span>
        <input
          type="range"
          min="0"
          max="100"
          value={effectiveOpacity}
          onChange={(e) =>
            dispatch(setOverlayOpacity(parseInt(e.target.value)))
          }
          className="flex-1 accent-text"
        />
        <span className="text-xs text-text-muted uppercase tracking-wider">
          B
        </span>
        <div className="w-px h-5 bg-border" />
        <button
          onClick={() =>
            dispatch(
              setOverlayBlend(blend === "normal" ? "difference" : "normal"),
            )
          }
          className={`px-2 py-1 text-xs rounded-[3px] flex items-center gap-1 ${
            blend === "difference"
              ? "bg-text text-white"
              : "bg-white border border-border-light text-text"
          }`}
        >
          <Zap size={12} /> diff blend
        </button>
        <span className="text-[10px] text-text-faint font-mono ml-auto">
          hold <kbd className="px-1 py-0.5 bg-cream-mid rounded text-text border border-border-light">F</kbd> to flash B
        </span>
      </div>

      {/* Stacked images */}
      <div className="relative border border-border rounded-[2px] overflow-hidden">
        <img src={leftUrl} alt="Prototype" className="block max-w-full" />
        <img
          src={rightUrl}
          alt="React"
          className="absolute inset-0 max-w-full"
          style={{
            opacity: effectiveOpacity / 100,
            mixBlendMode: blend === "difference" ? "difference" : "normal",
          }}
        />
      </div>
    </div>
  );
}
