import { z } from 'zod';
import type { FieldValues, Path, UseFormSetError } from 'react-hook-form';
import { ApiError } from './api';

/**
 * Maps backend field errors (422) onto react-hook-form fields.
 * Returns the message to show in a toast when the error isn't field-specific.
 */
export function applyServerErrors<T extends FieldValues>(err: unknown, setError: UseFormSetError<T>, fields: string[]): string | null {
  if (err instanceof ApiError) {
    let mapped = false;
    for (const fe of err.fieldErrors) {
      if (fields.includes(fe.field)) {
        setError(fe.field as Path<T>, { type: 'server', message: fe.message });
        mapped = true;
      }
    }
    if (mapped) return null;
    const first = err.fieldErrors[0];
    return first ? `${err.message}: ${first.message}` : err.message;
  }
  return err instanceof Error ? err.message : 'Terjadi kesalahan';
}

/** Converts form input to a number; empty input becomes undefined so the "required" message is used. */
const toNumber = (v: unknown) => {
  if (v === '' || v === null || v === undefined) return undefined;
  const n = typeof v === 'number' ? v : Number(String(v).replace(',', '.'));
  return Number.isNaN(n) ? v : n;
};

/** Zod number field for <input type="number"> / <select> values, with Indonesian messages. */
export function numberField(requiredMessage = 'Wajib diisi') {
  return z.preprocess(toNumber, z.number({ required_error: requiredMessage, invalid_type_error: 'Harus berupa angka' }));
}

/** Optional number field that defaults to 0 when left empty. */
export function optionalNumberField() {
  return z.preprocess((v) => toNumber(v) ?? 0, z.number({ invalid_type_error: 'Harus berupa angka' }));
}
