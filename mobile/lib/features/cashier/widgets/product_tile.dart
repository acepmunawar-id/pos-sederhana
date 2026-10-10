import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/product_image.dart';
import '../../catalog/models/product.dart';

class ProductTile extends StatelessWidget {
  const ProductTile({super.key, required this.product, required this.inCart, required this.onTap});

  final Product product;
  final int inCart;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final out = product.isOutOfStock;
    return Semantics(
      button: true,
      label: '${product.name}, ${formatRupiah(product.price)}${inCart > 0 ? ', $inCart di keranjang' : ''}',
      child: Material(
        color: c.card,
        borderRadius: BorderRadius.circular(AppRadius.lg),
        clipBehavior: Clip.antiAlias,
        child: InkWell(
          onTap: onTap,
          child: Opacity(
            opacity: out ? 0.5 : 1,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                AspectRatio(
                  aspectRatio: 1.25,
                  child: Stack(
                    fit: StackFit.expand,
                    children: [
                      ProductImage(url: product.imageUrl, radius: 0),
                      if (inCart > 0)
                        Positioned(
                          top: 8,
                          right: 8,
                          child: Container(
                            constraints: const BoxConstraints(minWidth: 30),
                            height: 30,
                            padding: const EdgeInsets.symmetric(horizontal: 8),
                            alignment: Alignment.center,
                            decoration: BoxDecoration(color: c.primary, borderRadius: BorderRadius.circular(15)),
                            child: Text('$inCart', style: AppText.headline.copyWith(color: c.onAccent)),
                          ),
                        ),
                      if (out)
                        Positioned(
                          left: 8,
                          bottom: 8,
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                            decoration: BoxDecoration(color: c.danger, borderRadius: BorderRadius.circular(6)),
                            child: Text('Habis', style: AppText.footnote.copyWith(color: c.onAccent, fontWeight: FontWeight.w600)),
                          ),
                        ),
                    ],
                  ),
                ),
                Padding(
                  padding: const EdgeInsets.fromLTRB(10, 8, 10, 10),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(product.name, maxLines: 2, overflow: TextOverflow.ellipsis, style: AppText.subheadline.copyWith(color: c.text, fontWeight: FontWeight.w600)),
                      const SizedBox(height: 4),
                      Text(formatRupiah(product.price), style: AppText.headline.copyWith(color: c.text)),
                      Text(
                        out ? 'Stok habis' : 'Stok ${product.stock}',
                        style: AppText.caption.copyWith(color: product.isLowStock ? c.warning : c.textSecondary),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
