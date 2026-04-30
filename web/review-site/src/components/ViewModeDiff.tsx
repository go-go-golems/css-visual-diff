import { Square } from "lucide-react";

export function ViewModeDiff({ diffUrl }: { diffUrl: string }) {
  return (
    <div>
      <div className="flex items-center gap-3 px-3 py-2 bg-cream-light border border-border rounded-[2px] mb-3 text-xs text-text-muted">
        <Square size={12} /> only the pixels that differ
      </div>
      <div className="border border-border rounded-[2px] overflow-hidden">
        <img
          src={diffUrl}
          alt="Diff only"
          className="block max-w-full"
          loading="lazy"
        />
      </div>
    </div>
  );
}
