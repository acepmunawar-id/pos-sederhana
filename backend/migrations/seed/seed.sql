-- Demo seed data. Safe to run multiple times.
-- Accounts:
--   owner@toko.id / owner123  (Owner)
--   kasir@toko.id / kasir123  (Kasir)

INSERT INTO users (role_id, name, email, password_hash)
SELECT r.id, v.name, v.email, v.password_hash
FROM (VALUES
    ('owner',   'Budi Santoso', 'owner@toko.id', '$2a$10$2QeRJTOrQrjNox2wuDoaguwuFS8K5fJgezq0MXKU7E/iDty.ANjwe'),
    ('cashier', 'Siti Aminah',  'kasir@toko.id', '$2a$10$WpCa0im84PzmlVX1tupcXO9357nBgNZNVgkX2AppJFs4sSEFyq04S')
) AS v(role_code, name, email, password_hash)
JOIN roles r ON r.code = v.role_code
WHERE NOT EXISTS (SELECT 1 FROM users u WHERE LOWER(u.email) = LOWER(v.email) AND u.deleted_at IS NULL);

UPDATE store_settings
SET name = 'Warung Kopi Nusantara',
    address = 'Jl. Merdeka No. 10, Bandung',
    phone = '0812-3456-7890',
    tax_rate_bp = 1000,
    receipt_footer = 'Terima kasih, sampai jumpa lagi!'
WHERE id = 1 AND name = 'Toko Saya';

INSERT INTO categories (name, description)
SELECT v.name, v.description
FROM (VALUES
    ('Minuman',  'Kopi, teh, dan minuman dingin'),
    ('Makanan',  'Makanan berat dan ringan'),
    ('Snack',    'Camilan kemasan'),
    ('Sembako',  'Kebutuhan pokok sehari-hari')
) AS v(name, description)
WHERE NOT EXISTS (SELECT 1 FROM categories c WHERE LOWER(c.name) = LOWER(v.name) AND c.deleted_at IS NULL);

INSERT INTO units (name, abbreviation)
SELECT v.name, v.abbreviation
FROM (VALUES ('Pieces', 'pcs'), ('Gelas', 'gls'), ('Porsi', 'prs'), ('Bungkus', 'bks'), ('Kilogram', 'kg')) AS v(name, abbreviation)
WHERE NOT EXISTS (SELECT 1 FROM units u WHERE LOWER(u.name) = LOWER(v.name) AND u.deleted_at IS NULL);

INSERT INTO products (category_id, unit_id, sku, barcode, name, price, cost, stock, min_stock)
SELECT c.id, u.id, v.sku, v.barcode, v.name, v.price, v.cost, v.stock, v.min_stock
FROM (VALUES
    ('Minuman', 'Gelas',    'MNM-001', '8990000000011', 'Kopi Susu Gula Aren', 18000, 7000,  100, 10),
    ('Minuman', 'Gelas',    'MNM-002', '8990000000028', 'Es Teh Manis',         5000, 1500,  150, 20),
    ('Minuman', 'Gelas',    'MNM-003', '8990000000035', 'Americano',           15000, 5000,   80, 10),
    ('Minuman', 'Pieces',   'MNM-004', '8992761111113', 'Air Mineral 600ml',    4000, 2500,    8, 12),
    ('Makanan', 'Porsi',    'MKN-001', '8990000000042', 'Nasi Goreng Spesial', 25000, 11000,  40,  5),
    ('Makanan', 'Porsi',    'MKN-002', '8990000000059', 'Mie Goreng',          20000, 8000,   35,  5),
    ('Makanan', 'Porsi',    'MKN-003', '8990000000066', 'Roti Bakar Coklat',   15000, 6000,    4,  5),
    ('Snack',   'Bungkus',  'SNK-001', '8991002101616', 'Keripik Kentang',     10000, 7000,   60, 10),
    ('Snack',   'Bungkus',  'SNK-002', '8996001600146', 'Kacang Kulit',         8000, 5500,   45, 10),
    ('Sembako', 'Kilogram', 'SMB-001', '8990000000073', 'Beras Premium 1kg',   16000, 13500,  25,  5),
    ('Sembako', 'Kilogram', 'SMB-002', '8990000000080', 'Gula Pasir 1kg',      18000, 15500,   3,  5)
) AS v(category, unit, sku, barcode, name, price, cost, stock, min_stock)
JOIN categories c ON c.name = v.category AND c.deleted_at IS NULL
JOIN units u ON u.name = v.unit AND u.deleted_at IS NULL
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE LOWER(p.sku) = LOWER(v.sku) AND p.deleted_at IS NULL)
ORDER BY v.sku;

-- Record the opening stock as an IN movement so the stock history is consistent.
INSERT INTO stock_movements (product_id, user_id, type, quantity, stock_before, stock_after, reference_type, note)
SELECT p.id, (SELECT id FROM users WHERE email = 'owner@toko.id'), 'IN', p.stock, 0, p.stock, 'SEED', 'Stok awal'
FROM products p
WHERE p.stock > 0
  AND NOT EXISTS (SELECT 1 FROM stock_movements m WHERE m.product_id = p.id);
