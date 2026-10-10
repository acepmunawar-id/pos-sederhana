const rupiah = new Intl.NumberFormat('id-ID', { maximumFractionDigits: 0 });

/** 25000 → "Rp 25.000" (same format as the mobile app and receipts). */
export function formatRupiah(value: number | null | undefined): string {
  const v = value ?? 0;
  return `${v < 0 ? '-' : ''}Rp ${rupiah.format(Math.abs(v))}`;
}

export function formatNumber(value: number | null | undefined): string {
  return rupiah.format(value ?? 0);
}

/** Basis points → "11%" / "12,5%". */
export function formatPercentBP(bp: number): string {
  return `${(bp / 100).toLocaleString('id-ID', { maximumFractionDigits: 2 })}%`;
}

/** "12,5" (percent typed by a user) → 1250 basis points. */
export function percentToBP(percent: number): number {
  return Math.round(percent * 100);
}

const dateFmt = new Intl.DateTimeFormat('id-ID', { day: '2-digit', month: 'short', year: 'numeric' });
const dateTimeFmt = new Intl.DateTimeFormat('id-ID', {
  day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
});
const shortDayFmt = new Intl.DateTimeFormat('id-ID', { weekday: 'short', day: 'numeric' });
const monthFmt = new Intl.DateTimeFormat('id-ID', { month: 'long', year: 'numeric' });

/** Parses "YYYY-MM-DD" as a local date (not UTC). */
function parseDay(value: string): Date {
  const [y, m, d] = value.split('-').map(Number);
  return new Date(y, (m ?? 1) - 1, d ?? 1);
}

export function formatDate(value: string | null | undefined): string {
  if (!value) return '-';
  return dateFmt.format(value.length === 10 ? parseDay(value) : new Date(value));
}

export function formatDateTime(value: string | null | undefined): string {
  if (!value) return '-';
  return dateTimeFmt.format(new Date(value));
}

export function formatShortDay(day: string): string {
  return shortDayFmt.format(parseDay(day));
}

/** "2026-10" → "Oktober 2026", "2026-10-10" → "10 Okt 2026". */
export function formatPeriod(period: string): string {
  return period.length === 7 ? monthFmt.format(parseDay(`${period}-01`)) : formatDate(period);
}

/** Date → "YYYY-MM-DD" in local time. */
export function toISODate(d: Date): string {
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
}

export function daysAgo(n: number): string {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return toISODate(d);
}

export const paymentMethodLabel: Record<string, string> = {
  CASH: 'Tunai',
  QRIS: 'QRIS',
  TRANSFER: 'Transfer',
  DEBIT: 'Debit',
};

export const roleLabel: Record<string, string> = {
  owner: 'Owner',
  admin: 'Admin',
  cashier: 'Kasir',
};
