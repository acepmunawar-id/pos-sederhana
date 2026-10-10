import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/layout/breakpoints.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/utils/haptics.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inputs.dart';
import '../../cart/providers/cart_controller.dart';
import '../../catalog/providers/catalog_providers.dart';
import '../../shift/providers/shift_controller.dart';
import '../../transactions/data/transaction_repository.dart';
import '../../transactions/providers/transaction_providers.dart';

enum PayMethod { cash, qris, transfer, debit }

extension on PayMethod {
  String get api => switch (this) { PayMethod.cash => 'CASH', PayMethod.qris => 'QRIS', PayMethod.transfer => 'TRANSFER', PayMethod.debit => 'DEBIT' };
  String get label => switch (this) { PayMethod.cash => 'Tunai', PayMethod.qris => 'QRIS', PayMethod.transfer => 'Transfer', PayMethod.debit => 'Debit' };
}

/// Quick-cash buttons: exact amount, up to two nearby round amounts, and the 50k / 100k notes.
List<int> quickCashOptions(int total) {
  final rounded = <int>{};
  for (final step in [5000, 10000]) {
    final r = ((total + step - 1) ~/ step) * step;
    if (r > total && r < 50000) rounded.add(r);
  }
  final notes = [50000, 100000].where((n) => n >= total);
  final all = {total, ...rounded.take(2), ...notes}.toList()..sort();
  return all;
}

/// Step 3 of a sale: choose payment and finish.
class PaymentScreen extends ConsumerStatefulWidget {
  const PaymentScreen({super.key});

  @override
  ConsumerState<PaymentScreen> createState() => _PaymentScreenState();
}

class _PaymentScreenState extends ConsumerState<PaymentScreen> {
  PayMethod _method = PayMethod.cash;
  final _cash = TextEditingController();
  final _reference = TextEditingController();
  bool _submitting = false;

  @override
  void dispose() {
    _cash.dispose();
    _reference.dispose();
    super.dispose();
  }

  void _setCash(int amount) {
    Haptics.tap();
    _cash.text = formatNumber(amount);
    setState(() {});
  }

