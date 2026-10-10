# POS UMKM — Backend (Go + Gin + PostgreSQL)

REST API untuk aplikasi Point of Sale UMKM: auth JWT, master data, stok, penjualan, shift kasir, laporan, dan pengaturan toko.

## Prasyarat

| Kebutuhan  | Versi minimum |
|------------|---------------|
| Go         | 1.26          |
| PostgreSQL | 14            |

## Cara Menjalankan

```bash
# 1. Siapkan database (sekali saja)
sudo -u postgres psql -c "CREATE USER pos WITH PASSWORD 'pos';"
sudo -u postgres psql -c "CREATE DATABASE pos_db OWNER pos;"

# 2. Konfigurasi
cp .env.example .env
# ganti JWT_ACCESS_SECRET, contoh: openssl rand -hex 32

# 3. Migration + data contoh
go run ./cmd/migrate up
go run ./cmd/migrate seed

# 4. Jalankan server
go run ./cmd/api
```

Server berjalan di `http://localhost:8080`:

- Health check: `GET /health`
- Swagger UI: <http://localhost:8080/swagger/index.html>
- Prefix API: `/api/v1`

### Akun Demo (dari seed)

| Role  | Email           | Password   |
|-------|-----------------|------------|
| Owner | `owner@toko.id` | `owner123` |
| Kasir | `kasir@toko.id` | `kasir123` |

Seed juga mengisi pengaturan toko (pajak 10%), 4 kategori, 5 satuan, dan 11 produk. Beberapa produk sengaja dibuat berstok menipis.

### Perintah Lain

| Perintah                       | Fungsi                                              |
|--------------------------------|-----------------------------------------------------|
| `go run ./cmd/migrate down 1`  | Rollback 1 migration (tanpa angka = semua)          |
| `go run ./cmd/migrate fresh`   | Hapus semua tabel, migrate ulang, lalu seed         |
| `go test ./... -count=1`       | Jalankan unit test                                  |
| `make cover`                   | Coverage unit test service                          |
| `make swagger`                 | Generate ulang dokumentasi (butuh `go install github.com/swaggo/swag/cmd/swag@v1.16.6`) |
| `make build`                   | Build binary ke `bin/api` dan `bin/migrate`         |

Jika `DB_AUTO_MIGRATE=true` (default), migration juga dijalankan otomatis setiap kali server start.

## Arsitektur

```
handler  →  service  →  domain (interface repository)  ←  repository/postgres
 (HTTP)     (bisnis)        (entity + kontrak)                (SQL pgx)
```

