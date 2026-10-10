import { Bar, BarChart as RBarChart, CartesianGrid, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { useThemeColors } from '@/lib/use-theme-colors';

export interface BarDatum {
  label: string;
  value: number;
  /** Extra line shown in the tooltip (e.g. transaction count). */
  detail?: string;
}

/** Compact axis labels: 1.250.000 → "1,3 jt". */
function compact(v: number): string {
  if (Math.abs(v) >= 1_000_000_000) return `${(v / 1_000_000_000).toLocaleString('id-ID', { maximumFractionDigits: 1 })} M`;
  if (Math.abs(v) >= 1_000_000) return `${(v / 1_000_000).toLocaleString('id-ID', { maximumFractionDigits: 1 })} jt`;
  if (Math.abs(v) >= 1_000) return `${(v / 1_000).toLocaleString('id-ID', { maximumFractionDigits: 0 })} rb`;
  return String(v);
}

/** Single-series vertical bar chart in the primary token colour, with hover tooltip. */
export function BarChart({
  data,
  formatValue,
  height = 260,
  ariaLabel,
}: {
  data: BarDatum[];
  formatValue: (v: number) => string;
  height?: number;
  ariaLabel: string;
}) {
  const c = useThemeColors();
  return (
    <div style={{ height }} role="img" aria-label={ariaLabel}>
      <ResponsiveContainer width="100%" height="100%">
        <RBarChart data={data} margin={{ top: 8, right: 8, bottom: 0, left: 0 }} barCategoryGap="28%">
          <CartesianGrid vertical={false} stroke={c.separator} strokeOpacity={0.5} />
          <XAxis dataKey="label" tickLine={false} axisLine={false} tick={{ fill: c.textSecondary, fontSize: 12 }} interval="preserveStartEnd" />
          <YAxis
            tickLine={false}
            axisLine={false}
            width={52}
            tick={{ fill: c.textSecondary, fontSize: 12 }}
            tickFormatter={compact}
          />
          <Tooltip
            cursor={{ fill: c.fill }}
            content={({ active, payload }) => {
              if (!active || !payload?.length) return null;
              const d = payload[0].payload as BarDatum;
              return (
                <div className="rounded-md bg-card px-3 py-2 shadow-modal">
                  <p className="text-caption text-text-secondary">{d.label}</p>
                  <p className="text-subheadline font-semibold tabular-nums text-text">{formatValue(d.value)}</p>
                  {d.detail && <p className="text-caption text-text-secondary">{d.detail}</p>}
                </div>
              );
            }}
          />
          <Bar dataKey="value" fill={c.primary} radius={[4, 4, 0, 0]} maxBarSize={44} isAnimationActive={false} />
        </RBarChart>
      </ResponsiveContainer>
    </div>
  );
}
