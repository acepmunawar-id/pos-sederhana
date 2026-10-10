const paymentMethodLabels = {'CASH': 'Tunai', 'QRIS': 'QRIS', 'TRANSFER': 'Transfer', 'DEBIT': 'Debit'};

class TransactionItem {
  const TransactionItem({required this.productName, required this.quantity, required this.unitPrice, required this.subtotal});
  final String productName;
  final int quantity;
  final int unitPrice;
  final int subtotal;

  factory TransactionItem.fromJson(Map<String, dynamic> j) => TransactionItem(
        productName: j['product_name'] as String,
        quantity: j['quantity'] as int,
        unitPrice: j['unit_price'] as int,
        subtotal: j['subtotal'] as int,
      );
}

class Payment {
  const Payment({required this.method, required this.amount, required this.tenderedAmount, this.referenceNumber});
  final String method;
  final int amount;
  final int tenderedAmount;
  final String? referenceNumber;

  String get label => paymentMethodLabels[method] ?? method;

  factory Payment.fromJson(Map<String, dynamic> j) => Payment(
        method: j['method'] as String,
        amount: j['amount'] as int,
        tenderedAmount: j['tendered_amount'] as int,
        referenceNumber: j['reference_number'] as String?,
      );
}

class Transaction {
  const Transaction({
    required this.id,
    required this.invoiceNumber,
    required this.cashierName,
    required this.subtotal,
    required this.discountAmount,
    required this.taxRateBP,
    required this.taxAmount,
    required this.total,
    required this.paidAmount,
    required this.changeAmount,
    required this.status,
    required this.itemCount,
    required this.items,
    required this.payments,
    required this.createdAt,
    this.voidReason,
  });

  final int id;
  final String invoiceNumber;
  final String cashierName;
  final int subtotal;
  final int discountAmount;
  final int taxRateBP;
  final int taxAmount;
  final int total;
  final int paidAmount;
  final int changeAmount;
  final String status;
  final int itemCount;
  final List<TransactionItem> items;
  final List<Payment> payments;
  final DateTime createdAt;
  final String? voidReason;

  bool get isVoided => status == 'VOIDED';
  String get paymentLabel => payments.map((p) => p.label).join(' + ');

  factory Transaction.fromJson(Map<String, dynamic> j) => Transaction(
        id: j['id'] as int,
        invoiceNumber: j['invoice_number'] as String,
        cashierName: (j['cashier'] as Map<String, dynamic>)['name'] as String,
        subtotal: j['subtotal'] as int,
        discountAmount: j['discount_amount'] as int,
        taxRateBP: j['tax_rate_bp'] as int,
        taxAmount: j['tax_amount'] as int,
        total: j['total'] as int,
        paidAmount: j['paid_amount'] as int,
        changeAmount: j['change_amount'] as int,
        status: j['status'] as String,
        itemCount: j['item_count'] as int,
        items: (j['items'] as List? ?? []).cast<Map<String, dynamic>>().map(TransactionItem.fromJson).toList(),
        payments: (j['payments'] as List? ?? []).cast<Map<String, dynamic>>().map(Payment.fromJson).toList(),
        createdAt: DateTime.parse(j['created_at'] as String),
        voidReason: j['void_reason'] as String?,
      );
}
