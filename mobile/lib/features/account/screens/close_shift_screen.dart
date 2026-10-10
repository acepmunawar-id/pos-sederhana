import 'package:flutter/cupertino.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inputs.dart';
import '../../../core/widgets/inset_group.dart';
import '../../cart/providers/cart_controller.dart';
import '../../shift/models/shift.dart';
import '../../shift/providers/shift_controller.dart';
import '../../transactions/models/transaction.dart';

/// Close the shift: count the drawer, see the difference, then confirm.
class CloseShiftScreen extends ConsumerStatefulWidget {
  const CloseShiftScreen({super.key});

  @override
  ConsumerState<CloseShiftScreen> createState() => _CloseShiftScreenState();
}

class _CloseShiftScreenState extends ConsumerState<CloseShiftScreen> {
  final _cash = TextEditingController();
  final _note = TextEditingController();
  bool _saving = false;
  Shift? _closed;

  @override
  void dispose() {
    _cash.dispose();
    _note.dispose();
    super.dispose();
  }

  Future<void> _close(int expected) async {
    final actual = parseRupiah(_cash.text);
    final diff = actual - expected;
    final ok = await confirmAction(
      context,
      title: 'Tutup shift sekarang?',
      message: diff == 0 ? 'Kas sesuai. Shift akan ditutup.' : 'Kas ${diff < 0 ? 'kurang' : 'lebih'} ${formatRupiah(diff.abs())}. Tetap tutup shift?',
      confirmLabel: 'Tutup Shift',
    );
    if (!ok) return;
    setState(() => _saving = true);
    try {
      final closed = await ref.read(currentShiftProvider.notifier).close(actual, note: _note.text.trim());
      ref.read(cartControllerProvider.notifier).clear();
      if (!mounted) return;
      setState(() => _closed = closed);
      showToast(context, 'Shift berhasil ditutup');
    } catch (e) {
      if (mounted) showToast(context, errorMessage(e), type: ToastType.error);
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_closed != null) return _Recap(shift: _closed!);
    final c = context.colors;
    final shift = ref.watch(currentShiftProvider);

    return AppPage(
      title: 'Tutup Shift',
      child: shift.when(
        loading: () => const Center(child: CupertinoActivityIndicator()),
        error: (e, _) => ErrorState(message: errorMessage(e), onRetry: () => ref.invalidate(currentShiftProvider)),
        data: (s) {
          if (s == null || s.summary == null) {
            return const EmptyState(icon: CupertinoIcons.clock, title: 'Tidak ada shift terbuka');
          }
          final sum = s.summary!;
          final actual = parseRupiah(_cash.text);
          final diff = actual - sum.expectedCash;
          final hasInput = _cash.text.isNotEmpty;
          final diffColor = !hasInput ? c.textSecondary : diff == 0 ? c.success : diff < 0 ? c.danger : c.warning;
          return ListView(
            padding: const EdgeInsets.only(bottom: AppSpacing.xxl),
            children: [
              InsetSection(header: 'Rekap shift', children: [
                InsetRow(title: 'Dibuka', value: formatDateTime(s.openedAt)),
                InsetRow(title: 'Jumlah transaksi', value: '${sum.totalTransactions}'),
                InsetRow(title: 'Total penjualan', value: formatRupiah(sum.totalSales)),
                for (final e in sum.paymentTotals.entries)
                  InsetRow(title: '· ${paymentMethodLabels[e.key] ?? e.key}', value: formatRupiah(e.value)),
              ]),
              InsetSection(header: 'Uang di laci', footer: 'Kas seharusnya = kas awal + penjualan tunai.', children: [
                InsetRow(title: 'Kas awal', value: formatRupiah(s.openingCash)),
                InsetRow(title: 'Penjualan tunai', value: formatRupiah(sum.cashSales)),
                InsetRow(title: 'Kas seharusnya', value: formatRupiah(sum.expectedCash)),
              ]),
              Padding(
                padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.xl, AppSpacing.lg, 0),
                child: AppTextField(
                  label: 'Hitung uang di laci sekarang',
                  controller: _cash,
                  prefix: 'Rp',
                  placeholder: '0',
                  large: true,
                  keyboardType: TextInputType.number,
                  inputFormatters: [ThousandsFormatter()],
                  onChanged: (_) => setState(() {}),
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.md, AppSpacing.lg, 0),
                child: Container(
                  padding: const EdgeInsets.all(AppSpacing.lg),
                  decoration: BoxDecoration(color: diffColor.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(AppRadius.lg)),
                  child: Column(children: [
                    Text(!hasInput ? 'Selisih' : diff == 0 ? 'Kas sesuai' : diff < 0 ? 'Kas kurang' : 'Kas lebih',
                        style: AppText.headline.copyWith(color: diffColor)),
                    FittedBox(child: Text(formatRupiah(hasInput ? diff.abs() : 0), style: AppText.amount.copyWith(color: diffColor))),
                  ]),
                ),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.md, AppSpacing.lg, 0),
                child: AppTextField(label: 'Catatan (opsional)', controller: _note, placeholder: 'Contoh: uang receh ditukar'),
              ),
              Padding(
                padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.xl, AppSpacing.lg, 0),
                child: AppButton(
                  label: 'Tutup Shift',
                  icon: CupertinoIcons.lock_fill,
                  variant: AppButtonVariant.danger,
                  loading: _saving,
                  onPressed: hasInput ? () => _close(sum.expectedCash) : null,
                ),
              ),
            ],
          );
        },
      ),
    );
  }
}

class _Recap extends StatelessWidget {
  const _Recap({required this.shift});
  final Shift shift;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final diff = shift.cashDifference ?? 0;
    return AppPage(
      title: 'Shift Ditutup',
      child: ListView(padding: const EdgeInsets.only(bottom: AppSpacing.xxl), children: [
        const SizedBox(height: AppSpacing.xl),
        Icon(CupertinoIcons.checkmark_circle_fill, color: c.success, size: 72),
        const SizedBox(height: AppSpacing.md),
        Text('Terima kasih!', textAlign: TextAlign.center, style: AppText.title.copyWith(color: c.text)),
        InsetSection(children: [
          InsetRow(title: 'Total penjualan', value: formatRupiah(shift.summary?.totalSales ?? 0)),
          InsetRow(title: 'Kas seharusnya', value: formatRupiah(shift.expectedCash ?? 0)),
          InsetRow(title: 'Kas aktual', value: formatRupiah(shift.actualCash ?? 0)),
          InsetRow(title: 'Selisih', value: '${diff > 0 ? '+' : ''}${formatRupiah(diff)}'),
        ]),
        Padding(
          padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.xl, AppSpacing.lg, 0),
          child: AppButton(label: 'Selesai', icon: CupertinoIcons.checkmark_alt, onPressed: () => context.go('/kasir')),
        ),
      ]),
    );
  }
}
