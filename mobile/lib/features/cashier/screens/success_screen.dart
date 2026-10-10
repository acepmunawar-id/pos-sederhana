import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/layout/breakpoints.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/feedback.dart';
import '../../receipt/receipt_service.dart';
import '../../transactions/data/transaction_repository.dart';
import '../../transactions/models/transaction.dart';

/// Shown after a successful checkout.
class SuccessScreen extends ConsumerStatefulWidget {
  const SuccessScreen({super.key, required this.transaction});
  final Transaction transaction;

  @override
  ConsumerState<SuccessScreen> createState() => _SuccessScreenState();
}

class _SuccessScreenState extends ConsumerState<SuccessScreen> {
  bool _busy = false;

  Future<void> _withReceipt(Future<void> Function(String text) action) async {
    setState(() => _busy = true);
    try {
      final text = await ref.read(transactionRepositoryProvider).receiptText(widget.transaction.id);
      await action(text);
    } catch (e) {
      if (mounted) showToast(context, errorMessage(e), type: ToastType.error);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final t = widget.transaction;
    const receipts = ReceiptService();
    return PopScope(
      canPop: false,
      child: Scaffold(
        backgroundColor: c.background,
        body: SafeArea(
          child: MaxWidth(
            maxWidth: Breakpoints.form,
            child: ListView(
            padding: const EdgeInsets.all(AppSpacing.xl),
            children: [
              const SizedBox(height: AppSpacing.xl),
              TweenAnimationBuilder<double>(
                tween: Tween(begin: 0.6, end: 1),
                duration: const Duration(milliseconds: 380),
                curve: Curves.elasticOut,
                builder: (_, s, child) => Transform.scale(scale: s, child: child),
                child: Icon(CupertinoIcons.checkmark_circle_fill, size: 96, color: c.success),
              ),
              const SizedBox(height: AppSpacing.lg),
              Text('Transaksi Berhasil', textAlign: TextAlign.center, style: AppText.title.copyWith(color: c.text)),
              const SizedBox(height: 4),
              Text(t.invoiceNumber, textAlign: TextAlign.center, style: AppText.subheadline.copyWith(color: c.textSecondary)),
              const SizedBox(height: AppSpacing.xl),
              Container(
                padding: const EdgeInsets.all(AppSpacing.xl),
                decoration: BoxDecoration(color: c.card, borderRadius: BorderRadius.circular(AppRadius.lg)),
                child: Column(children: [
                  _Big(label: 'Total', value: formatRupiah(t.total), color: c.text),
                  Padding(padding: const EdgeInsets.symmetric(vertical: AppSpacing.lg), child: Container(height: 0.5, color: c.separator)),
                  _Big(label: 'Kembalian', value: formatRupiah(t.changeAmount), color: c.success),
                  const SizedBox(height: AppSpacing.md),
                  Text('Dibayar ${t.paymentLabel} ${formatRupiah(t.paidAmount)}', style: AppText.subheadline.copyWith(color: c.textSecondary)),
                ]),
              ),
              const SizedBox(height: AppSpacing.xl),
              Row(children: [
                Expanded(
                  child: AppButton(
                    label: 'Bagikan',
                    icon: CupertinoIcons.share,
                    variant: AppButtonVariant.secondary,
                    loading: _busy,
                    onPressed: () => _withReceipt((text) => receipts.share(text, t.invoiceNumber)),
                  ),
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: AppButton(
                    label: 'Cetak',
                    icon: CupertinoIcons.printer,
                    variant: AppButtonVariant.secondary,
                    loading: _busy,
                    onPressed: () => _withReceipt((text) => receipts.print(text, t.invoiceNumber)),
                  ),
                ),
              ]),
              const SizedBox(height: AppSpacing.md),
              AppButton(
                label: 'Transaksi Baru',
                icon: CupertinoIcons.plus_circle_fill,
                height: 56,
                onPressed: () {
                  showToast(context, 'Siap untuk transaksi berikutnya', type: ToastType.info);
                  context.go('/kasir');
                },
              ),
            ],
          ),
          ),
        ),
      ),
    );
  }
}

class _Big extends StatelessWidget {
  const _Big({required this.label, required this.value, required this.color});
  final String label;
  final String value;
  final Color color;

  @override
  Widget build(BuildContext context) => Column(children: [
        Text(label, style: AppText.headline.copyWith(color: context.colors.textSecondary)),
        FittedBox(child: Text(value, style: AppText.amount.copyWith(color: color))),
      ]);
}
