# Blueprint: Multi-Context Task/Project Tracker

Dokumen ini adalah acuan pengerjaan proyek dari awal sampai selesai. Tujuan utama proyek: **mengasah skill Go sebagai backend engineer**, dengan pendekatan tanpa framework (stdlib murni) supaya fundamental HTTP handling & database access benar-benar dipahami, bukan sekadar dipakai.

---

## 1. Tujuan Proyek

- Alat pencatat task/to-do harian yang dipakai sendiri secara nyata (kerja PT IGP, kuliah, side project).
- Task dikelompokkan berdasarkan **context** (bukan sekadar checklist generik).
- Semua task aktif (belum selesai) selalu muncul, difilter secara dinamis dari frontend (Overdue / Hari ini / Akan datang / Tanpa deadline).
- Backend Go dipakai sebagai medium belajar mendalam: routing native, query manual ke Postgres, middleware manual, auth manual, filter dinamis yang aman dari SQL injection.

---

## 2. Tech Stack

| Layer | Pilihan | Alasan |
|---|---|---|
| Bahasa Backend | Go | Fokus pembelajaran utama |
| HTTP Layer | `net/http` stdlib murni (Go 1.22+, `ServeMux` pattern matching) | Tanpa framework — belajar routing, middleware, request/response handling dari dasar |
| Database Driver | `pgx` (murni, tanpa `sqlx`) | Native Postgres driver, scan manual → memahami mapping row-ke-struct secara eksplisit |
| Migration Tool | `golang-migrate` | Schema versioning lewat file `.sql`, bukan edit manual via DBeaver |
| Database | PostgreSQL (Supabase) | Direct connection (bukan pooled), gratis selamanya (dengan catatan auto-pause 7 hari idle) |
| Frontend | Astro (+ React island bila perlu interaktivitas) | Static-first, hanya hydrate bagian yang genuinely interactive |
| Auth | JWT sederhana (self-implemented) | Single-user personal tool, tetap diimplementasi untuk latihan |
| CLI Companion (opsional, fase lanjut) | `cobra` atau `urfave/cli` | Tambah task langsung dari terminal tanpa buka browser |

---

## 3. Arsitektur & Deployment

```
[ Astro Frontend ]  --HTTPS-->  [ Go Backend API ]  --pgx-->  [ PostgreSQL ]
   Cloudflare Pages                   Render                    Supabase
    / Vercel                       (cold start ~30-60s           (Direct Connection,
                                     saat idle >15 menit)          bukan pooled)
```

| Komponen | Platform Deploy | Catatan Operasional |
|---|---|---|
| Frontend (Astro) | Cloudflare Pages **atau** Vercel | Static asset, tidak ada sleep/cold-start. Pilih salah satu — keduanya setara untuk static site. |
| Backend (Go API) | Render (free tier) | Sleep setelah 15 menit idle, cold start 30–60 detik pada request pertama. Mitigasi: heartbeat ping (lihat bagian 7). |
| Database | Supabase (PostgreSQL, free tier) | Project auto-pause setelah 7 hari tanpa aktivitas. Wajib pakai **Direct Connection string** (port 5432), bukan pooled (port 6543), karena backend berjalan sebagai long-running process yang mengelola connection pool sendiri (`pgxpool`). |

**Prinsip biaya: full gratis selamanya.** Semua platform di atas free tier permanen (bukan trial), trade-off-nya di UX (cold start/pause), bukan di biaya.

---

## 4. Struktur Folder Backend

```
/cmd
  /api
    main.go              → entry point, setup server, wiring dependency, graceful shutdown
/internal
  /handler                → terima HTTP request, panggil service, tulis response
  /service                → business logic (validasi, orchestration)
  /repository              → query manual ke Postgres via pgx
  /model                   → struct (Task, Context, dll)
  /middleware               → logging, auth (JWT), CORS
  /router                    → setup ServeMux, daftar semua route
/migrations                 → file migrasi SQL (golang-migrate)
```

---

## 5. Skema Database (draft awal)

### Tabel `contexts`
```sql
CREATE TABLE contexts (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    color VARCHAR(7),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Tabel `tasks`
```sql
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    context_id INT REFERENCES contexts(id) ON DELETE SET NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'todo',  -- todo | in_progress | done
    due_date DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_context_id ON tasks(context_id);
