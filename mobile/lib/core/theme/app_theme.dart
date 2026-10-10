import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';

import 'app_colors.dart';

/// Apple HIG type scale (size / weight). SF Pro is used automatically on iOS;
/// Android falls back to the platform sans-serif.
class AppText {
  AppText._();

  static const largeTitle = TextStyle(fontSize: 34, height: 41 / 34, fontWeight: FontWeight.w700, letterSpacing: 0.37);
  static const title = TextStyle(fontSize: 28, height: 34 / 28, fontWeight: FontWeight.w700, letterSpacing: 0.36);
  static const headline = TextStyle(fontSize: 17, height: 22 / 17, fontWeight: FontWeight.w600, letterSpacing: -0.41);
  static const body = TextStyle(fontSize: 17, height: 22 / 17, fontWeight: FontWeight.w400, letterSpacing: -0.41);
  static const subheadline = TextStyle(fontSize: 15, height: 20 / 15, fontWeight: FontWeight.w400, letterSpacing: -0.24);
  static const footnote = TextStyle(fontSize: 13, height: 18 / 13, fontWeight: FontWeight.w400, letterSpacing: -0.08);
  static const caption = TextStyle(fontSize: 12, height: 16 / 12, fontWeight: FontWeight.w400);

  /// Large, bold money figures (total, change).
  static const amount = TextStyle(fontSize: 40, height: 1.1, fontWeight: FontWeight.w700, fontFeatures: [FontFeature.tabularFigures()]);
}

/// Radius & spacing tokens.
class AppRadius {
  AppRadius._();
  static const sm = 10.0;
  static const md = 12.0;
  static const lg = 14.0;
}

class AppSpacing {
  AppSpacing._();
  static const xs = 4.0;
  static const sm = 8.0;
  static const md = 12.0;
  static const lg = 16.0;
  static const xl = 20.0;
  static const xxl = 32.0;

  /// Apple minimum touch target.
  static const touch = 44.0;
}

class AppTheme {
  AppTheme._();

  static ThemeData light() => _build(AppColors.light, Brightness.light);
  static ThemeData dark() => _build(AppColors.dark, Brightness.dark);

  static ThemeData _build(AppColors c, Brightness b) {
    final base = ThemeData(brightness: b, useMaterial3: true);
    final textTheme = base.textTheme.apply(bodyColor: c.text, displayColor: c.text).copyWith(
          headlineLarge: AppText.largeTitle,
          headlineMedium: AppText.title,
          titleMedium: AppText.headline,
          bodyLarge: AppText.body,
          bodyMedium: AppText.subheadline,
          bodySmall: AppText.footnote,
          labelSmall: AppText.caption,
        );
    return base.copyWith(
      extensions: [c],
      platform: TargetPlatform.iOS, // iOS page transitions & scrolling physics everywhere
      scaffoldBackgroundColor: c.background,
      colorScheme: ColorScheme.fromSeed(seedColor: c.primary, brightness: b).copyWith(
        primary: c.primary,
        error: c.danger,
        surface: c.card,
        onSurface: c.text,
      ),
      textTheme: textTheme,
      cupertinoOverrideTheme: CupertinoThemeData(
        brightness: b,
        primaryColor: c.primary,
        scaffoldBackgroundColor: c.background,
        barBackgroundColor: c.card.withValues(alpha: 0.92),
        textTheme: CupertinoTextThemeData(primaryColor: c.primary, textStyle: AppText.body.copyWith(color: c.text)),
      ),
      dividerTheme: DividerThemeData(color: c.separator, thickness: 0.5, space: 0.5),
      splashFactory: NoSplash.splashFactory,
      highlightColor: c.fill,
      bottomSheetTheme: BottomSheetThemeData(
        backgroundColor: c.background,
        surfaceTintColor: Colors.transparent,
        showDragHandle: true,
        dragHandleColor: c.gray.withValues(alpha: 0.4),
        shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(AppRadius.lg))),
      ),
      snackBarTheme: SnackBarThemeData(
        behavior: SnackBarBehavior.floating,
        backgroundColor: c.card,
        contentTextStyle: AppText.subheadline.copyWith(color: c.text),
        elevation: 2,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(AppRadius.md)),
      ),
    );
  }
}
