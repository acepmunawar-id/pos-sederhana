import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/providers.dart';
import '../models/product.dart';

class CatalogRepository {
  CatalogRepository(this._api);
  final ApiClient _api;

  Future<Paged<Product>> products({
    String search = '',
    int? categoryId,
    bool? activeOnly,
    bool lowStock = false,
    String sort = 'name',
    String order = 'asc',
    int page = 1,
    int limit = 60,
  }) {
    return _api.getPage('/products', Product.fromJson, query: {
      'search': search,
      'category_id': categoryId,
      'is_active': activeOnly,
      'low_stock': lowStock ? true : null,
      'sort': sort,
      'order': order,
      'page': page,
      'limit': limit,
    });
  }

  Future<Product> byBarcode(String code) => _api.get('/products/barcode/${Uri.encodeComponent(code)}', Product.fromJson);

  Future<List<Category>> categories() async {
    final page = await _api.getPage('/categories', Category.fromJson, query: {'limit': 100, 'sort': 'name', 'order': 'asc'});
    return page.items;
  }
}

final catalogRepositoryProvider = Provider((ref) => CatalogRepository(ref.watch(apiClientProvider)));
