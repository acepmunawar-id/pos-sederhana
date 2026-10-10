package domain

import (
	"context"
	"time"

	"github.com/posumkm/backend/pkg/pagination"
)

type Category struct {
	ID           int64
	Name         string
	Description  *string
	ProductCount int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Unit struct {
	ID           int64
	Name         string
	Abbreviation string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Product struct {
	ID           int64
	CategoryID   int64
	CategoryName string
	UnitID       int64
	UnitName     string
	SKU          string
	Barcode      *string
	Name         string
	Price        int64
	Cost         int64
	Stock        int
	MinStock     int
	ImageURL     *string
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (p Product) IsLowStock() bool { return p.Stock <= p.MinStock }

type ProductFilter struct {
	pagination.Params
	CategoryID *int64
	IsActive   *bool
	LowStock   bool
}

type CategoryRepository interface {
	Create(ctx context.Context, c *Category) error
	Update(ctx context.Context, c *Category) error
	SoftDelete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Category, error)
	NameExists(ctx context.Context, name string, excludeID int64) (bool, error)
	CountProducts(ctx context.Context, id int64) (int64, error)
	List(ctx context.Context, p pagination.Params) ([]Category, int64, error)
}

type UnitRepository interface {
	Create(ctx context.Context, u *Unit) error
	Update(ctx context.Context, u *Unit) error
	SoftDelete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Unit, error)
	NameExists(ctx context.Context, name string, excludeID int64) (bool, error)
	CountProducts(ctx context.Context, id int64) (int64, error)
	List(ctx context.Context, p pagination.Params) ([]Unit, int64, error)
}

type ProductRepository interface {
	Create(ctx context.Context, p *Product) error
	Update(ctx context.Context, p *Product) error
	UpdateImage(ctx context.Context, id int64, url *string) error
	UpdateStock(ctx context.Context, id int64, stock int) error
	SoftDelete(ctx context.Context, id int64) error
	FindByID(ctx context.Context, id int64) (*Product, error)
	FindByBarcode(ctx context.Context, barcode string) (*Product, error)
	// FindByIDsForUpdate locks the rows (SELECT ... FOR UPDATE) in ascending id order.
	// Must be called inside TxManager.WithinTx.
	FindByIDsForUpdate(ctx context.Context, ids []int64) ([]Product, error)
	SKUExists(ctx context.Context, sku string, excludeID int64) (bool, error)
	BarcodeExists(ctx context.Context, barcode string, excludeID int64) (bool, error)
	CountLowStock(ctx context.Context) (int64, error)
	List(ctx context.Context, f ProductFilter) ([]Product, int64, error)
}
