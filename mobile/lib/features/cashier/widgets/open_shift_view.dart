import 'package:flutter/cupertino.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inputs.dart';
import '../../shift/providers/shift_controller.dart';

/// Shown on the Kasir tab when no shift is open yet.
class OpenShiftView extends ConsumerStatefulWidget {
  const OpenShiftView({super.key});

  @override
  ConsumerState<OpenShiftView> createState() => _OpenShiftViewState();
}

class _OpenShiftViewState extends ConsumerState<OpenShiftView> {
  final _cash = TextEditingController();
  bool _loading = false;

  @override
  void dispose() {
    _cash.dispose();
    super.dispose();
  }

  Future<void> _open() async {
    FocusScope.of(context).unfocus();
    setState(() => _loading = true);
    try {
      await ref.read(currentShiftProvider.notifier).open(parseRupiah(_cash.text));
      if (mounted) showToast(context, 'Shift dibuka. Selamat bekerja!');
    } catch (e) {
      if (mounted) showToast(context, errorMessage(e), type: ToastType.error);
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.xl),
        decoration: BoxDecoration(color: c.card, borderRadius: BorderRadius.circular(AppRadius.lg)),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Icon(CupertinoIcons.clock_fill, size: 48, color: c.primary),
            const SizedBox(height: AppSpacing.md),
            Text('Buka Shift Dulu', textAlign: TextAlign.center, style: AppText.title.copyWith(color: c.text)),
            const SizedBox(height: AppSpacing.xs),
            Text(
              'Hitung uang di laci kasir, lalu masukkan jumlahnya di bawah.',
              textAlign: TextAlign.center,
              style: AppText.subheadline.copyWith(color: c.textSecondary),
            ),
            const SizedBox(height: AppSpacing.xl),
            AppTextField(
              label: 'Uang awal di laci',
              controller: _cash,
              prefix: 'Rp',
              placeholder: '0',
              large: true,
              keyboardType: TextInputType.number,
              inputFormatters: [ThousandsFormatter()],
            ),
            const SizedBox(height: AppSpacing.lg),
            AppButton(label: 'Buka Shift', icon: CupertinoIcons.play_circle_fill, loading: _loading, onPressed: _open),
          ],
        ),
      ),
    );
  }
}
