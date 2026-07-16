# Architecture

## Ringkasan

Backend ini menggunakan **Feature-Based Architecture (FBA)** di level atas, dengan **Clean Architecture (layered)** di dalam masing-masing fitur. Pendekatan ini dipilih karena:

- Semua kode terkait satu fitur berada dalam satu folder → mudah dibaca, mudah onboarding kontributor baru (termasuk AI).
- Tetap punya separation of concern (handler, usecase, repository) sehingga business logic tidak tercampur dengan detail infrastruktur (HTTP, database).
- Satu folder fitur bisa "dipotong" jadi service terpisah di masa depan tanpa refactor besar, jika suatu saat dibutuhkan.

---

## Struktur Folder

```
pos-inventory-backend/
├── cmd/
│   └── api/
│       └── main.go                  # entry point, wiring semua dependency
├── internal/
│   ├── feature/
│   │   ├── auth/
│   │   │   ├── delivery/http/       # handler, route, request/response DTO
│   │   │   ├── usecase/             # business logic
│   │   │   ├── repository/          # implementasi akses data
│   │   │   ├── entity/              # domain model khusus fitur ini
│   │   │   └── contract.go          # interface/port yang dipakai/exposed fitur ini
│   │   ├── employee/
│   │   │   └── ...struktur sama...
│   │   ├── product/
│   │   │   └── ...struktur sama...
│   │   ├── inventory/
│   │   │   └── ...struktur sama...
│   │   └── transaction/
│   │       └── ...struktur sama...
│   ├── shared/                      # kode reusable lintas fitur
│   │   ├── entity/                  # domain model yang dipakai banyak fitur
│   │   ├── middleware/              # auth middleware, request logger, dsb
│   │   ├── response/                # standard API response wrapper
│   │   ├── apperror/                # custom error type terpusat
│   │   └── util/
│   └── infrastructure/
│       ├── database/                # koneksi db (pgx + sqlx), migration runner
│       ├── cache/                   # redis client
│       ├── logger/                  # setup zerolog/zap
│       └── config/                  # load env/config
├── migrations/                      # file migration golang-migrate
├── docs/                            # swagger output (auto-generated)
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── Jenkinsfile
└── go.mod
```

### Struktur di dalam satu fitur (Clean Architecture layer)

**Prinsip utama: 1 method = 1 file** di semua layer (handler, usecase, repository). Tujuannya menjaga LOC per file tetap pendek, review PR lebih fokus, dan menghindari merge conflict antar developer yang mengerjakan operation berbeda di fitur yang sama.

```
feature/product/
├── delivery/http/
│   ├── handler.go           # HANYA struct ProductHandler + constructor NewProductHandler
│   ├── route.go             # registrasi semua route milik fitur ini
│   ├── dto.go               # SEMUA request & response DTO fitur ini (digabung, bukan per method)
│   ├── get_product.go       # handler method GetProduct + Swagger annotation
│   ├── get_product_test.go  # test untuk get_product.go
│   ├── list_product.go      # handler method ListProducts + Swagger annotation
│   ├── list_product_test.go
│   ├── create_product.go    # handler method CreateProduct + Swagger annotation
│   ├── create_product_test.go
│   ├── update_product.go    # handler method UpdateProduct + Swagger annotation
│   ├── update_product_test.go
│   ├── delete_product.go    # handler method DeleteProduct + Swagger annotation
│   └── delete_product_test.go
├── usecase/
│   ├── usecase.go           # HANYA struct productUsecase + constructor NewProductUsecase
│   ├── interface.go         # definisi interface ProductUsecase (contract yang dipakai handler)
│   ├── get_product.go       # business logic: GetProduct
│   ├── get_product_test.go
│   ├── list_product.go      # business logic: ListProducts
│   ├── list_product_test.go
│   ├── create_product.go    # business logic: CreateProduct
│   ├── create_product_test.go
│   ├── update_product.go    # business logic: UpdateProduct
│   ├── update_product_test.go
│   ├── delete_product.go    # business logic: DeleteProduct
│   └── delete_product_test.go
├── repository/
│   ├── repository.go        # HANYA struct productRepository + constructor NewProductRepository
│   ├── interface.go         # definisi interface ProductRepository (contract yang dipakai usecase)
│   ├── get_product.go       # SQL: GetByID
│   ├── list_product.go      # SQL: List (dengan pagination)
│   ├── create_product.go    # SQL: Insert
│   ├── update_product.go    # SQL: Update
│   └── delete_product.go    # SQL: Delete (hard delete + catat ke audit log)
├── entity/
│   └── product.go           # domain model, tidak punya dependency ke layer lain
└── contract.go              # interface yang di-expose ke fitur lain (jika ada)
```

**Alur dependency (dari luar ke dalam):**

```
delivery (HTTP) → usecase (business logic) → repository (data access) → database
                        ↓
                    entity (domain model, tidak bergantung ke layer manapun)
```

Aturan: layer luar boleh bergantung ke layer dalam, **layer dalam tidak boleh bergantung ke layer luar**. `entity` adalah layer paling dalam dan tidak boleh mengimpor apapun dari `delivery`, `usecase`, atau `repository`.

