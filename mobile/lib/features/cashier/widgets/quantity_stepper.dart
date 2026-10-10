import 'package:flutter/cupertino.dart';
import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/haptics.dart';

/// − qty + control with 44pt buttons.
class QuantityStepper extends StatelessWidget {
  const QuantityStepper({super.key, required this.quantity, required this.onChanged, required this.max});

  final int quantity;
  final int max;
  final ValueChanged<int> onChanged;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    Widget btn(IconData icon, String label, VoidCallback? onTap) => Semantics(
          button: true,
          label: label,
          child: Material(
            color: c.fill,
            borderRadius: BorderRadius.circular(AppRadius.sm),
            child: InkWell(
              borderRadius: BorderRadius.circular(AppRadius.sm),
              onTap: onTap == null
                  ? null
                  : () {
                      Haptics.tap();
                      onTap();
                    },
              child: SizedBox(
                width: AppSpacing.touch,
                height: AppSpacing.touch,
                child: Icon(icon, size: 20, color: onTap == null ? c.gray : c.primary),
              ),
            ),
          ),
        );
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        btn(CupertinoIcons.minus, 'Kurangi', () => onChanged(quantity - 1)),
        SizedBox(
          width: 44,
          child: Text('$quantity', textAlign: TextAlign.center, style: AppText.headline.copyWith(color: c.text, fontSize: 20)),
        ),
        btn(CupertinoIcons.plus, 'Tambah', quantity < max ? () => onChanged(quantity + 1) : null),
      ],
    );
  }
}
