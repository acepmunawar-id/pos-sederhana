import 'package:flutter/material.dart';

/// iOS system palette — must stay in sync with web-cms/src/index.css.
/// Widgets read colours via `context.colors`, never hard-coded values.
@immutable
class AppColors extends ThemeExtension<AppColors> {
  const AppColors({
    required this.primary,
    required this.success,
    required this.danger,
    required this.warning,
    required this.yellow,
    required this.background,
    required this.card,
    required this.fill,
    required this.text,
    required this.textSecondary,
    required this.gray,
    required this.separator,
    required this.onAccent,
    required this.shadow,
  });

  final Color primary;
  final Color success;
  final Color danger;
  final Color warning;
  final Color yellow;
  final Color background;
  final Color card;
  final Color fill;
  final Color text;
  final Color textSecondary;
  final Color gray;
  final Color separator;

  /// Text/icons drawn on top of accent fills (primary, success, danger buttons).
  final Color onAccent;

  /// Very subtle elevation shadow.
  final Color shadow;

  static const light = AppColors(
    primary: Color(0xFF007AFF),
    success: Color(0xFF34C759),
    danger: Color(0xFFFF3B30),
    warning: Color(0xFFFF9500),
    yellow: Color(0xFFFFCC00),
    background: Color(0xFFF2F2F7),
    card: Color(0xFFFFFFFF),
    fill: Color.fromRGBO(118, 118, 128, 0.12),
    text: Color(0xFF000000),
    textSecondary: Color.fromRGBO(60, 60, 67, 0.6),
    gray: Color(0xFF8E8E93),
    separator: Color(0xFFC6C6C8),
    onAccent: Color(0xFFFFFFFF),
    shadow: Color(0x14000000),
  );

  static const dark = AppColors(
    primary: Color(0xFF0A84FF),
    success: Color(0xFF30D158),
    danger: Color(0xFFFF453A),
    warning: Color(0xFFFF9F0A),
    yellow: Color(0xFFFFD60A),
    background: Color(0xFF000000),
    card: Color(0xFF1C1C1E),
    fill: Color.fromRGBO(118, 118, 128, 0.24),
    text: Color(0xFFFFFFFF),
    textSecondary: Color.fromRGBO(235, 235, 245, 0.6),
    gray: Color(0xFF8E8E93),
    separator: Color(0xFF38383A),
    onAccent: Color(0xFFFFFFFF),
    shadow: Color(0x00000000),
  );

  @override
  AppColors copyWith({
    Color? primary,
    Color? success,
    Color? danger,
    Color? warning,
    Color? yellow,
    Color? background,
    Color? card,
    Color? fill,
    Color? text,
    Color? textSecondary,
    Color? gray,
    Color? separator,
    Color? onAccent,
    Color? shadow,
  }) {
    return AppColors(
      primary: primary ?? this.primary,
      success: success ?? this.success,
      danger: danger ?? this.danger,
      warning: warning ?? this.warning,
      yellow: yellow ?? this.yellow,
      background: background ?? this.background,
      card: card ?? this.card,
      fill: fill ?? this.fill,
      text: text ?? this.text,
      textSecondary: textSecondary ?? this.textSecondary,
      gray: gray ?? this.gray,
      separator: separator ?? this.separator,
      onAccent: onAccent ?? this.onAccent,
      shadow: shadow ?? this.shadow,
    );
  }

  @override
  AppColors lerp(ThemeExtension<AppColors>? other, double t) {
    if (other is! AppColors) return this;
    return AppColors(
      primary: Color.lerp(primary, other.primary, t)!,
      success: Color.lerp(success, other.success, t)!,
      danger: Color.lerp(danger, other.danger, t)!,
      warning: Color.lerp(warning, other.warning, t)!,
      yellow: Color.lerp(yellow, other.yellow, t)!,
      background: Color.lerp(background, other.background, t)!,
      card: Color.lerp(card, other.card, t)!,
      fill: Color.lerp(fill, other.fill, t)!,
      text: Color.lerp(text, other.text, t)!,
      textSecondary: Color.lerp(textSecondary, other.textSecondary, t)!,
      gray: Color.lerp(gray, other.gray, t)!,
      separator: Color.lerp(separator, other.separator, t)!,
      onAccent: Color.lerp(onAccent, other.onAccent, t)!,
      shadow: Color.lerp(shadow, other.shadow, t)!,
    );
  }
}

extension AppColorsX on BuildContext {
  AppColors get colors => Theme.of(this).extension<AppColors>()!;
}
