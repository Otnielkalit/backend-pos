# Coding Standards

Dokumen ini adalah acuan wajib untuk seluruh kode yang ditulis di project ini, baik oleh manusia maupun AI. Tujuannya menjaga konsistensi supaya kode dari siapapun/apapun terasa seperti ditulis oleh satu orang yang sama.

---

## 1. Format & Linting

- Semua kode **wajib** melewati `gofmt` / `goimports` sebelum commit.
- Gunakan `golangci-lint` dengan konfigurasi di `.golangci.yml` (linter minimal: `govet`, `errcheck`, `staticcheck`, `unused`).
- Tidak ada exception "skip lint" tanpa alasan yang dikomentari di kode (`//nolint:xxx // alasan`).

## 2. Naming Convention

### Package
- Nama package singular, lowercase, tanpa underscore: `product`, bukan `products` atau `Product`.
- Nama package = nama folder fitur.

### File
- `snake_case.go` — contoh: `product_usecase.go`, `product_repository.go`.
- File test selalu `_test.go` di folder yang sama dengan yang diuji.

### Variable & Function
- `camelCase` untuk unexported, `PascalCase` untuk exported.
- Nama boolean diawali `is`, `has`, `can`: `isActive`, `hasStock`, `canVoid`.
- Hindari singkatan tidak jelas (`prd` untuk `product`), kecuali sudah umum di Go (`ctx`, `err`, `req`, `res`, `id`).

### Struct & Interface
- Struct: `PascalCase` noun. Contoh: `Product`, `TransactionItem`.
- Interface: `PascalCase`, biasanya diakhiri sesuai perannya, bukan diawali `I`. Contoh: `ProductRepository`, `InventoryChecker` — **bukan** `IProductRepository`.
- Interface method satu fungsi sebaiknya dinamai sesuai aksi: `GetByID`, `Create`, `UpdateStock`.

### Konstanta
- `PascalCase` jika exported, dikelompokkan dengan `const ( ... )` block, bukan tersebar.
- Enum-like value menggunakan `type Status string` + konstanta bernama, bukan magic string:
```go
type TransactionStatus string

const (
    TransactionStatusDraft     TransactionStatus = "draft"
    TransactionStatusCompleted TransactionStatus = "completed"
    TransactionStatusVoided    TransactionStatus = "voided"
)
```

### Database
- Nama tabel: `snake_case`, plural. Contoh: `products`, `transaction_items`.
- Nama kolom: `snake_case`. Contoh: `created_at`, `store_id`.
- Foreign key: `<singular_table>_id`. Contoh: `product_id`, `store_id`.
- Setiap tabel wajib punya `id` (UUID), `created_at`, `updated_at`.
- **Tidak ada soft delete** (`deleted_at`) di project ini. Tabel transaksi dan stok bersifat immutable (gunakan kolom `status`). Data yang dihapus (produk, karyawan) dicatat di audit log sebelum dihapus permanen.

### API Endpoint
- REST, plural noun, `kebab-case` jika multi-kata: `/api/v1/products`, `/api/v1/stock-adjustments`.
- Versioning di path: `/api/v1/...`.
- Query param untuk filter/pagination: `snake_case` — `?store_id=xxx&page=1&limit=20`.

### JSON (request/response body)
- `snake_case` untuk semua field JSON, konsisten dengan kolom database. Gunakan tag json eksplisit di semua struct DTO:
```go
type ProductResponse struct {
    ID        string  `json:"id"`
    Name      string  `json:"name"`
    Price     float64 `json:"price"`
    CreatedAt string  `json:"created_at"`
}
```

## 3. Database Access Pattern

Gunakan `sqlx` untuk semua akses database. Tidak ada ORM. Semua SQL ditulis eksplisit di dalam folder `repository`.

```go
// Contoh query satu row
func (r *productRepository) GetByID(ctx context.Context, storeID, productID string) (*entity.Product, error) {
    var product entity.Product
    query := `SELECT id, name, price, store_id, created_at FROM products WHERE id = $1 AND store_id = $2`
    if err := r.db.GetContext(ctx, &product, query, productID, storeID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, apperror.NewNotFound("product not found")
        }
        return nil, fmt.Errorf("productRepository.GetByID: %w", err)
    }
    return &product, nil
}

// Contoh query banyak row
func (r *productRepository) List(ctx context.Context, storeID string, page, limit int) ([]entity.Product, int, error) {
    var products []entity.Product
    offset := (page - 1) * limit
    query := `SELECT id, name, price FROM products WHERE store_id = $1 ORDER BY name ASC LIMIT $2 OFFSET $3`
    if err := r.db.SelectContext(ctx, &products, query, storeID, limit, offset); err != nil {
        return nil, 0, fmt.Errorf("productRepository.List: %w", err)
    }
    // ... count query terpisah
    return products, total, nil
}
```

