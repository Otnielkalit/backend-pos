# Tech Stack

Dokumen ini menjelaskan teknologi yang digunakan di backend **POS Inventory Backend**, beserta alasan pemilihannya. Tujuannya supaya setiap kontributor (manusia maupun AI) memahami *kenapa* sebuah tool dipilih, bukan cuma *apa* yang dipakai.

---

## 1. Bahasa & Framework

| Teknologi | Fungsi | Alasan |
|---|---|---|
| **Go (Golang)** | Bahasa utama backend | Performa tinggi, static typing, kompilasi cepat, cocok untuk service yang butuh concurrency (transaksi, stok) |
| **Gin** | HTTP web framework | Berbasis `net/http` standar sehingga kompatibel dengan ekosistem middleware Go secara luas (observability, tracing, dsb), stabil, dan tim sudah familiar dari pengalaman kerja sehari-hari |

## 2. Database & Storage

| Teknologi | Fungsi | Alasan |
|---|---|---|
| **PostgreSQL** | Database utama (relational) | Mendukung JSONB untuk atribut fleksibel (varian produk), MVCC yang baik untuk concurrent write (potong stok bersamaan), window function & CTE untuk reporting |
| **pgx** | PostgreSQL driver (low-level) | Driver native Postgres paling performa, mendukung `pgx.Batch` untuk bulk query dan named parameter dengan baik |
| **sqlx** | Extension di atas `database/sql` | Scan result ke struct otomatis (`StructScan`, `Select`, `Get`) tanpa ORM — tetap pakai raw SQL tapi tidak boilerplate |
| **golang-migrate** | Database migration tool | Versioning schema database berbasis file SQL, dijalankan lewat CLI/Makefile, mudah di-rollback |
| **Redis** | Caching & rate limiting | Cache data yang sering dibaca (produk, harga), serta bisa dipakai untuk rate limiting endpoint sensitif |

> **Keputusan**: Tidak menggunakan ORM (GORM). Semua query ditulis sebagai raw SQL di dalam folder `repository`. Alasan: kontrol penuh atas query, tidak ada magic behavior, dan lebih mudah di-optimize untuk query reporting yang kompleks.

## 3. Autentikasi

| Teknologi | Fungsi | Alasan |
|---|---|---|
| **JWT (JSON Web Token)** | Session/auth token | Stateless, cocok untuk backend API yang diakses dari mobile/web app kasir |
| **OAuth2 (Google Sign-In)** | Login admin via Google | Kemudahan onboarding admin tanpa perlu bikin password baru |
| **Sign in with Apple** | Login admin via Apple | Wajib disediakan jika aplikasi client (iOS) menyediakan login pihak ketiga lain (ketentuan App Store) |

> **Keputusan**: Tidak menggunakan refresh token. Access token saja dengan expiry yang reasonable (misal: 24 jam untuk kasir, 8 jam untuk admin). Jika token expire, user login ulang. Token revocation dilakukan lewat blacklist sederhana di Redis (simpan `jti` dari token yang di-logout).

## 4. Dokumentasi API

| Teknologi | Fungsi | Alasan |
|---|---|---|
| **Swagger (swaggo/swag)** | API documentation | Generate dokumentasi dari annotation comment di kode, selalu sinkron dengan implementasi selama comment di-maintain |

## 5. Observability

| Teknologi | Fungsi | Alasan |
|---|---|---|
| **Grafana** | Dashboard & visualisasi | Satu tempat untuk memantau log, metrics, dan (nanti) tracing |
| **Loki + Promtail** | Log aggregation | Menampung structured log dari aplikasi (format JSON), terintegrasi langsung dengan Grafana |
| **Prometheus** | Metrics collection | Memantau request rate, latency, error rate, DB connection pool, dsb |
| **zerolog / zap** | Structured logging library (Go) | Output JSON yang mudah di-parse Loki, mendukung field kontekstual seperti `request_id` |

> Tracing (Tempo/Jaeger) belum diimplementasikan di fase ini karena arsitektur masih monolith. Akan dipertimbangkan ulang jika service dipecah.

## 6. CI/CD

| Teknologi | Fungsi | Alasan |
|---|---|---|
| **Jenkins** | Continuous Integration/Deployment | Sesuai infrastruktur yang tersedia saat ini, mendukung pipeline-as-code (`Jenkinsfile`) |

> Jika ke depan repository dipindah/dipusatkan di GitHub dan tidak ada kebutuhan khusus on-prem, GitHub Actions bisa dipertimbangkan sebagai alternatif yang lebih ringan untuk di-maintain.

## 7. Tooling Pendukung

| Teknologi | Fungsi | Alasan |
|---|---|---|
| **go-playground/validator** | Request validation | Validasi konsisten di level DTO/request, menghindari validasi manual berserakan di usecase |
| **Docker & Docker Compose** | Containerization | Environment development yang konsisten antar developer, memudahkan deployment |
| **Makefile** | Task runner | Standarisasi perintah umum (migration, build, run, test) supaya semua kontributor pakai command yang sama |
| **mockery** (atau manual mock) | Testing | Mocking interface untuk unit test yang terisolasi per fitur |

---

## Prinsip Pemilihan Teknologi

1. **Boring technology first** — pilih tool yang matang dan well-documented, bukan yang paling baru/trendy, kecuali ada alasan teknis kuat.
2. **Kompatibilitas ekosistem** di atas raw performance — misalnya Gin dipilih atas Fiber karena kompatibilitas `net/http`.
3. **Semua keputusan teknologi harus terdokumentasi di sini**, termasuk kapan sebuah tool digantikan tool lain dan alasannya (tambahkan di bagian *Changelog* di bawah).

---

## Changelog

| Tanggal | Perubahan | Alasan |
|---|---|---|
| 2026-07-16 | Pilih `pgx + sqlx` sebagai database access layer, tidak pakai GORM | Kontrol penuh atas SQL, lebih mudah optimize query reporting kompleks |
| 2026-07-16 | Tidak pakai refresh token — access token only | Simplicity first, user kasir tidak keberatan login ulang saat token expire |
| 2026-07-16 | Tidak ada soft delete di tabel manapun | Transaction dan stock adjustment bersifat immutable (gunakan status), produk dan karyawan yang dihapus cukup hard delete dengan audit log |