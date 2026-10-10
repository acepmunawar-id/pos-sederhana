import { cn } from '@/lib/cn';

/** iOS-style toggle with a visible text label. */
export function Switch({
  checked,
  onChange,
  label,
  description,
}: {
  checked: boolean;
  onChange: (v: boolean) => void;
  label: string;
  description?: string;
}) {
  return (
    <label className="flex min-h-touch cursor-pointer items-center justify-between gap-4">
      <span>
        <span className="block text-subheadline font-medium">{label}</span>
        {description && <span className="block text-caption text-text-secondary">{description}</span>}
      </span>
      <button
        type="button"
        role="switch"
        aria-checked={checked}
        onClick={() => onChange(!checked)}
        className={cn(
          'relative h-[31px] w-[51px] shrink-0 rounded-full transition-colors',
          checked ? 'bg-success' : 'bg-fill',
        )}
      >
        <span
          className={cn(
            'absolute left-0 top-[2px] h-[27px] w-[27px] rounded-full bg-white shadow transition-transform',
            checked ? 'translate-x-[22px]' : 'translate-x-[2px]',
          )}
        />
      </button>
    </label>
  );
}
