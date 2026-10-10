import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:pos_kasir/core/theme/app_theme.dart';
import 'package:pos_kasir/features/cashier/widgets/product_tile.dart';
import 'package:pos_kasir/features/cashier/widgets/quantity_stepper.dart';

import 'cart_test.dart' show product;

Widget wrap(Widget child, {bool dark = false}) => MaterialApp(
      theme: AppTheme.light(),
      darkTheme: AppTheme.dark(),
      themeMode: dark ? ThemeMode.dark : ThemeMode.light,
      home: Scaffold(body: Center(child: SizedBox(width: 180, height: 260, child: child))),
    );

void main() {
  setUpAll(() => initializeDateFormatting('id_ID'));

  testWidgets('ProductTile shows name, price, cart badge and handles tap', (tester) async {
    var taps = 0;
    await tester.pumpWidget(wrap(ProductTile(product: product(1, price: 18000), inCart: 3, onTap: () => taps++)));
    expect(find.text('Produk 1'), findsOneWidget);
    expect(find.text('Rp 18.000'), findsOneWidget);
    expect(find.text('3'), findsOneWidget);
    await tester.tap(find.byType(ProductTile));
    expect(taps, 1);
  });

  testWidgets('ProductTile marks out-of-stock products', (tester) async {
    await tester.pumpWidget(wrap(ProductTile(product: product(1, stock: 0), inCart: 0, onTap: () {}), dark: true));
    expect(find.text('Habis'), findsOneWidget);
  });

  testWidgets('QuantityStepper has 44pt targets and disables + at max', (tester) async {
    int? changed;
    await tester.pumpWidget(wrap(QuantityStepper(quantity: 2, max: 2, onChanged: (v) => changed = v)));
    final minus = find.bySemanticsLabel('Kurangi');
    expect(tester.getSize(find.descendant(of: minus, matching: find.byType(SizedBox)).first), const Size(44, 44));
    await tester.tap(find.bySemanticsLabel('Tambah'));
    expect(changed, isNull, reason: 'cannot exceed stock');
    await tester.tap(minus);
    expect(changed, 1);
  });
}
