package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/pagination"
	"github.com/xuri/excelize/v2"
)

const maxReportDays = 366

type ReportService struct {
	reports  domain.ReportRepository
	products domain.ProductRepository
	loc      *time.Location
	now      Clock
}

func NewReportService(reports domain.ReportRepository, products domain.ProductRepository, loc *time.Location, now Clock) *ReportService {
	return &ReportService{reports: reports, products: products, loc: loc, now: now}
}

// ParseRange turns inclusive YYYY-MM-DD dates (store timezone) into a half-open range.
// Missing dates default to the last 30 days ending today.
func (s *ReportService) ParseRange(from, to string) (domain.ReportRange, error) {
	today := startOfDay(s.now(), s.loc)
	end := today
	start := today.AddDate(0, 0, -29)
	var errs []apperror.FieldError
	if to != "" {
		t, err := time.ParseInLocation("2006-01-02", to, s.loc)
		if err != nil {
			errs = append(errs, apperror.FieldError{Field: "date_to", Message: "format tanggal harus YYYY-MM-DD"})
		}
		end = t
	}
	if from != "" {
		t, err := time.ParseInLocation("2006-01-02", from, s.loc)
		if err != nil {
			errs = append(errs, apperror.FieldError{Field: "date_from", Message: "format tanggal harus YYYY-MM-DD"})
		}
		start = t
	} else if to != "" {
		start = end.AddDate(0, 0, -29)
	}
	if len(errs) > 0 {
		return domain.ReportRange{}, apperror.Validation(errs)
	}
	if end.Before(start) {
		return domain.ReportRange{}, apperror.Validation([]apperror.FieldError{{Field: "date_to", Message: "tanggal akhir tidak boleh sebelum tanggal awal"}})
	}
	if end.Sub(start) > maxReportDays*24*time.Hour {
		return domain.ReportRange{}, apperror.Validation([]apperror.FieldError{{Field: "date_from", Message: "rentang laporan maksimal 1 tahun"}})
	}
	return domain.ReportRange{From: start, To: end.AddDate(0, 0, 1)}, nil
}

// Label formats a range back into its inclusive dates.
func (s *ReportService) Label(r domain.ReportRange) (string, string) {
	return r.From.In(s.loc).Format("2006-01-02"), r.To.In(s.loc).AddDate(0, 0, -1).Format("2006-01-02")
}

func (s *ReportService) Dashboard(ctx context.Context) (*dto.DashboardResponse, error) {
	today := startOfDay(s.now(), s.loc)
	todayRange := domain.ReportRange{From: today, To: today.AddDate(0, 0, 1)}
	yesterdayRange := domain.ReportRange{From: today.AddDate(0, 0, -1), To: today}
	weekRange := domain.ReportRange{From: today.AddDate(0, 0, -6), To: today.AddDate(0, 0, 1)}

	todaySum, err := s.reports.Summary(ctx, todayRange)
	if err != nil {
		return nil, err
	}
	ySum, err := s.reports.Summary(ctx, yesterdayRange)
	if err != nil {
		return nil, err
	}
	daily, err := s.reports.SalesByPeriod(ctx, weekRange, false)
	if err != nil {
		return nil, err
	}
	top, err := s.reports.TopProducts(ctx, weekRange, 5)
	if err != nil {
		return nil, err
	}
	active := true
	low, lowCount, err := s.products.List(ctx, domain.ProductFilter{
		Params:   pagination.Params{Page: 1, Limit: 5},
		IsActive: &active, LowStock: true,
	})
	if err != nil {
		return nil, err
	}

	// Fill days without sales so the chart always has 7 points.
	byDay := make(map[string]domain.SalesPeriodRow, len(daily))
	for _, d := range daily {
		byDay[d.Period] = d
	}
	points := make([]dto.DailyPoint, 0, 7)
	for d := weekRange.From; d.Before(weekRange.To); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		row := byDay[key]
		points = append(points, dto.DailyPoint{Date: key, TotalRevenue: row.TotalRevenue, TransactionCount: row.TransactionCount})
	}

	return &dto.DashboardResponse{
		Today:         dto.ToSalesSummaryResponse(*todaySum),
		Yesterday:     dto.ToSalesSummaryResponse(*ySum),
		Last7Days:     points,
		TopProducts:   dto.ToTopProductResponses(top),
		LowStock:      dto.ToProductResponses(low),
		LowStockCount: lowCount,
	}, nil
}

func (s *ReportService) Sales(ctx context.Context, r domain.ReportRange, monthly bool) (*dto.SalesReportResponse, error) {
	sum, err := s.reports.Summary(ctx, r)
	if err != nil {
		return nil, err
	}
	rows, err := s.reports.SalesByPeriod(ctx, r, monthly)
	if err != nil {
		return nil, err
	}
	from, to := s.Label(r)
	group := "daily"
	if monthly {
		group = "monthly"
	}
	out := &dto.SalesReportResponse{From: from, To: to, Group: group, Summary: dto.ToSalesSummaryResponse(*sum),
		Rows: make([]dto.SalesPeriodResponse, len(rows))}
	for i, row := range rows {
		out.Rows[i] = dto.SalesPeriodResponse{Period: row.Period, SalesSummaryResponse: dto.ToSalesSummaryResponse(row.SalesSummary)}
	}
	return out, nil
}

func (s *ReportService) GrossProfit(ctx context.Context, r domain.ReportRange) (*dto.SalesSummaryResponse, error) {
	sum, err := s.reports.Summary(ctx, r)
	if err != nil {
		return nil, err
	}
	res := dto.ToSalesSummaryResponse(*sum)
	return &res, nil
}

