import { formatPercentBP, formatRupiah } from '@/lib/format';

/** Live preview of the receipt header/footer while editing settings. */
export function ReceiptPreview({
  name,
  address,
  phone,
  footer,
  taxBP,
  logoUrl,
}: {
  name: string;
  address?: string;
  phone?: string;
  footer?: string;
  taxBP: number;
  logoUrl?: string | null;
}) {
  const subtotal = 43000;
  const tax = Math.round((subtotal * taxBP) / 10000);
  return (
    <div className="mx-auto w-full max-w-[280px] rounded-md bg-white p-4 font-mono text-[11px] leading-[16px] text-black shadow-card">
      {logoUrl && <img src={logoUrl} alt="" className="mx-auto mb-2 h-12 w-12 object-contain grayscale" />}
      <p className="text-center font-bold uppercase">{name || 'Nama Toko'}</p>
      {address && <p className="text-center">{address}</p>}
      {phone && <p className="text-center">Telp {phone}</p>}
      <p className="my-1 border-t border-dashed border-black/40" />
      <div className="flex justify-between"><span>Kopi Susu x2</span><span>36.000</span></div>
      <div className="flex justify-between"><span>Es Teh x1</span><span>7.000</span></div>
      <p className="my-1 border-t border-dashed border-black/40" />
      <div className="flex justify-between"><span>Subtotal</span><span>{formatRupiah(subtotal)}</span></div>
      {taxBP > 0 && <div className="flex justify-between"><span>Pajak {formatPercentBP(taxBP)}</span><span>{formatRupiah(tax)}</span></div>}
      <div className="flex justify-between font-bold"><span>TOTAL</span><span>{formatRupiah(subtotal + tax)}</span></div>
      <p className="my-1 border-t border-dashed border-black/40" />
      {footer && <p className="whitespace-pre-line text-center">{footer}</p>}
    </div>
  );
}
