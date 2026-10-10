import 'package:flutter/cupertino.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/app_scaffold.dart';
import '../../../core/widgets/feedback.dart';
import '../../cart/providers/cart_controller.dart';
import '../widgets/cart_contents.dart';
import '../widgets/total_bar.dart';

/// Step 2 of a sale on phones: review the cart. (Tablets show the cart beside the grid.)
class CartScreen extends ConsumerWidget {
  const CartScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final c = context.colors;
    final cart = ref.watch(cartControllerProvider);
    final totals = ref.watch(cartTotalsProvider);

    return AppPage(
      title: 'Keranjang',
      trailing: cart.isEmpty
          ? null
          : LabelButton(
              label: 'Kosongkan',
              color: c.danger,
              onPressed: () async {
                if (await confirmClearCart(context, ref) && context.mounted) context.pop();
              },
            ),
      bottom: cart.isEmpty
          ? null
          : TotalBar(
              label: 'Total bayar',
              total: totals.total,
              buttonLabel: 'Bayar',
              buttonIcon: CupertinoIcons.creditcard_fill,
              onPressed: () => context.push('/kasir/bayar'),
            ),
      child: cart.isEmpty
          ? Center(
              child: EmptyState(
                icon: CupertinoIcons.cart,
                title: 'Keranjang kosong',
                message: 'Pilih produk terlebih dahulu.',
                action: AppButton(label: 'Pilih Produk', expand: false, onPressed: () => context.pop()),
              ),
            )
          : const CartContents(),
    );
  }
}
