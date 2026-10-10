import 'package:flutter/widgets.dart';

import '../../../core/utils/haptics.dart';
import '../../../core/widgets/feedback.dart';
import '../../cart/providers/cart_controller.dart';
import '../../catalog/models/product.dart';

/// Gives toast + haptic feedback for an add-to-cart attempt.
void showAddFeedback(BuildContext context, Product p, AddResult result) {
  switch (result) {
    case AddResult.added:
      Haptics.add();
      showToast(context, '${p.name} ditambahkan', type: ToastType.info);
    case AddResult.outOfStock:
      showToast(context, 'Stok ${p.name} habis', type: ToastType.error);
    case AddResult.limitReached:
      showToast(context, 'Stok ${p.name} hanya ${p.stock}', type: ToastType.error);
  }
}
