import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:uuid/uuid.dart';

import '../../account/providers/settings_provider.dart';
import '../../catalog/models/product.dart';
import '../models/cart.dart';

/// Result of trying to add a product, so the UI can give the right feedback.
enum AddResult { added, outOfStock, limitReached }

class CartController extends Notifier<CartState> {
  @override
  CartState build() => const CartState();

  AddResult add(Product p) {
    final current = state.quantityOf(p.id);
    if (p.isOutOfStock) return AddResult.outOfStock;
    if (current >= p.stock) return AddResult.limitReached;
    final exists = state.items.any((i) => i.product.id == p.id);
    final items = exists
        ? [for (final i in state.items) i.product.id == p.id ? i.copyWith(quantity: i.quantity + 1) : i]
        : [...state.items, CartItem(product: p, quantity: 1)];
    state = state.copyWith(items: items, clearKey: true);
    return AddResult.added;
  }

  /// Sets the quantity (clamped to available stock). 0 removes the line.
  void setQuantity(int productId, int quantity) {
    if (quantity <= 0) return remove(productId);
    state = state.copyWith(
      items: [
        for (final i in state.items)
          if (i.product.id == productId) i.copyWith(quantity: quantity.clamp(1, i.product.stock)) else i,
      ],
      clearKey: true,
    );
  }

  void remove(int productId) {
    state = state.copyWith(items: state.items.where((i) => i.product.id != productId).toList(), clearKey: true);
  }

  void setDiscount(DiscountType type, int value) {
    state = value <= 0 ? state.copyWith(clearDiscount: true, clearKey: true) : state.copyWith(discountType: type, discountValue: value, clearKey: true);
  }

  void clearDiscount() => state = state.copyWith(clearDiscount: true, clearKey: true);

  /// Returns the idempotency key for this checkout attempt, creating it once.
  /// Changing the cart resets it, so a modified cart is never mistaken for a retry.
  String checkoutKey() {
    final key = state.checkoutKey ?? const Uuid().v4();
    if (state.checkoutKey == null) state = state.copyWith(checkoutKey: key);
    return key;
  }

  void clear() => state = const CartState();
}

final cartControllerProvider = NotifierProvider<CartController, CartState>(CartController.new);

/// Live totals using the store's tax rate.
final cartTotalsProvider = Provider<CartTotals>((ref) {
  final cart = ref.watch(cartControllerProvider);
  final taxBP = ref.watch(storeSettingsProvider).valueOrNull?.taxRateBP ?? 0;
  return CartTotals.compute(cart, taxBP);
});
