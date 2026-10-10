import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../catalog/data/catalog_repository.dart';
import '../../catalog/models/product.dart';

class StockQuery {
  const StockQuery({this.search = '', this.lowOnly = false});
  final String search;
  final bool lowOnly;

  @override
  bool operator ==(Object other) => other is StockQuery && other.search == search && other.lowOnly == lowOnly;

  @override
  int get hashCode => Object.hash(search, lowOnly);
}

/// Products ordered by stock (lowest first) for the Stok tab.
final stockListProvider = FutureProvider.family<List<Product>, StockQuery>((ref, q) async {
  final page = await ref.watch(catalogRepositoryProvider).products(
        search: q.search,
        lowStock: q.lowOnly,
        activeOnly: q.lowOnly ? true : null,
        sort: 'stock',
        order: 'asc',
        limit: 100,
      );
  return page.items;
});
