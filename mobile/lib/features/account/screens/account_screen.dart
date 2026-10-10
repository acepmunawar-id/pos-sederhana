import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/config/app_config.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inputs.dart';
import '../../../core/widgets/inset_group.dart';
import '../../auth/providers/auth_controller.dart';
import '../../cart/providers/cart_controller.dart';
import '../../shift/providers/shift_controller.dart';
import '../providers/settings_provider.dart';

class AccountScreen extends ConsumerWidget {
  const AccountScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final c = context.colors;
    final user = ref.watch(currentUserProvider);
    final shift = ref.watch(currentShiftProvider);
    final store = ref.watch(storeSettingsProvider);

    return LargeTitleScaffold(
      title: 'Akun',
      onRefresh: () async {
        await ref.read(currentShiftProvider.notifier).refresh();
        ref.invalidate(storeSettingsProvider);
      },
      slivers: [
        SliverToBoxAdapter(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, 0),
            child: Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(color: c.card, borderRadius: BorderRadius.circular(AppRadius.sm)),
              child: Row(children: [
                CircleAvatar(
                  radius: 30,
                  backgroundColor: c.primary,
                  child: Text(user.initials, style: AppText.title.copyWith(color: c.onAccent, fontSize: 22)),
                ),
                const SizedBox(width: AppSpacing.lg),
                Expanded(
                  child: Column(crossAxisAlignment: CrossAxisAlignment.start, children: [
                    Text(user.name, style: AppText.title.copyWith(color: c.text, fontSize: 22)),
                    Text('${user.roleName} · ${user.email}', style: AppText.subheadline.copyWith(color: c.textSecondary)),
                  ]),
                ),
              ]),
            ),
          ),
        ),
        SliverToBoxAdapter(
          child: shift.when(
            loading: () => const InsetSection(header: 'Shift', children: [InsetRow(title: 'Memuat…')]),
            error: (e, _) => InsetSection(header: 'Shift', children: [
              InsetRow(title: 'Gagal memuat shift', subtitle: errorMessage(e), onTap: () => ref.invalidate(currentShiftProvider)),
            ]),
            data: (s) => s == null
                ? InsetSection(header: 'Shift', footer: 'Buka shift dari tab Kasir untuk mulai berjualan.', children: [
                    InsetRow(icon: CupertinoIcons.clock, iconColor: c.gray, title: 'Belum ada shift yang dibuka'),
                  ])
                : InsetSection(header: 'Shift berjalan', children: [
                    InsetRow(icon: CupertinoIcons.clock_fill, iconColor: c.success, title: 'Dibuka', value: formatDateTime(s.openedAt)),
                    InsetRow(icon: CupertinoIcons.money_dollar_circle_fill, iconColor: c.primary, title: 'Kas awal', value: formatRupiah(s.openingCash)),
                    if (s.summary != null) ...[
                      InsetRow(icon: CupertinoIcons.cart_fill, iconColor: c.warning, title: 'Penjualan', value: formatRupiah(s.summary!.totalSales)),
                      InsetRow(icon: CupertinoIcons.doc_text_fill, iconColor: c.gray, title: 'Jumlah transaksi', value: '${s.summary!.totalTransactions}'),
                    ],
                    InsetRow(
                      icon: CupertinoIcons.lock_fill,
                      iconColor: c.danger,
                      title: 'Tutup Shift',
                      subtitle: 'Hitung kas & lihat rekap',
                      onTap: () => context.push('/akun/tutup-shift'),
                    ),
                  ]),
          ),
        ),
        SliverToBoxAdapter(
          child: InsetSection(header: 'Toko', children: [
            InsetRow(icon: CupertinoIcons.house_fill, iconColor: c.primary, title: 'Nama toko', value: store.valueOrNull?.name ?? '-'),
            InsetRow(
              icon: CupertinoIcons.percent,
              iconColor: c.warning,
              title: 'Pajak',
              value: store.valueOrNull == null ? '-' : formatPercentBP(store.value!.taxRateBP),
            ),
          ]),
        ),
        SliverToBoxAdapter(
          child: InsetSection(header: 'Keamanan', children: [
            InsetRow(icon: CupertinoIcons.lock_rotation, iconColor: c.gray, title: 'Ganti Password', onTap: () => _changePassword(context, ref)),
          ]),
        ),
        SliverToBoxAdapter(
          child: InsetSection(
            footer: 'Server: ${AppConfig.apiOrigin}',
            children: [
              InsetRow(
                title: 'Keluar',
                destructive: true,
                onTap: () async {
                  final cartNotEmpty = !ref.read(cartControllerProvider).isEmpty;
                  final ok = await confirmAction(
                    context,
                    title: 'Keluar dari akun?',
                    message: cartNotEmpty ? 'Keranjang yang belum dibayar akan dikosongkan.' : 'Anda perlu login lagi untuk berjualan.',
                    confirmLabel: 'Keluar',
                  );
                  if (!ok) return;
                  await ref.read(authControllerProvider.notifier).logout();
                  if (context.mounted) showToast(context, 'Anda telah keluar', type: ToastType.info);
                },
              ),
            ],
          ),
        ),
      ],
    );
  }

  void _changePassword(BuildContext context, WidgetRef ref) {
    showAppSheet(context, title: 'Ganti Password', child: const _ChangePasswordForm());
  }
}

class _ChangePasswordForm extends ConsumerStatefulWidget {
  const _ChangePasswordForm();

  @override
  ConsumerState<_ChangePasswordForm> createState() => _ChangePasswordFormState();
}

class _ChangePasswordFormState extends ConsumerState<_ChangePasswordForm> {
  final _old = TextEditingController();
  final _new = TextEditingController();
  final _confirm = TextEditingController();
  String? _error;
  bool _saving = false;

  @override
  void dispose() {
    _old.dispose();
    _new.dispose();
    _confirm.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (_old.text.isEmpty) return setState(() => _error = 'Isi password lama');
    if (_new.text.length < 8) return setState(() => _error = 'Password baru minimal 8 karakter');
    if (_new.text != _confirm.text) return setState(() => _error = 'Konfirmasi password tidak cocok');
    if (_new.text == _old.text) return setState(() => _error = 'Password baru harus berbeda');
    setState(() {
      _error = null;
      _saving = true;
    });
    try {
      final nav = Navigator.of(context);
      await ref.read(authControllerProvider.notifier).changePassword(_old.text, _new.text);
      nav.pop();
      if (mounted) showToast(context, 'Password diganti. Silakan login kembali.');
    } catch (e) {
      if (mounted) setState(() => _error = errorMessage(e));
    } finally {
      if (mounted) setState(() => _saving = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
      AppTextField(label: 'Password lama', controller: _old, obscure: true, autofocus: true),
      const SizedBox(height: AppSpacing.md),
      AppTextField(label: 'Password baru', controller: _new, obscure: true, placeholder: 'Minimal 8 karakter'),
      const SizedBox(height: AppSpacing.md),
      AppTextField(label: 'Ulangi password baru', controller: _confirm, obscure: true, error: _error),
      const SizedBox(height: AppSpacing.xl),
      AppButton(label: 'Simpan Password', icon: CupertinoIcons.checkmark_alt, loading: _saving, onPressed: _save),
    ]);
  }
}
