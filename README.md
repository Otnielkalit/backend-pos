# POS Inventory Backend

Backend API untuk sistem pencatatan stok & transaksi retail/grosir (minimarket, toko grosir, dsb).  
Aplikasi ini menangani **pencatatan transaksi dan manajemen stok** — proses pembayaran ditangani di luar sistem.

---

## Arsitektur

Project ini menggunakan **Feature-Based Architecture** di level atas, dengan **Clean Architecture** di dalam masing-masing fitur.

```
Feature-Based + Clean Architecture
├── Setiap fitur punya folder sendiri (auth, product, inventory, transaction, ...)
├── Di dalam tiap fitur: delivery → usecase → repository → entity
├── Komunikasi antar fitur hanya lewat interface (tidak ada direct import)
└── 1 method = 1 file di semua layer (handler, usecase, repository)
```

**Alur dependency:**
```
HTTP Request
    │
    ▼
delivery/http          ← handler, route, dto
    │
    ▼
usecase                ← business logic (tidak tahu HTTP/SQL)
    │
    ▼
repository             ← raw SQL via sqlx + pgx
    │
    ▼
PostgreSQL / Redis
```

**Tech stack utama:**

| Komponen | Teknologi |
|---|---|
| Language | Go 1.26+ |
| HTTP Framework | Gin |
| Database | PostgreSQL 16 |
| DB Access | sqlx + pgx (raw SQL, no ORM) |
| Cache | Redis 7 |
| Auth | JWT (access token only) |
| Logger | zerolog (JSON output) |
| API Docs | Swagger (swaggo) |
| Observability | Grafana + Loki + Prometheus |
| CI/CD | Jenkins |
| Containerization | Docker + Docker Compose |

---

## Prasyarat

Pastikan tools berikut sudah terinstall di lokal:

