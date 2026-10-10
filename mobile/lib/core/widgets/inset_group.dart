import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';

import '../theme/app_colors.dart';
import '../theme/app_theme.dart';

/// "Inset grouped" section, like the iPhone Settings app.
class InsetSection extends StatelessWidget {
  const InsetSection({super.key, this.header, this.footer, required this.children, this.margin});

  final String? header;
  final String? footer;
  final List<Widget> children;
  final EdgeInsets? margin;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final rows = <Widget>[];
    for (var i = 0; i < children.length; i++) {
      rows.add(children[i]);
      if (i < children.length - 1) {
        rows.add(Padding(
          padding: const EdgeInsets.only(left: AppSpacing.lg),
          child: Container(height: 0.5, color: c.separator),
        ));
      }
    }
    return Padding(
      padding: margin ?? const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, 0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (header != null)
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, 0, AppSpacing.lg, 6),
              child: Text(header!.toUpperCase(), style: AppText.footnote.copyWith(color: c.textSecondary)),
            ),
          ClipRRect(
            borderRadius: BorderRadius.circular(AppRadius.sm),
            child: ColoredBox(color: c.card, child: Column(children: rows)),
          ),
          if (footer != null)
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, 6, AppSpacing.lg, 0),
              child: Text(footer!, style: AppText.footnote.copyWith(color: c.textSecondary)),
            ),
        ],
      ),
    );
  }
}

/// A row inside [InsetSection]. Icons always come with a text label.
class InsetRow extends StatelessWidget {
  const InsetRow({
    super.key,
    required this.title,
    this.subtitle,
    this.value,
    this.icon,
    this.iconColor,
    this.onTap,
    this.destructive = false,
    this.trailing,
    this.showChevron,
  });

  final String title;
  final String? subtitle;
  final String? value;
  final IconData? icon;
  final Color? iconColor;
  final VoidCallback? onTap;
  final bool destructive;
  final Widget? trailing;
  final bool? showChevron;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final chevron = showChevron ?? (onTap != null && !destructive);
    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: onTap,
        child: ConstrainedBox(
          constraints: const BoxConstraints(minHeight: AppSpacing.touch),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: 10),
            child: Row(
              children: [
                if (icon != null) ...[
                  Container(
                    width: 30,
                    height: 30,
                    decoration: BoxDecoration(color: iconColor ?? c.primary, borderRadius: BorderRadius.circular(7)),
                    child: Icon(icon, size: 18, color: c.onAccent),
                  ),
                  const SizedBox(width: AppSpacing.md),
                ],
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(title, style: AppText.body.copyWith(color: destructive ? c.danger : c.text)),
                      if (subtitle != null)
                        Text(subtitle!, style: AppText.footnote.copyWith(color: c.textSecondary)),
                    ],
                  ),
                ),
                if (value != null)
                  Flexible(
                    child: Text(value!,
                        textAlign: TextAlign.right,
                        style: AppText.body.copyWith(color: c.textSecondary),
                        overflow: TextOverflow.ellipsis),
                  ),
                if (trailing != null) trailing!,
                if (chevron) ...[
                  const SizedBox(width: 6),
                  Icon(CupertinoIcons.chevron_forward, size: 18, color: c.gray.withValues(alpha: 0.6)),
                ],
              ],
            ),
          ),
        ),
      ),
    );
  }
}
