package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/response"
)

type SettingHandler struct {
	svc           *service.SettingService
	maxUploadSize int64
}

func NewSettingHandler(svc *service.SettingService, maxUploadSize int64) *SettingHandler {
	return &SettingHandler{svc: svc, maxUploadSize: maxUploadSize}
}

// Get godoc
// @Summary  Pengaturan toko
// @Tags     Settings
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} response.Body{data=dto.StoreSettingResponse}
// @Router   /settings [get]
func (h *SettingHandler) Get(c *gin.Context) {
	st, err := h.svc.Get(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToStoreSettingResponse(st))
}

// Update godoc
// @Summary  Ubah pengaturan toko
// @Description tax_rate_bp dalam basis point (1100 = 11%).
// @Tags     Settings
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.UpdateStoreSettingRequest true "Pengaturan"
// @Success  200 {object} response.Body{data=dto.StoreSettingResponse}
// @Router   /settings [put]
func (h *SettingHandler) Update(c *gin.Context) {
	var req dto.UpdateStoreSettingRequest
	if !bindJSON(c, &req) {
		return
	}
	st, err := h.svc.Update(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Pengaturan toko berhasil disimpan", dto.ToStoreSettingResponse(st))
}

// UploadLogo godoc
// @Summary  Unggah logo toko
// @Tags     Settings
// @Security BearerAuth
// @Accept   multipart/form-data
// @Produce  json
// @Param    logo formData file true "JPG/PNG/WEBP, maks 2 MB"
// @Success  200 {object} response.Body{data=dto.StoreSettingResponse}
// @Router   /settings/logo [post]
func (h *SettingHandler) UploadLogo(c *gin.Context) {
	file, contentType, ok := readImage(c, "logo", h.maxUploadSize)
	if !ok {
		return
	}
	defer file.Close()
	st, err := h.svc.UploadLogo(c.Request.Context(), contentType, file)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Logo berhasil diunggah", dto.ToStoreSettingResponse(st))
}
