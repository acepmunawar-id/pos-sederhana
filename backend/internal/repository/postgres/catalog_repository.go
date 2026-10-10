package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/posumkm/backend/internal/domain"
	"github.com/posumkm/backend/pkg/pagination"
)

// ---------- Category ----------

type CategoryRepository struct{ pool *pgxpool.Pool }

func NewCategoryRepository(pool *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{pool: pool}
}

const categorySelect = `SELECT c.id, c.name, c.description,
	(SELECT COUNT(*) FROM products p WHERE p.category_id = c.id AND p.deleted_at IS NULL),
	c.created_at, c.updated_at FROM categories c`

func scanCategory(row pgx.Row) (domain.Category, error) {
	var c domain.Category
	err := row.Scan(&c.ID, &c.Name, &c.Description, &c.ProductCount, &c.CreatedAt, &c.UpdatedAt)
	return c, mapErr(err)
}

func (r *CategoryRepository) Create(ctx context.Context, c *domain.Category) error {
	return mapErr(conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`,
		c.Name, c.Description).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt))
}

func (r *CategoryRepository) Update(ctx context.Context, c *domain.Category) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE categories SET name = $2, description = $3 WHERE id = $1 AND deleted_at IS NULL`,
		c.ID, c.Name, c.Description))
}

func (r *CategoryRepository) SoftDelete(ctx context.Context, id int64) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE categories SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id))
}

func (r *CategoryRepository) FindByID(ctx context.Context, id int64) (*domain.Category, error) {
	c, err := scanCategory(conn(ctx, r.pool).QueryRow(ctx, categorySelect+` WHERE c.id = $1 AND c.deleted_at IS NULL`, id))
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *CategoryRepository) NameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	return exists(ctx, conn(ctx, r.pool),
		`SELECT 1 FROM categories WHERE LOWER(name) = LOWER($1) AND id <> $2 AND deleted_at IS NULL`, name, excludeID)
}

func (r *CategoryRepository) CountProducts(ctx context.Context, id int64) (int64, error) {
	return count(ctx, conn(ctx, r.pool),
		`SELECT COUNT(*) FROM products WHERE category_id = $1 AND deleted_at IS NULL`, []any{id})
}

var categorySorts = map[string]string{"name": "c.name", "created_at": "c.created_at"}

func (r *CategoryRepository) List(ctx context.Context, p pagination.Params) ([]domain.Category, int64, error) {
	w := &where{}
	w.add("c.deleted_at IS NULL")
	w.search(p.Search, "c.name", "c.description")
	q := conn(ctx, r.pool)
	total, err := count(ctx, q, `SELECT COUNT(*) FROM categories c`+w.sql(), w.args)
	if err != nil {
		return nil, 0, err
	}
	limit, args := w.page(p)
	rows, err := q.Query(ctx, categorySelect+w.sql()+orderBy(p, categorySorts, "c.name ASC", "c.id")+limit, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Category, error) { return scanCategory(row) })
	return items, total, err
}

// ---------- Unit ----------

type UnitRepository struct{ pool *pgxpool.Pool }

func NewUnitRepository(pool *pgxpool.Pool) *UnitRepository { return &UnitRepository{pool: pool} }

const unitSelect = `SELECT u.id, u.name, u.abbreviation, u.created_at, u.updated_at FROM units u`

func scanUnit(row pgx.Row) (domain.Unit, error) {
	var u domain.Unit
	err := row.Scan(&u.ID, &u.Name, &u.Abbreviation, &u.CreatedAt, &u.UpdatedAt)
	return u, mapErr(err)
}

func (r *UnitRepository) Create(ctx context.Context, u *domain.Unit) error {
	return mapErr(conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO units (name, abbreviation) VALUES ($1, $2) RETURNING id, created_at, updated_at`,
		u.Name, u.Abbreviation).Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt))
}

func (r *UnitRepository) Update(ctx context.Context, u *domain.Unit) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE units SET name = $2, abbreviation = $3 WHERE id = $1 AND deleted_at IS NULL`,
		u.ID, u.Name, u.Abbreviation))
}

