package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/response"
)

// ---------- Category ----------

type CategoryHandler struct{ svc *service.CategoryService }

func NewCategoryHandler(svc *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

// List godoc
// @Summary  Daftar kategori
// @Tags     Categories
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman" default(10)
// @Param    search query string false "Cari nama/deskripsi"
// @Param    sort query string false "name|created_at"
// @Param    order query string false "asc|desc"
// @Success  200 {object} response.Body{data=[]dto.CategoryResponse}
// @Router   /categories [get]
func (h *CategoryHandler) List(c *gin.Context) {
	p := listParams(c)
	items, total, err := h.svc.List(c.Request.Context(), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToCategoryResponses(items), p.Meta(total))
}

// Get godoc
// @Summary  Detail kategori
// @Tags     Categories
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID kategori"
// @Success  200 {object} response.Body{data=dto.CategoryResponse}
// @Router   /categories/{id} [get]
func (h *CategoryHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToCategoryResponse(item))
}

// Create godoc
// @Summary  Tambah kategori
// @Tags     Categories
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CategoryRequest true "Data kategori"
// @Success  201 {object} response.Body{data=dto.CategoryResponse}
// @Router   /categories [post]
func (h *CategoryHandler) Create(c *gin.Context) {
	var req dto.CategoryRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Kategori berhasil ditambahkan", dto.ToCategoryResponse(item))
}

// Update godoc
// @Summary  Ubah kategori
// @Tags     Categories
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id path int true "ID kategori"
// @Param    body body dto.CategoryRequest true "Data kategori"
// @Success  200 {object} response.Body{data=dto.CategoryResponse}
// @Router   /categories/{id} [put]
func (h *CategoryHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.CategoryRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Kategori berhasil diperbarui", dto.ToCategoryResponse(item))
}

// Delete godoc
// @Summary  Hapus kategori (soft delete)
// @Tags     Categories
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID kategori"
// @Success  200 {object} response.Body
// @Failure  409 {object} response.ErrorBody "Masih dipakai produk"
// @Router   /categories/{id} [delete]
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Kategori berhasil dihapus", nil)
}

// ---------- Unit ----------

type UnitHandler struct{ svc *service.UnitService }

func NewUnitHandler(svc *service.UnitService) *UnitHandler { return &UnitHandler{svc: svc} }

// List godoc
// @Summary  Daftar satuan
// @Tags     Units
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman" default(10)
// @Param    search query string false "Cari nama/singkatan"
// @Param    sort query string false "name|abbreviation|created_at"
// @Param    order query string false "asc|desc"
// @Success  200 {object} response.Body{data=[]dto.UnitResponse}
// @Router   /units [get]
func (h *UnitHandler) List(c *gin.Context) {
	p := listParams(c)
	items, total, err := h.svc.List(c.Request.Context(), p)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToUnitResponses(items), p.Meta(total))
}

// Get godoc
// @Summary  Detail satuan
// @Tags     Units
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID satuan"
// @Success  200 {object} response.Body{data=dto.UnitResponse}
// @Router   /units/{id} [get]
func (h *UnitHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToUnitResponse(item))
}

// Create godoc
// @Summary  Tambah satuan
// @Tags     Units
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.UnitRequest true "Data satuan"
// @Success  201 {object} response.Body{data=dto.UnitResponse}
// @Router   /units [post]
func (h *UnitHandler) Create(c *gin.Context) {
	var req dto.UnitRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Satuan berhasil ditambahkan", dto.ToUnitResponse(item))
}

// Update godoc
// @Summary  Ubah satuan
// @Tags     Units
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id path int true "ID satuan"
// @Param    body body dto.UnitRequest true "Data satuan"
// @Success  200 {object} response.Body{data=dto.UnitResponse}
// @Router   /units/{id} [put]
func (h *UnitHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.UnitRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Satuan berhasil diperbarui", dto.ToUnitResponse(item))
}

// Delete godoc
// @Summary  Hapus satuan (soft delete)
// @Tags     Units
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID satuan"
// @Success  200 {object} response.Body
// @Router   /units/{id} [delete]
func (h *UnitHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Satuan berhasil dihapus", nil)
}

// ---------- Product ----------

type ProductHandler struct {
	svc           *service.ProductService
	maxUploadSize int64
}

func NewProductHandler(svc *service.ProductService, maxUploadSize int64) *ProductHandler {
	return &ProductHandler{svc: svc, maxUploadSize: maxUploadSize}
}

// List godoc
// @Summary  Daftar produk
// @Tags     Products
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman" default(10)
// @Param    search query string false "Cari nama/SKU/barcode"
// @Param    sort query string false "name|sku|price|cost|stock|category|created_at|updated_at"
// @Param    order query string false "asc|desc"
// @Param    category_id query int false "Filter kategori"
// @Param    is_active query bool false "Filter status aktif"
// @Param    low_stock query bool false "Hanya stok menipis"
// @Success  200 {object} response.Body{data=[]dto.ProductResponse}
// @Router   /products [get]
func (h *ProductHandler) List(c *gin.Context) {
	f, err := productFilter(c)
	if err != nil {
		response.Error(c, err)
		return
	}
	items, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToProductResponses(items), f.Meta(total))
}

