package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/internal/middleware"
	"github.com/posumkm/backend/internal/service"
	"github.com/posumkm/backend/pkg/response"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func client(c *gin.Context) service.ClientInfo {
	return service.ClientInfo{UserAgent: c.Request.UserAgent(), IP: c.ClientIP()}
}

// Login godoc
// @Summary      Login
// @Description  Masuk dengan email & password. Mengembalikan access token (JWT) dan refresh token.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.LoginRequest true "Kredensial"
// @Success      200 {object} response.Body{data=dto.TokenResponse}
// @Failure      401 {object} response.ErrorBody
// @Failure      422 {object} response.ErrorBody
// @Router       /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req, client(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Login berhasil", dto.ToTokenResponse(res))
}

// Refresh godoc
// @Summary      Perbarui token
// @Description  Tukar refresh token dengan pasangan token baru (rotasi). Refresh token lama tidak berlaku lagi.
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        body body dto.RefreshRequest true "Refresh token"
// @Success      200 {object} response.Body{data=dto.TokenResponse}
// @Failure      401 {object} response.ErrorBody
// @Router       /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req dto.RefreshRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.svc.Refresh(c.Request.Context(), req.RefreshToken, client(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Token diperbarui", dto.ToTokenResponse(res))
}

// Logout godoc
// @Summary      Logout
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.LogoutRequest true "Refresh token yang akan dicabut"
// @Success      200 {object} response.Body
// @Router       /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	var req dto.LogoutRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.Logout(c.Request.Context(), middleware.UserID(c), req.RefreshToken); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Logout berhasil", nil)
}

// Me godoc
// @Summary      Profil pengguna yang sedang login
// @Tags         Auth
// @Security     BearerAuth
// @Produce      json
// @Success      200 {object} response.Body{data=dto.UserResponse}
// @Router       /auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	u, err := h.svc.Me(c.Request.Context(), middleware.UserID(c))
	if err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Berhasil", dto.ToUserResponse(u))
}

// ChangePassword godoc
// @Summary      Ganti password
// @Description  Semua sesi lain akan logout setelah password diganti.
// @Tags         Auth
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body body dto.ChangePasswordRequest true "Password lama & baru"
// @Success      200 {object} response.Body
// @Failure      422 {object} response.ErrorBody
// @Router       /auth/password [put]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req dto.ChangePasswordRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), middleware.UserID(c), req); err != nil {
		response.Error(c, err)
		return
	}
	response.OK(c, "Password berhasil diganti. Silakan login kembali", nil)
}
