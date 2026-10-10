package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/internal/dto"
	"github.com/posumkm/backend/pkg/apperror"
	"github.com/posumkm/backend/pkg/pagination"
)

// ---------- Category ----------

type CategoryService struct{ repo domain.CategoryRepository }

func NewCategoryService(repo domain.CategoryRepository) *CategoryService {
	return &CategoryService{repo: repo}
}

func (s *CategoryService) List(ctx context.Context, p pagination.Params) ([]domain.Category, int64, error) {
	p.Normalize()
	return s.repo.List(ctx, p)
}

func (s *CategoryService) Get(ctx context.Context, id int64) (*domain.Category, error) {
	c, err := s.repo.FindByID(ctx, id)
	return c, notFound(err, "Kategori tidak ditemukan")
}

func (s *CategoryService) Create(ctx context.Context, req dto.CategoryRequest) (*domain.Category, error) {
	name := strings.TrimSpace(req.Name)
	if err := s.ensureNameFree(ctx, name, 0); err != nil {
		return nil, err
	}
	c := &domain.Category{Name: name, Description: trimPtr(req.Description)}
	if err := s.repo.Create(ctx, c); err != nil {
		return nil, duplicate(err, "Nama kategori sudah digunakan")
	}
	return s.Get(ctx, c.ID)
}

func (s *CategoryService) Update(ctx context.Context, id int64, req dto.CategoryRequest) (*domain.Category, error) {
	c, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if err := s.ensureNameFree(ctx, name, id); err != nil {
		return nil, err
	}
	c.Name, c.Description = name, trimPtr(req.Description)
	if err := s.repo.Update(ctx, c); err != nil {
		return nil, duplicate(notFound(err, "Kategori tidak ditemukan"), "Nama kategori sudah digunakan")
	}
	return s.Get(ctx, id)
}

func (s *CategoryService) Delete(ctx context.Context, id int64) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	n, err := s.repo.CountProducts(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return apperror.Conflict(fmt.Sprintf("Kategori masih dipakai oleh %d produk", n))
	}
	return notFound(s.repo.SoftDelete(ctx, id), "Kategori tidak ditemukan")
}

func (s *CategoryService) ensureNameFree(ctx context.Context, name string, excludeID int64) error {
	taken, err := s.repo.NameExists(ctx, name, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return apperror.Validation([]apperror.FieldError{{Field: "name", Message: "nama kategori sudah digunakan"}})
	}
	return nil
}

// ---------- Unit ----------

type UnitService struct{ repo domain.UnitRepository }

func NewUnitService(repo domain.UnitRepository) *UnitService { return &UnitService{repo: repo} }

func (s *UnitService) List(ctx context.Context, p pagination.Params) ([]domain.Unit, int64, error) {
	p.Normalize()
	return s.repo.List(ctx, p)
}

func (s *UnitService) Get(ctx context.Context, id int64) (*domain.Unit, error) {
	u, err := s.repo.FindByID(ctx, id)
	return u, notFound(err, "Satuan tidak ditemukan")
}

func (s *UnitService) Create(ctx context.Context, req dto.UnitRequest) (*domain.Unit, error) {
	name := strings.TrimSpace(req.Name)
	if err := s.ensureNameFree(ctx, name, 0); err != nil {
		return nil, err
	}
	u := &domain.Unit{Name: name, Abbreviation: strings.TrimSpace(req.Abbreviation)}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, duplicate(err, "Nama satuan sudah digunakan")
	}
	return u, nil
}

func (s *UnitService) Update(ctx context.Context, id int64, req dto.UnitRequest) (*domain.Unit, error) {
	u, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	name := strings.TrimSpace(req.Name)
	if err := s.ensureNameFree(ctx, name, id); err != nil {
		return nil, err
	}
	u.Name, u.Abbreviation = name, strings.TrimSpace(req.Abbreviation)
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, duplicate(notFound(err, "Satuan tidak ditemukan"), "Nama satuan sudah digunakan")
	}
	return s.Get(ctx, id)
}

func (s *UnitService) Delete(ctx context.Context, id int64) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	n, err := s.repo.CountProducts(ctx, id)
	if err != nil {
		return err
	}
	if n > 0 {
		return apperror.Conflict(fmt.Sprintf("Satuan masih dipakai oleh %d produk", n))
	}
	return notFound(s.repo.SoftDelete(ctx, id), "Satuan tidak ditemukan")
}

func (s *UnitService) ensureNameFree(ctx context.Context, name string, excludeID int64) error {
	taken, err := s.repo.NameExists(ctx, name, excludeID)
	if err != nil {
		return err
	}
	if taken {
		return apperror.Validation([]apperror.FieldError{{Field: "name", Message: "nama satuan sudah digunakan"}})
	}
	return nil
}

// ---------- Product ----------

type ProductService struct {
	products   domain.ProductRepository
	categories domain.CategoryRepository
	units      domain.UnitRepository
	movements  domain.StockMovementRepository
	tx         domain.TxManager
	storage    domain.FileStorage
}

