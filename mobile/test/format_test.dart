import 'package:flutter_test/flutter_test.dart';
import 'package:intl/date_symbol_data_local.dart';
import 'package:pos_kasir/core/utils/format.dart';
import 'package:pos_kasir/features/cashier/screens/payment_screen.dart';

void main() {
  setUpAll(() => initializeDateFormatting('id_ID'));

  test('formatRupiah uses Indonesian thousands separators', () {
    expect(formatRupiah(0), 'Rp 0');
    expect(formatRupiah(25000), 'Rp 25.000');
    expect(formatRupiah(1250000), 'Rp 1.250.000');
    expect(formatRupiah(-5000), '-Rp 5.000');
  });

  test('parseRupiah ignores formatting', () {
    expect(parseRupiah('50.000'), 50000);
    expect(parseRupiah('Rp 1.250.000'), 1250000);
    expect(parseRupiah(''), 0);
  });

  test('formatPercentBP', () {
    expect(formatPercentBP(1100), '11%');
    expect(formatPercentBP(1250), '12,5%');
  });

  test('quick cash options start with exact amount and include 50k/100k notes', () {
    expect(quickCashOptions(45954), [45954, 50000, 100000]);
    expect(quickCashOptions(22000), [22000, 25000, 30000, 50000, 100000]);
    expect(quickCashOptions(50000), [50000, 100000]);
    expect(quickCashOptions(150000).first, 150000);
  });
}
