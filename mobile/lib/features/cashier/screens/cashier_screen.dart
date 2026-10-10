import 'dart:async';

import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/layout/breakpoints.dart';
import '../../../core/network/api_exception.dart';
import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../account/providers/settings_provider.dart';
import '../../cart/providers/cart_controller.dart';
import '../../catalog/providers/catalog_providers.dart';
import '../../shift/providers/shift_controller.dart';
import '../widgets/cart_contents.dart';
import '../widgets/cart_feedback.dart';
import '../widgets/open_shift_view.dart';
import '../widgets/product_tile.dart';
import '../widgets/total_bar.dart';

/// Step 1 of a sale: pick products.
class CashierScreen extends ConsumerStatefulWidget {
  const CashierScreen({super.key});

  @override
  ConsumerState<CashierScreen> createState() => _CashierScreenState();
}

class _CashierScreenState extends ConsumerState<CashierScreen> {
  final _search = TextEditingController();
  Timer? _debounce;
  ProductQuery _query = const ProductQuery();

  @override
  void dispose() {
    _debounce?.cancel();
    _search.dispose();
    super.dispose();
  }

  void _onSearch(String v) {
    _debounce?.cancel();
    _debounce = Timer(const Duration(milliseconds: 300), () {
      setState(() => _query = ProductQuery(search: v.trim(), categoryId: _query.categoryId));
    });
  }

  Future<void> _refresh() async {
    ref.invalidate(categoriesProvider);
    ref.invalidate(storeSettingsProvider);
    ref.invalidate(productListProvider(_query));
    await ref.read(productListProvider(_query).future);
  }

  @override
  Widget build(BuildContext context) {
    final shift = ref.watch(currentShiftProvider);
    ref.watch(storeSettingsProvider); // warm up tax rate for cart totals

    return shift.when(
      loading: () => const LargeTitleScaffold(title: 'Kasir', slivers: [SliverToBoxAdapter(child: _GridSkeleton())]),
      error: (e, _) => LargeTitleScaffold(
        title: 'Kasir',
        slivers: [SliverFillRemaining(hasScrollBody: false, child: ErrorState(message: errorMessage(e), onRetry: () => ref.invalidate(currentShiftProvider)))],
      ),
      data: (s) => s == null
          ? const LargeTitleScaffold(title: 'Kasir', slivers: [SliverToBoxAdapter(child: OpenShiftView())])
          : _buildSelling(context),
    );
  }

