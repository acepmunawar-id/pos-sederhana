import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';

import '../theme/app_colors.dart';
import '../theme/app_theme.dart';
import '../utils/haptics.dart';

/// Labelled text field in the iOS "fill" style.
class AppTextField extends StatelessWidget {
  const AppTextField({
    super.key,
    required this.label,
    this.controller,
    this.placeholder,
    this.obscure = false,
    this.keyboardType,
    this.inputFormatters,
    this.prefix,
    this.error,
    this.autofocus = false,
    this.onChanged,
    this.onSubmitted,
    this.textInputAction,
    this.autofillHints,
    this.maxLines = 1,
    this.large = false,
  });

  final String label;
  final TextEditingController? controller;
  final String? placeholder;
  final bool obscure;
  final TextInputType? keyboardType;
  final List<TextInputFormatter>? inputFormatters;
  final String? prefix;
  final String? error;
  final bool autofocus;
  final ValueChanged<String>? onChanged;
  final ValueChanged<String>? onSubmitted;
  final TextInputAction? textInputAction;
  final Iterable<String>? autofillHints;
  final int maxLines;

  /// Big, bold text for money entry.
  final bool large;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final style = (large ? AppText.title : AppText.body).copyWith(color: c.text);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(left: 4, bottom: 6),
          child: Text(label, style: AppText.footnote.copyWith(color: c.textSecondary, fontWeight: FontWeight.w500)),
        ),
        CupertinoTextField(
          controller: controller,
          placeholder: placeholder,
          obscureText: obscure,
          keyboardType: keyboardType,
          inputFormatters: inputFormatters,
          autofocus: autofocus,
          onChanged: onChanged,
          onSubmitted: onSubmitted,
          textInputAction: textInputAction,
          autofillHints: autofillHints,
          maxLines: maxLines,
          minLines: 1,
          style: style,
          placeholderStyle: style.copyWith(color: c.gray),
          padding: EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: large ? 12 : 13),
          prefix: prefix == null
              ? null
              : Padding(
                  padding: const EdgeInsets.only(left: AppSpacing.md),
                  child: Text(prefix!, style: style.copyWith(color: c.textSecondary)),
                ),
          decoration: BoxDecoration(
            color: c.card,
            borderRadius: BorderRadius.circular(AppRadius.sm),
            border: Border.all(color: error != null ? c.danger : c.separator.withValues(alpha: 0.6), width: error != null ? 1.5 : 0.5),
          ),
        ),
        if (error != null)
          Padding(
            padding: const EdgeInsets.only(left: 4, top: 6),
            child: Text(error!, style: AppText.footnote.copyWith(color: c.danger)),
          ),
      ],
    );
  }
}

/// Formats digits as "50.000" while typing.
class ThousandsFormatter extends TextInputFormatter {
  static final _fmt = NumberFormat.decimalPattern('id_ID');

  @override
  TextEditingValue formatEditUpdate(TextEditingValue oldValue, TextEditingValue newValue) {
    final digits = newValue.text.replaceAll(RegExp(r'[^0-9]'), '');
    if (digits.isEmpty) return const TextEditingValue();
    final capped = digits.length > 12 ? digits.substring(0, 12) : digits;
    final text = _fmt.format(int.parse(capped));
    return TextEditingValue(text: text, selection: TextSelection.collapsed(offset: text.length));
  }
}

/// iOS segmented control with haptic feedback.
class AppSegmented<T extends Object> extends StatelessWidget {
  const AppSegmented({super.key, required this.value, required this.segments, required this.onChanged});

  final T value;
  final Map<T, String> segments;
  final ValueChanged<T> onChanged;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return SizedBox(
      width: double.infinity,
      child: CupertinoSlidingSegmentedControl<T>(
        groupValue: value,
        backgroundColor: c.fill,
        thumbColor: c.card,
        padding: const EdgeInsets.all(2),
        onValueChanged: (v) {
          if (v == null) return;
          Haptics.tap();
          onChanged(v);
        },
        children: {
          for (final e in segments.entries)
            e.key: Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Text(e.value, style: AppText.subheadline.copyWith(color: c.text, fontWeight: FontWeight.w600)),
            ),
        },
      ),
    );
  }
}

/// Opens a rounded bottom sheet that moves above the keyboard.
Future<T?> showAppSheet<T>(BuildContext context, {required String title, required Widget child}) {
  return showModalBottomSheet<T>(
    context: context,
    isScrollControlled: true,
    useSafeArea: true,
    builder: (ctx) {
      final c = ctx.colors;
      return Padding(
        padding: EdgeInsets.only(bottom: MediaQuery.viewInsetsOf(ctx).bottom),
        child: SingleChildScrollView(
          padding: const EdgeInsets.fromLTRB(AppSpacing.lg, 0, AppSpacing.lg, AppSpacing.xl),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(title, textAlign: TextAlign.center, style: AppText.headline.copyWith(color: c.text)),
              const SizedBox(height: AppSpacing.lg),
              child,
            ],
          ),
        ),
      );
    },
  );
}