func (r *UnitRepository) SoftDelete(ctx context.Context, id int64) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE units SET deleted_at = NOW() WHERE id = $1 AND deleted_at IS NULL`, id))
}

func (r *UnitRepository) FindByID(ctx context.Context, id int64) (*domain.Unit, error) {
	u, err := scanUnit(conn(ctx, r.pool).QueryRow(ctx, unitSelect+` WHERE u.id = $1 AND u.deleted_at IS NULL`, id))
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UnitRepository) NameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	return exists(ctx, conn(ctx, r.pool),
		`SELECT 1 FROM units WHERE LOWER(name) = LOWER($1) AND id <> $2 AND deleted_at IS NULL`, name, excludeID)
}

func (r *UnitRepository) CountProducts(ctx context.Context, id int64) (int64, error) {
	return count(ctx, conn(ctx, r.pool),
		`SELECT COUNT(*) FROM products WHERE unit_id = $1 AND deleted_at IS NULL`, []any{id})
}

var unitSorts = map[string]string{"name": "u.name", "abbreviation": "u.abbreviation", "created_at": "u.created_at"}

func (r *UnitRepository) List(ctx context.Context, p pagination.Params) ([]domain.Unit, int64, error) {
	w := &where{}
	w.add("u.deleted_at IS NULL")
	w.search(p.Search, "u.name", "u.abbreviation")
	q := conn(ctx, r.pool)
	total, err := count(ctx, q, `SELECT COUNT(*) FROM units u`+w.sql(), w.args)
	if err != nil {
		return nil, 0, err
	}
	limit, args := w.page(p)
	rows, err := q.Query(ctx, unitSelect+w.sql()+orderBy(p, unitSorts, "u.name ASC", "u.id")+limit, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Unit, error) { return scanUnit(row) })
	return items, total, err
}

// ---------- Product ----------

type ProductRepository struct{ pool *pgxpool.Pool }

func NewProductRepository(pool *pgxpool.Pool) *ProductRepository {
	return &ProductRepository{pool: pool}
}

const productColumns = `p.id, p.category_id, c.name, p.unit_id, un.abbreviation, p.sku, p.barcode, p.name,
	p.price, p.cost, p.stock, p.min_stock, p.image_url, p.is_active, p.created_at, p.updated_at`

const productFrom = ` FROM products p
	JOIN categories c ON c.id = p.category_id
	JOIN units un ON un.id = p.unit_id`

func scanProduct(row pgx.Row) (domain.Product, error) {
	var p domain.Product
	err := row.Scan(&p.ID, &p.CategoryID, &p.CategoryName, &p.UnitID, &p.UnitName, &p.SKU, &p.Barcode, &p.Name,
		&p.Price, &p.Cost, &p.Stock, &p.MinStock, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, mapErr(err)
}

func (r *ProductRepository) Create(ctx context.Context, p *domain.Product) error {
	return mapErr(conn(ctx, r.pool).QueryRow(ctx,
		`INSERT INTO products (category_id, unit_id, sku, barcode, name, price, cost, stock, min_stock, image_url, is_active)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id, created_at, updated_at`,
		p.CategoryID, p.UnitID, p.SKU, p.Barcode, p.Name, p.Price, p.Cost, p.Stock, p.MinStock, p.ImageURL, p.IsActive).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt))
}

// Update changes master data only; stock is changed exclusively through UpdateStock.
func (r *ProductRepository) Update(ctx context.Context, p *domain.Product) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE products SET category_id = $2, unit_id = $3, sku = $4, barcode = $5, name = $6,
		 price = $7, cost = $8, min_stock = $9, is_active = $10
		 WHERE id = $1 AND deleted_at IS NULL`,
		p.ID, p.CategoryID, p.UnitID, p.SKU, p.Barcode, p.Name, p.Price, p.Cost, p.MinStock, p.IsActive))
}

func (r *ProductRepository) UpdateImage(ctx context.Context, id int64, url *string) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE products SET image_url = $2 WHERE id = $1 AND deleted_at IS NULL`, id, url))
}

func (r *ProductRepository) UpdateStock(ctx context.Context, id int64, stock int) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE products SET stock = $2 WHERE id = $1`, id, stock))
}