CREATE INDEX idx_tasks_due_date ON tasks(due_date);
```

**Catatan desain:**
- `status` pakai VARCHAR (bukan native ENUM Postgres) agar fleksibel menambah value baru tanpa migration rumit; validasi value di-enforce di layer aplikasi Go.
- `completed_at` terpisah dari `status` untuk kebutuhan analytics (laporan task selesai per periode).
- `context_id` awalnya direncanakan `SET NULL`, namun **telah diubah menjadi `ON DELETE CASCADE`** sesuai kebutuhan user. Jadi, jika sebuah context dihapus, seluruh task dan subtask di dalamnya akan otomatis ikut terhapus.
- Default view "task aktif" = `WHERE status != 'done'` — sudah otomatis mencakup overdue, hari ini, dan task masa depan tanpa logic carry-over tambahan.

### Tabel `subtasks`
```sql
CREATE TABLE subtasks (
    id SERIAL PRIMARY KEY,
    task_id INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    is_done BOOLEAN NOT NULL DEFAULT false,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_subtasks_task_id ON subtasks(task_id);
```

**Catatan desain subtask:**
- `task_id` pakai `ON DELETE CASCADE` (berbeda dari `tasks.context_id` yang `SET NULL`) — subtask tidak punya arti berdiri sendiri tanpa parent task, jadi wajar ikut terhapus.
- `is_done` boolean sederhana (bukan `status` bertingkat seperti task utama) — subtask cukup dua state: selesai/belum.
- `sort_order` untuk menjaga urutan checklist sesuai kehendak user, bukan sekadar `created_at`.
- **Keputusan desain terbuka**: apakah task utama otomatis jadi `done` ketika semua subtask-nya selesai? Jika ya, logic ini ditulis di **service layer** (bukan database trigger) agar tetap terkontrol dan mudah ditelusuri.

---

## 6. Prinsip Engineering yang Wajib Dipegang di Setiap Fase

Ini bukan sekadar checklist fitur — ini kebiasaan yang harus terbawa di setiap baris kode, di project ini maupun project Go berikutnya:

1. **Fail fast.** Kalau ada konfigurasi/dependency yang gagal (env var kosong, koneksi DB gagal), program harus berhenti jelas di awal (`log.Fatal`), bukan lanjut jalan "setengah sehat" dan baru error membingungkan saat runtime.
2. **Error wrapping, bukan error swallowing.** Selalu `fmt.Errorf("konteks: %w", err)` saat meneruskan error ke pemanggil — jangan pernah `if err != nil { return nil }` tanpa jejak. Error harus traceable sampai ke akar penyebabnya.
3. **Eksplisit di atas magic.** Tidak ada abstraksi yang menyembunyikan apa yang benar-benar terjadi (tanpa ORM, tanpa framework routing). Setiap mapping data, setiap query, ditulis dan dipahami baris per baris.
4. **Resource harus dibersihkan secara sadar.** Setiap `Open`/`New`/`Begin` (koneksi, transaction, file) punya pasangan `Close`/`Rollback`/`Commit` yang eksplisit dan dijamin jalan (`defer`), termasuk di jalur error.
5. **Context menyertai setiap operasi I/O.** Query DB, request eksternal, semua menerima `context.Context` agar bisa dibatalkan/timeout — tidak ada operasi I/O yang "menggantung tanpa batas".
6. **Validasi di boundary, bukan di tengah.** Input dari luar (request body, query param) divalidasi begitu masuk ke handler, sebelum menyentuh business logic atau database.
7. **Security bukan pengecualian.** Query dinamis selalu parameterized, tidak pernah concatenate string SQL dari input user — bukan hanya "untuk sekarang", tapi standar tetap.
8. **Schema berubah lewat migration file, bukan lewat DBeaver manual.** Setiap perubahan struktur database punya jejak versi di git.

---

## 7. Rencana Fase Pengerjaan (Detail)

### Fase 1 — Fondasi Backend ✅ (selesai)
1. Inisialisasi Go module (`go mod init`) dan struktur folder sesuai Bagian 4.
2. Setup `.env` untuk development lokal (`DATABASE_URL`, `PORT`), pastikan masuk `.gitignore`.
3. Tulis `internal/model` — struct dasar (`Context`, `Task`, `Subtask`) dengan mapping nullable field ke pointer.
4. Tulis `internal/repository/db.go` — `pgxpool.New()` + `Ping()` eksplisit (fail fast saat koneksi gagal).
5. Tulis migration awal (`golang-migrate`): tabel `contexts`, `tasks`, `subtasks` beserta index.
6. Tulis `cmd/api/main.go` minimal: load env, buka pool, satu endpoint `/health`, `ListenAndServe`.
7. Verifikasi end-to-end: server nyala, `/health` merespons, migration berhasil diterapkan (dicek via `psql`/DBeaver).

### Fase 2 — CRUD Dasar & Query Layer ✅ (selesai)
1. Tulis `internal/repository/task_repository.go` — fungsi `Create`, `GetByID`, `List` (tanpa filter dulu), `Update`, `Delete`, semua pakai `pgx` dengan `Scan()` manual.
2. Tulis `internal/repository/context_repository.go` — CRUD dasar untuk `contexts` (sudah mendukung Create, Update, Delete secara penuh).
3. Tulis `internal/service/task_service.go` — orkestrasi antara handler dan repository, tempat business rule diletakkan (misal: set `completed_at` otomatis saat status berubah jadi `done`).
4. Tulis `internal/handler/task_handler.go` — parse request (JSON decode, path value), panggil service, tulis response JSON. Termasuk validasi input dasar (title tidak boleh kosong, status harus salah satu dari enum yang valid).
5. Buat helper response standar (`internal/handler/response.go`) — format sukses dan error yang konsisten di semua endpoint, supaya tidak menulis ulang pola yang sama di tiap handler.
6. Tulis `internal/router/router.go` — daftarkan semua route task & context ke satu `ServeMux`, dipisah per file resource agar rapi saat endpoint bertambah.
7. Uji manual tiap endpoint dengan `curl` atau REST client (create, list, update status, delete) sebelum lanjut.

### Fase 3 — Filter Dinamis, Subtask, dan Pagination ✅ (selesai)
1. Desain struct filter (`TaskFilter`) yang menampung parameter opsional: `status`, `context_id`, `overdue`, `date_range`.
2. Tulis query builder aman di repository — WHERE clause dirakit dari kombinasi filter yang ada, selalu parameterized (`$1`, `$2`, dst), dengan whitelist field yang boleh difilter.
3. Implementasi pagination keyset (berbasis `id`/`created_at`, bukan `OFFSET`) di query list task.
4. Tulis `internal/repository/subtask_repository.go` — CRUD subtask, terhubung ke `task_id`.
5. Putuskan dan implementasikan logic: apakah task otomatis `done` saat semua subtask selesai (di service layer, bukan trigger DB).
6. Uji query dengan `EXPLAIN ANALYZE` setelah data dummy cukup banyak — verifikasi index benar-benar terpakai, bukan berasumsi.

### Fase 4 — Middleware & Auth ✅ (selesai)
1. Tulis `internal/middleware/logging.go` — pola decorator (`func(http.Handler) http.Handler`), catat method, path, durasi tiap request.
2. Tulis `internal/middleware/auth.go` — validasi JWT dari header `Authorization`, inject identitas user ke `context.Context` request. Tambahkan implementasi CORS (`cors.go`).
3. Terapkan middleware chaining di `router.go` — bedakan route publik (login) vs route terproteksi.
4. Tambahkan graceful shutdown di `main.go` (`http.Server` + `Shutdown(ctx)` merespons sinyal `SIGTERM`/`SIGINT`) — request yang sedang berjalan diberi kesempatan selesai sebelum proses benar-benar berhenti.

### Fase 5 — Frontend (Astro & React) ✅ (selesai)
1. Setup project Astro, koneksi ke API Go lewat `fetch` (base URL dari environment variable, beda untuk dev vs production).
2. Halaman list task — grouping visual (Overdue / Hari ini / Akan datang / Tanpa deadline) dihitung di client dari satu response API, sesuai desain di Bagian 5.
3. Form tambah/edit task + subtask checklist, sebagai React island (bagian yang genuinely interactive).
4. Halaman/section per context (filter berdasarkan context dari UI, lengkap dengan manajemen warna).
5. **UI/UX Polish (Google Forms Style)**: Tampilan antarmuka sudah dirancang sangat responsif, merambat secara dinamis ke ukuran layar mobile/desktop. Menggunakan desain "kartu" putih, input teks elegan (*bottom border* tipis), serta pewarnaan yang menenangkan.
6. **Sistem Modal Validasi Global**: Setiap interaksi aksi (Simpan, Ubah, Hapus, Selesai, Unduh PDF, Logout) diikat dengan modal konfirmasi interaktif untuk mencegah *human error*.
7. **Halaman Autentikasi (Login)**: Halaman *login* responsif berpusat di tengah dengan manajemen *state* tombol *Loading*.

### Fase 6 — Deploy & Mitigasi Free-Tier (Tertunda / Dalam Proses)
1. Push schema ke Supabase: jalankan migration yang sama persis (`migrate -database "$SUPABASE_URL" -path migrations up`) ke Direct Connection Supabase.
2. Deploy backend ke Render (set environment variable dari dashboard, bukan `.env`).
3. Deploy frontend ke Cloudflare Pages/Vercel, arahkan base URL API ke domain Render.
4. Setup GitHub Actions heartbeat cron (ping backend + query kecil ke DB) sesuai Bagian 7 (versi lama, cek Bagian 8 sekarang).
5. Verifikasi end-to-end di production: buka frontend, coba create/list/update task, pastikan CORS antara domain frontend dan backend sudah benar.

### Fase 7 — Enhancement (opsional, setelah semua fase inti selesai)
1. CLI companion (`cobra`) — tambah task langsung dari terminal, memanggil API Go yang sama.
2. Aggregation report — task selesai per context per minggu/bulan, dari query `GROUP BY` di repository.
3. **Ekspor PDF per Context ✅ (selesai)** — implementasi library `go-pdf/fpdf` untuk meng-generate dokumen/laporan enterprise standard berdasarkan daftar task pada sebuah context (lengkap dengan checklist subtask, tanggal, dan format tabel tebal).
4. Evaluasi ulang struktur setelah endpoint bertambah banyak — refactor `router/` jika mulai berantakan (bukan ganti ke framework, sesuai keputusan Bagian 9).

---

## 8. Mitigasi Free-Tier (Cold Start & Auto-Pause)

- **Render** (backend): tanpa aktivitas 15 menit → sleep, cold start 30–60 detik pada request berikutnya.
- **Supabase** (database): tanpa API request 7 hari → project auto-pause, butuh resume manual dari dashboard.
- **Mitigasi**: satu GitHub Actions cron job yang ping endpoint backend secara berkala (tiap 3–4 hari), di mana endpoint tersebut juga melakukan query kecil ke Supabase — sehingga dua-duanya tetap "aktif" dengan satu mekanisme.

---

## 9. Keputusan Teknis yang Sudah Terkunci

- ❌ Tidak pakai HTTP framework (`gin`, `chi`, dll) — pakai `net/http` stdlib murni.
- ❌ Tidak pakai `sqlx` atau ORM — pakai `pgx` murni, scan manual.
- ✅ Migration schema lewat file `.sql` (`golang-migrate`), bukan edit manual via DBeaver.
- ✅ DBeaver tetap dipakai untuk inspeksi/debug data manual (connect via Direct Connection string Supabase) — terpisah dari cara aplikasi Go mengakses database.
- ✅ Frontend static-first (Astro), interactivity minimal via React island bila perlu.

---

## 10. Keputusan Final (update)

- ✅ **Sub-task/checklist**: dibutuhkan. Skema `subtasks` sudah ditambahkan di Bagian 5.
- ✅ **Router**: tetap `ServeMux` stdlib untuk seterusnya — performa matching-nya (trie-based, Go 1.22+) sudah setara framework seperti `chi`/`gin` meski endpoint bertambah banyak. Yang membedakan framework hanyalah ergonomi penulisan (route grouping, middleware per-group), yang bisa direplikasi sendiri lewat struktur package `router/` yang rapi (file per resource, digabung di satu entry point). Tidak perlu migrasi ke framework di kemudian hari.
- ❌ **Time tracking (start/stop timer)**: tidak diperlukan, dihapus dari scope.
