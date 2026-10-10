import 'dart:async';

import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';

import '../theme/app_colors.dart';
import '../theme/app_theme.dart';
import '../utils/haptics.dart';
import 'app_button.dart';

enum ToastType { success, error, info }

OverlayEntry? _currentToast;

/// Short message at the top of the screen, always paired with haptic feedback.
void showToast(BuildContext context, String message, {ToastType type = ToastType.success}) {
  switch (type) {
    case ToastType.success:
      Haptics.success();
    case ToastType.error:
      Haptics.error();
    case ToastType.info:
      Haptics.tap();
  }
  final overlay = Overlay.of(context, rootOverlay: true);
  _currentToast?.remove();
  final entry = OverlayEntry(builder: (_) => _Toast(message: message, type: type));
  _currentToast = entry;
  overlay.insert(entry);
  Timer(Duration(milliseconds: type == ToastType.error ? 3500 : 2200), () {
    if (_currentToast == entry) {
      entry.remove();
      _currentToast = null;
    }
  });
}

class _Toast extends StatelessWidget {
  const _Toast({required this.message, required this.type});
  final String message;
  final ToastType type;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final (icon, color) = switch (type) {
      ToastType.success => (CupertinoIcons.check_mark_circled_solid, c.success),
      ToastType.error => (CupertinoIcons.xmark_circle_fill, c.danger),
      ToastType.info => (CupertinoIcons.info_circle_fill, c.primary),
    };
    return SafeArea(
      child: Align(
        alignment: Alignment.topCenter,
        child: TweenAnimationBuilder<double>(
          tween: Tween(begin: 0, end: 1),
          duration: const Duration(milliseconds: 220),
          curve: Curves.easeOutCubic,
          builder: (context, t, child) => Opacity(opacity: t, child: Transform.translate(offset: Offset(0, -16 * (1 - t)), child: child)),
          child: Container(
            margin: const EdgeInsets.all(AppSpacing.md),
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
            decoration: BoxDecoration(
              color: c.card,
              borderRadius: BorderRadius.circular(AppRadius.lg),
              boxShadow: [BoxShadow(color: c.shadow, blurRadius: 18, offset: const Offset(0, 6))],
              border: Border.all(color: c.separator.withValues(alpha: 0.4), width: 0.5),
            ),
            child: Material(
              color: Colors.transparent,
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(icon, color: color, size: 24),
                  const SizedBox(width: AppSpacing.sm),
                  Flexible(child: Text(message, style: AppText.subheadline.copyWith(color: c.text, fontWeight: FontWeight.w500))),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

/// iOS alert asking for confirmation. Returns true when the user confirms.
Future<bool> confirmAction(
  BuildContext context, {
  required String title,
  required String message,
  String confirmLabel = 'Ya',
  bool destructive = true,
}) async {
  Haptics.warning();
  final ok = await showCupertinoDialog<bool>(
    context: context,
    builder: (ctx) => CupertinoAlertDialog(
      title: Text(title),
      content: Padding(padding: const EdgeInsets.only(top: 6), child: Text(message)),
      actions: [
        CupertinoDialogAction(onPressed: () => Navigator.pop(ctx, false), child: const Text('Batal')),
        CupertinoDialogAction(
          isDestructiveAction: destructive,
          isDefaultAction: !destructive,
          onPressed: () => Navigator.pop(ctx, true),
          child: Text(confirmLabel),
        ),
      ],
    ),
  );
  return ok ?? false;
}

class EmptyState extends StatelessWidget {
  const EmptyState({super.key, required this.icon, required this.title, this.message, this.action});

  final IconData icon;
  final String title;
  final String? message;
  final Widget? action;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.xxl, vertical: 48),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 56, color: c.gray.withValues(alpha: 0.7)),
          const SizedBox(height: AppSpacing.md),
          Text(title, textAlign: TextAlign.center, style: AppText.headline.copyWith(color: c.text)),
          if (message != null) ...[
            const SizedBox(height: AppSpacing.xs),
            Text(message!, textAlign: TextAlign.center, style: AppText.subheadline.copyWith(color: c.textSecondary)),
          ],
          if (action != null) ...[const SizedBox(height: AppSpacing.lg), action!],
        ],
      ),
    );
  }
}

class ErrorState extends StatelessWidget {
  const ErrorState({super.key, required this.message, required this.onRetry});

  final String message;
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) {
    return EmptyState(
      icon: CupertinoIcons.wifi_exclamationmark,
      title: 'Gagal memuat data',
      message: message,
      action: AppButton(
        label: 'Coba Lagi',
        icon: CupertinoIcons.arrow_clockwise,
        variant: AppButtonVariant.secondary,
        expand: false,
        onPressed: onRetry,
      ),
    );
  }
}

/// Pulsing placeholder block used while data loads.
class Skeleton extends StatefulWidget {
  const Skeleton({super.key, this.width, this.height = 16, this.radius = 6});
  final double? width;
  final double height;
  final double radius;

  @override
  State<Skeleton> createState() => _SkeletonState();
}

class _SkeletonState extends State<Skeleton> with SingleTickerProviderStateMixin {
  late final _ctrl = AnimationController(vsync: this, duration: const Duration(milliseconds: 900))..repeat(reverse: true);

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return FadeTransition(
      opacity: Tween(begin: 0.45, end: 1.0).animate(_ctrl),
      child: Container(
        width: widget.width,
        height: widget.height,
        decoration: BoxDecoration(color: c.fill, borderRadius: BorderRadius.circular(widget.radius)),
      ),
    );
  }
}
