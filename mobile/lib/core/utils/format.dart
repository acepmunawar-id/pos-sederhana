import 'package:intl/intl.dart';

import '../config/app_config.dart';

final _thousands = NumberFormat.decimalPattern('id_ID');

/// 25000 → "Rp 25.000" (same format as web and receipts).
String formatRupiah(int value) {
  final sign = value < 0 ? '-' : '';
  return '${sign}Rp ${_thousands.format(value.abs())}';
}

String formatNumber(int value) => _thousands.format(value);

/// Basis points → "11%" / "12,5%".
String formatPercentBP(int bp) {
  final v = bp / 100;
  return '${NumberFormat('#,##0.##', 'id_ID').format(v)}%';
}

/// round-half-up(amount × bp / 10000) — mirrors the backend exactly.
int mulBP(int amount, int bp) => (amount * bp + 5000) ~/ 10000;

String formatDateTime(DateTime dt) => DateFormat('d MMM yyyy, HH.mm', 'id_ID').format(dt.toLocal());
String formatTime(DateTime dt) => DateFormat('HH.mm', 'id_ID').format(dt.toLocal());
String formatDayHeader(DateTime dt) => DateFormat('EEEE, d MMMM yyyy', 'id_ID').format(dt.toLocal());

/// Parses user-typed money ("50.000", "50000") into rupiah.
int parseRupiah(String input) => int.tryParse(input.replaceAll(RegExp(r'[^0-9]'), '')) ?? 0;

/// The backend may return image URLs on "localhost", which an emulator or phone can't reach.
/// Rewrite such URLs to the API host the app is configured with.
String? resolveImageUrl(String? url) {
  if (url == null || url.isEmpty) return null;
  final uri = Uri.tryParse(url);
  if (uri == null) return null;
  if (uri.host == 'localhost' || uri.host == '127.0.0.1') {
    final api = AppConfig.apiOrigin;
    return uri.replace(scheme: api.scheme, host: api.host, port: api.port).toString();
  }
  return url;
}
