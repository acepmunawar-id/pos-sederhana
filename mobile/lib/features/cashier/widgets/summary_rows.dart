import 'package:flutter/material.dart';

import '../../../core/theme/app_colors.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/format.dart';

class SummaryRow extends StatelessWidget {
  const SummaryRow({super.key, required this.label, required this.value, this.valueColor, this.bold = false});

  final String label;
  final String value;
  final Color? valueColor;
  final bool bold;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    final style = (bold ? AppText.headline : AppText.body);
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 5),
      child: Row(
        children: [
          Expanded(child: Text(label, style: style.copyWith(color: bold ? c.text : c.textSecondary))),
          Text(value, style: style.copyWith(color: valueColor ?? c.text, fontFeatures: const [FontFeature.tabularFigures()])),
        ],
      ),
    );
  }
}

/// Subtotal / discount / tax lines.
class TotalsBreakdown extends StatelessWidget {
  const TotalsBreakdown({super.key, required this.subtotal, required this.discount, required this.taxRateBP, required this.tax, this.discountLabel});

  final int subtotal;
  final int discount;
  final int taxRateBP;
  final int tax;
  final String? discountLabel;

  @override
  Widget build(BuildContext context) {
    final c = context.colors;
    return Column(
      children: [
        SummaryRow(label: 'Subtotal', value: formatRupiah(subtotal)),
        if (discount > 0) SummaryRow(label: discountLabel ?? 'Diskon', value: '-${formatRupiah(discount)}', valueColor: c.danger),
        if (taxRateBP > 0) SummaryRow(label: 'Pajak ${formatPercentBP(taxRateBP)}', value: formatRupiah(tax)),
      ],
    );
  }
}
