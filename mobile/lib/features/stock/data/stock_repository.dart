import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/providers.dart';

class StockRepository {
  StockRepository(this._api);
  final ApiClient _api;

  Future<void> stockIn(int productId, int quantity, {String? note}) => _api.post('/stock/in', body: {
        'items': [
          {'product_id': productId, 'quantity': quantity},
        ],
        if (note != null && note.isNotEmpty) 'note': note,
      });

  Future<void> adjust(int productId, int actualStock, String note) =>
      _api.post('/stock/adjustments', body: {'product_id': productId, 'actual_stock': actualStock, 'note': note});
}

final stockRepositoryProvider = Provider((ref) => StockRepository(ref.watch(apiClientProvider)));