func (r *ProductRepository) SoftDelete(ctx context.Context, id int64) error {
	return affectedOrNotFound(conn(ctx, r.pool).Exec(ctx,
		`UPDATE products SET deleted_at = NOW(), is_active = FALSE WHERE id = $1 AND deleted_at IS NULL`, id))
}

func (r *ProductRepository) FindByID(ctx context.Context, id int64) (*domain.Product, error) {
	p, err := scanProduct(conn(ctx, r.pool).QueryRow(ctx,
		`SELECT `+productColumns+productFrom+` WHERE p.id = $1 AND p.deleted_at IS NULL`, id))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) FindByBarcode(ctx context.Context, barcode string) (*domain.Product, error) {
	p, err := scanProduct(conn(ctx, r.pool).QueryRow(ctx,
		`SELECT `+productColumns+productFrom+` WHERE (p.barcode = $1 OR LOWER(p.sku) = LOWER($1)) AND p.deleted_at IS NULL
		 ORDER BY (p.barcode = $1) DESC LIMIT 1`, barcode))
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepository) FindByIDsForUpdate(ctx context.Context, ids []int64) ([]domain.Product, error) {
	// Lock product rows only (OF p) in a deterministic order to avoid deadlocks between checkouts.
	rows, err := conn(ctx, r.pool).Query(ctx,
		`SELECT `+productColumns+productFrom+` WHERE p.id = ANY($1) AND p.deleted_at IS NULL
		 ORDER BY p.id FOR UPDATE OF p`, ids)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Product, error) { return scanProduct(row) })
}

func (r *ProductRepository) SKUExists(ctx context.Context, sku string, excludeID int64) (bool, error) {
	return exists(ctx, conn(ctx, r.pool),
		`SELECT 1 FROM products WHERE LOWER(sku) = LOWER($1) AND id <> $2 AND deleted_at IS NULL`, sku, excludeID)
}

func (r *ProductRepository) BarcodeExists(ctx context.Context, barcode string, excludeID int64) (bool, error) {
	return exists(ctx, conn(ctx, r.pool),
		`SELECT 1 FROM products WHERE barcode = $1 AND id <> $2 AND deleted_at IS NULL`, barcode, excludeID)
}

func (r *ProductRepository) CountLowStock(ctx context.Context) (int64, error) {
	return count(ctx, conn(ctx, r.pool),
		`SELECT COUNT(*) FROM products WHERE deleted_at IS NULL AND is_active AND stock <= min_stock`, nil)
}

var productSorts = map[string]string{
	"name": "p.name", "sku": "p.sku", "price": "p.price", "cost": "p.cost", "stock": "p.stock",
	"category": "c.name", "created_at": "p.created_at", "updated_at": "p.updated_at",
}

func (r *ProductRepository) List(ctx context.Context, f domain.ProductFilter) ([]domain.Product, int64, error) {
	w := &where{}
	w.add("p.deleted_at IS NULL")
	w.search(f.Search, "p.name", "p.sku", "p.barcode")
	if f.CategoryID != nil {
		w.add("p.category_id = ?", *f.CategoryID)
	}
	if f.IsActive != nil {
		w.add("p.is_active = ?", *f.IsActive)
	}
	fallback := "p.name ASC"
	if f.LowStock {
		w.add("p.stock <= p.min_stock")
		fallback = "(p.stock - p.min_stock) ASC"
	}
	q := conn(ctx, r.pool)
	total, err := count(ctx, q, `SELECT COUNT(*)`+productFrom+w.sql(), w.args)
	if err != nil {
		return nil, 0, err
	}
	limit, args := w.page(f.Params)
	rows, err := q.Query(ctx, `SELECT `+productColumns+productFrom+w.sql()+orderBy(f.Params, productSorts, fallback, "p.id")+limit, args...)
	if err != nil {
		return nil, 0, err
	}
	items, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Product, error) { return scanProduct(row) })
	return items, total, err
}