---

## Komunikasi Antar Fitur

Fitur **tidak boleh saling import struct/repository secara langsung**. Jika fitur A butuh data dari fitur B (contoh: `transaction` butuh cek stok dari `inventory`):

1. Fitur A (`transaction`) mendefinisikan interface kecil di `contract.go` miliknya sendiri, sesuai kebutuhannya. Contoh: `InventoryChecker` dengan method `CheckStock(ctx, productID) (int, error)`.
2. Fitur B (`inventory`) menyediakan implementasi yang memenuhi interface tersebut.
3. Wiring (menghubungkan implementasi ke interface) dilakukan di `main.go`, bukan di dalam fitur itu sendiri.

> Prinsip Go: **interface didefinisikan di sisi consumer, bukan di sisi provider.** Ini mencegah circular dependency dan menjaga fitur tetap independen/loosely coupled.

Jika ada entity yang mulai dibutuhkan oleh lebih dari satu fitur, pindahkan ke `shared/entity` — jangan duplikat.

---

## Multi-Actor Authentication (Admin & Employee)

Sistem punya dua tipe aktor: `admin` dan `employee`, masing-masing punya tabel dan flow auth berbeda, tapi disatukan lewat JWT claim yang seragam:

```json
{
  "actor_id": "uuid",
  "actor_type": "admin | employee",
  "store_id": "uuid",
  "role": "owner | employee",
  "exp": 1234567890
}
```

Middleware otorisasi bekerja berdasarkan claim ini, bukan berdasarkan tabel asal aktor. Setiap query data harus di-scope dengan `store_id` dari claim tersebut — tidak ada endpoint yang mengambil data lintas toko.

---

## Standard Response Format

Semua endpoint API mengembalikan format response yang konsisten:

```json
{
  "success": true,
  "message": "Product created successfully",
  "data": { },
  "meta": null
}
```

Untuk error:

```json
{
  "success": false,
  "message": "Product not found",
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "details": null
  }
}
```

Wrapper ini didefinisikan sekali di `shared/response` dan dipakai semua handler.

---

## Error Handling

Error terpusat lewat custom type `AppError` di `shared/apperror`, berisi `Code`, `Message`, `HTTPStatus`. Usecase mengembalikan `AppError`, handler tinggal translate ke response tanpa if-else berlapis. Detail lengkap ada di `CODING_STANDARDS.md`.

---

## Multi-Tenant Architecture

Sistem ini menggunakan model **Single-Instance Multi-Tenant**: satu binary dan satu database melayani banyak toko sekaligus. Isolasi data antar toko dilakukan sepenuhnya lewat `store_id` dari JWT claim.

- Setiap admin/employee hanya bisa mengakses data toko mereka sendiri.
- Tidak ada endpoint yang mengembalikan data lintas toko.
- Semua query repository yang menyentuh data toko **wajib** menyertakan `WHERE store_id = $n` — tidak ada pengecualian.
- `store_id` diambil dari JWT claim yang sudah divalidasi middleware, bukan dari request body/query param (mencegah manipulasi klien).

---

## Strategi Pagination

Pagination disesuaikan per entity berdasarkan volume data yang diantisipasi:

| Entity | Strategi | Alasan |
|---|---|---|
| `products`, `employees`, `categories` | **Offset-based** | Volume per toko kecil (<ribuan), UI butuh navigasi halaman, query sederhana |
| `transactions`, `stock_adjustments`, `audit_logs` | **Cursor-based** | Berpotensi jutaan baris, offset tidak efisien di page tinggi, data append-only |

Cursor untuk `transactions` dan `audit_logs` menggunakan composite `(created_at, id)` agar bisa di-index dan deterministik saat ada data dengan timestamp sama.

Contoh response pagination offset-based:
```json
{
  "data": [...],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 150
  }
}
```

Contoh response pagination cursor-based:
```json
{
  "data": [...],
  "meta": {
    "next_cursor": "2026-07-16T10:00:00Z_uuid-xxx",
    "has_more": true
  }
}
```

---

## Prinsip Umum

1. **Business logic tidak boleh tahu soal HTTP atau SQL.** Usecase hanya menerima/mengembalikan entity dan primitive/DTO, bukan `*gin.Context` atau `*sql.Rows`.
2. **Semua akses database melalui repository**, tidak ada query SQL langsung di usecase atau handler.
3. **Context (`context.Context`) selalu di-pass** dari handler sampai ke repository, untuk keperluan timeout dan tracing di masa depan.
4. **Concurrency-sensitive operation** (potong stok) wajib menggunakan row-level locking (`SELECT ... FOR UPDATE`) di level repository.
5. **Idempotency** pada endpoint pencatatan transaksi menggunakan client-generated `transaction_id` (UUID) dengan unique constraint di database.

---

## Yang Belum Diputuskan (Open Decisions)

Bagian ini sengaja ditulis eksplisit supaya AI/kontributor baru tidak berasumsi sendiri. Update bagian ini begitu keputusan dibuat.

- [ ] Kebijakan retensi audit log (disimpan selamanya atau ada retention policy?)