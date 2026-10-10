/** Horizontal share bars (single hue) — readable alternative to a pie chart. */
export function ProportionBars({
  rows,
  formatValue,
}: {
  rows: { label: string; value: number; detail?: string }[];
  formatValue: (v: number) => string;
}) {
  const max = Math.max(1, ...rows.map((r) => r.value));
  const total = rows.reduce((s, r) => s + r.value, 0);
  return (
    <ul className="space-y-3">
      {rows.map((r) => (
        <li key={r.label} title={`${r.label}: ${formatValue(r.value)}`}>
          <div className="mb-1 flex items-baseline justify-between gap-3">
            <span className="text-subheadline font-medium">{r.label}</span>
            <span className="text-subheadline tabular-nums">
              {formatValue(r.value)}
              <span className="ml-2 text-footnote text-text-secondary">
                {total > 0 ? Math.round((r.value / total) * 100) : 0}%
              </span>
            </span>
          </div>
          <div className="h-2 rounded-full bg-fill">
            <div className="h-2 rounded-full bg-primary" style={{ width: `${(r.value / max) * 100}%` }} />
          </div>
          {r.detail && <p className="mt-1 text-caption text-text-secondary">{r.detail}</p>}
        </li>
      ))}
    </ul>
  );
}
