import 'package:flutter/material.dart';

import '../../../core/layout/breakpoints.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';

/// Bottom bar that always shows the total and the next-step button.
class TotalBar extends StatelessWidget {
  const TotalBar({
    super.key,
    required this.label,
    required this.total,
    required this.buttonLabel,
    required this.buttonIcon,
    required this.onPressed,
    this.loading = false,
  });

  final String label;
  final int total;
  final String buttonLabel;
  final IconData buttonIcon;
  final VoidCallback? onPressed;
  final bool loading;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Container(
      decoration: BoxDecoration(
        color: c.card,
        border: Border(top: BorderSide(color: c.separator.withValues(alpha: 0.6), width: 0.5)),
      ),
      child: SafeArea(
        top: false,
        child: MaxWidth(
          maxWidth: Breakpoints.form,
          child: Padding(
          padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.md, AppSpacing.lg, AppSpacing.md),
          child: Row(
            children: [
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Text(label, style: AppText.footnote.copyWith(color: c.textSecondary)),
                    FittedBox(
                      fit: BoxFit.scaleDown,
                      alignment: Alignment.centerLeft,
                      child: Text(formatRupiah(total), style: AppText.title.copyWith(color: c.text, fontFeatures: const [FontFeature.tabularFigures()])),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: AppSpacing.md),
              AppButton(label: buttonLabel, icon: buttonIcon, expand: false, loading: loading, onPressed: onPressed),
            ],
          ),
        ),
        ),
      ),
    );
  }
}
