import type { SummaryRow, CompareData } from "../types";

export function MetaTab({
  row,
  compareData,
}: {
  row: SummaryRow;
  compareData: CompareData | null;
}) {
  return (
    <div className="p-4 space-y-4 text-xs">
      {/* Bounds */}
      <Section title="Bounds">
        {compareData ? (
          <>
            <KV
              k="prototype"
              v={`${compareData.bounds.left.width}×${Math.round(compareData.bounds.left.height)}`}
            />
            <KV
              k="react"
              v={`${compareData.bounds.right.width}×${Math.round(compareData.bounds.right.height)}`}
            />
            <KV
              k="delta"
              v={`+${Math.round(compareData.bounds.delta.height)}px tall`}
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
      {compareData && (
        <Section title="Sources">
          <div className="space-y-1">
            <div>
              <div className="text-[10px] uppercase tracking-wider text-text-muted mb-0.5">
                prototype
              </div>
              <a
                href={compareData.left.url}
                className="font-mono text-[11px] text-text break-all hover:underline"
              >
                {compareData.left.url}
              </a>
            </div>
            <div className="pt-2">
              <div className="text-[10px] uppercase tracking-wider text-text-muted mb-0.5">
                react
              </div>
              <a
                href={compareData.right.url}
                className="font-mono text-[11px] text-text break-all hover:underline"
              >
                {compareData.right.url}
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
