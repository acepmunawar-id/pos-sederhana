import { useState, type ReactNode } from 'react';
import { Link, useParams } from 'react-router-dom';
import { ArrowLeft, Ban, Copy, FileText } from 'lucide-react';
import { PageHeader } from '@/components/layout/PageHeader';
import { Button, Card, ConfirmDialog, ErrorState, Modal, Skeleton, SkeletonCard, Textarea, useToast } from '@/components/ui';
import { errorMessage } from '@/lib/api';
import { formatDateTime, formatPercentBP, formatRupiah, paymentMethodLabel } from '@/lib/format';
import { useAuth } from '@/features/auth/hooks/useAuth';
import { TransactionStatusBadge } from '../components/StatusBadge';
import { useReceipt, useTransaction, useVoidTransaction } from '../hooks/useTransactions';

function Row({ label, value, strong }: { label: string; value: ReactNode; strong?: boolean }) {
  return (
    <div className="flex items-center justify-between py-1.5">
      <span className={strong ? 'text-headline' : 'text-subheadline text-text-secondary'}>{label}</span>
      <span className={strong ? 'text-title tabular-nums' : 'text-subheadline tabular-nums'}>{value}</span>
    </div>
  );
}

export default function TransactionDetailPage() {
  const id = Number(useParams().id);
  const toast = useToast();
  const { hasRole } = useAuth();
  const canVoid = hasRole('owner', 'admin');
  const query = useTransaction(id);
  const voidMutation = useVoidTransaction();
  const [voidOpen, setVoidOpen] = useState(false);
  const [reason, setReason] = useState('');
  const [receiptOpen, setReceiptOpen] = useState(false);
  const receipt = useReceipt(id, receiptOpen);

  const back = (
    <Link to="/transactions" className="mb-3 inline-flex h-touch items-center gap-1 text-subheadline font-medium text-primary">
      <ArrowLeft className="h-4 w-4" /> Daftar transaksi
    </Link>
  );

  if (query.isLoading) {
    return (
      <>
        {back}
        <Skeleton className="mb-6 h-10 w-72" />
        <div className="grid gap-4 lg:grid-cols-3">
          <SkeletonCard className="lg:col-span-2" lines={6} />
          <SkeletonCard lines={6} />
        </div>
      </>
    );
  }
  if (query.isError || !query.data) {
    return (
      <>
        {back}
        <ErrorState message={errorMessage(query.error)} onRetry={() => query.refetch()} />
      </>
    );
  }

  const t = query.data;
  const submitVoid = () =>
    voidMutation.mutate(
      { id: t.id, reason: reason.trim() },
      {
        onSuccess: () => {
          toast.success('Transaksi dibatalkan dan stok dikembalikan');
          setVoidOpen(false);
          setReason('');
        },
        onError: (err) => toast.error(errorMessage(err)),
      },
    );

  return (
    <>
      {back}
      <PageHeader
        title={t.invoice_number}
        description={`${formatDateTime(t.created_at)} · Kasir ${t.cashier.name}`}
        actions={
          <>
            <TransactionStatusBadge status={t.status} />
            <Button variant="secondary" icon={<FileText className="h-4 w-4" />} onClick={() => setReceiptOpen(true)}>
              Lihat Struk
            </Button>
            {canVoid && t.status === 'COMPLETED' && (
              <Button variant="danger" icon={<Ban className="h-4 w-4" />} onClick={() => setVoidOpen(true)}>
                Void Transaksi
              </Button>
            )}
          </>
        }
      />

      {t.status === 'VOIDED' && (
        <div className="mb-4 rounded-xl bg-danger/10 px-5 py-4 text-subheadline text-danger" role="status">
          Dibatalkan oleh <b>{t.voided_by?.name}</b> pada {formatDateTime(t.voided_at)}. Alasan: {t.void_reason}
        </div>
      )}

      <div className="grid gap-4 lg:grid-cols-3">
        <Card title="Item" className="lg:col-span-2" padded={false}>
          <ul className="divide-y divide-separator/50">
            {t.items?.map((it) => (
              <li key={it.id} className="flex items-center justify-between gap-4 px-5 py-3">
                <div className="min-w-0">
                  <p className="truncate font-semibold">{it.product_name}</p>
                  <p className="text-footnote text-text-secondary">
                    {it.quantity} × {formatRupiah(it.unit_price)} · {it.sku}
                  </p>
                </div>
                <span className="font-semibold tabular-nums">{formatRupiah(it.subtotal)}</span>
              </li>
            ))}
          </ul>
          {t.note && <p className="border-t border-separator/50 px-5 py-3 text-footnote text-text-secondary">Catatan: {t.note}</p>}
        </Card>

        <div className="space-y-4">
          <Card title="Ringkasan">
            <Row label={`Subtotal (${t.item_count} item)`} value={formatRupiah(t.subtotal)} />
            {t.discount_amount > 0 && (
              <Row
                label={`Diskon${t.discount_type === 'PERCENT' ? ` ${formatPercentBP(t.discount_value)}` : ''}`}
                value={<span className="text-danger">-{formatRupiah(t.discount_amount)}</span>}
              />
            )}
            {t.tax_amount > 0 && <Row label={`Pajak ${formatPercentBP(t.tax_rate_bp)}`} value={formatRupiah(t.tax_amount)} />}
            <div className="my-2 border-t border-separator/60" />
            <Row label="Total" value={formatRupiah(t.total)} strong />
            {t.gross_profit !== undefined && <Row label="Laba kotor" value={<span className="text-success">{formatRupiah(t.gross_profit)}</span>} />}
          </Card>
          <Card title="Pembayaran">
            {t.payments.map((p) => (
              <Row
                key={p.method}
                label={`${paymentMethodLabel[p.method]}${p.reference_number ? ` · ${p.reference_number}` : ''}`}
                value={formatRupiah(p.tendered_amount)}
              />
            ))}
            <Row label="Kembalian" value={<span className="font-semibold">{formatRupiah(t.change_amount)}</span>} />
          </Card>
        </div>
      </div>

      <ConfirmDialog
        open={voidOpen}
        title="Void transaksi ini?"
        message={<>Transaksi <b>{t.invoice_number}</b> senilai <b>{formatRupiah(t.total)}</b> akan dibatalkan dan stok semua item dikembalikan. Tindakan ini tidak bisa diurungkan.</>}
        confirmLabel="Void Transaksi"
        loading={voidMutation.isPending}
        onClose={() => setVoidOpen(false)}
        onConfirm={() => {
          if (reason.trim().length < 3) {
            toast.error('Tulis alasan pembatalan (minimal 3 karakter)');
            return;
          }
          submitVoid();
        }}
      >
        <Textarea label="Alasan pembatalan" required placeholder="Contoh: pelanggan batal membeli" value={reason} onChange={(e) => setReason(e.target.value)} maxLength={500} />
      </ConfirmDialog>

      <Modal
        open={receiptOpen}
        onClose={() => setReceiptOpen(false)}
        title="Struk"
        size="sm"
        footer={
          <Button
            variant="secondary"
            icon={<Copy className="h-4 w-4" />}
            disabled={!receipt.data}
            onClick={async () => {
              await navigator.clipboard.writeText(receipt.data!.text);
              toast.success('Teks struk disalin');
            }}
          >
            Salin Teks
          </Button>
        }
      >
        {receipt.isLoading ? (
          <Skeleton className="h-80 w-full" />
        ) : receipt.isError ? (
          <ErrorState message={errorMessage(receipt.error)} onRetry={() => receipt.refetch()} />
        ) : (
          <pre className="overflow-x-auto rounded-md bg-fill p-4 font-mono text-[12px] leading-[18px]">{receipt.data?.text}</pre>
        )}
      </Modal>
    </>
  );
}
