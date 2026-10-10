package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/posumkm/backend/internal/domain"
)

const receiptWidth = 32 // 58mm thermal paper

var paymentLabels = map[string]string{
	domain.PaymentCash: "Tunai", domain.PaymentQRIS: "QRIS", domain.PaymentTransfer: "Transfer", domain.PaymentDebit: "Debit",
}

// FormatRupiah formats an amount as "Rp 25.000" (negative: "-Rp 25.000").
func FormatRupiah(v int64) string {
	sign := ""
	if v < 0 {
		sign, v = "-", -v
	}
	s := strconv.FormatInt(v, 10)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return sign + "Rp " + b.String()
}

// FormatPercentBP formats basis points as a percentage: 1100 → "11%", 1250 → "12,5%".
func FormatPercentBP(bp int64) string {
	whole, frac := bp/100, bp%100
	if frac == 0 {
		return fmt.Sprintf("%d%%", whole)
	}
	return strings.TrimRight(fmt.Sprintf("%d,%02d", whole, frac), "0") + "%"
}

// BuildReceiptText renders a plain-text receipt suitable for sharing or thermal printing.
func BuildReceiptText(st *domain.StoreSetting, t *domain.Transaction, loc *time.Location) string {
	var b strings.Builder
	line := strings.Repeat("-", receiptWidth)
	center := func(s string) {
		for _, part := range wrap(s, receiptWidth) {
			pad := (receiptWidth - utf8.RuneCountInString(part)) / 2
			b.WriteString(strings.Repeat(" ", pad) + part + "\n")
		}
	}
	row := func(left, right string) {
		space := receiptWidth - utf8.RuneCountInString(left) - utf8.RuneCountInString(right)
		if space < 1 {
			b.WriteString(left + "\n")
			space = receiptWidth - utf8.RuneCountInString(right)
			left = ""
		}
		b.WriteString(left + strings.Repeat(" ", space) + right + "\n")
	}

	center(strings.ToUpper(st.Name))
	if st.Address != nil {
		center(*st.Address)
	}
	if st.Phone != nil {
		center("Telp " + *st.Phone)
	}
	b.WriteString(line + "\n")
	row("No", t.InvoiceNumber)
	row("Tanggal", t.CreatedAt.In(loc).Format("02/01/2006 15:04"))
	row("Kasir", t.CashierName)
	b.WriteString(line + "\n")

	for _, it := range t.Items {
		for _, part := range wrap(it.ProductName, receiptWidth) {
			b.WriteString(part + "\n")
		}
		row(fmt.Sprintf("  %d x %s", it.Quantity, strings.TrimPrefix(FormatRupiah(it.UnitPrice), "Rp ")),
			strings.TrimPrefix(FormatRupiah(it.Subtotal), "Rp "))
	}
	b.WriteString(line + "\n")
	row("Subtotal", FormatRupiah(t.Subtotal))
	if t.DiscountAmount > 0 {
		label := "Diskon"
		if t.DiscountType != nil && *t.DiscountType == domain.DiscountPercent {
			label += " " + FormatPercentBP(t.DiscountValue)
		}
		row(label, "-"+FormatRupiah(t.DiscountAmount))
	}
	if t.TaxAmount > 0 {
		row("Pajak "+FormatPercentBP(int64(t.TaxRateBP)), FormatRupiah(t.TaxAmount))
	}
	row("TOTAL", FormatRupiah(t.Total))
	for _, p := range t.Payments {
		row(paymentLabels[p.Method], FormatRupiah(p.TenderedAmount))
	}
	row("Kembalian", FormatRupiah(t.ChangeAmount))
	b.WriteString(line + "\n")
	if t.Status == domain.TxVoided {
		center("*** TRANSAKSI DIBATALKAN ***")
		b.WriteString(line + "\n")
	}
	if st.ReceiptFooter != nil {
		center(*st.ReceiptFooter)
	}
	return b.String()
}

// wrap splits s into lines of at most width runes, breaking on spaces where possible.
func wrap(s string, width int) []string {
	var out []string
	var cur []rune
	for _, word := range strings.Fields(s) {
		w := []rune(word)
		for len(w) > width { // hard-break very long words
			if len(cur) > 0 {
				out, cur = append(out, string(cur)), nil
			}
			out, w = append(out, string(w[:width])), w[width:]
		}
		switch {
		case len(cur) == 0:
			cur = w
		case len(cur)+1+len(w) <= width:
			cur = append(append(cur, ' '), w...)
		default:
			out, cur = append(out, string(cur)), w
		}
	}
	if len(cur) > 0 {
		out = append(out, string(cur))
	}
	return out
}
