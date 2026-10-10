import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:pos_kasir/features/cart/models/cart.dart';
import 'package:pos_kasir/features/cart/providers/cart_controller.dart';
import 'package:pos_kasir/features/catalog/models/product.dart';

Product product(int id, {int price = 10000, int stock = 10}) => Product(
      id: id,
      sku: 'SKU-$id',
      name: 'Produk $id',
      categoryId: 1,
      categoryName: 'Minuman',
      unitName: 'pcs',
      price: price,
      stock: stock,
      minStock: 1,
      isActive: true,
    );

void main() {
  group('CartController', () {
    late ProviderContainer container;
    late CartController cart;

    setUp(() {
      container = ProviderContainer();
      cart = container.read(cartControllerProvider.notifier);
    });
    tearDown(() => container.dispose());

    test('adding the same product increments its quantity', () {
      final p = product(1);
      expect(cart.add(p), AddResult.added);
      expect(cart.add(p), AddResult.added);
      final state = container.read(cartControllerProvider);
      expect(state.items, hasLength(1));
      expect(state.itemCount, 2);
    });

    test('cannot add beyond available stock', () {
      final p = product(1, stock: 1);
      expect(cart.add(p), AddResult.added);
      expect(cart.add(p), AddResult.limitReached);
      expect(cart.add(product(2, stock: 0)), AddResult.outOfStock);
    });

    test('setQuantity clamps to stock and 0 removes the line', () {
      cart.add(product(1, stock: 5));
      cart.setQuantity(1, 99);
      expect(container.read(cartControllerProvider).items.single.quantity, 5);
      cart.setQuantity(1, 0);
      expect(container.read(cartControllerProvider).isEmpty, isTrue);
    });

    test('checkout key is stable until the cart changes', () {
      cart.add(product(1));
      final k1 = cart.checkoutKey();
      expect(cart.checkoutKey(), k1, reason: 'retries must reuse the same key');
      cart.add(product(2));
      expect(cart.checkoutKey(), isNot(k1), reason: 'a changed cart is a new transaction');
    });
  });

  group('CartTotals (mirrors backend rounding)', () {
    CartState cartOf(List<CartItem> items, {DiscountType? type, int value = 0}) =>
        CartState(items: items, discountType: type, discountValue: value);

    test('percent discount and tax', () {
      final state = cartOf([
        CartItem(product: product(1, price: 18000), quantity: 2),
        CartItem(product: product(2, price: 5000), quantity: 2),
      ], type: DiscountType.percent, value: 1000);
      final t = CartTotals.compute(state, 1100);
      expect(t.subtotal, 46000);
      expect(t.discount, 4600);
      expect(t.tax, 4554);
      expect(t.total, 45954);
    });

    test('nominal discount is capped at subtotal', () {
      final t = CartTotals.compute(
        cartOf([CartItem(product: product(1, price: 5000), quantity: 1)], type: DiscountType.nominal, value: 9000),
        1000,
      );
      expect(t.discount, 5000);
      expect(t.total, 0);
    });

    test('half-up rounding on 12.5%', () {
      final t = CartTotals.compute(
        cartOf([CartItem(product: product(1, price: 12345), quantity: 1)], type: DiscountType.percent, value: 1250),
        0,
      );
      expect(t.discount, 1543);
      expect(t.total, 10802);
    });
  });
}
