import '../../../core/utils/format.dart';
import '../../catalog/models/product.dart';

enum DiscountType { nominal, percent }

extension DiscountTypeApi on DiscountType {
  String get api => this == DiscountType.nominal ? 'NOMINAL' : 'PERCENT';
}

class CartItem {
  const CartItem({required this.product, required this.quantity});
  final Product product;
  final int quantity;

  int get subtotal => product.price * quantity;

  CartItem copyWith({int? quantity}) => CartItem(product: product, quantity: quantity ?? this.quantity);
}

class CartState {
  const CartState({this.items = const [], this.discountType, this.discountValue = 0, this.checkoutKey});

  final List<CartItem> items;
  final DiscountType? discountType;

  /// Nominal: rupiah. Percent: basis points (1000 = 10%).
  final int discountValue;

  /// Idempotency-Key for the pending checkout; kept across retries, reset after success.
  final String? checkoutKey;

  bool get isEmpty => items.isEmpty;
  int get itemCount => items.fold(0, (s, i) => s + i.quantity);
  int quantityOf(int productId) => items.where((i) => i.product.id == productId).fold(0, (s, i) => s + i.quantity);

  CartState copyWith({
    List<CartItem>? items,
    DiscountType? discountType,
    int? discountValue,
    String? checkoutKey,
    bool clearDiscount = false,
    bool clearKey = false,
  }) {
    return CartState(
      items: items ?? this.items,
      discountType: clearDiscount ? null : (discountType ?? this.discountType),
      discountValue: clearDiscount ? 0 : (discountValue ?? this.discountValue),
      checkoutKey: clearKey ? null : (checkoutKey ?? this.checkoutKey),
    );
  }
}

/// Price breakdown. Mirrors the backend: discount on subtotal, tax on (subtotal − discount),
/// all integer rupiah with half-up rounding. The server remains the source of truth.
class CartTotals {
  const CartTotals({required this.subtotal, required this.discount, required this.taxRateBP, required this.tax, required this.total});

  final int subtotal;
  final int discount;
  final int taxRateBP;
  final int tax;
  final int total;

  factory CartTotals.compute(CartState cart, int taxRateBP) {
    final subtotal = cart.items.fold(0, (s, i) => s + i.subtotal);
    var discount = 0;
    if (cart.discountType != null && cart.discountValue > 0) {
      discount = cart.discountType == DiscountType.nominal ? cart.discountValue : mulBP(subtotal, cart.discountValue);
      if (discount > subtotal) discount = subtotal;
    }
    final taxable = subtotal - discount;
    final tax = mulBP(taxable, taxRateBP);
    return CartTotals(subtotal: subtotal, discount: discount, taxRateBP: taxRateBP, tax: tax, total: taxable + tax);
  }
}