func NewProductService(products domain.ProductRepository, categories domain.CategoryRepository,
	units domain.UnitRepository, movements domain.StockMovementRepository, tx domain.TxManager,
	storage domain.FileStorage) *ProductService {
	return &ProductService{products: products, categories: categories, units: units, movements: movements,
		tx: tx, storage: storage}
}

func (s *ProductService) List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int64, error) {
	f.Normalize()
	return s.products.List(ctx, f)
}

func (s *ProductService) Get(ctx context.Context, id int64) (*domain.Product, error) {
	p, err := s.products.FindByID(ctx, id)
	return p, notFound(err, "Produk tidak ditemukan")
}

func (s *ProductService) GetByBarcode(ctx context.Context, code string) (*domain.Product, error) {
	p, err := s.products.FindByBarcode(ctx, strings.TrimSpace(code))
	return p, notFound(err, "Produk dengan barcode tersebut tidak ditemukan")
}

// Create stores the product; a non-zero initial stock is recorded as an IN movement.
func (s *ProductService) Create(ctx context.Context, actor Actor, req dto.CreateProductRequest) (*domain.Product, error) {
	p := &domain.Product{
		CategoryID: req.CategoryID, UnitID: req.UnitID, SKU: strings.TrimSpace(req.SKU),
		Barcode: trimPtr(req.Barcode), Name: strings.TrimSpace(req.Name), Price: req.Price, Cost: req.Cost,
		Stock: req.InitialStock, MinStock: req.MinStock, IsActive: req.IsActive == nil || *req.IsActive,
	}
	if err := s.validate(ctx, p); err != nil {
		return nil, err
	}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		if err := s.products.Create(ctx, p); err != nil {
			return duplicate(err, "SKU atau barcode sudah digunakan")
		}
		if p.Stock == 0 {
			return nil
		}
		note := "Stok awal"
		return s.movements.Create(ctx, &domain.StockMovement{
			ProductID: p.ID, UserID: actor.UserID, Type: domain.MovementIn, Quantity: p.Stock,
			StockBefore: 0, StockAfter: p.Stock, Note: &note,
		})
	})
	if err != nil {
		return nil, err
	}
	return s.Get(ctx, p.ID)
}

func (s *ProductService) Update(ctx context.Context, id int64, req dto.UpdateProductRequest) (*domain.Product, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	p.CategoryID, p.UnitID, p.SKU, p.Barcode = req.CategoryID, req.UnitID, strings.TrimSpace(req.SKU), trimPtr(req.Barcode)
	p.Name, p.Price, p.Cost, p.MinStock, p.IsActive = strings.TrimSpace(req.Name), req.Price, req.Cost, req.MinStock, req.IsActive
	if err := s.validate(ctx, p); err != nil {
		return nil, err
	}
	if err := s.products.Update(ctx, p); err != nil {
		return nil, duplicate(notFound(err, "Produk tidak ditemukan"), "SKU atau barcode sudah digunakan")
	}
	return s.Get(ctx, id)
}

func (s *ProductService) Delete(ctx context.Context, id int64) error {
	return notFound(s.products.SoftDelete(ctx, id), "Produk tidak ditemukan")
}

// AllowedImageTypes maps accepted MIME types to file extensions.
var AllowedImageTypes = map[string]string{"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp"}

func (s *ProductService) UploadImage(ctx context.Context, id int64, contentType string, r io.Reader) (*domain.Product, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	ext, ok := AllowedImageTypes[contentType]
	if !ok {
		return nil, apperror.Validation([]apperror.FieldError{{Field: "image", Message: "format gambar harus JPG, PNG, atau WEBP"}})
	}
	url, err := s.storage.Save(ctx, "products", ext, r)
	if err != nil {
		return nil, err
	}
	if err := s.products.UpdateImage(ctx, id, &url); err != nil {
		_ = s.storage.Delete(ctx, url)
		return nil, err
	}
	if p.ImageURL != nil {
		_ = s.storage.Delete(ctx, *p.ImageURL)
	}
	return s.Get(ctx, id)
}

func (s *ProductService) validate(ctx context.Context, p *domain.Product) error {
	var errs []apperror.FieldError
	if _, err := s.categories.FindByID(ctx, p.CategoryID); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		errs = append(errs, apperror.FieldError{Field: "category_id", Message: "kategori tidak ditemukan"})
	}
	if _, err := s.units.FindByID(ctx, p.UnitID); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			return err
		}
		errs = append(errs, apperror.FieldError{Field: "unit_id", Message: "satuan tidak ditemukan"})
	}
	taken, err := s.products.SKUExists(ctx, p.SKU, p.ID)
	if err != nil {
		return err
	}
	if taken {
		errs = append(errs, apperror.FieldError{Field: "sku", Message: "SKU sudah digunakan"})
	}
	if p.Barcode != nil {
		taken, err := s.products.BarcodeExists(ctx, *p.Barcode, p.ID)
		if err != nil {
			return err
		}
		if taken {
			errs = append(errs, apperror.FieldError{Field: "barcode", Message: "barcode sudah digunakan"})
		}
	}
	if len(errs) > 0 {
		return apperror.Validation(errs)
	}
	return nil
}

// trimPtr trims a string pointer and turns empty strings into nil.
func trimPtr(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}
