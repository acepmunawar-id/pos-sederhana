package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/response"
)

type ShiftHandler struct {
	svc *service.ShiftService
	loc *time.Location
}

func NewShiftHandler(svc *service.ShiftService, loc *time.Location) *ShiftHandler {
	return &ShiftHandler{svc: svc, loc: loc}
}

// Open godoc
// @Summary  Buka shift
// @Tags     Shifts
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.OpenShiftRequest true "Kas awal"
// @Success  201 {object} response.Body{data=dto.ShiftResponse}
// @Failure  409 {object} response.ErrorBody "Masih ada shift terbuka"
// @Router   /shifts/open [post]
func (h *ShiftHandler) Open(c *gin.Context) {
	var req dto.OpenShiftRequest
	if !bindJSON(c, &req) {
		return
	}
	sh, err := h.svc.Open(c.Request.Context(), actor(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Shift berhasil dibuka", dto.ToShiftResponse(sh))
}

// Current godoc
// @Summary  Shift yang sedang berjalan
// @Tags     Shifts
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} response.Body{data=dto.ShiftDetailResponse}
// @Failure  404 {object} response.ErrorBody "Belum buka shift"
// @Router   /shifts/current [get]
func (h *ShiftHandler) Current(c *gin.Context) {
	sh, sum, err := h.svc.Current(c.Request.Context(), actor(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToShiftDetailResponse(sh, sum))
}

// Close godoc
// @Summary  Tutup shift
// @Description Hitung kas yang diharapkan (kas awal + penjualan tunai) dan selisih terhadap kas aktual.
// @Tags     Shifts
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CloseShiftRequest true "Kas aktual di laci"
// @Success  200 {object} response.Body{data=dto.ShiftDetailResponse}
// @Router   /shifts/current/close [post]
func (h *ShiftHandler) Close(c *gin.Context) {
	var req dto.CloseShiftRequest
	if !bindJSON(c, &req) {
		return
	}
	sh, sum, err := h.svc.Close(c.Request.Context(), actor(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Shift berhasil ditutup", dto.ToShiftDetailResponse(sh, sum))
}

// List godoc
// @Summary  Riwayat shift
// @Description Kasir hanya melihat shift miliknya sendiri.
// @Tags     Shifts
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman" default(10)
// @Param    search query string false "Cari nama kasir"
// @Param    sort query string false "opened_at|closed_at|cashier"
// @Param    order query string false "asc|desc"
// @Param    user_id query int false "Filter kasir"
// @Param    status query string false "OPEN|CLOSED"
// @Param    date_from query string false "YYYY-MM-DD"
// @Param    date_to query string false "YYYY-MM-DD"
// @Success  200 {object} response.Body{data=[]dto.ShiftResponse}
// @Router   /shifts [get]
func (h *ShiftHandler) List(c *gin.Context) {
	f := domain.ShiftFilter{Params: listParams(c)}
	var err error
	if f.UserID, err = queryInt64(c, "user_id"); err != nil {
		response.Error(c, err)
		return
	}
	if f.Status, err = queryEnum(c, "status", domain.ShiftOpen, domain.ShiftClosed); err != nil {
		response.Error(c, err)
		return
	}
	if f.DateFrom, f.DateTo, err = dateFilter(c, h.loc); err != nil {
		response.Error(c, err)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), actor(c), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToShiftResponses(items), f.Meta(total))
}

// Get godoc
// @Summary  Detail shift
// @Tags     Shifts
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID shift"
// @Success  200 {object} response.Body{data=dto.ShiftDetailResponse}
// @Router   /shifts/{id} [get]
func (h *ShiftHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	sh, sum, err := h.svc.Get(c.Request.Context(), actor(c), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToShiftDetailResponse(sh, sum))
}
