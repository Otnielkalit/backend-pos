# AI Rules

Dokumen ini adalah instruksi wajib untuk AI assistant (Claude, Copilot, Cursor, atau tools sejenis) yang membantu development di project ini. Tujuannya supaya output AI konsisten dengan arsitektur dan standar yang sudah ditetapkan, bukan mengikuti pola generik dari training data.

**Sebelum mengerjakan task apapun, AI wajib membaca:**
1. `ARCHITECTURE.md` — untuk memahami struktur folder dan aturan dependency
2. `CODING_STANDARDS.md` — untuk konvensi naming, error handling, dan format kode
3. `TECH_STACK.md` — untuk mengetahui library/tool yang sudah disetujui dipakai

---

## 1. Prinsip Utama

- **Ikuti pola yang sudah ada di codebase**, jangan perkenalkan pola/library baru tanpa diminta eksplisit. Jika ada 2 cara melakukan sesuatu, pakai cara yang sudah dipakai di fitur lain, bukan cara "terbaik menurut AI".
- **Jangan mengubah struktur folder** yang sudah ditetapkan di `ARCHITECTURE.md` tanpa konfirmasi eksplisit dari developer.
- **Jangan menambahkan dependency/library baru** di `go.mod` tanpa menyebutkan secara eksplisit ke developer bahwa ini dependency baru dan alasannya.
- Ketika ragu antara solusi cepat vs solusi yang konsisten dengan arsitektur, **selalu pilih yang konsisten dengan arsitektur**, walau lebih verbose.

## 2. Saat Membuat Fitur Baru

1. Ikuti struktur folder fitur yang sudah ada persis: `delivery/http`, `usecase`, `repository`, `entity`, `contract.go`.
2. Jangan langsung import struct/repository dari fitur lain — buat interface di `contract.go` fitur yang membutuhkan (lihat aturan komunikasi antar fitur di `ARCHITECTURE.md`).
3. Setiap handler baru yang expose ke publik **wajib** disertai Swagger annotation lengkap (lihat contoh di `CODING_STANDARDS.md` bagian 7).
4. Setiap usecase baru **wajib** disertai unit test dengan mock repository, minimal untuk skenario sukses dan skenario error utama.
5. Semua endpoint baru wajib mengikuti standard response format (`shared/response`) — jangan buat format response custom per endpoint.
6. Semua query database yang menyentuh data toko wajib di-scope dengan `store_id` dari JWT claim — tidak ada pengecualian.

## 3. Saat Mengubah Kode yang Sudah Ada

- **Minimal, targeted changes** — jangan melakukan rewrite besar-besaran kalau task-nya hanya perbaikan kecil. Jangan "sekalian merapikan" kode lain yang tidak diminta.
- Jika menemukan kode yang tidak sesuai standar saat sedang mengerjakan task lain, **laporkan ke developer** (sebagai catatan/saran), jangan langsung diubah di luar scope task.
- Jangan menghapus komentar `// TODO` atau `// Yang Belum Diputuskan` kecuali task tersebut memang menyelesaikan TODO itu.

## 4. Larangan Eksplisit untuk AI

- ❌ Jangan generate SQL mentah di luar folder `repository`.
- ❌ Jangan pass `*gin.Context` ke usecase — selalu gunakan `context.Context`.
- ❌ Jangan hardcode credential, API key, atau secret apapun di kode. Selalu gunakan environment variable dan tambahkan contohnya (tanpa value asli) ke `.env.example`.
- ❌ Jangan membuat abstraksi/interface baru yang tidak dibutuhkan saat ini ("just in case"). Ikuti prinsip YAGNI kecuali developer eksplisit minta persiapan untuk kebutuhan masa depan.
- ❌ Jangan mengasumsikan requirement bisnis yang belum dikonfirmasi (contoh: kebijakan void transaksi, retention audit log). Jika requirement belum jelas dari dokumen atau task, **tanyakan ke developer**, jangan berasumsi sendiri.
- ❌ Jangan mengubah isi `ARCHITECTURE.md`, `CODING_STANDARDS.md`, atau `TECH_STACK.md` tanpa diminta eksplisit — dokumen ini adalah keputusan yang disengaja, bukan draft.

## 5. Format Output yang Diharapkan

- Saat AI membuat/mengubah kode, sertakan penjelasan singkat **apa yang diubah dan kenapa**, bukan hanya kode mentah.
- Saat AI ragu terhadap suatu keputusan desain (misal: nama kolom, struktur response), **tanyakan terlebih dahulu** daripada menebak dan melanjutkan.
- Saat AI menambahkan migration baru, gunakan `make migrate-create name=<deskripsi>` sebagai referensi, dan sertakan file `up` dan `down` yang lengkap (bisa di-rollback).

## 6. Update Dokumen Ini

Jika ada keputusan arsitektur atau aturan baru yang disepakati selama development (misalnya keputusan di bagian "Yang Belum Diputuskan" di `ARCHITECTURE.md` sudah difinalisasi), dokumen-dokumen ini **wajib diupdate** di PR yang sama — jangan biarkan dokumentasi basi (out of sync) dengan kode aktual.

---

> Dokumen ini adalah living document. Update sesuai perkembangan project, dan pastikan seluruh tim (termasuk sesi AI baru) selalu merujuk ke versi terbaru.