  Future<void> _pay(int total) async {
    final cart = ref.read(cartControllerProvider);
    final tendered = _method == PayMethod.cash ? parseRupiah(_cash.text) : total;
    if (_method == PayMethod.cash && tendered < total) {
      showToast(context, 'Uang kurang ${formatRupiah(total - tendered)}', type: ToastType.error);
      return;
    }
    setState(() => _submitting = true);
    try {
      final key = ref.read(cartControllerProvider.notifier).checkoutKey();
      final tx = await ref.read(transactionRepositoryProvider).checkout(
            idempotencyKey: key,
            cart: cart,
            payments: [
              CheckoutPayment(
                method: _method.api,
                amount: tendered,
                referenceNumber: _method == PayMethod.cash ? null : _reference.text.trim(),
              ),
            ],
          );
      ref.read(cartControllerProvider.notifier).clear();
      ref.invalidate(productListProvider);
      ref.invalidate(transactionListProvider);
      ref.read(currentShiftProvider.notifier).refresh();
      if (!mounted) return;
      Haptics.success();
      context.go('/kasir/selesai', extra: tx);
    } catch (e) {
      if (!mounted) return;
      showToast(context, errorMessage(e), type: ToastType.error);
      // Stock may have changed under us; refresh so the cart limits are accurate.
      if (e is ApiException && e.status == 422) ref.invalidate(productListProvider);
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final totals = ref.watch(cartTotalsProvider);
    final total = totals.total;
    final tendered = parseRupiah(_cash.text);
    final change = tendered - total;
    final isCash = _method == PayMethod.cash;
    final canPay = !isCash || tendered >= total;

    return AppPage(
      title: 'Pembayaran',
      bottom: SafeArea(
        top: false,
        child: MaxWidth(
          maxWidth: Breakpoints.form,
          child: Padding(
          padding: const EdgeInsets.all(AppSpacing.lg),
          child: AppButton(
            label: 'Selesaikan Pembayaran',
            icon: CupertinoIcons.checkmark_seal_fill,
            variant: AppButtonVariant.success,
            height: 56,
            loading: _submitting,
            onPressed: canPay ? () => _pay(total) : null,
          ),
        ),
        ),
      ),
      child: ListView(
        padding: const EdgeInsets.all(AppSpacing.lg),
        children: [
          Container(
            padding: const EdgeInsets.symmetric(vertical: AppSpacing.xl, horizontal: AppSpacing.lg),
            decoration: BoxDecoration(color: c.card, borderRadius: BorderRadius.circular(AppRadius.lg)),
            child: Column(children: [
              Text('Total Belanja', style: AppText.subheadline.copyWith(color: c.textSecondary)),
              const SizedBox(height: 4),
              FittedBox(child: Text(formatRupiah(total), style: AppText.amount.copyWith(color: c.text))),
            ]),
          ),
          const SizedBox(height: AppSpacing.xl),
          const _Label('Metode bayar'),
          AppSegmented<PayMethod>(
            value: _method,
            segments: {for (final m in PayMethod.values) m: m.label},
            onChanged: (m) => setState(() => _method = m),
          ),
          const SizedBox(height: AppSpacing.xl),
          if (isCash) ...[
            AppTextField(
              label: 'Uang diterima',
              controller: _cash,
              prefix: 'Rp',
              placeholder: '0',
              large: true,
              keyboardType: TextInputType.number,
              inputFormatters: [ThousandsFormatter()],
              onChanged: (_) => setState(() {}),
            ),
            const SizedBox(height: AppSpacing.md),
            Wrap(
              spacing: AppSpacing.sm,
              runSpacing: AppSpacing.sm,
              children: [
                for (final (i, amount) in quickCashOptions(total).indexed)
                  _QuickCash(label: i == 0 ? 'Uang Pas' : formatNumber(amount), selected: tendered == amount, onTap: () => _setCash(amount)),
              ],
            ),
            const SizedBox(height: AppSpacing.xl),
            Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(
                color: tendered == 0 ? c.fill : (change >= 0 ? c.success : c.danger).withValues(alpha: 0.12),
                borderRadius: BorderRadius.circular(AppRadius.lg),
              ),
              child: Column(children: [
                Text(
                  tendered == 0 ? 'Kembalian' : change >= 0 ? 'Kembalian' : 'Uang kurang',
                  style: AppText.headline.copyWith(color: tendered == 0 ? c.textSecondary : change >= 0 ? c.success : c.danger),
                ),
                const SizedBox(height: 4),
                FittedBox(
                  child: Text(
                    formatRupiah(tendered == 0 ? 0 : change.abs()),
                    style: AppText.amount.copyWith(color: tendered == 0 ? c.textSecondary : change >= 0 ? c.success : c.danger),
                  ),
                ),
              ]),
            ),
          ] else ...[
            Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(color: c.primary.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(AppRadius.lg)),
              child: Row(children: [
                Icon(CupertinoIcons.info_circle_fill, color: c.primary),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Text(
                    'Pastikan pembayaran ${_method.label} sebesar ${formatRupiah(total)} sudah diterima sebelum menekan tombol di bawah.',
                    style: AppText.subheadline.copyWith(color: c.text),
                  ),
                ),
              ]),
            ),
            const SizedBox(height: AppSpacing.lg),
            AppTextField(label: 'No. referensi (opsional)', controller: _reference, placeholder: 'Contoh: 4 digit akhir / kode transaksi'),
          ],
        ],
      ),
    );
  }
}

class _Label extends StatelessWidget {
  const _Label(this.text);
  final String text;

  @override
  Widget build(BuildContext context) => Padding(
        padding: const EdgeInsets.only(left: 4, bottom: 6),
        child: Text(text, style: AppText.footnote.copyWith(color: context.colors.textSecondary, fontWeight: FontWeight.w500)),
      );
}

class _QuickCash extends StatelessWidget {
  const _QuickCash({required this.label, required this.selected, required this.onTap});
  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Material(
      color: selected ? c.primary : c.card,
      borderRadius: BorderRadius.circular(AppRadius.md),
      child: InkWell(
        borderRadius: BorderRadius.circular(AppRadius.md),
        onTap: onTap,
        child: ConstrainedBox(
          constraints: const BoxConstraints(minHeight: 48, minWidth: 96),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
            child: Center(widthFactor: 1, heightFactor: 1, child: Text(label, style: AppText.headline.copyWith(color: selected ? c.onAccent : c.primary))),
          ),
        ),
      ),
    );
  }
}
