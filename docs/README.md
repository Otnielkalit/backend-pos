# POS Inventory Backend

Backend API untuk sistem pencatatan stok & transaksi retail/grosir (tipe toko seperti grosir, minimarket, dsb). Aplikasi ini **hanya menangani pencatatan** — proses pembayaran ditangani di luar sistem.

## Fitur Utama (Fase Awal)

- Autentikasi Admin (email/password, Google Sign-In, Sign in with Apple)
- Manajemen Karyawan oleh Admin (satu toko = satu admin)
- Manajemen Produk
- Pencatatan Stok (masuk/keluar)
- Pencatatan Transaksi (recording only, tanpa payment processing)
- Audit log untuk aksi-aksi penting

## Tech Stack

Lihat detail lengkap beserta alasan pemilihannya di [`TECH_STACK.md`](./TECH_STACK.md). Ringkasnya:

- **Go + Gin** — HTTP API
- **PostgreSQL** — database utama
- **Redis** — caching
- **golang-migrate** — database migration
- **Swagger (swaggo)** — API documentation
- **Grafana + Loki + Prometheus** — observability
- **Jenkins** — CI/CD
- **Docker** — containerization

## Arsitektur

Project ini menggunakan **Feature-Based Architecture** dengan **Clean Architecture** di dalam tiap fitur. Penjelasan lengkap, struktur folder, dan aturan komunikasi antar fitur ada di [`ARCHITECTURE.md`](./ARCHITECTURE.md).

## Coding Standards

Semua konvensi penamaan, format kode, error handling, dan aturan commit ada di [`CODING_STANDARDS.md`](./CODING_STANDARDS.md). **Wajib dibaca sebelum kontribusi kode**, termasuk oleh AI assistant (lihat [`AI_RULES.md`](./AI_RULES.md)).

## Persiapan Environment

### Prasyarat
- Go 1.22+
- Docker & Docker Compose
- `golang-migrate` CLI
- PostgreSQL client (opsional, untuk akses manual)

### Instalasi

```bash
# 1. Clone repository
git clone <repo-url>
cd pos-inventory-backend

# 2. Copy environment variable
cp .env.example .env

# 3. Jalankan dependency (Postgres, Redis) via Docker Compose
docker compose up -d

# 4. Jalankan migration
make migrate-up

# 5. Jalankan aplikasi
go run cmd/api/main.go
```

### Environment Variables

Lihat `.env.example` untuk daftar lengkap. Minimal yang dibutuhkan:

```
APP_PORT=8080
DB_URL=postgres://user:pass@localhost:5432/pos_db?sslmode=disable
REDIS_URL=localhost:6379
JWT_SECRET=change-me
GOOGLE_CLIENT_ID=
APPLE_CLIENT_ID=
```

### Perintah Umum (Makefile)

```bash
make run              # jalankan aplikasi
make migrate-up        # jalankan migration
make migrate-down      # rollback 1 migration
make migrate-create name=<nama_migration>   # buat file migration baru
make test              # jalankan seluruh unit test
make lint              # jalankan golangci-lint
make swagger           # generate ulang dokumentasi swagger
```

## Dokumentasi API

Setelah aplikasi berjalan, dokumentasi Swagger dapat diakses di:

```
http://localhost:8080/swagger/index.html
```

## Struktur Dokumentasi Project

| Dokumen | Isi |
|---|---|
| [`README.md`](./README.md) | Pengenalan project (dokumen ini) |
| [`TECH_STACK.md`](./TECH_STACK.md) | Daftar teknologi dan alasan pemilihan |
| [`ARCHITECTURE.md`](./ARCHITECTURE.md) | Struktur arsitektur dan aturan desain |
| [`CODING_STANDARDS.md`](./CODING_STANDARDS.md) | Konvensi kode, naming, commit |
| [`AI_RULES.md`](./AI_RULES.md) | Aturan kerja sama dengan AI assistant |

## Kontribusi

1. Buat branch baru dari `main` mengikuti konvensi di `CODING_STANDARDS.md`.
2. Pastikan `make lint` dan `make test` lolos sebelum membuat pull request.
3. Update dokumentasi terkait (`ARCHITECTURE.md`, Swagger annotation) jika ada perubahan struktur/endpoint.