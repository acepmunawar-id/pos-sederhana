import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../data/catalog_repository.dart';
import '../models/product.dart';

final categoriesProvider = FutureProvider<List<Category>>((ref) => ref.watch(catalogRepositoryProvider).categories());

/// Filter for the cashier product grid.
class ProductQuery {
  const ProductQuery({this.search = '', this.categoryId});
  final String search;
  final int? categoryId;

  @override
  bool operator ==(Object other) => other is ProductQuery && other.search == search && other.categoryId == categoryId;

  @override
  int get hashCode => Object.hash(search, categoryId);
}

class ProductListState {
  const ProductListState({required this.items, required this.total, required this.page, this.loadingMore = false});
  final List<Product> items;
  final int total;
  final int page;
  final bool loadingMore;

  bool get hasMore => items.length < total;

  ProductListState copyWith({List<Product>? items, int? total, int? page, bool? loadingMore}) => ProductListState(
        items: items ?? this.items,
        total: total ?? this.total,
        page: page ?? this.page,
        loadingMore: loadingMore ?? this.loadingMore,
      );
}

/// Active products for the cashier grid, with incremental loading.
class ProductListController extends FamilyAsyncNotifier<ProductListState, ProductQuery> {
  static const _pageSize = 60;

  @override
  Future<ProductListState> build(ProductQuery query) async {
    final res = await ref
        .read(catalogRepositoryProvider)
        .products(search: query.search, categoryId: query.categoryId, activeOnly: true, limit: _pageSize);
    return ProductListState(items: res.items, total: res.meta.total, page: 1);
  }

  Future<void> loadMore() async {
    final current = state.valueOrNull;
    if (current == null || !current.hasMore || current.loadingMore) return;
    state = AsyncData(current.copyWith(loadingMore: true));
    try {
      final res = await ref.read(catalogRepositoryProvider).products(
          search: arg.search, categoryId: arg.categoryId, activeOnly: true, page: current.page + 1, limit: _pageSize);
      state = AsyncData(current.copyWith(items: [...current.items, ...res.items], page: current.page + 1, total: res.meta.total, loadingMore: false));
    } catch (_) {
      state = AsyncData(current.copyWith(loadingMore: false));
    }
  }
}

final productListProvider =
    AsyncNotifierProvider.family<ProductListController, ProductListState, ProductQuery>(ProductListController.new);
