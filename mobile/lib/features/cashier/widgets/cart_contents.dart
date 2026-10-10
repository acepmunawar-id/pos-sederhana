import 'package:flutter/cupertino.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inset_group.dart';
import '../../../core/widgets/product_image.dart';
import '../../cart/models/cart.dart';
import '../../cart/providers/cart_controller.dart';
import 'discount_sheet.dart';
import 'quantity_stepper.dart';
import 'summary_rows.dart';

Future<bool> _confirmRemove(BuildContext context, CartItem item) => confirmAction(
      context,
      title: 'Hapus dari keranjang?',
      message: '${item.product.name} (${item.quantity}×) akan dihapus.',
      confirmLabel: 'Hapus',
    );

/// Asks for confirmation, then empties the cart. Returns true if emptied.
Future<bool> confirmClearCart(BuildContext context, WidgetRef ref) async {
  final ok = await confirmAction(context, title: 'Kosongkan keranjang?', message: 'Semua barang di keranjang akan dihapus.', confirmLabel: 'Kosongkan');
  if (ok) {
    ref.read(cartControllerProvider.notifier).clear();
    if (context.mounted) showToast(context, 'Keranjang dikosongkan');
  }
  return ok;
}

/// Cart lines (swipe to delete, +/− stepper), discount row and totals.
/// Used by the phone cart screen and the tablet/desktop side panel.
class CartContents extends ConsumerWidget {
  const CartContents({super.key, this.compact = false});

  /// Narrow side panel: smaller paddings.
  final bool compact;

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final c = context.colors;
    final cart = ref.watch(cartControllerProvider);
    final totals = ref.watch(cartTotalsProvider);
    final ctrl = ref.read(cartControllerProvider.notifier);
    final hPad = compact ? AppSpacing.md : AppSpacing.lg;

    return ListView(
      padding: const EdgeInsets.only(bottom: AppSpacing.xl),
      children: [
        Padding(
          padding: EdgeInsets.fromLTRB(hPad + 4, AppSpacing.lg, hPad, 0),
          child: Text('Geser ke kiri untuk menghapus', style: AppText.footnote.copyWith(color: c.textSecondary)),
        ),
        Padding(
          padding: EdgeInsets.fromLTRB(hPad, 6, hPad, 0),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(AppRadius.sm),
            child: ColoredBox(
              color: c.card,
              child: Column(
                children: [
                  for (final (i, item) in cart.items.indexed) ...[
                    if (i > 0) Padding(padding: const EdgeInsets.only(left: 76), child: Container(height: 0.5, color: c.separator)),
                    Dismissible(
                      key: ValueKey(item.product.id),
                      direction: DismissDirection.endToStart,
                      confirmDismiss: (_) => _confirmRemove(context, item),
                      onDismissed: (_) {
                        ctrl.remove(item.product.id);
                        showToast(context, '${item.product.name} dihapus');
                      },
                      background: Container(
                        color: c.danger,
                        alignment: Alignment.centerRight,
                        padding: const EdgeInsets.only(right: AppSpacing.xl),
                        child: Row(mainAxisSize: MainAxisSize.min, children: [
                          Icon(CupertinoIcons.trash_fill, color: c.onAccent),
                          const SizedBox(width: 6),
                          Text('Hapus', style: AppText.headline.copyWith(color: c.onAccent)),
                        ]),
                      ),
                      child: _CartLine(
                        item: item,
                        onQuantity: (q) async {
                          if (q == 0) {
                            if (await _confirmRemove(context, item)) {
                              ctrl.remove(item.product.id);
                              if (context.mounted) showToast(context, '${item.product.name} dihapus');
                            }
                            return;
                          }
                          ctrl.setQuantity(item.product.id, q);
                        },
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ),
        ),
        InsetSection(
          margin: EdgeInsets.fromLTRB(hPad, AppSpacing.lg, hPad, 0),
          children: [
            InsetRow(
              icon: CupertinoIcons.tag_fill,
              iconColor: c.warning,
              title: cart.discountValue > 0 ? 'Ubah diskon' : 'Tambah diskon',
              value: totals.discount > 0 ? '-${formatRupiah(totals.discount)}' : null,
              onTap: () => showDiscountSheet(context),
            ),
          ],
        ),
        Padding(
          padding: EdgeInsets.fromLTRB(hPad, AppSpacing.lg, hPad, 0),
          child: Container(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg, vertical: AppSpacing.md),
            decoration: BoxDecoration(color: c.card, borderRadius: BorderRadius.circular(AppRadius.sm)),
            child: TotalsBreakdown(
              subtotal: totals.subtotal,
              discount: totals.discount,
              taxRateBP: totals.taxRateBP,
              tax: totals.tax,
              discountLabel: cart.discountType == DiscountType.percent ? 'Diskon ${formatPercentBP(cart.discountValue)}' : 'Diskon',
            ),
          ),
        ),
      ],
    );
  }
}

class _CartLine extends StatelessWidget {
  const _CartLine({required this.item, required this.onQuantity});
  final CartItem item;
  final ValueChanged<int> onQuantity;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Padding(
      padding: const EdgeInsets.all(AppSpacing.md),
      child: Row(
        children: [
          ProductImage(url: item.product.imageUrl, size: 52),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(item.product.name, maxLines: 2, overflow: TextOverflow.ellipsis, style: AppText.body.copyWith(color: c.text, fontWeight: FontWeight.w500)),
                Text(formatRupiah(item.product.price), style: AppText.footnote.copyWith(color: c.textSecondary)),
                const SizedBox(height: 2),
                Text(formatRupiah(item.subtotal), style: AppText.headline.copyWith(color: c.text)),
              ],
            ),
          ),
          QuantityStepper(quantity: item.quantity, max: item.product.stock, onChanged: onQuantity),
        ],
      ),
    );
  }
}

