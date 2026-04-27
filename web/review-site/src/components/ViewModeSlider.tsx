import { useState, useRef, useEffect } from "react";
import { useSelector, useDispatch } from "react-redux";
import type { RootState, AppDispatch } from "../store";
import { setSliderPosition } from "../store/slices/viewSlice";
import { GripVertical, CornerDownRight } from "lucide-react";

export function ViewModeSlider({
  leftUrl,
  rightUrl,
}: {
  leftUrl: string;
  rightUrl: string;
}) {
  const dispatch = useDispatch<AppDispatch>();
  const position = useSelector((s: RootState) => s.view.sliderPosition);
  const [dragging, setDragging] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!dragging) return;
    const move = (e: MouseEvent) => {
      if (!containerRef.current) return;
      const rect = containerRef.current.getBoundingClientRect();
      const pos = ((e.clientX - rect.left) / rect.width) * 100;
      dispatch(setSliderPosition(Math.max(0, Math.min(100, pos))));
    };
    const up = () => setDragging(false);
    window.addEventListener("mousemove", move);
    window.addEventListener("mouseup", up);
    return () => {
      window.removeEventListener("mousemove", move);
      window.removeEventListener("mouseup", up);
    };
  }, [dragging, dispatch]);

  return (
    <div>
      <div className="flex items-center gap-3 px-3 py-2 bg-cream-light border border-border rounded-[2px] mb-3 text-xs text-text-muted">
        <CornerDownRight size={12} /> drag the handle to sweep
      </div>
      <div
        ref={containerRef}
        className="relative select-none border border-border rounded-[2px] overflow-hidden"
      >
        {/* Prototype (left portion) */}
        <img
          src={leftUrl}
          alt="Prototype"
          className="block max-w-full"
          style={{ clipPath: `inset(0 ${100 - position}% 0 0)` }}
        />

        {/* React (right portion) */}
        <img
          src={rightUrl}
          alt="React"
          className="absolute inset-0 max-w-full"
          style={{ clipPath: `inset(0 0 0 ${position}%)`, pointerEvents: "none" }}
        />

        {/* Slider handle */}
        <div
          className="absolute top-0 bottom-0 z-20"
          style={{ left: `${position}%`, transform: "translateX(-50%)" }}
        >
          <div className="w-px h-full bg-text/40" />
          <button
            onMouseDown={(e) => {
              e.stopPropagation();
              setDragging(true);
            }}
            className="absolute top-1/2 -translate-y-1/2 -translate-x-1/2 left-1/2 w-7 h-7 rounded-full bg-text text-white flex items-center justify-center shadow-lg cursor-ew-resize"
          >
            <GripVertical size={14} />
          </button>
        </div>
      </div>
    </div>
  );
}
