import 'package:flutter/cupertino.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';
import '../../../core/widgets/app_button.dart';
import '../../../core/widgets/feedback.dart';
import '../../../core/widgets/inputs.dart';
import '../../cart/models/cart.dart';
import '../../cart/providers/cart_controller.dart';

Future<void> showDiscountSheet(BuildContext context) =>
    showAppSheet(context, title: 'Diskon', child: const _DiscountForm());

class _DiscountForm extends ConsumerStatefulWidget {
  const _DiscountForm();

  @override
  ConsumerState<_DiscountForm> createState() => _DiscountFormState();
}

class _DiscountFormState extends ConsumerState<_DiscountForm> {
  late DiscountType _type;
  late final TextEditingController _value;
  String? _error;

  @override
  void initState() {
    super.initState();
    final cart = ref.read(cartControllerProvider);
    _type = cart.discountType ?? DiscountType.nominal;
    _value = TextEditingController(
      text: cart.discountValue == 0
          ? ''
          : _type == DiscountType.nominal
              ? formatNumber(cart.discountValue)
              : (cart.discountValue / 100).toString().replaceAll('.0', '').replaceAll('.', ','),
    );
  }

  @override
  void dispose() {
    _value.dispose();
    super.dispose();
  }

  void _apply() {
    final subtotal = ref.read(cartTotalsProvider).subtotal;
    int value;
    if (_type == DiscountType.nominal) {
      value = parseRupiah(_value.text);
      if (value > subtotal) return setState(() => _error = 'Diskon tidak boleh melebihi ${formatRupiah(subtotal)}');
    } else {
      final pct = double.tryParse(_value.text.replaceAll(',', '.')) ?? 0;
      if (pct > 100) return setState(() => _error = 'Diskon maksimal 100%');
      value = (pct * 100).round();
    }
    ref.read(cartControllerProvider.notifier).setDiscount(_type, value);
    Navigator.pop(context);
    showToast(context, value == 0 ? 'Diskon dihapus' : 'Diskon diterapkan');
  }

  @override
  Widget build(BuildContext context) {
    final hasDiscount = ref.watch(cartControllerProvider).discountValue > 0;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        AppSegmented<DiscountType>(
          value: _type,
          segments: const {DiscountType.nominal: 'Rupiah (Rp)', DiscountType.percent: 'Persen (%)'},
          onChanged: (t) => setState(() {
            _type = t;
            _value.clear();
            _error = null;
          }),
        ),
        const SizedBox(height: AppSpacing.lg),
        AppTextField(
          key: ValueKey(_type),
          label: _type == DiscountType.nominal ? 'Potongan harga' : 'Persentase diskon',
          controller: _value,
          prefix: _type == DiscountType.nominal ? 'Rp' : null,
          placeholder: _type == DiscountType.nominal ? '0' : 'Contoh: 10',
          large: true,
          autofocus: true,
          error: _error,
          keyboardType: _type == DiscountType.nominal ? TextInputType.number : const TextInputType.numberWithOptions(decimal: true),
          inputFormatters: _type == DiscountType.nominal
              ? [ThousandsFormatter()]
              : [FilteringTextInputFormatter.allow(RegExp(r'^\d{0,3}([.,]\d{0,2})?'))],
        ),
        const SizedBox(height: AppSpacing.xl),
        AppButton(label: 'Terapkan Diskon', icon: CupertinoIcons.checkmark_alt, onPressed: _apply),
        if (hasDiscount) ...[
          const SizedBox(height: AppSpacing.sm),
          AppButton(
            label: 'Hapus Diskon',
            icon: CupertinoIcons.trash,
            variant: AppButtonVariant.secondary,
            onPressed: () {
              ref.read(cartControllerProvider.notifier).clearDiscount();
              Navigator.pop(context);
              showToast(context, 'Diskon dihapus');
            },
          ),
        ],
      ],
    );
  }
}
