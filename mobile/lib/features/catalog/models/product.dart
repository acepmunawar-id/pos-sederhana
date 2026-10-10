class Product {
  const Product({
    required this.id,
    required this.sku,
    required this.name,
    required this.categoryId,
    required this.categoryName,
    required this.unitName,
    required this.price,
    required this.stock,
    required this.minStock,
    required this.isActive,
    this.barcode,
    this.imageUrl,
  });

  final int id;
  final String sku;
  final String? barcode;
  final String name;
  final int categoryId;
  final String categoryName;
  final String unitName;
  final int price;
  final int stock;
  final int minStock;
  final String? imageUrl;
  final bool isActive;

  bool get isLowStock => stock <= minStock;
  bool get isOutOfStock => stock <= 0;

  factory Product.fromJson(Map<String, dynamic> j) {
    final cat = j['category'] as Map<String, dynamic>;
    final unit = j['unit'] as Map<String, dynamic>;
    return Product(
      id: j['id'] as int,
      sku: j['sku'] as String,
      barcode: j['barcode'] as String?,
      name: j['name'] as String,
      categoryId: cat['id'] as int,
      categoryName: cat['name'] as String,
      unitName: unit['name'] as String,
      price: j['price'] as int,
      stock: j['stock'] as int,
      minStock: j['min_stock'] as int,
      imageUrl: j['image_url'] as String?,
      isActive: j['is_active'] as bool,
    );
  }
}

class Category {
  const Category({required this.id, required this.name});
  final int id;
  final String name;

  factory Category.fromJson(Map<String, dynamic> j) => Category(id: j['id'] as int, name: j['name'] as String);
}
