import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/network/api_client.dart';
import '../../../core/providers.dart';
import '../../cart/models/cart.dart';
import '../models/transaction.dart';

class CheckoutPayment {
  const CheckoutPayment({required this.method, required this.amount, this.referenceNumber});
  final String method;
  final int amount;
  final String? referenceNumber;

  Map<String, dynamic> toJson() => {
        'method': method,
        'amount': amount,
        if (referenceNumber != null && referenceNumber!.isNotEmpty) 'reference_number': referenceNumber,
      };
}

class TransactionRepository {
  TransactionRepository(this._api);
  final ApiClient _api;

  /// Submits the sale. Re-sending the same [idempotencyKey] returns the original transaction,
  /// so a retry after a network timeout never creates a duplicate.
  Future<Transaction> checkout({
    required String idempotencyKey,
    required CartState cart,
    required List<CheckoutPayment> payments,
  }) async {
    final res = await _api.post(
      '/transactions/checkout',
      headers: {'Idempotency-Key': idempotencyKey},
      body: {
        'items': [for (final i in cart.items) {'product_id': i.product.id, 'quantity': i.quantity}],
        if (cart.discountType != null && cart.discountValue > 0) ...{
          'discount_type': cart.discountType!.api,
          'discount_value': cart.discountValue,
        },
        'payments': [for (final p in payments) p.toJson()],
      },
    );
    return Transaction.fromJson(res['data'] as Map<String, dynamic>);
  }

  Future<Paged<Transaction>> list({String search = '', String? dateFrom, String? dateTo, int page = 1}) =>
      _api.getPage('/transactions', Transaction.fromJson, query: {
        'search': search,
        'date_from': dateFrom,
        'date_to': dateTo,
        'page': page,
        'limit': 30,
        'sort': 'created_at',
        'order': 'desc',
      });

  Future<Transaction> get(int id) => _api.get('/transactions/$id', Transaction.fromJson);

  /// Plain-text receipt rendered by the server (32 columns).
  Future<String> receiptText(int id) async => (await _api.get('/transactions/$id/receipt', (j) => j))['text'] as String;
}

final transactionRepositoryProvider = Provider((ref) => TransactionRepository(ref.watch(apiClientProvider)));