- **handler/**: parsing request, validasi DTO, memanggil service, menulis response. Tidak ada business logic maupun query.
- **service/**: seluruh aturan bisnis. Service hanya bergantung pada interface di `domain/`, sehingga bisa diuji dengan mock.
- **domain/**: entity dan interface repository (satu per entity), plus `TxManager`.
- **repository/postgres/**: implementasi interface dengan SQL murni (pgx). Tanpa business logic.
- **dto/**: request/response JSON, terpisah dari entity.
- **middleware/**: auth JWT, role guard, request ID, logging, recovery.
- **pkg/**: config, logger, response, validator, apperror, pagination, jwt, hash.

Dependency injection dilakukan lewat constructor di `cmd/api/main.go`.

### Transaksi Database

`TxManager.WithinTx(ctx, fn)` menyimpan `pgx.Tx` di `context`. Setiap repository memakai `conn(ctx, pool)`, sehingga semua repository yang dipanggil di dalam `fn` otomatis ikut dalam transaksi yang sama. Service tetap tidak mengenal pgx.

## Keputusan Penting

| Topik | Implementasi |
|-------|--------------|
| Uang | `BIGINT` rupiah, tidak ada float |
| Persentase | Basis point (`1100` = 11%), dipakai untuk pajak & diskon persen. Pembulatan setengah ke atas dengan integer |
| Checkout atomic | Satu DB transaction: lock produk `SELECT ... FOR UPDATE OF p ORDER BY id` (urutan tetap → bebas deadlock), validasi stok, simpan transaksi + item + pembayaran, kurangi stok, catat `stock_movements` |
| Idempotency | Header `Idempotency-Key` wajib (8–100 karakter), disimpan di kolom unik. Request ulang mengembalikan transaksi lama (HTTP 200 + header `Idempotent-Replayed: true`). Jika dua request bersamaan, pemenangnya dikembalikan ke keduanya |
| Nomor invoice | `INV/YYYYMMDD/0001` dari tabel `invoice_sequences` (upsert per tanggal di zona waktu toko). Ikut DB transaction, jadi nomor tidak loncat walau checkout gagal |
| Pembayaran | Bisa split (mis. QRIS + tunai). `amount` = porsi tagihan, `tendered_amount` = uang yang diserahkan. Kembalian hanya dari tunai |
| Diskon & pajak | Diskon (nominal/persen) dari subtotal. Pajak dihitung dari subtotal setelah diskon |
| Snapshot harga | `transaction_items` menyimpan nama, SKU, harga jual & modal saat transaksi, sehingga laporan laba tetap akurat |
| Void | Hanya Owner/Admin. Stok dikembalikan dengan movement `VOID`, transaksi ditandai `VOIDED` |
| Shift | Satu shift terbuka per kasir (partial unique index). Kas diharapkan = kas awal + tunai bersih. Selisih = kas aktual − diharapkan |
| Refresh token | Token acak, yang disimpan hanya hash SHA-256-nya. Dirotasi setiap refresh. Token lama yang dipakai ulang → semua sesi user dicabut |
| Soft delete | `deleted_at` pada users, categories, units, products. Unique index parsial `WHERE deleted_at IS NULL` |
| Upload | Tipe file dideteksi dari isi (bukan header klien), maks 2 MB, disimpan di `UPLOAD_DIR` dan disajikan di `/uploads` |

## Format Response

```json
{ "success": true, "message": "Berhasil", "data": [ ... ], "meta": { "page": 1, "limit": 10, "total": 42 } }
{ "success": false, "message": "Data yang dikirim tidak valid", "errors": [ { "field": "email", "message": "format email tidak valid" } ] }
```

Semua endpoint list menerima `page`, `limit` (maks 100), `search`, `sort`, `order` (`asc|desc`) ditambah filter khusus per modul. Daftar lengkapnya ada di Swagger.

## Hak Akses

| Modul | Owner | Admin | Kasir |
|-------|:-----:|:-----:|:-----:|
| Login, profil, ganti password | ✓ | ✓ | ✓ |
| Lihat produk/kategori/satuan, stok, riwayat stok | ✓ | ✓ | ✓ |
| Tambah/ubah/hapus master data, stok masuk, penyesuaian | ✓ | ✓ | – |
| Buka/tutup shift, checkout, struk | ✓ | ✓ | ✓ |
| Lihat transaksi & shift | semua | semua | milik sendiri |
| Void transaksi | ✓ | ✓ | – |
| Laporan & dashboard | ✓ | ✓ | – |
| Kelola user | ✓ | ✓ (tidak bisa mengelola Owner) | – |
| Ubah pengaturan toko & logo | ✓ | – | – |

Admin juga tidak bisa memberi role Owner. Sistem selalu menyisakan minimal satu Owner aktif, dan user tidak bisa menghapus atau menonaktifkan akunnya sendiri.

## Contoh Checkout

```bash
curl -X POST http://localhost:8080/api/v1/transactions/checkout \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: 6f1c2b7e-0d8a-4b8e-9b51-2d3f1a0c9e11" \
  -d '{
    "items": [ { "product_id": 4, "quantity": 2 }, { "product_id": 5, "quantity": 3 } ],
    "discount_type": "PERCENT",
    "discount_value": 1000,
    "payments": [ { "method": "CASH", "amount": 100000 } ]
  }'
```

## Unit Test

Test service ada di `internal/service/*_test.go` dan memakai mock repository (`testify/mock`) tanpa database. Yang diuji antara lain:

- perhitungan total, diskon, pajak, dan kembalian
- pembayaran split
- stok kurang, produk nonaktif atau tidak ditemukan
- replay idempotency dan race key ganda
- void beserta pengembalian stok
- rotasi refresh token dan deteksi token yang dipakai ulang
- aturan role pada manajemen user
- selisih kas saat tutup shift
- stok masuk dan penyesuaian
- validasi produk
- format Rupiah dan lebar struk
