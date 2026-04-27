import { useState } from "react";
import type { SummaryRow, CompareData } from "../types";
import { ChevronDown, ChevronRight } from "lucide-react";

export function StylesTab({
  compareData,
}: {
  row: SummaryRow;
  compareData: CompareData | null;
}) {
  const [expanded, setExpanded] = useState(true);
  const styles = compareData?.styles.filter((s) => s.changed) ?? [];
  const attrs = compareData?.attributes.filter((a) => a.changed) ?? [];

  return (
    <div>
      {/* CSS styles */}
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full px-4 py-2.5 text-left text-xs uppercase tracking-wider font-semibold text-text-muted flex items-center gap-1.5 border-b border-border hover:bg-cream-light"
      >
        {expanded ? <ChevronDown size={12} /> : <ChevronRight size={12} />}
        computed style diffs ({styles.length})
      </button>
      {expanded && (
        <div className="divide-y divide-cream-mid">
          {styles.map((s) => (
            <div key={s.name} className="px-4 py-2.5">
              <div className="font-mono text-[11px] font-semibold text-text mb-1">
                {s.name}
              </div>
              <div className="space-y-0.5">
                <div className="flex items-start gap-2 font-mono text-[11px]">
                  <span className="text-text-faint shrink-0">L</span>
                  <span className="text-text-muted break-all">{s.left}</span>
                </div>
                <div className="flex items-start gap-2 font-mono text-[11px]">
                  <span className="text-red shrink-0">R</span>
                  <span className="text-text break-all">{s.right}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Attributes */}
      {attrs.length > 0 && (
        <>
          <div className="px-4 py-2.5 text-xs uppercase tracking-wider font-semibold text-text-muted border-t border-b border-border">
            attribute diffs ({attrs.length})
          </div>
          {attrs.map((a) => (
            <div key={a.name} className="px-4 py-2.5">
              <div className="font-mono text-[11px] font-semibold mb-1">
                {a.name}
              </div>
              <div className="font-mono text-[11px] text-text break-all">
                <span className="text-text-faint">+</span> {a.right}
              </div>
            </div>
          ))}
        </>
      )}
    </div>
  );
}
