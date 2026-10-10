import 'package:flutter/cupertino.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inset_group.dart';
import '../../cashier/widgets/summary_rows.dart';
import '../../receipt/receipt_service.dart';
import '../data/transaction_repository.dart';
import '../providers/transaction_providers.dart';

class TransactionDetailScreen extends ConsumerStatefulWidget {
  const TransactionDetailScreen({super.key, required this.id});
  final int id;

  @override
  ConsumerState<TransactionDetailScreen> createState() => _TransactionDetailScreenState();
}

class _TransactionDetailScreenState extends ConsumerState<TransactionDetailScreen> {
  bool _busy = false;

  Future<void> _receipt(Future<void> Function(String text, String invoice) action, String invoice) async {
    setState(() => _busy = true);
    try {
      final text = await ref.read(transactionRepositoryProvider).receiptText(widget.id);
      await action(text, invoice);
    } catch (e) {
      if (mounted) showToast(context, errorMessage(e), type: ToastType.error);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final tx = ref.watch(transactionDetailProvider(widget.id));
    const receipts = ReceiptService();

    return AppPage(
      title: 'Detail Transaksi',
      child: tx.when(
        loading: () => const Center(child: CupertinoActivityIndicator()),
        error: (e, _) => ErrorState(message: errorMessage(e), onRetry: () => ref.invalidate(transactionDetailProvider(widget.id))),
        data: (t) => ListView(
          padding: const EdgeInsets.only(bottom: AppSpacing.xxl),
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.xl, AppSpacing.lg, 0),
              child: Column(children: [
                Text(t.invoiceNumber, style: AppText.headline.copyWith(color: c.textSecondary)),
                const SizedBox(height: 4),
                FittedBox(child: Text(formatRupiah(t.total), style: AppText.amount.copyWith(color: t.isVoided ? c.gray : c.text))),
                const SizedBox(height: 4),
                Text(formatDayHeader(t.createdAt), style: AppText.footnote.copyWith(color: c.textSecondary)),
                if (t.isVoided) ...[
                  const SizedBox(height: AppSpacing.md),
                  Container(
                    padding: const EdgeInsets.all(AppSpacing.md),
                    decoration: BoxDecoration(color: c.danger.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(AppRadius.sm)),
                    child: Text('Transaksi dibatalkan${t.voidReason != null ? ': ${t.voidReason}' : ''}',
                        textAlign: TextAlign.center, style: AppText.subheadline.copyWith(color: c.danger)),
                  ),
                ],
              ]),
            ),
            InsetSection(header: 'Barang (${t.itemCount})', children: [
              for (final it in t.items)
                InsetRow(title: it.productName, subtitle: '${it.quantity} × ${formatRupiah(it.unitPrice)}', value: formatRupiah(it.subtotal)),
            ]),
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, 0),
              child: Container(
                padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
                decoration: BoxDecoration(color: c.card, borderRadius: BorderRadius.circular(AppRadius.sm)),
                child: Column(children: [
                  TotalsBreakdown(subtotal: t.subtotal, discount: t.discountAmount, taxRateBP: t.taxRateBP, tax: t.taxAmount),
                  SummaryRow(label: 'Total', value: formatRupiah(t.total), bold: true),
                ]),
              ),
            ),
            InsetSection(header: 'Pembayaran', children: [
              for (final p in t.payments) InsetRow(title: p.label, subtitle: p.referenceNumber, value: formatRupiah(p.tenderedAmount)),
              InsetRow(title: 'Kembalian', value: formatRupiah(t.changeAmount)),
              InsetRow(title: 'Kasir', value: t.cashierName),
            ]),
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.xl, AppSpacing.lg, 0),
              child: Row(children: [
                Expanded(
                  child: AppButton(
                    label: 'Bagikan Struk',
                    icon: CupertinoIcons.share,
                    variant: AppButtonVariant.secondary,
                    loading: _busy,
                    onPressed: () => _receipt(receipts.share, t.invoiceNumber),
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: AppButton(
                    label: 'Cetak',
                    icon: CupertinoIcons.printer,
                    variant: AppButtonVariant.secondary,
                    loading: _busy,
                    onPressed: () => _receipt(receipts.print, t.invoiceNumber),
                  ),
                ),
              ]),
            ),
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.xl, AppSpacing.md, AppSpacing.xl, 0),
              child: Text(
                'Pembatalan (void) hanya dapat dilakukan oleh Owner/Admin melalui dashboard.',
                textAlign: TextAlign.center,
                style: AppText.footnote.copyWith(color: c.textSecondary),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
