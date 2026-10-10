import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/cupertino.dart';

import '../theme/app_colors.dart';
import '../utils/format.dart';

class ProductImage extends StatelessWidget {
  const ProductImage({super.key, required this.url, this.size, this.radius = 10});

  final String? url;
  final double? size;
  final double radius;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final resolved = resolveImageUrl(url);
    final placeholder = Container(
      width: size,
      height: size,
      color: c.fill,
      alignment: Alignment.center,
      child: Icon(CupertinoIcons.cube_box, color: c.gray, size: (size ?? 64) * 0.4),
    );
    return ClipRRect(
      borderRadius: BorderRadius.circular(radius),
      child: resolved == null
          ? placeholder
          : CachedNetworkImage(
              imageUrl: resolved,
              width: size,
              height: size,
              fit: BoxFit.cover,
              placeholder: (_, __) => placeholder,
              errorWidget: (_, __, ___) => placeholder,
            ),
    );
  }
}
