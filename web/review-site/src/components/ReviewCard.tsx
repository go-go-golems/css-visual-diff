import { useState, useEffect, useRef, useCallback } from "react";
import { useSelector, useDispatch } from "react-redux";
import type { AppDispatch, RootState } from "../store";
import type { SummaryRow, CompareData } from "../types";
import { selectReview } from "../store/slices/reviewSlice";
import { setStatus, setNote } from "../store/slices/reviewSlice";
import { addPin, selectPins } from "../store/slices/commentsSlice";
import { ViewModeSideBySide } from "./ViewModeSideBySide";
import { ViewModeOverlay } from "./ViewModeOverlay";
import { ViewModeSlider } from "./ViewModeSlider";
import { ViewModeDiff } from "./ViewModeDiff";
import { CommentPin } from "./CommentPin";
import { FileText } from "lucide-react";
import type { ReviewStatus } from "../types";
import { compareJsonUrl } from "../utils/paths";

const STATUS_OPTIONS: { value: ReviewStatus; label: string }[] = [
  { value: "unreviewed", label: "Unreviewed" },
  { value: "accepted", label: "Accepted" },
  { value: "needs-work", label: "Needs work" },
  { value: "fixed", label: "Fixed" },
  { value: "wont-fix", label: "Won't fix" },
];

const CLASSIFICATION_COLORS: Record<string, string> = {
  accepted: "bg-green-100 text-green-800",
  review: "bg-amber-100 text-amber-800",
  "tune-required": "bg-orange-100 text-orange-800",
  "major-mismatch": "bg-red-100 text-red-800",
};

