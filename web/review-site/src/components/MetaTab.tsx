import type { SummaryRow, CompareData } from "../types";
import { compareBounds, compareSourceUrls } from "../utils/compareData";

export function MetaTab({
  row,
  compareData,
}: {
  row: SummaryRow;
  compareData: CompareData | null;
}) {
  const bounds = compareBounds(compareData);
  const sources = compareSourceUrls(compareData);

  return (
    <div className="p-4 space-y-4 text-xs">
      {/* Bounds */}
      <Section title="Bounds">
        {bounds ? (
          <>
            <KV
              k="prototype"
              v={`${bounds.left.width}×${Math.round(bounds.left.height)}`}
            />
            <KV
              k="react"
              v={`${bounds.right.width}×${Math.round(bounds.right.height)}`}
            />
            <KV
              k="delta"
              v={`+${Math.round(bounds.delta.height)}px tall`}
              highlight
            />
          </>
        ) : (
          <KV k="—" v="no compare.json loaded" />
        )}
      </Section>

      {/* Pixel diff */}
      <Section title="Pixel diff">
        <KV
          k="changed"
          v={row.changedPixels.toLocaleString()}
        />
        <KV
          k="changed%"
          v={`${row.changedPercent.toFixed(2)}%`}
        />
        {row.totalPixels && (
          <KV k="total" v={row.totalPixels.toLocaleString()} />
        )}
      </Section>

      {/* Selectors */}
      <Section title="Selectors">
        <div className="font-mono text-[11px] text-text">
          <div>
            <span className="text-text-faint">L: </span>
            {row.leftSelector}
          </div>
          <div>
            <span className="text-text-faint">R: </span>
            {row.rightSelector}
          </div>
        </div>
      </Section>

      {/* Sources */}
      {sources && (
        <Section title="Sources">
          <div className="space-y-1">
            <div>
              <div className="text-[10px] uppercase tracking-wider text-text-muted mb-0.5">
                prototype
              </div>
              <a
                href={sources.left}
                className="font-mono text-[11px] text-text break-all hover:underline"
              >
                {sources.left}
              </a>
            </div>
            <div className="pt-2">
              <div className="text-[10px] uppercase tracking-wider text-text-muted mb-0.5">
                react
              </div>
              <a
                href={sources.right}
                className="font-mono text-[11px] text-text break-all hover:underline"
              >
                {sources.right}
              </a>
            </div>
          </div>
        </Section>
      )}
    </div>
  );
}

function Section({
  title,
  children,
}: {
  title: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <div className="text-[10px] uppercase tracking-wider font-semibold text-text-muted mb-1.5">
        {title}
      </div>
      <div className="space-y-1">{children}</div>
    </div>
  );
}

function KV({
  k,
  v,
  highlight,
}: {
  k: string;
  v: string;
  highlight?: boolean;
}) {
  return (
    <div className="flex items-baseline justify-between gap-3">
      <span className="text-text-muted font-mono text-[11px]">{k}</span>
      <span
        className={`font-mono text-[11px] ${
          highlight ? "text-red font-semibold" : "text-text"
        }`}
      >
        {v}
      </span>
    </div>
  );
}
