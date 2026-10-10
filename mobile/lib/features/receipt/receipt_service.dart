import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';
import 'package:share_plus/share_plus.dart';

/// Shares or prints the server-rendered receipt text.
/// (Bluetooth thermal printing is planned for phase 2.)
class ReceiptService {
  const ReceiptService();

  Future<void> share(String receiptText, String invoiceNumber) =>
      Share.share(receiptText, subject: 'Struk $invoiceNumber');

  /// Opens the system print dialog (AirPrint / Android print service) with a 58mm-roll PDF.
  Future<void> print(String receiptText, String invoiceNumber) async {
    final font = await PdfGoogleFonts.robotoMonoRegular().catchError((_) => pw.Font.courier());
    final doc = pw.Document(title: 'Struk $invoiceNumber');
    doc.addPage(pw.Page(
      pageFormat: const PdfPageFormat(58 * PdfPageFormat.mm, double.infinity, marginAll: 3 * PdfPageFormat.mm),
      build: (_) => pw.Text(receiptText, style: pw.TextStyle(font: font, fontSize: 6.4, lineSpacing: 1.2)),
    ));
    await Printing.layoutPdf(name: 'Struk-$invoiceNumber', onLayout: (_) => doc.save());
  }
}