- [Go 1.26+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [`golang-migrate` CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)
- [`golangci-lint`](https://golangci-lint.run/usage/install/) (untuk linting)
- [`swag` CLI](https://github.com/swaggo/swag) (untuk generate Swagger docs)

```bash
# Install golang-migrate
brew install golang-migrate

# Install golangci-lint
brew install golangci-lint

# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest
```

---

## Clone & Setup

### 1. Clone repository

```bash
git clone https://github.com/Otnielkalit/backend-pos.git
cd backend-pos
```

### 2. Download dependencies

```bash
go mod download
```

### 3. Setup environment variables

```bash
cp .env.example .env
```

Buka `.env` dan isi nilai yang diperlukan:

```env
APP_PORT=8080
APP_ENV=development
APP_NAME=pos-backend

# Database — sesuaikan dengan konfigurasi lokal atau Docker Compose
DB_URL=postgres://posuser:pospassword@localhost:5432/pos_db?sslmode=disable

# Redis
REDIS_URL=redis://localhost:6379

# JWT — ganti dengan random string minimal 32 karakter
JWT_SECRET=change-me-to-a-long-random-secret-at-least-32-chars
```

### 4. Jalankan services (Postgres + Redis + Observability stack)

```bash
make docker-up
```

Ini akan menjalankan:
- **PostgreSQL** → `localhost:5432`
- **Redis** → `localhost:6379`
- **Grafana** → `http://localhost:3000` (login: `admin` / `admin`)
- **Prometheus** → `http://localhost:9090`
- **Loki** → `http://localhost:3100`

### 5. Jalankan database migration

```bash
make migrate-up
```

---

## Menjalankan Aplikasi

```bash
make run
```

Atau langsung dengan Go:

```bash
go run cmd/api/main.go
```

Server akan berjalan di `http://localhost:8080`.

---

## Cara Akses Setelah Running

### Health Check

```bash
curl http://localhost:8080/health
# Response: {"status":"ok"}
```

### API Endpoints

Semua endpoint API berada di bawah prefix `/api/v1`:

```
http://localhost:8080/api/v1/...
```

> Lihat daftar lengkap endpoint di **Swagger UI** di bawah.

### Swagger UI (API Documentation)

```
http://localhost:8080/swagger/index.html
```

> **Regenerate docs** setiap ada perubahan handler/annotation:
> ```bash
> make swagger
> ```

### Grafana Dashboard (Observability)

```
http://localhost:3000
```

- Username: `admin`
- Password: `admin`
- Datasources sudah di-provision otomatis (Prometheus + Loki)

---

## Perintah Umum (Makefile)

```bash
make run              # Jalankan aplikasi
make build            # Build binary ke ./bin/
make test             # Jalankan semua unit test + coverage report
make lint             # Jalankan golangci-lint
make swagger          # Generate/update Swagger docs dari annotation

make migrate-up       # Apply semua pending migration
make migrate-down     # Rollback 1 migration terakhir
make migrate-create name=<nama>   # Buat file migration baru
make migrate-force version=<ver>  # Force migration version (jika dirty)

make docker-up        # Start semua Docker services (background)
make docker-down      # Stop semua Docker services
make docker-logs      # Follow log semua services

make clean            # Hapus binary dan coverage report
make help             # Tampilkan semua perintah yang tersedia
```

---

## Membuat Migration Baru

```bash
# Buat file migration
make migrate-create name=create_products_table

# Akan menghasilkan dua file di folder migrations/:
# 000001_create_products_table.up.sql   ← tulis query CREATE TABLE di sini
# 000001_create_products_table.down.sql ← tulis query DROP TABLE di sini

# Apply migration
make migrate-up

# Rollback jika perlu
make migrate-down
```

**Aturan migration:**
- File `.up.sql` dan `.down.sql` wajib ada dan bisa di-rollback
- Jangan pernah edit migration yang sudah di-apply di `main` — buat migration baru
- Urutan tabel: dari parent ke child (stores → admins → employees → products → inventory → transactions)

---

## Struktur Folder

```
backend-pos/
├── cmd/
│   └── api/
│       └── main.go                  # Entry point, dependency wiring
├── internal/
│   ├── feature/
│   │   ├── auth/                    # Autentikasi (admin & employee)
│   │   ├── employee/                # Manajemen karyawan
│   │   ├── product/                 # Manajemen produk
│   │   ├── inventory/               # Pencatatan stok masuk/keluar
│   │   └── transaction/             # Pencatatan transaksi
│   ├── shared/
│   │   ├── apperror/                # Centralized error type
│   │   ├── middleware/              # Auth, logger, CORS middleware
│   │   ├── response/                # Standard API response wrapper
│   │   └── entity/                  # Shared domain model (JWT claims, dll)
│   └── infrastructure/
│       ├── config/                  # Environment config loader
│       ├── database/                # PostgreSQL connection (pgx + sqlx)
│       ├── cache/                   # Redis client
│       └── logger/                  # zerolog setup
├── migrations/                      # SQL migration files (golang-migrate)
├── observability/
│   ├── prometheus/                  # Prometheus scrape config
│   ├── promtail/                    # Log shipping config
│   └── grafana/                     # Grafana datasource provisioning
├── .env.example                     # Template environment variables
├── .golangci.yml                    # Linter configuration
├── docker-compose.yml               # Local dev services
├── Dockerfile                       # Multi-stage production build
├── Jenkinsfile                      # CI/CD pipeline
└── Makefile                         # Standar perintah project
```

---

## Standard API Response

Semua endpoint mengembalikan format response yang konsisten:

**Success:**
```json
{
  "success": true,
  "message": "Product created successfully",
  "data": { },
  "meta": null
}
```

**Error:**
```json
{
  "success": false,
  "message": "Product not found",
  "error": {
    "code": "NOT_FOUND",
    "details": null
  }
}
```

---

## Kontribusi

1. Buat branch baru dari `main`:
   ```bash
   git checkout -b feature/nama-fitur
   # atau
   git checkout -b fix/nama-bug
   ```

2. Pastikan lint dan test lolos sebelum push:
   ```bash
   make lint
   make test
   ```

3. Buat Pull Request ke `main` dengan deskripsi perubahan yang jelas.

**Format commit message (Conventional Commits):**
```
feat(product): add stock adjustment endpoint
fix(auth): fix token expiry validation
docs(readme): update setup instructions
```

---

## License

MIT
