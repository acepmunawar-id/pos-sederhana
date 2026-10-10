import 'package:flutter/widgets.dart';

/// Phone (< 600), tablet (600–1023), desktop (≥ 1024).
enum ScreenSize { phone, tablet, desktop }

class Breakpoints {
  Breakpoints._();
  static const tablet = 600.0;
  static const desktop = 1024.0;

  /// From this width the app uses a sidebar and the split cashier layout.
  static const wide = 900.0;

  /// Max width of reading/form content on large screens.
  static const content = 760.0;
  static const form = 600.0;
}

extension ScreenSizeX on BuildContext {
  double get screenWidth => MediaQuery.sizeOf(this).width;

  ScreenSize get screenSize {
    final w = screenWidth;
    if (w >= Breakpoints.desktop) return ScreenSize.desktop;
    if (w >= Breakpoints.tablet) return ScreenSize.tablet;
    return ScreenSize.phone;
  }

  bool get isWide => screenWidth >= Breakpoints.wide;
}

/// Centers [child] and limits its width on large screens.
class MaxWidth extends StatelessWidget {
  const MaxWidth({super.key, required this.child, this.maxWidth = Breakpoints.content, this.alignment = Alignment.topCenter});

  final Widget child;
  final double maxWidth;
  final AlignmentGeometry alignment;

  @override
  Widget build(BuildContext context) => Align(
        alignment: alignment,
        // Size to the child's height so it also works inside bottom bars.
        heightFactor: 1,
        child: ConstrainedBox(constraints: BoxConstraints(maxWidth: maxWidth), child: child),
      );
}
