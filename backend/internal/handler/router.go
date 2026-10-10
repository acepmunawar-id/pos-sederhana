package handler

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/middleware"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/jwt"
	"github.com/posumkm/backend/pkg/response"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handlers struct {
	Auth        *AuthHandler
	User        *UserHandler
	Category    *CategoryHandler
	Unit        *UnitHandler
	Product     *ProductHandler
	Stock       *StockHandler
	Shift       *ShiftHandler
	Transaction *TransactionHandler
	Report      *ReportHandler
	Setting     *SettingHandler
}

type RouterConfig struct {
	Logger         *slog.Logger
	Tokens         *jwt.Manager
	AllowedOrigins []string
	UploadDir      string
	Health         func(ctx context.Context) error
}

func NewRouter(cfg RouterConfig, h Handlers) *gin.Engine {
	r := gin.New()
	r.Use(middleware.RequestID(), middleware.Logger(cfg.Logger), middleware.Recovery(cfg.Logger))
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Idempotency-Key", "X-Request-ID"},
		ExposeHeaders:    []string{"Content-Disposition", "X-Request-ID", "Idempotent-Replayed"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	r.MaxMultipartMemory = 8 << 20

	r.GET("/health", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := cfg.Health(ctx); err != nil {
			response.Error(c, apperror.New(http.StatusServiceUnavailable, "Database tidak dapat dihubungi"))
			return
		}
		response.OK(c, "OK", gin.H{"status": "up"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.Static("/uploads", cfg.UploadDir)
	r.NoRoute(func(c *gin.Context) { response.Error(c, apperror.NotFound("Endpoint tidak ditemukan")) })

	manager := middleware.RequireRoles(domain.RoleOwner, domain.RoleAdmin)
	ownerOnly := middleware.RequireRoles(domain.RoleOwner)

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		auth.POST("/login", h.Auth.Login)
		auth.POST("/refresh", h.Auth.Refresh)
	}

	p := api.Group("", middleware.Auth(cfg.Tokens))
	{
		p.POST("/auth/logout", h.Auth.Logout)
		p.GET("/auth/me", h.Auth.Me)
		p.PUT("/auth/password", h.Auth.ChangePassword)

		p.GET("/roles", manager, h.User.ListRoles)
		users := p.Group("/users", manager)
		users.GET("", h.User.List)
		users.POST("", h.User.Create)
		users.GET("/:id", h.User.Get)
		users.PUT("/:id", h.User.Update)
		users.DELETE("/:id", h.User.Delete)

		p.GET("/categories", h.Category.List)
		p.GET("/categories/:id", h.Category.Get)
		p.POST("/categories", manager, h.Category.Create)
		p.PUT("/categories/:id", manager, h.Category.Update)
		p.DELETE("/categories/:id", manager, h.Category.Delete)

		p.GET("/units", h.Unit.List)
		p.GET("/units/:id", h.Unit.Get)
		p.POST("/units", manager, h.Unit.Create)
		p.PUT("/units/:id", manager, h.Unit.Update)
		p.DELETE("/units/:id", manager, h.Unit.Delete)

		p.GET("/products", h.Product.List)
		p.GET("/products/barcode/:code", h.Product.GetByBarcode)
		p.GET("/products/:id", h.Product.Get)
		p.POST("/products", manager, h.Product.Create)
		p.PUT("/products/:id", manager, h.Product.Update)
		p.DELETE("/products/:id", manager, h.Product.Delete)
		p.POST("/products/:id/image", manager, h.Product.UploadImage)

		p.POST("/stock/in", manager, h.Stock.StockIn)
		p.POST("/stock/adjustments", manager, h.Stock.Adjust)
		p.GET("/stock/movements", h.Stock.Movements)
		p.GET("/stock/low", h.Stock.LowStock)

		p.POST("/shifts/open", h.Shift.Open)
		p.GET("/shifts/current", h.Shift.Current)
		p.POST("/shifts/current/close", h.Shift.Close)
		p.GET("/shifts", h.Shift.List)
		p.GET("/shifts/:id", h.Shift.Get)

		p.POST("/transactions/checkout", h.Transaction.Checkout)
		p.GET("/transactions", h.Transaction.List)
		p.GET("/transactions/:id", h.Transaction.Get)
		p.GET("/transactions/:id/receipt", h.Transaction.Receipt)
		p.POST("/transactions/:id/void", manager, h.Transaction.Void)

		reports := p.Group("/reports", manager)
		reports.GET("/dashboard", h.Report.Dashboard)
		reports.GET("/sales", h.Report.Sales)
		reports.GET("/gross-profit", h.Report.GrossProfit)
		reports.GET("/top-products", h.Report.TopProducts)
		reports.GET("/payment-methods", h.Report.PaymentMethods)
		reports.GET("/cashiers", h.Report.Cashiers)
		reports.GET("/export", h.Report.Export)

		p.GET("/settings", h.Setting.Get)
		p.PUT("/settings", ownerOnly, h.Setting.Update)
		p.POST("/settings/logo", ownerOnly, h.Setting.UploadLogo)
	}
	return r
}
