# POS UMKM — Aplikasi Kasir (Flutter)

Aplikasi kasir bergaya iPhone (Apple Human Interface Guidelines) untuk pengguna yang belum terbiasa dengan teknologi. Satu transaksi selesai dalam 3 langkah: **Pilih produk → Keranjang → Bayar**.

## Prasyarat

- Flutter 3.24+ (diuji dengan Flutter 3.47 / Dart 3.13)
- Android Studio (emulator/SDK) atau Xcode (simulator iOS)
- Backend sudah berjalan (lihat `../backend/README.md`)

## Cara Menjalankan

```bash
cp .env.example .env        # atur API_URL sesuai perangkat (lihat tabel di bawah)
flutter pub get
flutter run --dart-define-from-file=.env
```

| Perangkat         | `API_URL`                                       |
|-------------------|-------------------------------------------------|
| Android emulator  | `http://10.0.2.2:8080/api/v1`                   |
| iOS simulator     | `http://localhost:8080/api/v1`                  |
| HP fisik          | `http://<IP-laptop-di-WiFi-yang-sama>:8080/api/v1` |
| Browser (demo)    | `http://localhost:8080/api/v1`, lalu tambahkan origin web ke `CORS_ALLOWED_ORIGINS` backend |

Akun demo: `kasir@toko.id` / `kasir123`, atau `owner@toko.id` / `owner123` (bisa juga stok masuk dan penyesuaian stok).

| Perintah | Fungsi |
|----------|--------|
| `flutter analyze` | Lint & analisis statis |
| `flutter test` | Unit test (keranjang, pembulatan, format) dan widget test |
| `flutter build apk --dart-define-from-file=.env` | Build APK Android |
| `flutter build ios --dart-define-from-file=.env` | Build iOS (butuh macOS) |
| `flutter build web --dart-define-from-file=.env` | Build versi web untuk demo |

> URL gambar produk dari backend yang menunjuk ke `localhost` otomatis diarahkan ke host `API_URL`, sehingga gambar tetap tampil di emulator dan HP.

## Fitur

**Tab bar:** Kasir · Transaksi · Stok · Akun (setiap ikon selalu disertai label).

1. **Kasir**
   - Jika shift belum dibuka, kasir diminta mengisi kas awal terlebih dahulu.
   - Grid produk (gambar, nama, harga, stok) dengan pencarian, filter kategori berbentuk chip, dan scan barcode (`mobile_scanner`).
   - Ketuk produk untuk menambahkannya ke keranjang, dengan badge jumlah, toast, dan getaran (haptic).
   - Total belanja selalu terlihat di bagian bawah.
2. **Keranjang**
   - Ubah jumlah dengan tombol −/+ (44 pt), atau geser item ke kiri untuk menghapus (dengan konfirmasi).
   - Bisa menambah diskon nominal atau persen.
   - Rincian subtotal, diskon, dan pajak.
3. **Bayar**
   - Pilih metode: Tunai, QRIS, Transfer, atau Debit (segmented control).
   - Tombol nominal cepat: **Uang Pas, 50.000, 100.000**, plus pembulatan terdekat.
   - Kembalian dihitung otomatis dan ditampilkan besar.
4. **Selesai**
   - Layar sukses dengan nomor invoice, total, dan kembalian.
   - Struk bisa dibagikan (`share_plus`) atau dicetak lewat dialog cetak sistem dalam format PDF 58 mm (`printing`).
   - Tombol **Transaksi Baru** untuk kembali ke layar kasir.

Fitur lain:
- **Transaksi:** riwayat hari ini, 7 hari, atau semua; detail transaksi beserta struk.
- **Stok:** daftar stok (yang paling sedikit tampil di atas) dan filter *Stok Menipis*. Owner/Admin bisa mencatat stok masuk atau penyesuaian dari bottom sheet.
- **Akun:** profil, rekap shift berjalan, **Tutup Shift** (hitung kas, lihat selisih, konfirmasi), info toko, ganti password, dan logout.

