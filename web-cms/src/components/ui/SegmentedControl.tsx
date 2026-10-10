import { cn } from '@/lib/cn';

export interface Segment<V extends string> {
  value: V;
  label: string;
}

/** iOS-style segmented control, used for tabs and quick filters. */
export function SegmentedControl<V extends string>({
  segments,
  value,
  onChange,
  className,
}: {
  segments: Segment<V>[];
  value: V;
  onChange: (v: V) => void;
  className?: string;
}) {
  return (
    <div role="tablist" className={cn('inline-flex max-w-full overflow-x-auto rounded-md bg-fill p-0.5 [scrollbar-width:none]', className)}>
      {segments.map((s) => (
        <button
          key={s.value}
          type="button"
          role="tab"
          aria-selected={s.value === value}
          onClick={() => onChange(s.value)}
          className={cn(
            'min-h-8 shrink-0 whitespace-nowrap rounded-[8px] px-3.5 py-1.5 text-footnote font-semibold transition',
            s.value === value ? 'bg-card text-text shadow-card' : 'text-text-secondary hover:text-text',
          )}
        >
          {s.label}
        </button>
      ))}
    </div>
  );
}
