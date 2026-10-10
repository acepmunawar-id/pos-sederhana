import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';

import '../theme/app_colors.dart';
import '../theme/app_theme.dart';

enum AppButtonVariant { primary, secondary, danger, success }

/// Large, labelled button (≥ 50pt tall) used for primary actions.
class AppButton extends StatelessWidget {
  const AppButton({
    super.key,
    required this.label,
    required this.onPressed,
    this.icon,
    this.variant = AppButtonVariant.primary,
    this.loading = false,
    this.expand = true,
    this.height = 52,
  });

  final String label;
  final VoidCallback? onPressed;
  final IconData? icon;
  final AppButtonVariant variant;
  final bool loading;
  final bool expand;
  final double height;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final (bg, fg) = switch (variant) {
      AppButtonVariant.primary => (c.primary, c.onAccent),
      AppButtonVariant.success => (c.success, c.onAccent),
      AppButtonVariant.danger => (c.danger, c.onAccent),
      AppButtonVariant.secondary => (c.fill, c.primary),
    };
    final disabled = onPressed == null || loading;
    final child = Row(
      mainAxisSize: expand ? MainAxisSize.max : MainAxisSize.min,
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        if (loading)
          CupertinoActivityIndicator(color: fg)
        else if (icon != null)
          Icon(icon, color: fg, size: 20),
        if (loading || icon != null) const SizedBox(width: AppSpacing.sm),
        Flexible(child: Text(label, style: AppText.headline.copyWith(color: fg), overflow: TextOverflow.ellipsis)),
      ],
    );
    return Opacity(
      opacity: disabled && !loading ? 0.45 : 1,
      child: Material(
        color: bg,
        borderRadius: BorderRadius.circular(AppRadius.md),
        child: InkWell(
          borderRadius: BorderRadius.circular(AppRadius.md),
          onTap: disabled ? null : onPressed,
          child: ConstrainedBox(
            constraints: BoxConstraints(minHeight: height, minWidth: AppSpacing.touch),
            child: Padding(padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg), child: Center(heightFactor: 1, child: child)),
          ),
        ),
      ),
    );
  }
}

/// Text-style button with icon + label (e.g. in nav bars), with a 44pt hit area.
class LabelButton extends StatelessWidget {
  const LabelButton({super.key, required this.label, required this.onPressed, this.icon, this.color});

  final String label;
  final IconData? icon;
  final VoidCallback? onPressed;
  final Color? color;

  @override
  Widget build(BuildContext context) {
    final fg = color ?? context.colors.primary;
    return CupertinoButton(
      padding: const EdgeInsets.symmetric(horizontal: 8),
      minimumSize: const Size(AppSpacing.touch, AppSpacing.touch),
      onPressed: onPressed,
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          if (icon != null) ...[Icon(icon, size: 20, color: fg), const SizedBox(width: 4)],
          Text(label, style: AppText.body.copyWith(color: fg)),
        ],
      ),
    );
  }
}
