package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/response"
)

type StockHandler struct {
	svc *service.StockService
	loc *time.Location
}

func NewStockHandler(svc *service.StockService, loc *time.Location) *StockHandler {
	return &StockHandler{svc: svc, loc: loc}
}

// StockIn godoc
// @Summary  Stok masuk
// @Description Tambah stok satu atau beberapa produk sekaligus (mis. barang datang dari supplier).
// @Tags     Stock
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.StockInRequest true "Daftar produk & jumlah"
// @Success  201 {object} response.Body{data=[]dto.StockMovementResponse}
// @Router   /stock/in [post]
func (h *StockHandler) StockIn(c *gin.Context) {
	var req dto.StockInRequest
	if !bindJSON(c, &req) {
		return
	}
	items, err := h.svc.StockIn(c.Request.Context(), actor(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Stok masuk berhasil dicatat", dto.ToStockMovementResponses(items))
}

// Adjust godoc
// @Summary  Penyesuaian stok (stock opname)
// @Description Set stok ke jumlah fisik hasil hitung. Selisih dicatat sebagai pergerakan ADJUSTMENT.
// @Tags     Stock
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.StockAdjustmentRequest true "Stok aktual & alasan"
// @Success  201 {object} response.Body{data=dto.StockMovementResponse}
// @Router   /stock/adjustments [post]
func (h *StockHandler) Adjust(c *gin.Context) {
	var req dto.StockAdjustmentRequest
	if !bindJSON(c, &req) {
		return
	}
	m, err := h.svc.Adjust(c.Request.Context(), actor(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Stok berhasil disesuaikan", dto.ToStockMovementResponse(m))
}

// Movements godoc
// @Summary  Riwayat pergerakan stok
// @Tags     Stock
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman" default(10)
// @Param    search query string false "Cari nama/SKU produk atau catatan"
// @Param    sort query string false "created_at|quantity|product|type"
// @Param    order query string false "asc|desc"
// @Param    product_id query int false "Filter produk"
// @Param    type query string false "IN|ADJUSTMENT|SALE|VOID"
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Success  200 {object} response.Body{data=[]dto.StockMovementResponse}
// @Router   /stock/movements [get]
func (h *StockHandler) Movements(c *gin.Context) {
	f := domain.StockMovementFilter{Params: listParams(c)}
	var err error
	if f.ProductID, err = queryInt64(c, "product_id"); err != nil {
		response.Error(c, err)
		return
	}
	if f.Type, err = queryEnum(c, "type", domain.MovementIn, domain.MovementAdjustment, domain.MovementSale, domain.MovementVoid); err != nil {
		response.Error(c, err)
		return
	}
	if f.DateFrom, f.DateTo, err = dateFilter(c, h.loc); err != nil {
		response.Error(c, err)
		return
	}
	items, total, err := h.svc.Movements(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToStockMovementResponses(items), f.Meta(total))
}

// LowStock godoc
// @Summary  Produk dengan stok menipis
// @Description Produk aktif dengan stok <= stok minimum.
// @Tags     Stock
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman" default(10)
// @Param    search query string false "Cari nama/SKU"
// @Param    category_id query int false "Filter kategori"
// @Success  200 {object} response.Body{data=[]dto.ProductResponse}
// @Router   /stock/low [get]
func (h *StockHandler) LowStock(c *gin.Context) {
	f, err := productFilter(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	items, total, err := h.svc.LowStock(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToProductResponses(items), f.Meta(total))
}