export function ReviewCard({
  row,
}: {
  row: SummaryRow;
}) {
  const dispatch = useDispatch<AppDispatch>();
  const viewMode = useSelector((s: RootState) => s.view.mode);
  const commentMode = useSelector((s: RootState) => s.view.commentMode);
  const commentDraftType = useSelector(
    (s: RootState) => s.view.commentDraftType,
  );
  const review = useSelector(selectReview(row.page, row.section));
  const pins = useSelector(selectPins(row.page, row.section));
  const [compareData, setCompareData] = useState<CompareData | null>(null);
  const [expanded, setExpanded] = useState(false);
  const [activePinId, setActivePinId] = useState<string | null>(null);
  const canvasRef = useRef<HTMLDivElement>(null);

  // Lazily load compare.json when expanded
  useEffect(() => {
    if (!expanded || compareData) return;
    fetch(compareJsonUrl(row.page, row.section))
      .then((res) => (res.ok ? res.json() : Promise.reject(res.status)))
      .then(setCompareData)
      .catch(console.error);
  }, [expanded, compareData, row.page, row.section]);

  const handleCanvasClick = useCallback(
    (e: React.MouseEvent) => {
      if (!commentMode || !canvasRef.current) return;
      const rect = canvasRef.current.getBoundingClientRect();
      const x = ((e.clientX - rect.left) / rect.width) * 100;
      const y = ((e.clientY - rect.top) / rect.height) * 100;
      const side =
        viewMode === "side-by-side" ? "left" as const : "merged" as const;
      const newPin = {
        id: `pin-${Date.now()}`,
        x,
        y,
        side,
        type: commentDraftType,
        text: "",
      };
      dispatch(
        addPin({ page: row.page, section: row.section, pin: newPin }),
      );
      setActivePinId(newPin.id);
      dispatch({ type: "view/setCommentMode", payload: false });
    },
    [commentMode, commentDraftType, viewMode, dispatch, row.page, row.section],
  );

  return (
    <section
      className={`bg-white border rounded-lg overflow-hidden shadow-sm ${
        row.classification === "tune-required"
          ? "border-orange-300"
          : row.classification === "major-mismatch"
            ? "border-red-300"
            : "border-border"
      }`}
      data-card={`${row.page}/${row.section}`}
    >
      {/* Header */}
      <div
        className="flex items-center justify-between px-4 py-3 cursor-pointer hover:bg-cream-light"
        onClick={() => setExpanded(!expanded)}
      >
        <div className="flex items-center gap-3">
          <span className="font-mono font-semibold">
            {row.page} <span className="text-text-faint">/</span>{" "}
            {row.section}
          </span>
          <span
            className={`text-[11px] uppercase tracking-wider rounded-[3px] px-2 py-0.5 font-semibold ${CLASSIFICATION_COLORS[row.classification] || "bg-gray-100 text-gray-800"}`}
          >
            {row.classification}
          </span>
          <span className="font-mono text-xs text-text-muted">
            {row.changedPercent.toFixed(2)}%
          </span>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={review.status}
            onClick={(e) => e.stopPropagation()}
            onChange={(e) =>
              dispatch(
                setStatus({
                  page: row.page,
                  section: row.section,
                  status: e.target.value as ReviewStatus,
                }),
              )
            }
            className="text-xs border border-border-light rounded-[3px] px-2 py-1 bg-white"
          >
            {STATUS_OPTIONS.map((o) => (
              <option key={o.value} value={o.value}>
                {o.label}
              </option>
            ))}
          </select>
          <span className="text-text-faint text-sm">
            {expanded ? "▾" : "▸"}
          </span>
        </div>
      </div>

      {/* Expanded content */}
      {expanded && (
        <div className="border-t border-border">
          {/* Canvas */}
          <div
            ref={canvasRef}
            onClick={handleCanvasClick}
            className={`relative p-4 ${
              commentMode ? "cursor-crosshair" : ""
            }`}
          >
            {viewMode === "side-by-side" && (
              <ViewModeSideBySide
                leftUrl={row.leftRegionPath}
                rightUrl={row.rightRegionPath}
                leftLabel={row.leftSelector}
                rightLabel={row.rightSelector}
              />
            )}
            {viewMode === "overlay" && (
              <ViewModeOverlay
                leftUrl={row.leftRegionPath}
                rightUrl={row.rightRegionPath}
              />
            )}
            {viewMode === "slider" && (
              <ViewModeSlider
                leftUrl={row.leftRegionPath}
                rightUrl={row.rightRegionPath}
              />
            )}
            {viewMode === "diff" && (
              <ViewModeDiff diffUrl={row.diffOnlyPath} />
            )}

            {/* Comment pins */}
            {pins.map((pin, i) => (
              <CommentPin
                key={pin.id}
                pin={pin}
                index={i}
                active={activePinId === pin.id}
                onClick={() =>
                  setActivePinId(
                    activePinId === pin.id ? null : pin.id,
                  )
                }
              />
            ))}
          </div>

          {/* Note textarea */}
          <div className="px-4 pb-4">
            <div className="flex items-center gap-2 mb-2">
              <FileText size={13} className="text-text-muted" />
              <span className="text-xs uppercase tracking-wider font-semibold text-text-muted">
                General observation
              </span>
            </div>
            <textarea
              value={review.note}
              onChange={(e) =>
                dispatch(
                  setNote({
                    page: row.page,
                    section: row.section,
                    note: e.target.value,
                  }),
                )
              }
              className="w-full text-sm border border-border-light rounded-[3px] px-3 py-2 bg-cream-light outline-none resize-none text-text placeholder:text-text-faint"
              rows={3}
              placeholder={`Reviewer notes for ${row.page}/${row.section}…`}
            />
          </div>

          {/* Artifact links */}
          <div className="px-4 pb-3 flex gap-4 text-xs">
            <a
              href={row.leftRegionPath}
              className="text-red hover:underline"
              target="_blank"
              rel="noreferrer"
            >
              left/prototype
            </a>
            <a
              href={row.rightRegionPath}
              className="text-red hover:underline"
              target="_blank"
              rel="noreferrer"
            >
              right/react
            </a>
            <a
              href={row.diffOnlyPath}
              className="text-red hover:underline"
              target="_blank"
              rel="noreferrer"
            >
              diff_only
            </a>
            <a
              href={row.artifactJson}
              className="text-red hover:underline"
              target="_blank"
              rel="noreferrer"
            >
              compare.json
            </a>
          </div>
        </div>
      )}
    </section>
  );
}
