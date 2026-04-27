import { useDispatch } from "react-redux";
import type { AppDispatch } from "../store";
import type { SummaryRow, CommentPin, CommentType } from "../types";
import { updatePin, deletePin } from "../store/slices/commentsSlice";
import { Pin, Trash2 } from "lucide-react";

const COMMENT_TYPES: Record<CommentType, { label: string; color: string }> = {
  issue: { label: "issue", color: "#B91C1C" },
  note: { label: "note", color: "#B68545" },
  question: { label: "question", color: "#2D6F62" },
  praise: { label: "praise", color: "#714464" },
};

export function CommentsTab({
  pins,
  row,
}: {
  pins: CommentPin[];
  row: SummaryRow;
}) {
  const dispatch = useDispatch<AppDispatch>();

  if (pins.length === 0) {
    return (
      <div className="p-6 text-center text-text-muted text-sm">
        <Pin size={18} className="mx-auto mb-2 opacity-50" />
        No pins yet. Click <span className="font-medium">Add comment</span>{" "}
        then click on the canvas.
      </div>
    );
  }

  return (
    <div className="divide-y divide-border">
      {pins.map((pin, i) => {
        const meta = COMMENT_TYPES[pin.type];
        const sideLabel =
          pin.side === "left"
            ? "prototype"
            : pin.side === "right"
              ? "react"
              : "merged";
        return (
          <div key={pin.id} className="p-3 hover:bg-cream-light">
            <div className="flex items-start gap-2">
              <span
                className="flex items-center justify-center w-6 h-6 rounded-full text-white text-xs font-semibold shrink-0 font-mono"
                style={{ background: meta.color }}
              >
                {i + 1}
              </span>
              <div className="flex-1 min-w-0">
                <div className="flex items-center gap-2 mb-1">
                  <select
                    value={pin.type}
                    onChange={(e) =>
                      dispatch(
                        updatePin({
                          page: row.page,
                          section: row.section,
                          id: pin.id,
                          patch: { type: e.target.value as CommentType },
                        }),
                      )
                    }
                    className="text-[11px] uppercase tracking-wider font-semibold bg-transparent outline-none cursor-pointer"
                    style={{ color: meta.color }}
                  >
                    {Object.entries(COMMENT_TYPES).map(([k, t]) => (
                      <option key={k} value={k}>
                        {t.label}
                      </option>
                    ))}
                  </select>
                  <span className="text-[10px] text-text-faint font-mono">
                    {sideLabel} · {pin.x.toFixed(0)},{pin.y.toFixed(0)}
                  </span>
                  <button
                    onClick={() =>
                      dispatch(
                        deletePin({
                          page: row.page,
                          section: row.section,
                          id: pin.id,
                        }),
                      )
                    }
                    className="ml-auto text-text-faint hover:text-red"
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
                <textarea
                  value={pin.text}
                  onChange={(e) =>
                    dispatch(
                      updatePin({
                        page: row.page,
                        section: row.section,
                        id: pin.id,
                        patch: { text: e.target.value },
                      }),
                    )
                  }
                  placeholder="What's wrong here?"
                  className="w-full text-sm bg-transparent outline-none resize-none text-text placeholder:text-text-faint"
                  rows={pin.text.length > 80 ? 3 : 2}
                />
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
}
