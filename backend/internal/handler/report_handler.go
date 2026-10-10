package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/response"
)

type ReportHandler struct{ svc *service.ReportService }

func NewReportHandler(svc *service.ReportService) *ReportHandler { return &ReportHandler{svc: svc} }

// Dashboard godoc
// @Summary  Ringkasan dashboard
// @Description Penjualan hari ini & kemarin, grafik 7 hari, 5 produk terlaris (7 hari), stok menipis.
// @Tags     Reports
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} response.Body{data=dto.DashboardResponse}
// @Router   /reports/dashboard [get]
func (h *ReportHandler) Dashboard(c *gin.Context) {
	res, err := h.svc.Dashboard(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", res)
}

// Sales godoc
// @Summary  Laporan penjualan harian / bulanan
// @Tags     Reports
// @Security BearerAuth
// @Produce  json
// @Param    date_from query string false "YYYY-MM-DD (default 30 hari terakhir)"
// @Param    date_to query string false "YYYY-MM-DD (default hari ini)"
// @Param    group query string false "daily|monthly" default(daily)
// @Success  200 {object} response.Body{data=dto.SalesReportResponse}
// @Router   /reports/sales [get]
func (h *ReportHandler) Sales(c *gin.Context) {
	rg, err := h.svc.ParseRange(c.Query("date_from"), c.Query("date_to"))
	if err != nil {
		response.Error(c, err)
		return
	}
	group := c.DefaultQuery("group", "daily")
	if group != "daily" && group != "monthly" {
		response.Error(c, apperror.Validation([]apperror.FieldError{{Field: "group", Message: "harus salah satu dari: daily, monthly"}}))
		return
	}
	res, err := h.svc.Sales(c.Request.Context(), rg, group == "monthly")
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", res)
}

// GrossProfit godoc
// @Summary  Laba kotor
// @Description Laba kotor = (penjualan − diskon) − HPP. Pajak tidak dihitung sebagai pendapatan.
// @Tags     Reports
// @Security BearerAuth
// @Produce  json
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Success  200 {object} response.Body{data=dto.SalesSummaryResponse}
// @Router   /reports/gross-profit [get]
func (h *ReportHandler) GrossProfit(c *gin.Context) {
	rg, err := h.svc.ParseRange(c.Query("date_from"), c.Query("date_to"))
	if err != nil {
		response.Error(c, err)
		return
	}
	res, err := h.svc.GrossProfit(c.Request.Context(), rg)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", res)
}

// TopProducts godoc
// @Summary  Produk terlaris
// @Tags     Reports
// @Security BearerAuth
// @Produce  json
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Param    limit query int false "Jumlah produk (maks 100)" default(10)
// @Success  200 {object} response.Body{data=[]dto.TopProductResponse}
// @Router   /reports/top-products [get]
func (h *ReportHandler) TopProducts(c *gin.Context) {
	rg, err := h.svc.ParseRange(c.Query("date_from"), c.Query("date_to"))
	if err != nil {
		response.Error(c, err)
		return
	}
	limit, _ := strconv.Atoi(c.Query("limit"))
	rows, err := h.svc.TopProducts(c.Request.Context(), rg, limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToTopProductResponses(rows))
}

// PaymentMethods godoc
// @Summary  Penjualan per metode bayar
// @Tags     Reports
// @Security BearerAuth
// @Produce  json
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Success  200 {object} response.Body{data=[]dto.PaymentMethodResponse}
// @Router   /reports/payment-methods [get]
func (h *ReportHandler) PaymentMethods(c *gin.Context) {
	rg, err := h.svc.ParseRange(c.Query("date_from"), c.Query("date_to"))
	if err != nil {
		response.Error(c, err)
		return
	}
	rows, err := h.svc.PaymentMethods(c.Request.Context(), rg)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToPaymentMethodResponses(rows))
}

// Cashiers godoc
// @Summary  Penjualan per kasir
// @Tags     Reports
// @Security BearerAuth
// @Produce  json
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Success  200 {object} response.Body{data=[]dto.CashierReportResponse}
// @Router   /reports/cashiers [get]
func (h *ReportHandler) Cashiers(c *gin.Context) {
	rg, err := h.svc.ParseRange(c.Query("date_from"), c.Query("date_to"))
	if err != nil {
		response.Error(c, err)
		return
	}
	rows, err := h.svc.Cashiers(c.Request.Context(), rg)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToCashierReportResponses(rows))
}

// Export godoc
// @Summary  Export laporan ke CSV / Excel
// @Tags     Reports
// @Security BearerAuth
// @Produce  octet-stream
// @Param    type query string true "sales-daily|sales-monthly|top-products|payment-methods|cashiers"
// @Param    format query string false "csv|xlsx" default(xlsx)
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Success  200 {file} file
// @Router   /reports/export [get]
func (h *ReportHandler) Export(c *gin.Context) {
	rg, err := h.svc.ParseRange(c.Query("date_from"), c.Query("date_to"))
	if err != nil {
		response.Error(c, err)
		return
	}
	file, err := h.svc.Export(c.Request.Context(), c.Query("type"), c.DefaultQuery("format", "xlsx"), rg)
	if err != nil {
		response.Error(c, err)
		return
	}
	c.Header("Content-Disposition", `attachment; filename="`+file.Filename+`"`)
	c.Header("Access-Control-Expose-Headers", "Content-Disposition")
	c.Data(200, file.ContentType, file.Data)
}
