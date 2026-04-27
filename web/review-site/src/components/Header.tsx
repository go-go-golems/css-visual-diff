import { useDispatch, useSelector } from "react-redux";
import type { AppDispatch, RootState } from "../store";
import { selectFilteredRows } from "../store/slices/cardsSlice";
import { setViewMode, setCommentMode } from "../store/slices/viewSlice";
import type { ViewMode } from "../store/slices/viewSlice";
import {
  SplitSquareHorizontal,
  Layers,
  ArrowLeftRight,
  Square,
  Sparkles,
  Pin,
} from "lucide-react";

const VIEW_MODES: {
  id: ViewMode;
  label: string;
  icon: React.ElementType;
}[] = [
  { id: "side-by-side", label: "Side-by-side", icon: SplitSquareHorizontal },
  { id: "overlay", label: "Overlay", icon: Layers },
  { id: "slider", label: "Slider", icon: ArrowLeftRight },
  { id: "diff", label: "Diff only", icon: Square },
];

export function Header() {
  const dispatch = useDispatch<AppDispatch>();
  const counts = useSelector((s: RootState) => s.cards.classificationCounts);
  const worst = useSelector((s: RootState) => s.cards.worstClassification);
  const pageCount = useSelector((s: RootState) => s.cards.pageCount);
  const filteredRows = useSelector(selectFilteredRows);
  const viewMode = useSelector((s: RootState) => s.view.mode);
  const commentMode = useSelector((s: RootState) => s.view.commentMode);
  const totalCards = filteredRows.length;

  return (
    <header className="border-b border-border bg-cream/90 backdrop-blur-sm sticky top-0 z-40">
      <div className="max-w-[1400px] mx-auto px-6 py-3 flex items-center gap-4">
        <div className="flex items-center gap-2">
          <div className="w-7 h-7 rounded-[2px] bg-text flex items-center justify-center">
            <span className="font-display text-white italic text-base">c</span>
          </div>
          <span className="font-display text-lg tracking-tight">
            css-visual-diff review
          </span>
        </div>

        <div className="h-5 w-px bg-border" />

        <div className="flex items-center gap-3 text-sm">
          <span className="text-text-muted">{pageCount} pages</span>
          <span className="text-text-faint">·</span>
          <span className="text-text-muted">{totalCards} sections</span>
          <span className="text-text-faint">·</span>
          <span className="text-text-muted font-mono text-xs">
            worst: {worst}
          </span>
        </div>

        <div className="ml-auto flex items-center gap-2">
          <button
            onClick={() => alert("Export — TODO")}
            className="flex items-center gap-1.5 px-3 py-1.5 rounded-[3px] bg-text text-white text-sm font-medium hover:bg-black"
          >
            <Sparkles size={14} /> Send to LLM
          </button>
        </div>
      </div>

      <div className="border-t border-border">
        <div className="max-w-[1400px] mx-auto px-6 py-2 flex items-center gap-2">
          <div className="flex items-center bg-white border border-border rounded-[3px] p-0.5">
            {VIEW_MODES.map((m) => {
              const Icon = m.icon;
              const active = viewMode === m.id;
              return (
                <button
                  key={m.id}
                  onClick={() => dispatch(setViewMode(m.id))}
                  title={m.label}
                  className={`flex items-center gap-1.5 px-2.5 py-1 rounded-[2px] text-xs font-medium transition-colors ${
                    active
                      ? "bg-text text-white"
                      : "text-text hover:bg-cream-mid"
                  }`}
                >
                  <Icon size={13} /> {m.label}
                </button>
              );
            })}
          </div>

          <div className="h-5 w-px bg-border" />

          <button
            onClick={() => dispatch(setCommentMode(!commentMode))}
            className={`flex items-center gap-1.5 px-2.5 py-1 rounded-[3px] text-xs font-medium transition-colors ${
              commentMode
                ? "bg-red text-white"
                : "bg-white border border-border text-text hover:bg-cream-mid"
            }`}
          >
            <Pin size={13} />{" "}
            {commentMode ? "click to drop pin…" : "Add comment"}
          </button>

          <div className="h-5 w-px bg-border" />

          {(
            ["tune-required", "review", "accepted", "major-mismatch"] as const
          ).map((cls) => {
            const count = counts[cls] ?? 0;
            if (count === 0) return null;
            return (
              <span
                key={cls}
                className="text-[10px] font-mono px-1.5 py-0.5 rounded-full bg-cream-mid text-text-muted"
              >
                {count} {cls}
              </span>
            );
          })}
        </div>
      </div>
    </header>
  );
}