  Widget _buildSelling(BuildContext context) {
    final c = context.colors;
    final cart = ref.watch(cartControllerProvider);
    final totals = ref.watch(cartTotalsProvider);
    final products = ref.watch(productListProvider(_query));
    final categories = ref.watch(categoriesProvider).valueOrNull ?? const [];
    final wide = context.isWide;

    final grid = NotificationListener<ScrollNotification>(
      onNotification: (n) {
        if (n.metrics.extentAfter < 400) ref.read(productListProvider(_query).notifier).loadMore();
        return false;
      },
      child: LargeTitleScaffold(
        title: 'Kasir',
        maxContentWidth: null,
        onRefresh: _refresh,
        trailing: LabelButton(label: 'Scan', icon: CupertinoIcons.barcode_viewfinder, onPressed: () => context.push('/kasir/scan')),
        bottom: cart.isEmpty || wide
            ? null
            : TotalBar(
                label: '${cart.itemCount} item di keranjang',
                total: totals.total,
                buttonLabel: 'Keranjang',
                buttonIcon: CupertinoIcons.cart_fill,
                onPressed: () => context.push('/kasir/keranjang'),
              ),
        slivers: [
          SliverToBoxAdapter(
            child: Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.sm, AppSpacing.lg, AppSpacing.sm),
              child: CupertinoSearchTextField(
                controller: _search,
                placeholder: 'Cari nama produk atau SKU',
                onChanged: _onSearch,
                style: AppText.body.copyWith(color: c.text),
                backgroundColor: c.fill,
              ),
            ),
          ),
          SliverToBoxAdapter(
            child: SizedBox(
              height: 52,
              child: ListView(
                scrollDirection: Axis.horizontal,
                padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: 4),
                children: [
                  _Chip(label: 'Semua', selected: _query.categoryId == null, onTap: () => setState(() => _query = ProductQuery(search: _query.search))),
                  for (final cat in categories)
                    _Chip(
                      label: cat.name,
                      selected: _query.categoryId == cat.id,
                      onTap: () => setState(() => _query = ProductQuery(search: _query.search, categoryId: cat.id)),
                    ),
                ],
              ),
            ),
          ),
          ...products.when(
            loading: () => [const SliverToBoxAdapter(child: _GridSkeleton())],
            error: (e, _) => [
              SliverToBoxAdapter(child: ErrorState(message: errorMessage(e), onRetry: () => ref.invalidate(productListProvider(_query)))),
            ],
            data: (state) => state.items.isEmpty
                ? [
                    SliverToBoxAdapter(
                      child: EmptyState(
                        icon: CupertinoIcons.search,
                        title: _query.search.isEmpty ? 'Belum ada produk' : 'Produk tidak ditemukan',
                        message: _query.search.isEmpty ? 'Tambahkan produk dari dashboard web.' : 'Coba kata kunci lain atau pindai barcode.',
                      ),
                    ),
                  ]
                : [
                    SliverPadding(
                      padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.sm, AppSpacing.lg, 0),
                      sliver: SliverLayoutBuilder(builder: (context, constraints) {
                        // Columns follow the available width (phone 2, tablet 3–4, desktop 5+);
                        // tile height = image (1.25:1) + text block, so cards never have empty space.
                        final w = constraints.crossAxisExtent;
                        final columns = (w / 150).floor().clamp(2, 8);
                        final tileW = (w - (columns - 1) * AppSpacing.md) / columns;
                        return SliverGrid(
                          gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                            crossAxisCount: columns,
                            mainAxisSpacing: AppSpacing.md,
                            crossAxisSpacing: AppSpacing.md,
                            mainAxisExtent: tileW / 1.25 + 100,
                          ),
                          delegate: SliverChildBuilderDelegate(
                            childCount: state.items.length,
                            (context, i) {
                              final p = state.items[i];
                              return ProductTile(
                                product: p,
                                inCart: cart.quantityOf(p.id),
                                onTap: () => showAddFeedback(context, p, ref.read(cartControllerProvider.notifier).add(p)),
                              );
                            },
                          ),
                        );
                      }),
                    ),
                    if (state.loadingMore)
                      const SliverToBoxAdapter(child: Padding(padding: EdgeInsets.all(AppSpacing.lg), child: CupertinoActivityIndicator())),
                  ],
          ),
        ],
      ),
    );

    if (!wide) return grid;
    return Row(children: [
      Expanded(child: grid),
      Container(width: 0.5, color: c.separator),
      SizedBox(width: context.screenSize == ScreenSize.desktop ? 420 : 360, child: const CartSidePanel()),
    ]);
  }
}

class _Chip extends StatelessWidget {
  const _Chip({required this.label, required this.selected, required this.onTap});
  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Padding(
      padding: const EdgeInsets.only(right: AppSpacing.sm),
      child: Material(
        color: selected ? c.primary : c.card,
        borderRadius: BorderRadius.circular(22),
        child: InkWell(
          borderRadius: BorderRadius.circular(22),
          onTap: onTap,
          child: ConstrainedBox(
            constraints: const BoxConstraints(minHeight: AppSpacing.touch, minWidth: AppSpacing.touch),
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
              child: Center(
                child: Text(label, style: AppText.subheadline.copyWith(color: selected ? c.onAccent : c.text, fontWeight: FontWeight.w600)),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _GridSkeleton extends StatelessWidget {
  const _GridSkeleton();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: GridView.count(
        crossAxisCount: 2,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        mainAxisSpacing: AppSpacing.md,
        crossAxisSpacing: AppSpacing.md,
        childAspectRatio: 0.72,
        children: List.generate(6, (_) => const Skeleton(radius: AppRadius.lg)),
      ),
    );
  }
}