func (s *ReportService) TopProducts(ctx context.Context, r domain.ReportRange, limit int) ([]domain.TopProductRow, error) {
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.reports.TopProducts(ctx, r, limit)
}

func (s *ReportService) PaymentMethods(ctx context.Context, r domain.ReportRange) ([]domain.PaymentMethodRow, error) {
	return s.reports.ByPaymentMethod(ctx, r)
}

func (s *ReportService) Cashiers(ctx context.Context, r domain.ReportRange) ([]domain.CashierRow, error) {
	return s.reports.ByCashier(ctx, r)
}

// ExportFile is a generated report download.
type ExportFile struct {
	Filename    string
	ContentType string
	Data        []byte
}

var ExportTypes = []string{"sales-daily", "sales-monthly", "top-products", "payment-methods", "cashiers"}

// Export renders a report as CSV or XLSX. Money columns are plain integers (rupiah)
// so spreadsheets can sum them.
func (s *ReportService) Export(ctx context.Context, reportType, format string, r domain.ReportRange) (*ExportFile, error) {
	header, rows, err := s.exportTable(ctx, reportType, r)
	if err != nil {
		return nil, err
	}
	from, to := s.Label(r)
	base := fmt.Sprintf("laporan-%s_%s_%s", reportType, from, to)
	switch format {
	case "csv":
		var buf bytes.Buffer
		buf.WriteString("\xEF\xBB\xBF") // BOM so Excel detects UTF-8
		w := csv.NewWriter(&buf)
		_ = w.Write(header)
		for _, row := range rows {
			rec := make([]string, len(row))
			for i, v := range row {
				rec[i] = fmt.Sprint(v)
			}
			_ = w.Write(rec)
		}
		w.Flush()
		if err := w.Error(); err != nil {
			return nil, err
		}
		return &ExportFile{Filename: base + ".csv", ContentType: "text/csv; charset=utf-8", Data: buf.Bytes()}, nil
	case "xlsx":
		data, err := buildXLSX(reportType, header, rows)
		if err != nil {
			return nil, err
		}
		return &ExportFile{Filename: base + ".xlsx",
			ContentType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", Data: data}, nil
	}
	return nil, apperror.Validation([]apperror.FieldError{{Field: "format", Message: "harus salah satu dari: csv, xlsx"}})
}

func (s *ReportService) exportTable(ctx context.Context, reportType string, r domain.ReportRange) ([]string, [][]any, error) {
	switch reportType {
	case "sales-daily", "sales-monthly":
		data, err := s.reports.SalesByPeriod(ctx, r, reportType == "sales-monthly")
		if err != nil {
			return nil, nil, err
		}
		header := []string{"Periode", "Jumlah Transaksi", "Penjualan Kotor", "Diskon", "Penjualan Bersih", "Pajak",
			"Total Diterima", "HPP", "Laba Kotor"}
		rows := make([][]any, len(data))
		for i, d := range data {
			rows[i] = []any{d.Period, d.TransactionCount, d.GrossSales, d.DiscountTotal, d.NetSales, d.TaxTotal,
				d.TotalRevenue, d.CostTotal, d.GrossProfit}
		}
		return header, rows, nil
	case "top-products":
		data, err := s.reports.TopProducts(ctx, r, 100)
		if err != nil {
			return nil, nil, err
		}
		rows := make([][]any, len(data))
		for i, d := range data {
			rows[i] = []any{i + 1, d.SKU, d.ProductName, d.Quantity, d.Revenue, d.Profit}
		}
		return []string{"Peringkat", "SKU", "Produk", "Qty Terjual", "Pendapatan", "Laba Kotor"}, rows, nil
	case "payment-methods":
		data, err := s.reports.ByPaymentMethod(ctx, r)
		if err != nil {
			return nil, nil, err
		}
		rows := make([][]any, len(data))
		for i, d := range data {
			rows[i] = []any{paymentLabels[d.Method], d.TransactionCount, d.Amount}
		}
		return []string{"Metode Bayar", "Jumlah Transaksi", "Nominal"}, rows, nil
	case "cashiers":
		data, err := s.reports.ByCashier(ctx, r)
		if err != nil {
			return nil, nil, err
		}
		rows := make([][]any, len(data))
		for i, d := range data {
			rows[i] = []any{d.UserName, d.TransactionCount, d.TotalSales}
		}
		return []string{"Kasir", "Jumlah Transaksi", "Total Penjualan"}, rows, nil
	}
	return nil, nil, apperror.Validation([]apperror.FieldError{{Field: "type",
		Message: "harus salah satu dari: sales-daily, sales-monthly, top-products, payment-methods, cashiers"}})
}

func buildXLSX(title string, header []string, rows [][]any) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()
	sheet := "Laporan"
	if err := f.SetSheetName("Sheet1", sheet); err != nil {
		return nil, err
	}
	bold, err := f.NewStyle(&excelize.Style{Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E5E5EA"}, Pattern: 1}})
	if err != nil {
		return nil, err
	}
	money, err := f.NewStyle(&excelize.Style{NumFmt: 3}) // #,##0
	if err != nil {
		return nil, err
	}
	for c, h := range header {
		cell, _ := excelize.CoordinatesToCellName(c+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, bold)
		col, _ := excelize.ColumnNumberToName(c + 1)
		_ = f.SetColWidth(sheet, col, col, 18)
	}
	for r, row := range rows {
		for c, v := range row {
			cell, _ := excelize.CoordinatesToCellName(c+1, r+2)
			_ = f.SetCellValue(sheet, cell, v)
			if _, isInt := v.(int64); isInt {
				_ = f.SetCellStyle(sheet, cell, cell, money)
			}
		}
	}
	_ = f.SetDocProps(&excelize.DocProperties{Title: title, Creator: "POS UMKM"})
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
