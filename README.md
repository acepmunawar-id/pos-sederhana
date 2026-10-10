# pos-sederhana
Full-stack Point of Sale system for small businesses (UMKM): REST API with repository architecture, React + Tailwind admin dashboard, and Flutter cashier app with iOS-style design.

## Struktur Proyek

| Folder      | Isi                                              | Teknologi                          |
|-------------|--------------------------------------------------|------------------------------------|
| `backend/`  | REST API                                         | Go + Gin + PostgreSQL              |
| `web-cms/`  | Dashboard Owner/Admin                            | React + TypeScript + Vite + Tailwind |
| `mobile/`   | Aplikasi kasir                                   | Flutter                            |

## Prasyarat

| Kebutuhan  | Versi minimum | Dipakai oleh |
|------------|---------------|--------------|
| Go         | 1.26          | backend      |
| PostgreSQL | 14            | backend      |
| Node.js    | 20 (npm 9+)   | web-cms      |
| Flutter    | 3.24          | mobile       |

## Cara Menjalankan

Jalankan secara berurutan: **backend → web CMS → aplikasi mobile**. Web CMS dan aplikasi mobile membutuhkan backend yang sudah berjalan.

### 1. Backend (`http://localhost:8080`)

```bash
cd backend

# Siapkan database (sekali saja)
sudo -u postgres psql -c "CREATE USER pos WITH PASSWORD 'pos';"
sudo -u postgres psql -c "CREATE DATABASE pos_db OWNER pos;"

# Konfigurasi, lalu ganti JWT_ACCESS_SECRET (contoh: openssl rand -hex 32)
cp .env.example .env

# Migration + data contoh
go run ./cmd/migrate up
go run ./cmd/migrate seed

# Jalankan server
go run ./cmd/api
```

- Health check: <http://localhost:8080/health>
- Swagger UI: <http://localhost:8080/swagger/index.html>

### 2. Web CMS (`http://localhost:5173`)

Di terminal baru:

```bash
cd web-cms
cp .env.example .env    # sesuaikan VITE_API_URL bila backend tidak di localhost:8080
npm install
npm run dev
```

> Jika proyek berada di drive Windows yang diakses dari WSL (`/mnt/c`, `/mnt/d`), set `VITE_USE_POLLING=true` di `web-cms/.env` agar hot reload berfungsi.

### 3. Aplikasi Kasir Mobile

Di terminal baru:

```bash
cd mobile
cp .env.example .env    # atur API_URL sesuai perangkat (lihat tabel di bawah)
flutter pub get
flutter run --dart-define-from-file=.env
```

| Perangkat         | `API_URL`                                          |
|-------------------|----------------------------------------------------|
| Android emulator  | `http://10.0.2.2:8080/api/v1`                      |
| iOS simulator     | `http://localhost:8080/api/v1`                     |
| HP fisik          | `http://<IP-laptop-di-WiFi-yang-sama>:8080/api/v1` |
| Browser (demo)    | `http://localhost:8080/api/v1`, lalu tambahkan origin web ke `CORS_ALLOWED_ORIGINS` di `backend/.env` |

## Akun Demo

Tersedia setelah menjalankan `go run ./cmd/migrate seed`:

| Role  | Email           | Password   |
|-------|-----------------|------------|
| Owner | `owner@toko.id` | `owner123` |
| Kasir | `kasir@toko.id` | `kasir123` |

## Dokumentasi Lengkap

Detail arsitektur, perintah lain, dan fitur ada di README masing-masing:

- [backend/README.md](backend/README.md)
- [web-cms/README.md](web-cms/README.md)
- [mobile/README.md](mobile/README.md)
