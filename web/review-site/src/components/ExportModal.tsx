import { useState, useMemo } from "react";
import { useSelector } from "react-redux";
import type { RootState } from "../store";
import { selectFilteredRows } from "../store/slices/cardsSlice";
import { buildExportMarkdown } from "../utils/export";
import {
  Sparkles,
  X,
  Copy,
  Check,
  Camera,
} from "lucide-react";

interface ExportModalProps {
  open: boolean;
  onClose: () => void;
}

export function ExportModal({ open, onClose }: ExportModalProps) {
  const rows = useSelector(selectFilteredRows);
  const [includeImage, setIncludeImage] = useState(true);
  const [copied, setCopied] = useState(false);
  const [scope, setScope] = useState<"all" | "reviewed">("all");

  const reviews = useSelector((s: RootState) => s.review.cards);
  const commentPins = useSelector((s: RootState) => s.comments.pins);

  const markdown = useMemo(() => {
    const filteredRows =
      scope === "reviewed"
        ? rows.filter((row) => {
            const key = `${row.page}/${row.section}`;
            const r = reviews[key];
            return r && (r.status !== "unreviewed" || r.note || (r.comments && r.comments.length > 0));
          })
        : rows;

    return filteredRows
      .map((row) => {
        const key = `${row.page}/${row.section}`;
        const review = reviews[key] ?? {
          page: row.page,
          section: row.section,
          status: "unreviewed" as const,
          note: "",
          comments: commentPins[key] ?? [],
          updatedAt: "",
        };
        return buildExportMarkdown(row, null, review);
      })
      .join("\n\n---\n\n");
  }, [rows, reviews, commentPins, scope]);

  if (!open) return null;

  const handleCopy = () => {
    navigator.clipboard.writeText(markdown);
    setCopied(true);
    setTimeout(() => setCopied(false), 1600);
  };

  const reviewedCount = rows.filter((row) => {
    const key = `${row.page}/${row.section}`;
    const r = reviews[key];
    return r && (r.status !== "unreviewed" || r.note);
  }).length;

  return (
    <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-6">
      <div className="bg-cream w-full max-w-3xl max-h-[85vh] rounded-[2px] shadow-2xl border border-border flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-3 border-b border-border">
          <div className="flex items-center gap-2">
            <Sparkles size={16} />
            <span className="font-display text-lg">LLM handoff</span>
            <span className="text-xs text-text-muted font-mono ml-2">
              markdown + yaml
            </span>
          </div>
          <button
            onClick={onClose}
            className="text-text-muted hover:text-text"
          >
            <X size={18} />
          </button>
        </div>

        {/* Options */}
        <div className="px-5 py-3 border-b border-border flex items-center gap-4">
          <label className="flex items-center gap-2 text-sm cursor-pointer">
            <input
              type="checkbox"
              checked={includeImage}
              onChange={(e) => setIncludeImage(e.target.checked)}
              className="accent-text"
            />
            <Camera size={13} /> attach annotated screenshot
          </label>
          <div className="flex items-center gap-2 ml-4 text-xs">
            <button
              onClick={() => setScope("all")}
              className={`px-2 py-1 rounded-[3px] ${
                scope === "all"
                  ? "bg-text text-white"
                  : "bg-cream-mid text-text-muted"
              }`}
            >
              All ({rows.length})
            </button>
            <button
              onClick={() => setScope("reviewed")}
              className={`px-2 py-1 rounded-[3px] ${
                scope === "reviewed"
                  ? "bg-text text-white"
                  : "bg-cream-mid text-text-muted"
              }`}
            >
              Reviewed ({reviewedCount})
            </button>
          </div>
          <span className="text-xs text-text-faint font-mono ml-auto">
            {rows.length} cards
          </span>
        </div>

        {/* Preview */}
        <div className="flex-1 overflow-auto p-5">
          <pre className="font-mono text-[12px] leading-relaxed text-text whitespace-pre-wrap bg-white border border-border rounded-[2px] p-4">
            {markdown}
          </pre>
        </div>

        {/* Footer */}
        <div className="border-t border-border px-5 py-3 flex items-center gap-2 justify-end">
          <button
            onClick={onClose}
            className="px-3 py-1.5 text-sm rounded-[3px] border border-border hover:bg-white"
          >
            Cancel
          </button>
          <button
            onClick={handleCopy}
            className="flex items-center gap-1.5 px-3 py-1.5 text-sm rounded-[3px] bg-text text-white font-medium hover:bg-black"
          >
            {copied ? <Check size={14} /> : <Copy size={14} />}
            {copied ? "Copied" : "Copy markdown"}
          </button>
        </div>
      </div>
    </div>
  );
}
