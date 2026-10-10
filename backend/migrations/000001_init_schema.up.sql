-- Core schema for POS UMKM.
-- Money columns are BIGINT rupiah. Percentages are basis points (1100 = 11%).

CREATE OR REPLACE FUNCTION set_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE roles (
    id          BIGSERIAL PRIMARY KEY,
    code        VARCHAR(20)  NOT NULL UNIQUE,
    name        VARCHAR(50)  NOT NULL,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE TABLE users (
    id             BIGSERIAL PRIMARY KEY,
    role_id        BIGINT       NOT NULL REFERENCES roles(id),
    name           VARCHAR(100) NOT NULL,
    email          VARCHAR(150) NOT NULL,
    password_hash  VARCHAR(255) NOT NULL,
    is_active      BOOLEAN      NOT NULL DEFAULT TRUE,
    last_login_at  TIMESTAMPTZ,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMPTZ
);
CREATE UNIQUE INDEX ux_users_email ON users (LOWER(email)) WHERE deleted_at IS NULL;
CREATE INDEX ix_users_role ON users (role_id);

CREATE TABLE refresh_tokens (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  CHAR(64)     NOT NULL UNIQUE,
    user_agent  VARCHAR(255),
    ip_address  VARCHAR(64),
    expires_at  TIMESTAMPTZ  NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_refresh_tokens_user ON refresh_tokens (user_id);

CREATE TABLE categories (
    id           BIGSERIAL PRIMARY KEY,
    name         VARCHAR(100) NOT NULL,
    description  TEXT,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);
CREATE UNIQUE INDEX ux_categories_name ON categories (LOWER(name)) WHERE deleted_at IS NULL;

CREATE TABLE units (
    id            BIGSERIAL PRIMARY KEY,
    name          VARCHAR(50) NOT NULL,
    abbreviation  VARCHAR(10) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at    TIMESTAMPTZ
);
CREATE UNIQUE INDEX ux_units_name ON units (LOWER(name)) WHERE deleted_at IS NULL;

CREATE TABLE products (
    id           BIGSERIAL PRIMARY KEY,
    category_id  BIGINT       NOT NULL REFERENCES categories(id),
    unit_id      BIGINT       NOT NULL REFERENCES units(id),
    sku          VARCHAR(50)  NOT NULL,
    barcode      VARCHAR(50),
    name         VARCHAR(150) NOT NULL,
    price        BIGINT       NOT NULL CHECK (price >= 0),
    cost         BIGINT       NOT NULL CHECK (cost >= 0),
    stock        INTEGER      NOT NULL DEFAULT 0 CHECK (stock >= 0),
    min_stock    INTEGER      NOT NULL DEFAULT 0 CHECK (min_stock >= 0),
    image_url    VARCHAR(500),
    is_active    BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMPTZ
);
CREATE UNIQUE INDEX ux_products_sku ON products (LOWER(sku)) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX ux_products_barcode ON products (barcode) WHERE deleted_at IS NULL AND barcode IS NOT NULL;
CREATE INDEX ix_products_category ON products (category_id);
CREATE INDEX ix_products_name ON products (LOWER(name));

CREATE TABLE shifts (
    id                  BIGSERIAL PRIMARY KEY,
    user_id             BIGINT      NOT NULL REFERENCES users(id),
    status              VARCHAR(10) NOT NULL CHECK (status IN ('OPEN', 'CLOSED')),
    opening_cash        BIGINT      NOT NULL CHECK (opening_cash >= 0),
    expected_cash       BIGINT,
    actual_cash         BIGINT,
    cash_difference     BIGINT,
    total_sales         BIGINT,
    total_transactions  INTEGER,
    opening_note        TEXT,
    closing_note        TEXT,
    opened_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    closed_at           TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- A cashier can only have one open shift at a time.
CREATE UNIQUE INDEX ux_shifts_open_per_user ON shifts (user_id) WHERE status = 'OPEN';

CREATE TABLE invoice_sequences (
    seq_date     DATE        PRIMARY KEY,
    last_number  INTEGER     NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    id               BIGSERIAL PRIMARY KEY,
    invoice_number   VARCHAR(30)  NOT NULL UNIQUE,
    idempotency_key  VARCHAR(100) NOT NULL UNIQUE,
    shift_id         BIGINT       NOT NULL REFERENCES shifts(id),
    user_id          BIGINT       NOT NULL REFERENCES users(id),
    subtotal         BIGINT       NOT NULL CHECK (subtotal >= 0),
    discount_type    VARCHAR(10)  CHECK (discount_type IN ('NOMINAL', 'PERCENT')),
    discount_value   BIGINT       NOT NULL DEFAULT 0 CHECK (discount_value >= 0),
    discount_amount  BIGINT       NOT NULL DEFAULT 0 CHECK (discount_amount >= 0),
    tax_rate_bp      INTEGER      NOT NULL DEFAULT 0 CHECK (tax_rate_bp >= 0),
    tax_amount       BIGINT       NOT NULL DEFAULT 0 CHECK (tax_amount >= 0),
    total            BIGINT       NOT NULL CHECK (total >= 0),
    paid_amount      BIGINT       NOT NULL CHECK (paid_amount >= 0),
    change_amount    BIGINT       NOT NULL DEFAULT 0 CHECK (change_amount >= 0),
    total_cost       BIGINT       NOT NULL DEFAULT 0 CHECK (total_cost >= 0),
    status           VARCHAR(10)  NOT NULL CHECK (status IN ('COMPLETED', 'VOIDED')),
    note             TEXT,
    voided_at        TIMESTAMPTZ,
    voided_by        BIGINT REFERENCES users(id),
    void_reason      TEXT,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_transactions_created_at ON transactions (created_at);
CREATE INDEX ix_transactions_shift ON transactions (shift_id);
CREATE INDEX ix_transactions_user ON transactions (user_id);
CREATE INDEX ix_transactions_status ON transactions (status);

CREATE TABLE transaction_items (
    id              BIGSERIAL PRIMARY KEY,
    transaction_id  BIGINT       NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    product_id      BIGINT       NOT NULL REFERENCES products(id),
    product_name    VARCHAR(150) NOT NULL,
    sku             VARCHAR(50)  NOT NULL,
    quantity        INTEGER      NOT NULL CHECK (quantity > 0),
    unit_price      BIGINT       NOT NULL CHECK (unit_price >= 0),
    unit_cost       BIGINT       NOT NULL CHECK (unit_cost >= 0),
    subtotal        BIGINT       NOT NULL CHECK (subtotal >= 0),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_transaction_items_tx ON transaction_items (transaction_id);
CREATE INDEX ix_transaction_items_product ON transaction_items (product_id);

-- amount = portion of the total settled by this method (sums to transactions.total);
-- tendered_amount = money actually handed over (only differs for CASH, the change is returned).
CREATE TABLE payments (
    id                BIGSERIAL PRIMARY KEY,
    transaction_id    BIGINT      NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    method            VARCHAR(10) NOT NULL CHECK (method IN ('CASH', 'QRIS', 'TRANSFER', 'DEBIT')),
    amount            BIGINT      NOT NULL CHECK (amount >= 0),
    tendered_amount   BIGINT      NOT NULL CHECK (tendered_amount >= 0),
    reference_number  VARCHAR(100),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (transaction_id, method)
);

CREATE TABLE stock_movements (
    id              BIGSERIAL PRIMARY KEY,
    product_id      BIGINT      NOT NULL REFERENCES products(id),
    user_id         BIGINT      NOT NULL REFERENCES users(id),
    type            VARCHAR(12) NOT NULL CHECK (type IN ('IN', 'ADJUSTMENT', 'SALE', 'VOID')),
    quantity        INTEGER     NOT NULL CHECK (quantity <> 0),
    stock_before    INTEGER     NOT NULL,
    stock_after     INTEGER     NOT NULL,
    reference_type  VARCHAR(20),
    reference_id    BIGINT,
    note            TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ix_stock_movements_product ON stock_movements (product_id, created_at DESC);
CREATE INDEX ix_stock_movements_created_at ON stock_movements (created_at);

CREATE TABLE store_settings (
    id              BIGINT       PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    name            VARCHAR(100) NOT NULL,
    address         TEXT,
    phone           VARCHAR(30),
    logo_url        VARCHAR(500),
    tax_rate_bp     INTEGER      NOT NULL DEFAULT 0 CHECK (tax_rate_bp BETWEEN 0 AND 10000),
    receipt_footer  TEXT,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

DO $$
DECLARE t TEXT;
BEGIN
    FOREACH t IN ARRAY ARRAY['roles','users','refresh_tokens','categories','units','products','shifts',
                             'invoice_sequences','transactions','transaction_items','payments',
                             'stock_movements','store_settings']
    LOOP
        EXECUTE format('CREATE TRIGGER trg_%1$s_updated_at BEFORE UPDATE ON %1$s
                        FOR EACH ROW EXECUTE FUNCTION set_updated_at()', t);
    END LOOP;
END $$;

INSERT INTO roles (code, name) VALUES ('owner', 'Owner'), ('admin', 'Admin'), ('cashier', 'Kasir');
INSERT INTO store_settings (id, name, tax_rate_bp, receipt_footer)
VALUES (1, 'Toko Saya', 0, 'Terima kasih atas kunjungan Anda');