/// Right-hand cart column used on tablets and desktop next to the product grid.
class CartSidePanel extends ConsumerWidget {
  const CartSidePanel({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final c = context.colors;
    final cart = ref.watch(cartControllerProvider);
    final totals = ref.watch(cartTotalsProvider);
    return ColoredBox(
      color: c.background,
      child: SafeArea(
        left: false,
        child: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(AppSpacing.lg, AppSpacing.lg, AppSpacing.sm, 0),
              child: Row(children: [
                Expanded(child: Text('Keranjang', style: AppText.title.copyWith(color: c.text))),
                if (!cart.isEmpty)
                  LabelButton(label: 'Kosongkan', color: c.danger, onPressed: () => confirmClearCart(context, ref)),
              ]),
            ),
            Expanded(
              child: cart.isEmpty
                  ? const Center(
                      child: EmptyState(icon: CupertinoIcons.cart, title: 'Keranjang kosong', message: 'Ketuk produk di sebelah kiri untuk menambahkan.'),
                    )
                  : const CartContents(compact: true),
            ),
            Container(
              decoration: BoxDecoration(
                color: c.card,
                border: Border(top: BorderSide(color: c.separator.withValues(alpha: 0.6), width: 0.5)),
              ),
              padding: const EdgeInsets.all(AppSpacing.lg),
              child: Column(crossAxisAlignment: CrossAxisAlignment.stretch, children: [
                Row(children: [
                  Expanded(child: Text('Total (${cart.itemCount} item)', style: AppText.headline.copyWith(color: c.textSecondary))),
                  Text(formatRupiah(totals.total), style: AppText.title.copyWith(color: c.text)),
                ]),
                const SizedBox(height: AppSpacing.md),
                AppButton(
                  label: 'Bayar',
                  icon: CupertinoIcons.creditcard_fill,
                  height: 56,
                  onPressed: cart.isEmpty ? null : () => context.push('/kasir/bayar'),
                ),
              ]),
            ),
          ],
        ),
      ),
    );
  }
}