Aturan wajib:
- Selalu gunakan `GetContext`, `SelectContext`, `ExecContext` (bukan versi tanpa context).
- Selalu sertakan `store_id` sebagai kondisi WHERE untuk query yang menyentuh data toko.
- Wrap error repository dengan `fmt.Errorf("<repositoryName>.<methodName>: %w", err)` untuk stack trace yang informatif.

## 4. Konvensi Pagination

### Offset-based (untuk `products`, `employees`)
Request query param: `?page=1&limit=20` (default limit 20, max 100).
Response `meta`:
```go
type PaginationMeta struct {
    Page  int `json:"page"`
    Limit int `json:"limit"`
    Total int `json:"total"`
}
```

### Cursor-based (untuk `transactions`, `stock_adjustments`, `audit_logs`)
Request query param: `?cursor=<value>&limit=20`. Cursor pertama kosong (fetch awal).
Cursor format: `<created_at_RFC3339>_<id>` — contoh: `2026-07-16T10:00:00Z_uuid-xxx`.
Response `meta`:
```go
type CursorMeta struct {
    NextCursor string `json:"next_cursor"` // kosong jika tidak ada halaman berikutnya
    HasMore    bool   `json:"has_more"`
}
```

## 5. Struktur Fungsi

- Satu fungsi idealnya < 40 baris. Jika lebih, pertimbangkan ekstrak ke fungsi/helper lain.
- Early return untuk validasi/error, hindari nested if berlapis:
```go
// Good
func (u *productUsecase) Create(ctx context.Context, req CreateProductRequest) (*Product, error) {
    if req.Name == "" {
        return nil, apperror.NewBadRequest("name is required")
    }
    if req.Price <= 0 {
        return nil, apperror.NewBadRequest("price must be greater than 0")
    }
    // proses utama
}
```

## 6. Error Handling

- Semua error dari usecase/repository dibungkus dengan context tambahan menggunakan `fmt.Errorf("...: %w", err)`, jangan pernah `panic` untuk expected error.
- Gunakan custom error `AppError` (`shared/apperror`) untuk error yang perlu diterjemahkan ke HTTP status tertentu.
- Jangan pernah `_ = err` (ignore error) tanpa alasan eksplisit dan komentar.

```go
type AppError struct {
    Code       string
    Message    string
    HTTPStatus int
}

func NewNotFound(message string) *AppError {
    return &AppError{Code: "NOT_FOUND", Message: message, HTTPStatus: http.StatusNotFound}
}
```

## 7. Testing

- Setiap usecase wajib punya unit test dengan mock repository.
- Nama test: `Test<FunctionName>_<Scenario>`, contoh: `TestCreateProduct_ShouldReturnErrorWhenPriceIsZero`.
- Gunakan table-driven test untuk skenario dengan banyak variasi input.
- Target minimal coverage untuk usecase layer: 70% (disesuaikan seiring project berjalan).

## 8. Commit & Branch

### Commit message — Conventional Commits
```
<type>(<scope>): <description>

feat(product): add stock adjustment endpoint
fix(auth): fix token expiry validation
refactor(transaction): extract validation to separate function
docs(readme): update setup instructions
```
Type yang dipakai: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`, `perf`.

### Branch naming
```
feature/<nama-fitur-singkat>
fix/<nama-bug-singkat>
```
Contoh: `feature/stock-adjustment`, `fix/token-expiry`.

## 9. Comment & Dokumentasi Kode

- Comment menjelaskan **kenapa**, bukan **apa** (kode sudah menjelaskan apa).
- Setiap handler yang exposed via Swagger wajib punya annotation lengkap:
```go
// CreateProduct godoc
// @Summary      Create a new product
// @Description  Create a new product for the authenticated store
// @Tags         product
// @Accept       json
// @Produce      json
// @Param        request body CreateProductRequest true "Product payload"
// @Success      201 {object} response.Success{data=ProductResponse}
// @Failure      400 {object} response.Error
// @Router       /products [post]
func (h *ProductHandler) CreateProduct(c *gin.Context) { ... }
```

## 10. Larangan

- ❌ Tidak ada SQL mentah di luar folder `repository`.
- ❌ Tidak ada `*gin.Context` yang di-pass ke usecase (gunakan `context.Context`).
- ❌ Tidak ada fitur yang langsung import struct/repository fitur lain (harus lewat interface).
- ❌ Tidak ada magic string/number tanpa konstanta bernama.
- ❌ Tidak ada credential/secret hardcoded di kode (selalu lewat environment variable).