class ShiftSummary {
  const ShiftSummary({
    required this.totalSales,
    required this.totalTransactions,
    required this.voidedCount,
    required this.cashSales,
    required this.expectedCash,
    required this.paymentTotals,
  });

  final int totalSales;
  final int totalTransactions;
  final int voidedCount;
  final int cashSales;
  final int expectedCash;
  final Map<String, int> paymentTotals;

  factory ShiftSummary.fromJson(Map<String, dynamic> j) => ShiftSummary(
        totalSales: j['total_sales'] as int,
        totalTransactions: j['total_transactions'] as int,
        voidedCount: j['voided_count'] as int,
        cashSales: j['cash_sales'] as int,
        expectedCash: j['expected_cash'] as int,
        paymentTotals: (j['payment_totals'] as Map<String, dynamic>? ?? {}).map((k, v) => MapEntry(k, v as int)),
      );
}

class Shift {
  const Shift({
    required this.id,
    required this.status,
    required this.openingCash,
    required this.openedAt,
    this.closedAt,
    this.expectedCash,
    this.actualCash,
    this.cashDifference,
    this.summary,
  });

  final int id;
  final String status;
  final int openingCash;
  final DateTime openedAt;
  final DateTime? closedAt;
  final int? expectedCash;
  final int? actualCash;
  final int? cashDifference;
  final ShiftSummary? summary;

  bool get isOpen => status == 'OPEN';

  factory Shift.fromJson(Map<String, dynamic> j) => Shift(
        id: j['id'] as int,
        status: j['status'] as String,
        openingCash: j['opening_cash'] as int,
        openedAt: DateTime.parse(j['opened_at'] as String),
        closedAt: j['closed_at'] == null ? null : DateTime.parse(j['closed_at'] as String),
        expectedCash: j['expected_cash'] as int?,
        actualCash: j['actual_cash'] as int?,
        cashDifference: j['cash_difference'] as int?,
        summary: j['summary'] == null ? null : ShiftSummary.fromJson(j['summary'] as Map<String, dynamic>),
      );
}
