import 'package:flutter/services.dart';

/// Haptic feedback vocabulary used across the app.
class Haptics {
  Haptics._();

  static Future<void> tap() => HapticFeedback.selectionClick();
  static Future<void> add() => HapticFeedback.lightImpact();
  static Future<void> success() => HapticFeedback.mediumImpact();
  static Future<void> warning() => HapticFeedback.heavyImpact();
  static Future<void> error() async {
    await HapticFeedback.heavyImpact();
    await Future<void>.delayed(const Duration(milliseconds: 90));
    await HapticFeedback.heavyImpact();
  }
}