Setiap aksi berbahaya (hapus item, kosongkan keranjang, tutup shift, logout) wajib dikonfirmasi lewat dialog iOS. Setiap aksi memberi toast dan getaran.

### Responsif: Ponsel, Tablet, Desktop

| Lebar layar | Tata letak |
|-------------|------------|
| < 900 px (ponsel) | Tab bar di bawah, grid 2 kolom, keranjang di layar terpisah |
| ≥ 900 px (tablet/desktop) | Sidebar ala iPad di kiri, layar kasir *split view* (grid produk + panel keranjang di kanan) sehingga transaksi cukup 2 langkah |
| Semua ukuran | Jumlah kolom grid menyesuaikan lebar (2–8). Konten daftar maksimal 760 px dan form maksimal 600 px, di tengah layar |

Breakpoint didefinisikan di `lib/core/layout/breakpoints.dart`.

### Idempotency

Saat membayar, aplikasi membuat `Idempotency-Key` (UUID) dan menyimpannya di state keranjang. Jika koneksi putus lalu kasir menekan *Selesaikan* lagi, key yang sama dikirim ulang dan server mengembalikan transaksi yang sama, sehingga tidak terjadi transaksi dobel. Key direset otomatis bila isi keranjang berubah.

## Desain

- **Palet iOS** (light & dark otomatis) ada di `lib/core/theme/app_colors.dart`. Nilainya identik dengan web CMS (`web-cms/src/index.css`). Komponen membaca warna lewat `context.colors`, tidak ada warna hardcode di widget.
- **Tipografi HIG:** Large Title 34, Title 28, Headline 17 semibold, Body 17, Subheadline 15, Footnote 13, Caption 12 (`AppText`).
- **Gaya iOS:** header Large Title (`CupertinoSliverNavigationBar`), daftar *inset grouped*, bottom sheet untuk form, segmented control, radius 10–14, bayangan sangat tipis, ikon `CupertinoIcons`.
- **Target sentuh** minimal 44×44 pt. Angka penting (total, kembalian) ditampilkan 40 pt tebal. Format uang: `Rp 25.000`.

## Struktur

```
lib/
├── main.dart, app.dart
├── core/
│   ├── config/      API_URL (dart-define)
│   ├── layout/      breakpoint & MaxWidth
│   ├── network/     ApiClient (Dio + refresh token otomatis), ApiException
│   ├── storage/     token di Keychain/Keystore (flutter_secure_storage)
│   ├── theme/       AppColors (light/dark), AppText, radius & spacing
│   ├── utils/       format Rupiah/tanggal, haptics
│   └── widgets/     LargeTitleScaffold, InsetSection/Row, AppButton, toast,
│                    dialog konfirmasi, EmptyState, ErrorState, Skeleton, input, segmented
├── router/          go_router + tab bar / sidebar
└── features/
    ├── auth/        login, sesi
    ├── catalog/     produk & kategori
    ├── cart/        state keranjang (Riverpod) & kalkulasi total
    ├── cashier/     layar Kasir, Keranjang, Bayar, Selesai, Scanner
    ├── shift/       buka/tutup shift
    ├── transactions/ riwayat & detail
    ├── stock/       stok & stok menipis
    ├── account/     akun, pengaturan toko, tutup shift
    └── receipt/     bagikan / cetak struk
```

## Pemetaan Library

Spesifikasi awal menyebut library Expo/React Native. Karena stack yang dipilih adalah Flutter, dipakai padanannya:

| Spesifikasi | Flutter |
|-------------|---------|
| Expo Router | `go_router` (`StatefulShellRoute` untuk tab) |
| TanStack Query | `flutter_riverpod` (`FutureProvider` / `AsyncNotifier`) |
| Zustand (keranjang) | Riverpod `Notifier` |
| expo-camera | `mobile_scanner` |
| expo-haptics | `HapticFeedback` (bawaan Flutter) |
| Ionicons | `CupertinoIcons` (ikon gaya iOS) |

## Fase 2 (Belum Dikerjakan)

Mode offline + sinkronisasi, dan printer thermal Bluetooth.
