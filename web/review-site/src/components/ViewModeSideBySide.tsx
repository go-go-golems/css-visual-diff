export function ViewModeSideBySide({
  leftUrl,
  rightUrl,
  leftLabel,
  rightLabel,
}: {
  leftUrl: string;
  rightUrl: string;
  leftLabel?: string;
  rightLabel?: string;
}) {
  return (
    <div className="flex gap-4">
      <div className="flex-1">
        <div className="flex items-center gap-2 px-3 py-1.5 text-xs border border-border rounded-t-[2px] -mb-px bg-cream-light text-text-muted">
          <span className="uppercase tracking-wider font-semibold">
            prototype
          </span>
          {leftLabel && (
            <span className="font-mono text-[10px] truncate opacity-80">
              {leftLabel}
            </span>
          )}
        </div>
        <div className="border border-t-0 border-border rounded-b-[2px] overflow-auto">
          <img src={leftUrl} alt="Prototype" className="block max-w-full" loading="lazy" />
        </div>
      </div>
      <div className="flex-1">
        <div className="flex items-center gap-2 px-3 py-1.5 text-xs border border-text rounded-t-[2px] -mb-px bg-text text-cream">
          <span className="uppercase tracking-wider font-semibold">react</span>
          {rightLabel && (
            <span className="font-mono text-[10px] truncate opacity-80">
              {rightLabel}
            </span>
          )}
        </div>
        <div className="border border-t-0 border-text rounded-b-[2px] overflow-auto">
          <img src={rightUrl} alt="React" className="block max-w-full" loading="lazy" />
        </div>
      </div>
    </div>
  );
}
