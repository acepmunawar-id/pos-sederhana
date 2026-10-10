import { Pencil, Trash2 } from 'lucide-react';

/** Edit / delete buttons for table rows (icon + label, 44pt touch targets). */
export function RowActions({ onEdit, onDelete }: { onEdit?: () => void; onDelete?: () => void }) {
  return (
    <div className="flex justify-end gap-1" onClick={(e) => e.stopPropagation()}>
      {onEdit && (
        <button
          type="button"
          onClick={onEdit}
          className="inline-flex h-touch items-center gap-1.5 rounded-md px-2.5 text-footnote font-semibold text-primary hover:bg-fill"
        >
          <Pencil className="h-4 w-4" /> Ubah
        </button>
      )}
      {onDelete && (
        <button
          type="button"
          onClick={onDelete}
          className="inline-flex h-touch items-center gap-1.5 rounded-md px-2.5 text-footnote font-semibold text-danger hover:bg-fill"
        >
          <Trash2 className="h-4 w-4" /> Hapus
        </button>
      )}
    </div>
  );
}
