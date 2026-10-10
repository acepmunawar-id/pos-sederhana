import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:mobile_scanner/mobile_scanner.dart';

import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/feedback.dart';
import '../../cart/providers/cart_controller.dart';
import '../../catalog/data/catalog_repository.dart';
import '../widgets/cart_feedback.dart';

/// Scans a barcode and adds the matching product to the cart.
class ScannerScreen extends ConsumerStatefulWidget {
  const ScannerScreen({super.key});

  @override
  ConsumerState<ScannerScreen> createState() => _ScannerScreenState();
}

class _ScannerScreenState extends ConsumerState<ScannerScreen> {
  final _controller = MobileScannerController(detectionSpeed: DetectionSpeed.noDuplicates, formats: const [BarcodeFormat.all]);
  bool _busy = false;
  bool _torch = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _onDetect(BarcodeCapture capture) async {
    final code = capture.barcodes.map((b) => b.rawValue).whereType<String>().firstOrNull;
    if (code == null || _busy) return;
    setState(() => _busy = true);
    try {
      final product = await ref.read(catalogRepositoryProvider).byBarcode(code);
      if (!mounted) return;
      final result = ref.read(cartControllerProvider.notifier).add(product);
      showAddFeedback(context, product, result);
      if (result == AddResult.added) context.pop();
    } catch (e) {
      if (mounted) showToast(context, e is ApiException && e.status == 404 ? 'Barcode $code tidak terdaftar' : errorMessage(e), type: ToastType.error);
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Scaffold(
      backgroundColor: c.background,
      body: Stack(
        fit: StackFit.expand,
        children: [
          MobileScanner(
            controller: _controller,
            onDetect: _onDetect,
            errorBuilder: (context, error, _) => Center(
              child: EmptyState(
                icon: CupertinoIcons.camera,
                title: 'Kamera tidak bisa dibuka',
                message: error.errorCode == MobileScannerErrorCode.permissionDenied
                    ? 'Izinkan akses kamera di Pengaturan HP, lalu coba lagi.'
                    : 'Terjadi kesalahan pada kamera.',
              ),
            ),
          ),
          Center(
            child: Container(
              width: 260,
              height: 160,
              decoration: BoxDecoration(border: Border.all(color: c.onAccent, width: 3), borderRadius: BorderRadius.circular(AppRadius.lg)),
            ),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      LabelButton(label: 'Tutup', icon: CupertinoIcons.xmark, color: c.onAccent, onPressed: () => context.pop()),
                      LabelButton(
                        label: _torch ? 'Lampu Mati' : 'Lampu',
                        icon: _torch ? CupertinoIcons.lightbulb_fill : CupertinoIcons.lightbulb,
                        color: c.onAccent,
                        onPressed: () {
                          _controller.toggleTorch();
                          setState(() => _torch = !_torch);
                        },
                      ),
                    ],
                  ),
                  const Spacer(),
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
                    decoration: BoxDecoration(color: c.card, borderRadius: BorderRadius.circular(AppRadius.lg)),
                    child: Row(mainAxisSize: MainAxisSize.min, children: [
                      if (_busy) const CupertinoActivityIndicator() else Icon(CupertinoIcons.barcode, color: c.primary),
                      const SizedBox(width: AppSpacing.sm),
                      Text(_busy ? 'Mencari produk…' : 'Arahkan kamera ke barcode', style: AppText.headline.copyWith(color: c.text)),
                    ]),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