func productFilter(c *gin.Context) (domain.ProductFilter, error) {
	f := domain.ProductFilter{Params: listParams(c)}
	var err error
	if f.CategoryID, err = queryInt64(c, "category_id"); err != nil {
		return f, err
	}
	if f.IsActive, err = queryBool(c, "is_active"); err != nil {
		return f, err
	}
	low, err := queryBool(c, "low_stock")
	if err != nil {
		return f, err
	}
	f.LowStock = low != nil && *low
	return f, nil
}

// Get godoc
// @Summary  Detail produk
// @Tags     Products
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID produk"
// @Success  200 {object} response.Body{data=dto.ProductResponse}
// @Router   /products/{id} [get]
func (h *ProductHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	item, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToProductResponse(item))
}

// GetByBarcode godoc
// @Summary  Cari produk berdasarkan barcode atau SKU
// @Tags     Products
// @Security BearerAuth
// @Produce  json
// @Param    code path string true "Barcode / SKU"
// @Success  200 {object} response.Body{data=dto.ProductResponse}
// @Failure  404 {object} response.ErrorBody
// @Router   /products/barcode/{code} [get]
func (h *ProductHandler) GetByBarcode(c *gin.Context) {
	item, err := h.svc.GetByBarcode(c.Request.Context(), c.Param("code"))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToProductResponse(item))
}

// Create godoc
// @Summary  Tambah produk
// @Tags     Products
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateProductRequest true "Data produk (harga dalam rupiah)"
// @Success  201 {object} response.Body{data=dto.ProductResponse}
// @Failure  422 {object} response.ErrorBody
// @Router   /products [post]
func (h *ProductHandler) Create(c *gin.Context) {
	var req dto.CreateProductRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.svc.Create(c.Request.Context(), actor(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Produk berhasil ditambahkan", dto.ToProductResponse(item))
}

// Update godoc
// @Summary  Ubah produk
// @Description Stok tidak bisa diubah di sini; gunakan endpoint stok masuk / penyesuaian.
// @Tags     Products
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id path int true "ID produk"
// @Param    body body dto.UpdateProductRequest true "Data produk"
// @Success  200 {object} response.Body{data=dto.ProductResponse}
// @Router   /products/{id} [put]
func (h *ProductHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.UpdateProductRequest
	if !bindJSON(c, &req) {
		return
	}
	item, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Produk berhasil diperbarui", dto.ToProductResponse(item))
}

// Delete godoc
// @Summary  Hapus produk (soft delete)
// @Tags     Products
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID produk"
// @Success  200 {object} response.Body
// @Router   /products/{id} [delete]
func (h *ProductHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Produk berhasil dihapus", nil)
}

// UploadImage godoc
// @Summary  Unggah gambar produk
// @Tags     Products
// @Security BearerAuth
// @Accept   multipart/form-data
// @Produce  json
// @Param    id path int true "ID produk"
// @Param    image formData file true "JPG/PNG/WEBP, maks 2 MB"
// @Success  200 {object} response.Body{data=dto.ProductResponse}
// @Router   /products/{id}/image [post]
func (h *ProductHandler) UploadImage(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	file, contentType, ok := readImage(c, "image", h.maxUploadSize)
	if !ok {
		return
	}
	defer file.Close()
	item, err := h.svc.UploadImage(c.Request.Context(), id, contentType, file)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Gambar produk berhasil diunggah", dto.ToProductResponse(item))
}

type readSeekCloser interface {
	Read(p []byte) (int, error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
}

// readImage opens a multipart file and sniffs its real content type (the client header is not trusted).
func readImage(c *gin.Context, field string, maxSize int64) (readSeekCloser, string, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize+1<<20)
	fh, err := c.FormFile(field)
	if err != nil {
		response.Error(c, apperror.Validation([]apperror.FieldError{{Field: field, Message: "file wajib diunggah"}}))
		return nil, "", false
	}
	if fh.Size > maxSize {
		response.Error(c, apperror.Validation([]apperror.FieldError{{Field: field,
			Message: "ukuran file maksimal " + formatMB(maxSize)}}))
		return nil, "", false
	}
	f, err := fh.Open()
	if err != nil {
		response.Error(c, err)
		return nil, "", false
	}
	head := make([]byte, 512)
	n, _ := f.Read(head)
	if _, err := f.Seek(0, 0); err != nil {
		f.Close()
		response.Error(c, err)
		return nil, "", false
	}
	return f, http.DetectContentType(head[:n]), true
}

func formatMB(b int64) string {
	return fmt.Sprintf("%d MB", b>>20)
}
