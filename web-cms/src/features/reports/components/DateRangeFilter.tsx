import { Input, SegmentedControl } from '@/components/ui';
import { daysAgo, toISODate } from '@/lib/format';
import type { DateRange } from '../types';

type Preset = 'today' | '7d' | '30d' | 'month' | 'custom';

export function presetRange(p: Exclude<Preset, 'custom'>): DateRange {
  const today = toISODate(new Date());
  switch (p) {
    case 'today':
      return { date_from: today, date_to: today };
    case '7d':
      return { date_from: daysAgo(6), date_to: today };
    case '30d':
      return { date_from: daysAgo(29), date_to: today };
    case 'month': {
      const d = new Date();
      return { date_from: toISODate(new Date(d.getFullYear(), d.getMonth(), 1)), date_to: today };
    }
  }
}

function detectPreset(r: DateRange): Preset {
  for (const p of ['today', '7d', '30d', 'month'] as const) {
    const pr = presetRange(p);
    if (pr.date_from === r.date_from && pr.date_to === r.date_to) return p;
  }
  return 'custom';
}

export function DateRangeFilter({ value, onChange }: { value: DateRange; onChange: (r: DateRange) => void }) {
  const preset = detectPreset(value);
  return (
    <div className="flex flex-wrap items-center gap-3">
      <SegmentedControl<Preset>
        value={preset}
        onChange={(p) => p !== 'custom' && onChange(presetRange(p))}
        segments={[
          { value: 'today', label: 'Hari ini' },
          { value: '7d', label: '7 hari' },
          { value: '30d', label: '30 hari' },
          { value: 'month', label: 'Bulan ini' },
          { value: 'custom', label: 'Pilih tanggal' },
        ]}
      />
      <div className="flex items-center gap-2">
        <Input
          type="date"
          aria-label="Dari tanggal"
          className="w-40"
          value={value.date_from}
          max={value.date_to}
          onChange={(e) => e.target.value && onChange({ ...value, date_from: e.target.value })}
        />
        <span className="text-text-secondary">–</span>
        <Input
          type="date"
          aria-label="Sampai tanggal"
          className="w-40"
          value={value.date_to}
          min={value.date_from}
          onChange={(e) => e.target.value && onChange({ ...value, date_to: e.target.value })}
        />
      </div>
    </div>
  );
}
