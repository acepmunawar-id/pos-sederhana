package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/response"
)

type UserHandler struct{ svc *service.UserService }

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc: svc} }

// ListRoles godoc
// @Summary  Daftar role
// @Tags     Users
// @Security BearerAuth
// @Produce  json
// @Success  200 {object} response.Body{data=[]dto.RoleResponse}
// @Router   /roles [get]
func (h *UserHandler) ListRoles(c *gin.Context) {
	roles, err := h.svc.ListRoles(c.Request.Context())
	if err != nil {
		response.Error(c, err)
		return
	}
	out := make([]dto.RoleResponse, len(roles))
	for i, r := range roles {
		out[i] = dto.ToRoleResponse(r)
	}
	response.OK(c, "Berhasil", out)
}

// List godoc
// @Summary  Daftar pengguna
// @Tags     Users
// @Security BearerAuth
// @Produce  json
// @Param    page query int false "Halaman" default(1)
// @Param    limit query int false "Jumlah per halaman (maks 100)" default(10)
// @Param    search query string false "Cari nama/email"
// @Param    sort query string false "name|email|created_at|last_login_at"
// @Param    order query string false "asc|desc"
// @Param    role query string false "owner|admin|cashier"
// @Param    is_active query bool false "Filter status aktif"
// @Success  200 {object} response.Body{data=[]dto.UserResponse}
// @Router   /users [get]
func (h *UserHandler) List(c *gin.Context) {
	isActive, err := queryBool(c, "is_active")
	if err != nil {
		response.Error(c, err)
		return
	}
	role, err := queryEnum(c, "role", "OWNER", "ADMIN", "CASHIER")
	if err != nil {
		response.Error(c, err)
		return
	}
	f := domain.UserFilter{Params: listParams(c), IsActive: isActive}
	if role != "" {
		f.RoleCode = map[string]string{"OWNER": domain.RoleOwner, "ADMIN": domain.RoleAdmin, "CASHIER": domain.RoleCashier}[role]
	}
	users, total, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Paginated(c, "Berhasil", dto.ToUserResponses(users), f.Meta(total))
}

// Get godoc
// @Summary  Detail pengguna
// @Tags     Users
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID pengguna"
// @Success  200 {object} response.Body{data=dto.UserResponse}
// @Failure  404 {object} response.ErrorBody
// @Router   /users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToUserResponse(u))
}

// Create godoc
// @Summary  Tambah pengguna
// @Tags     Users
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    body body dto.CreateUserRequest true "Data pengguna"
// @Success  201 {object} response.Body{data=dto.UserResponse}
// @Failure  422 {object} response.ErrorBody
// @Router   /users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req dto.CreateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	u, err := h.svc.Create(c.Request.Context(), actor(c), req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Created(c, "Pengguna berhasil ditambahkan", dto.ToUserResponse(u))
}

// Update godoc
// @Summary  Ubah pengguna
// @Tags     Users
// @Security BearerAuth
// @Accept   json
// @Produce  json
// @Param    id path int true "ID pengguna"
// @Param    body body dto.UpdateUserRequest true "Data pengguna"
// @Success  200 {object} response.Body{data=dto.UserResponse}
// @Router   /users/{id} [put]
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	u, err := h.svc.Update(c.Request.Context(), actor(c), id, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Pengguna berhasil diperbarui", dto.ToUserResponse(u))
}

// Delete godoc
// @Summary  Hapus pengguna (soft delete)
// @Tags     Users
// @Security BearerAuth
// @Produce  json
// @Param    id path int true "ID pengguna"
// @Success  200 {object} response.Body
// @Router   /users/{id} [delete]
func (h *UserHandler) Delete(c *gin.Context) {
	id, ok := pathID(c)
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), actor(c), id); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Pengguna berhasil dihapus", nil)
}
