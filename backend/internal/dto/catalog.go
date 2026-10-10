package dto

import (
	"time"

	"github.com/posumkm/backend/internal/domain"
)

type CategoryRequest struct {
	Name        string  `json:"name" validate:"required,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
}

type CategoryResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	ProductCount int64     `json:"product_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToCategoryResponse(c *domain.Category) CategoryResponse {
	return CategoryResponse{ID: c.ID, Name: c.Name, Description: c.Description, ProductCount: c.ProductCount,
		CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt}
}

func ToCategoryResponses(items []domain.Category) []CategoryResponse {
	out := make([]CategoryResponse, len(items))
	for i := range items {
		out[i] = ToCategoryResponse(&items[i])
	}
	return out
}

type UnitRequest struct {
	Name         string `json:"name" validate:"required,min=1,max=50"`
	Abbreviation string `json:"abbreviation" validate:"required,min=1,max=10"`
}

type UnitResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Abbreviation string    `json:"abbreviation"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToUnitResponse(u *domain.Unit) UnitResponse {
	return UnitResponse{ID: u.ID, Name: u.Name, Abbreviation: u.Abbreviation, CreatedAt: u.CreatedAt, UpdatedAt: u.UpdatedAt}
}

func ToUnitResponses(items []domain.Unit) []UnitResponse {
	out := make([]UnitResponse, len(items))
	for i := range items {
		out[i] = ToUnitResponse(&items[i])
	}
	return out
}

type CreateProductRequest struct {
	CategoryID   int64   `json:"category_id" validate:"required,gt=0"`
	UnitID       int64   `json:"unit_id" validate:"required,gt=0"`
	SKU          string  `json:"sku" validate:"required,min=1,max=50"`
	Barcode      *string `json:"barcode" validate:"omitempty,max=50"`
	Name         string  `json:"name" validate:"required,min=2,max=150"`
	Price        int64   `json:"price" validate:"gte=0" example:"18000"`
	Cost         int64   `json:"cost" validate:"gte=0" example:"7000"`
	InitialStock int     `json:"initial_stock" validate:"gte=0"`
	MinStock     int     `json:"min_stock" validate:"gte=0"`
	IsActive     *bool   `json:"is_active"`
}

type UpdateProductRequest struct {
	CategoryID int64   `json:"category_id" validate:"required,gt=0"`
	UnitID     int64   `json:"unit_id" validate:"required,gt=0"`
	SKU        string  `json:"sku" validate:"required,min=1,max=50"`
	Barcode    *string `json:"barcode" validate:"omitempty,max=50"`
	Name       string  `json:"name" validate:"required,min=2,max=150"`
	Price      int64   `json:"price" validate:"gte=0"`
	Cost       int64   `json:"cost" validate:"gte=0"`
	MinStock   int     `json:"min_stock" validate:"gte=0"`
	IsActive   bool    `json:"is_active"`
}

type IDName struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ProductResponse struct {
	ID         int64     `json:"id"`
	SKU        string    `json:"sku"`
	Barcode    *string   `json:"barcode"`
	Name       string    `json:"name"`
	Category   IDName    `json:"category"`
	Unit       IDName    `json:"unit"`
	Price      int64     `json:"price"`
	Cost       int64     `json:"cost"`
	Stock      int       `json:"stock"`
	MinStock   int       `json:"min_stock"`
	IsLowStock bool      `json:"is_low_stock"`
	ImageURL   *string   `json:"image_url"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ToProductResponse(p *domain.Product) ProductResponse {
	return ProductResponse{
		ID: p.ID, SKU: p.SKU, Barcode: p.Barcode, Name: p.Name,
		Category: IDName{ID: p.CategoryID, Name: p.CategoryName},
		Unit:     IDName{ID: p.UnitID, Name: p.UnitName},
		Price:    p.Price, Cost: p.Cost, Stock: p.Stock, MinStock: p.MinStock, IsLowStock: p.IsLowStock(),
		ImageURL: p.ImageURL, IsActive: p.IsActive, CreatedAt: p.CreatedAt, UpdatedAt: p.UpdatedAt,
	}
}

func ToProductResponses(items []domain.Product) []ProductResponse {
	out := make([]ProductResponse, len(items))
	for i := range items {
		out[i] = ToProductResponse(&items[i])
	}
	return out
}
