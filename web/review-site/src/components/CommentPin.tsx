import type { CommentPin as CommentPinType, CommentType } from "../types";

const COMMENT_COLORS: Record<CommentType, string> = {
  issue: "#B91C1C",
  note: "#B68545",
  question: "#2D6F62",
  praise: "#714464",
};

export function CommentPin({
  pin,
  index,
  active,
  onClick,
}: {
  pin: CommentPinType;
  index: number;
  active: boolean;
  onClick: () => void;
}) {
  const color = COMMENT_COLORS[pin.type];
  return (
    <button
      onClick={(e) => {
        e.stopPropagation();
        onClick();
      }}
      className="absolute -translate-x-1/2 -translate-y-1/2 z-30"
      style={{ left: `${pin.x}%`, top: `${pin.y}%` }}
    >
      <span
        className={`flex items-center justify-center w-7 h-7 rounded-full text-white font-semibold text-xs shadow-lg ring-2 transition-all ${
          active ? "scale-110 ring-white" : "ring-white/80 hover:scale-105"
        }`}
        style={{
          background: color,
          outline: active ? `2px solid ${color}` : "none",
          outlineOffset: "2px",
        }}
      >
        {index + 1}
      </span>
    </button>
  );
}
