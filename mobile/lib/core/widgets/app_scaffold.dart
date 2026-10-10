import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';

import '../layout/breakpoints.dart';
import '../theme/app_colors.dart';

/// Page with an iOS Large Title navigation bar that collapses on scroll.
class LargeTitleScaffold extends StatelessWidget {
  const LargeTitleScaffold({
    super.key,
    required this.title,
    required this.slivers,
    this.trailing,
    this.leading,
    this.bottom,
    this.onRefresh,
    this.controller,
    this.maxContentWidth = Breakpoints.content,
  });

  final String title;
  final List<Widget> slivers;
  final Widget? trailing;
  final Widget? leading;

  /// Pinned widget below the content (e.g. the cart total bar).
  final Widget? bottom;
  final Future<void> Function()? onRefresh;
  final ScrollController? controller;

  /// Content is centred within this width on wide screens (null = full width).
  final double? maxContentWidth;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final width = MediaQuery.sizeOf(context).width;
    final side = maxContentWidth == null ? 0.0 : ((width - maxContentWidth!) / 2).clamp(0.0, double.infinity);
    return Scaffold(
      backgroundColor: c.background,
      body: CustomScrollView(
        controller: controller,
        physics: const BouncingScrollPhysics(parent: AlwaysScrollableScrollPhysics()),
        slivers: [
          CupertinoSliverNavigationBar(
            largeTitle: Text(title),
            leading: leading,
            trailing: trailing,
            backgroundColor: c.background.withValues(alpha: 0.92),
            border: Border(bottom: BorderSide(color: c.separator.withValues(alpha: 0.5), width: 0.5)),
            stretch: true,
          ),
          if (onRefresh != null) CupertinoSliverRefreshControl(onRefresh: onRefresh),
          SliverPadding(
            padding: EdgeInsets.symmetric(horizontal: side),
            sliver: SliverMainAxisGroup(slivers: slivers),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 24)),
        ],
      ),
      bottomNavigationBar: bottom,
    );
  }
}

/// Standard (small title) page used for pushed screens.
class AppPage extends StatelessWidget {
  const AppPage({super.key, required this.title, required this.child, this.trailing, this.bottom});

  final String title;
  final Widget child;
  final Widget? trailing;
  final Widget? bottom;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Scaffold(
      backgroundColor: c.background,
      appBar: CupertinoNavigationBar(
        middle: Text(title),
        trailing: trailing,
        previousPageTitle: 'Kembali',
        backgroundColor: c.background.withValues(alpha: 0.92),
        border: Border(bottom: BorderSide(color: c.separator.withValues(alpha: 0.5), width: 0.5)),
      ),
      body: MaxWidth(maxWidth: Breakpoints.form, child: child),
      bottomNavigationBar: bottom,
    );
  }
}
