import { useState, useRef, useCallback, useEffect } from "react";

/**
 * ZoomPan wraps an image (or any content) and provides:
 * - scroll wheel to zoom in/out
 * - drag to pan
 * - double-click to reset
 * - displays current zoom level and pixel offset
 */
export function ZoomPan({
  children,
  className = "",
}: {
  children: React.ReactNode;
  className?: string;
}) {
  const [zoom, setZoom] = useState(1);
  const [offset, setOffset] = useState({ x: 0, y: 0 });
  const [dragging, setDragging] = useState(false);
  const [dragStart, setDragStart] = useState({ x: 0, y: 0 });
  const [offsetStart, setOffsetStart] = useState({ x: 0, y: 0 });
  const containerRef = useRef<HTMLDivElement>(null);

  const MIN_ZOOM = 0.25;
  const MAX_ZOOM = 8;
  const ZOOM_STEP = 0.15;

  // Wheel zoom: zoom toward cursor position
  const handleWheel = useCallback(
    (e: WheelEvent) => {
      e.preventDefault();
      const rect = containerRef.current?.getBoundingClientRect();
      if (!rect) return;

      const mouseX = e.clientX - rect.left;
      const mouseY = e.clientY - rect.top;

      const delta = e.deltaY > 0 ? -ZOOM_STEP : ZOOM_STEP;
      const newZoom = Math.max(MIN_ZOOM, Math.min(MAX_ZOOM, zoom * (1 + delta)));

      // Zoom toward cursor: adjust offset so the point under cursor stays fixed
      const scale = newZoom / zoom;
      const newOffsetX = mouseX - scale * (mouseX - offset.x);
      const newOffsetY = mouseY - scale * (mouseY - offset.y);

      setZoom(newZoom);
      setOffset({ x: newOffsetX, y: newOffsetY });
    },
    [zoom, offset],
  );

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    el.addEventListener("wheel", handleWheel, { passive: false });
    return () => el.removeEventListener("wheel", handleWheel);
  }, [handleWheel]);

  // Drag to pan
  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      // Only pan with middle button or when holding shift+left
      if (e.button === 1 || (e.button === 0 && e.shiftKey)) {
        e.preventDefault();
        setDragging(true);
        setDragStart({ x: e.clientX, y: e.clientY });
        setOffsetStart({ ...offset });
      }
    },
    [offset],
  );

  useEffect(() => {
    if (!dragging) return;
    const move = (e: MouseEvent) => {
      const dx = e.clientX - dragStart.x;
      const dy = e.clientY - dragStart.y;
      setOffset({
        x: offsetStart.x + dx,
        y: offsetStart.y + dy,
      });
    };
    const up = () => setDragging(false);
    window.addEventListener("mousemove", move);
    window.addEventListener("mouseup", up);
    return () => {
      window.removeEventListener("mousemove", move);
      window.removeEventListener("mouseup", up);
    };
  }, [dragging, dragStart, offsetStart]);

  // Double-click to reset
  const handleDoubleClick = useCallback(() => {
    setZoom(1);
    setOffset({ x: 0, y: 0 });
  }, []);

  const cursorStyle = dragging
    ? "grabbing"
    : zoom > 1
      ? "grab"
      : "default";

  return (
    <div className={`relative ${className}`}>
      <div
        ref={containerRef}
        onMouseDown={handleMouseDown}
        onDoubleClick={handleDoubleClick}
        className="overflow-hidden border border-border rounded-[2px] bg-cream-light"
        style={{ cursor: cursorStyle }}
      >
        <div
          style={{
            transform: `translate(${offset.x}px, ${offset.y}px) scale(${zoom})`,
            transformOrigin: "0 0",
            transition: dragging ? "none" : "transform 0.1s ease-out",
          }}
        >
          {children}
        </div>
      </div>
      {/* Zoom indicator */}
      {zoom !== 1 && (
        <div className="absolute bottom-2 left-2 flex items-center gap-2 bg-text/80 text-white text-[10px] font-mono px-2 py-1 rounded-[2px] backdrop-blur-sm">
          <span>{(zoom * 100).toFixed(0)}%</span>
          {offset.x !== 0 || offset.y !== 0 ? (
            <span>
              Δ{Math.round(offset.x)}, Δ{Math.round(offset.y)}
            </span>
          ) : null}
          <span className="text-text-faint">dbl-click reset</span>
        </div>
      )}
    </div>
  );
}
