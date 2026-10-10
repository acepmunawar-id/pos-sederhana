package dto

import "github.com/posumkm/backend/internal/domain"

type SalesSummaryResponse struct {
	TransactionCount int64 `json:"transaction_count"`
	GrossSales       int64 `json:"gross_sales"`
	DiscountTotal    int64 `json:"discount_total"`
	TaxTotal         int64 `json:"tax_total"`
	NetSales         int64 `json:"net_sales"`
	TotalRevenue     int64 `json:"total_revenue"`
	CostTotal        int64 `json:"cost_total"`
	GrossProfit      int64 `json:"gross_profit"`
	// Gross margin in basis points (2500 = 25%).
	MarginBP int64 `json:"margin_bp"`
}

func ToSalesSummaryResponse(s domain.SalesSummary) SalesSummaryResponse {
	var margin int64
	if s.NetSales > 0 {
		margin = s.GrossProfit * domain.BasisPoints / s.NetSales
	}
	return SalesSummaryResponse{
		TransactionCount: s.TransactionCount, GrossSales: s.GrossSales, DiscountTotal: s.DiscountTotal,
		TaxTotal: s.TaxTotal, NetSales: s.NetSales, TotalRevenue: s.TotalRevenue, CostTotal: s.CostTotal,
		GrossProfit: s.GrossProfit, MarginBP: margin,
	}
}

type SalesPeriodResponse struct {
	Period string `json:"period" example:"2026-10-10"`
	SalesSummaryResponse
}

type SalesReportResponse struct {
	From    string                `json:"from"`
	To      string                `json:"to"`
	Group   string                `json:"group" enums:"daily,monthly"`
	Summary SalesSummaryResponse  `json:"summary"`
	Rows    []SalesPeriodResponse `json:"rows"`
}

type TopProductResponse struct {
	ProductID   int64  `json:"product_id"`
	ProductName string `json:"product_name"`
	SKU         string `json:"sku"`
	Quantity    int64  `json:"quantity"`
	Revenue     int64  `json:"revenue"`
	Profit      int64  `json:"profit"`
}

func ToTopProductResponses(rows []domain.TopProductRow) []TopProductResponse {
	out := make([]TopProductResponse, len(rows))
	for i, r := range rows {
		out[i] = TopProductResponse{ProductID: r.ProductID, ProductName: r.ProductName, SKU: r.SKU,
			Quantity: r.Quantity, Revenue: r.Revenue, Profit: r.Profit}
	}
	return out
}

type PaymentMethodResponse struct {
	Method           string `json:"method"`
	TransactionCount int64  `json:"transaction_count"`
	Amount           int64  `json:"amount"`
}

func ToPaymentMethodResponses(rows []domain.PaymentMethodRow) []PaymentMethodResponse {
	out := make([]PaymentMethodResponse, len(rows))
	for i, r := range rows {
		out[i] = PaymentMethodResponse{Method: r.Method, TransactionCount: r.TransactionCount, Amount: r.Amount}
	}
	return out
}

type CashierReportResponse struct {
	UserID           int64  `json:"user_id"`
	UserName         string `json:"user_name"`
	TransactionCount int64  `json:"transaction_count"`
	TotalSales       int64  `json:"total_sales"`
}

func ToCashierReportResponses(rows []domain.CashierRow) []CashierReportResponse {
	out := make([]CashierReportResponse, len(rows))
	for i, r := range rows {
		out[i] = CashierReportResponse{UserID: r.UserID, UserName: r.UserName,
			TransactionCount: r.TransactionCount, TotalSales: r.TotalSales}
	}
	return out
}

type DailyPoint struct {
	Date             string `json:"date" example:"2026-10-10"`
	TotalRevenue     int64  `json:"total_revenue"`
	TransactionCount int64  `json:"transaction_count"`
}

type DashboardResponse struct {
	Today         SalesSummaryResponse `json:"today"`
	Yesterday     SalesSummaryResponse `json:"yesterday"`
	Last7Days     []DailyPoint         `json:"last_7_days"`
	TopProducts   []TopProductResponse `json:"top_products"`
	LowStock      []ProductResponse    `json:"low_stock"`
	LowStockCount int64                `json:"low_stock_count"`
}
