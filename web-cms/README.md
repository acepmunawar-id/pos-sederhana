# POS UMKM — Web CMS

Dashboard untuk Owner dan Admin (serta tampilan terbatas untuk Kasir). Dibangun dengan React, TypeScript, Vite, dan Tailwind CSS.

## Prasyarat

- Node.js 20 atau lebih baru (npm 9+)
- Backend sudah berjalan (lihat `../backend/README.md`)

## Cara Menjalankan

```bash
cp .env.example .env    # sesuaikan VITE_API_URL bila backend tidak di localhost:8080
npm install
npm run dev             # http://localhost:5173
```

Login dengan akun demo dari seed backend: `owner@toko.id` / `owner123` atau `kasir@toko.id` / `kasir123`.

> **WSL dengan proyek di drive Windows** (`/mnt/c`, `/mnt/d`): set `VITE_USE_POLLING=true` di `.env`. Tanpa itu, hot reload tidak mendeteksi perubahan file.

| Perintah            | Fungsi                                       |
|---------------------|----------------------------------------------|
| `npm run dev`       | Server development dengan hot reload         |
| `npm run build`     | Typecheck lalu build produksi ke `dist/`     |
| `npm run preview`   | Menyajikan hasil build secara lokal          |
| `npm run typecheck` | Pemeriksaan TypeScript saja                  |

Pastikan `CORS_ALLOWED_ORIGINS` di backend memuat origin web ini (default `http://localhost:5173`).

## Halaman

| Halaman | Owner | Admin | Kasir |
|---------|:-----:|:-----:|:-----:|
| Dashboard: penjualan & transaksi hari ini, grafik 7 hari, produk terlaris, stok menipis | ✓ | ✓ | – |
| Produk: tabel, filter, CRUD lewat modal, upload gambar | ✓ | ✓ | lihat saja |
| Kategori & Satuan | ✓ | ✓ | – |
| Stok: riwayat pergerakan, stok menipis, stok masuk, penyesuaian | ✓ | ✓ | lihat saja |
| Transaksi: daftar, detail, struk, void | ✓ | ✓ (void) | milik sendiri |
| Shift Kasir: rekap kas & selisih | ✓ | ✓ | milik sendiri |
| Laporan: harian/bulanan, laba kotor, produk terlaris, metode bayar, per kasir, export CSV/Excel | ✓ | ✓ | – |
| Pengguna & Role | ✓ | ✓ (tanpa akun Owner) | – |
| Pengaturan Toko: nama, alamat, logo, pajak, catatan kaki struk (dengan pratinjau) | ✓ | – | – |

Menu sidebar dan route guard memakai satu sumber yang sama, yaitu `src/routes/menu.ts`. Backend tetap memeriksa hak akses di setiap endpoint.

## Struktur

```
src/
├── components/
│   ├── ui/        Button, Input, Textarea, Select, Modal, ConfirmDialog, Table, Badge, Card, StatCard,
│   │              Pagination, EmptyState, ErrorState, Skeleton, Toast, SegmentedControl, SearchInput,
│   │              Switch, ImagePicker, RowActions, BarChart, ProportionBars
│   └── layout/    AppLayout, Sidebar, Topbar, PageHeader
├── features/      auth, dashboard, products, categories, stock, transactions, reports, users, settings
│                  (masing-masing: pages/, components/, hooks/, api.ts, types.ts)
├── lib/           api.ts (axios + refresh token), format.ts (Rupiah, tanggal), form.ts, use-list-state.ts, …
├── routes/        index.tsx (router + guard), menu.ts (menu per role)
└── types/         tipe envelope API bersama
```

## Catatan Implementasi

- **Design token**: warna (palet iOS, light dan dark), radius 8–14, skala tipografi iOS, dan spacing (`touch` = 44px) didefinisikan di `tailwind.config.ts`. Nilai warnanya berupa CSS variable di `src/index.css` dan sama persis dengan `mobile/lib/theme/app_colors.dart`. Dark mode mengikuti pengaturan sistem.
- **Refresh token**: interceptor Axios menangkap respons 401, memanggil `/auth/refresh` sekali saja walaupun banyak request gagal bersamaan, lalu mengulang request. Jika refresh gagal, sesi dihapus dan pengguna diarahkan ke halaman login.
- **Data fetching**: TanStack Query. Setiap mutasi meng-invalidate data terkait, misalnya void transaksi menyegarkan stok, dashboard, dan laporan.
- **Form**: React Hook Form + Zod. Error validasi dari server (422) dipetakan ke field yang sesuai.
- **State halaman**: setiap halaman punya tiga state: loading (skeleton), kosong (empty state), dan error (dengan tombol coba lagi).
- **List**: pagination, pencarian (debounce 350 ms), filter, dan sort dikirim ke server lewat hook `useListState`.
- **Export**: file CSV/Excel dibuat oleh backend lalu diunduh dengan nama file dari header `Content-Disposition`.
- **Responsif (ponsel, tablet, desktop)**:
  - **Desktop (≥ 1024px):** sidebar tetap dan tabel lengkap.
  - **Tablet (768–1023px):** sidebar menjadi drawer (tombol *Menu*) dan kolom sekunder disembunyikan.
  - **Ponsel (< 768px):** komponen `Table` otomatis berubah menjadi daftar kartu (judul + pasangan label/nilai + tombol aksi), kartu statistik dua kolom, segmented control bisa digeser, modal tampil sebagai bottom sheet, dan tidak ada scroll horizontal.
