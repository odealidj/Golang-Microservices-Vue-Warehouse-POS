# FMCG Multi-Warehouse & Retail POS Platform

Sistem distribusi multi-gudang dan kasir (**Point of Sale**) untuk jaringan grosir & ritel FMCG / sembako kemasan, dibangun dengan **7 microservices Go (Fiber + GORM)**, **Database-per-Service (PostgreSQL)**, **Redis**, **RabbitMQ**, **Midtrans Snap**, dan **Vue.js 3 SPA**.

---

## Ringkasan Bisnis & Studi Kasus

Aplikasi ini memodelkan alur supply chain ritel modern dari **Gudang Pusat/Regional (`Warehouse`)** ke **Outlet/Toko Cabang (`Merchant`)** hingga **Transaksi Pelanggan (`POS`)**:

- **Back-Office (Manager)**: mengelola katalog master produk ber-barcode, kategori, stok di beberapa gudang, alokasi distribusi stok dari gudang ke outlet merchant, serta manajemen user & role (`Manager` / `Keeper`).
- **Front-Office / POS (Keeper)**: mengelola stok outlet yang dipegangnya, memproses checkout 3 langkah (data pelanggan → scan barcode kamera via **QuaggaJS** / pilih produk dengan validasi stok real-time → review subtotal + **PPN 11%** + pembayaran digital **Midtrans Snap**), serta memantau omzet outlet.

---

## Arsitektur Sistem

```mermaid
flowchart LR
    UI["Vue 3 + Pinia SPA<br/>(QuaggaJS Barcode + Midtrans Snap)"] -->|HTTPS / REST + JWT| GW["API Gateway (:8080)<br/>JWT Auth + Rate Limiter"]

    GW --> US["User Service (:8081)"]
    GW --> PS["Product Service (:8082)"]
    GW --> WS["Warehouse Service (:8083)"]
    GW --> MS["Merchant Service (:8084)"]
    GW --> TS["Transaction Service (:8085)"]
    GW --> NS["Notification Service (:8086)"]

    US --> DB1[(PostgreSQL User)]
    PS --> DB2[(PostgreSQL Product)]
    WS --> DB3[(PostgreSQL Warehouse)]
    MS --> DB4[(PostgreSQL Merchant)]
    TS --> DB5[(PostgreSQL Transaction)]
    NS --> DB6[(PostgreSQL Notification)]

    GW <--> REDIS[(Redis<br/>Rate Limit & Cache-Aside)]
    MS <--> REDIS

    US -->|welcome_email| RMQ[[RabbitMQ]]
    MS -->|warehouse_events<br/>stock.reduction| RMQ
    TS -->|business_events<br/>merchant.stock.reduced| RMQ

    RMQ --> NS
    RMQ --> WS
    RMQ --> MS

    TS <-->|Snap Token & Webhook| MID[Midtrans Payment Gateway]
    PS -->|Image Upload| SUP[Supabase Storage]
    MS -->|Image Upload| SUP
    WS -->|Image Upload| SUP
    US -->|Image Upload| SUP
```

### Daftar Microservices

| Service | Port | Tanggung Jawab Utama |
|---|---|---|
| `api-gateway` | `8080` | Reverse proxy, JWT verification, header sanitization, Redis/in-memory rate limiting |
| `user-service` | `8081` | Autentikasi, manajemen user, RBAC (`Manager`, `Keeper`), auto-seeder, publish email |
| `product-service` | `8082` | Master produk, barcode lookup, kategori, upload media ke Supabase Storage |
| `warehouse-service` | `8083` | Manajemen gudang & stok gudang, consumer `stock.reduction` |
| `merchant-service` | `8084` | Manajemen outlet merchant, stok outlet, cache-aside Redis, consumer `merchant.stock.reduced` |
| `transaction-service` | `8085` | Checkout POS, server-side pricing + PPN 11%, integrasi Midtrans Snap & verifikasi webhook |
| `notification-service` | `8086` | Consumer RabbitMQ untuk pengiriman email via SMTP |

---

## Alur Event-Driven (RabbitMQ)

1. **User Onboarding**: `user-service` membuat user baru (tanpa mengirim plaintext password ke broker) → publish ke queue email → `notification-service` mengirim email selamat datang via SMTP.
2. **Distribusi Stok Gudang → Outlet**: `merchant-service` mengalokasikan stok produk dari gudang ke merchant → publish event `stock.reduction` ke exchange `warehouse_events` → `warehouse-service` mengurangi stok gudang terkait.
3. **Checkout POS → Sinkronisasi Stok Outlet**: `transaction-service` memvalidasi stok ke `merchant-service`, menghitung harga otoritatif & PPN 11% dari `product-service`, menyimpan transaksi + membuat Snap token Midtrans → publish event `merchant.stock.reduced` ke exchange `business_events` → `merchant-service` mengurangi stok outlet secara asinkron.

---

## Engineering & Security Highlights

- **Zero Trust Gateway Boundary**: `api-gateway` selalu memvalidasi signed JWT untuk semua protected route (termasuk komunikasi internal antar-service melalui token sistem) dan men-strip header internal/identitas (`X-Internal-Request`, `X-Gateway`, `X-User-*`) dari caller eksternal sebelum meneruskan request.
- **Resilient Rate Limiting**: `api-gateway` menggunakan Redis rate limiter saat Redis tersedia dan otomatis melakukan *graceful fallback* ke in-memory rate limiter jika Redis tidak aktif.
- **Authoritative Server-Side Pricing**: `transaction-service` mengambil harga produk langsung dari `product-service` dan menghitung subtotal + PPN 11% di sisi server (tidak mempercayai nominal harga dari payload client).
- **Verified Payment Webhooks**: Callback Midtrans memverifikasi `signature_key` menggunakan `SHA512(order_id + status_code + gross_amount + server_key)` dengan perbandingan *constant-time*.
- **Clean Architecture**: Setiap service memisahkan layer `controller` → `usecase` → `repository` → `model` dengan dependency injection di `app/container.go`.
- **Observability & Load Testing (Modular Branch)**: Metrics Prometheus + Grafana opsional (`feat/metrics-prometheus`), Distributed Tracing OpenTelemetry + Jaeger (`feat/distributed-tracing`), dan skrip k6 load testing (`feat/load-testing`). Lihat [`docs/05-observability-dan-pengujian.md`](docs/05-observability-dan-pengujian.md).

---

## Branch Fitur

Repository ini memisahkan fitur observability dan load testing ke 3 branch independen dari `main`:

| Branch | Fitur | Komponen Utama |
|---|---|---|
| `main` | Core Platform | 7 microservices Go, Vue 3 SPA, unit test, integration test, Docker Compose, GitHub Actions CI, `docs/` |
| `feat/metrics-prometheus` | Metrics & Monitoring | `backend/pkg/observability/metrics.go` (`GET /metrics`), Prometheus (`:9090`), Grafana opsional (`:3001` via `--profile grafana`) |
| `feat/distributed-tracing` | Distributed Tracing | `backend/pkg/observability/tracing.go` (`InitTracing`, `TracingMiddleware`, `NewTracedHTTPClient`, `X-Trace-ID`), Jaeger UI (`:16686`) |
| `feat/load-testing` | Load Testing | Skrip k6 `backend/tests/load/smoke_and_pos.js` & target `make load-test` |

---

## Struktur Repository

```text
.
├── backend/
│   ├── api-gateway/
│   ├── user-service/
│   ├── product-service/
│   ├── warehouse-service/
│   ├── merchant-service/
│   ├── transaction-service/
│   ├── notification-service/
│   ├── tests/integration/      # E2E integration test lintas service (build tag: integration)
│   ├── pkg/observability/      # [Branch fitur] Shared metrics Prometheus & OpenTelemetry tracing
│   ├── observability/          # [feat/metrics-prometheus] Konfigurasi Prometheus & Grafana datasource
│   ├── tests/load/             # [feat/load-testing] Skrip load testing k6
│   └── docker-compose.yml
├── frontend/                   # Vue 3 + Vite + Pinia + Tailwind CSS + QuaggaJS
├── docs/
│   ├── README.md                           # Indeks dokumentasi arsitektur & teknis
│   ├── 01-gambaran-umum-sistem.md          # System overview, container architecture & katalog service
│   ├── 02-proses-bisnis.md                 # Diagram proses bisnis, RBAC & siklus status pembayaran
│   ├── 03-proses-teknis.md                 # Sequence diagram, topologi RabbitMQ & peta routing gateway
│   ├── 04-erd-dan-data-model.md            # ERD 6 database PostgreSQL & relasi logis lintas service
│   ├── 05-observability-dan-pengujian.md   # Panduan Metrics, Distributed Tracing & k6 Load Testing
│   └── postman_collection.json             # Koleksi endpoint API untuk pengujian
├── .github/workflows/ci.yml    # GitHub Actions CI (Go test/vet, frontend test & Vite build)
└── Makefile                    # Shortcut setup, Docker, testing, dan build
```

---

## Cara Menjalankan Project

### Prasyarat
- Docker & Docker Compose
- Go `1.24+` (opsional, untuk menjalankan test / development lokal)
- Node.js `20+` & npm (untuk frontend)

### 1. Siapkan Environment Variables
```bash
make setup-env
```
Sesuaikan kredensial opsional di file `.env` masing-masing service jika ingin mengaktifkan fitur eksternal:
- `MIDTRANS_SERVER_KEY`, `MIDTRANS_CLIENT_KEY` di `backend/transaction-service/.env` dan `VITE_MIDTRANS_CLIENT_KEY` di `frontend/.env`
- `SUPABASE_URL`, `SUPABASE_KEY`, `SUPABASE_BUCKET` untuk upload gambar
- `EMAIL_HOST`, `EMAIL_USER`, `EMAIL_PASSWORD` di `backend/notification-service/.env`

### 2. Jalankan Backend (Docker Compose)
```bash
make build
make up
make ps
```
API Gateway akan aktif di `http://localhost:8080` (`GET /health`) dan RabbitMQ Management UI di `http://localhost:15672` (`guest`/`guest`).
*(Pada branch observability terpisah tersedia juga Prometheus `:9090`, Grafana opsional `:3001`, dan Jaeger UI `:16686`.)*

Akun default hasil auto-seeder (`user-service`):
- **Email**: `manager@mail.com`
- **Password**: `manager123`

### 3. Jalankan Frontend (Vue 3)
```bash
cd frontend
npm install
npm run dev
```
Akses aplikasi melalui browser di `http://localhost:3000`.

---

## Pengujian & Verifikasi

Menjalankan seluruh unit test Go (`go test -race`), static check (`go vet`), linter (`golangci-lint`), unit test frontend (`node --test` untuk Pinia store transaksi), dan build produksi frontend:
```bash
make test
```
Atau per bagian (termasuk vulnerability scan `govulncheck` dan regenerasi mock **GoMock**):
```bash
make test-backend
make vet-backend
make lint-backend
make vulncheck-backend
make generate-mocks
make verify-mocks
make test-integration-gateway
make test-frontend
make build-frontend
```

### GitHub Actions CI ([`.github/workflows/ci.yml`](.github/workflows/ci.yml))
Pipeline otomatis berjalan pada `push` (`main`, `master`, `feat/**`), `pull_request` (`main`, `master`), dan `workflow_dispatch`, mencakup 4 job utama:
- **Backend Matrix (7 services)**: `go test -race -v ./...`, `go test -tags=integration -race -v ./...` (`api-gateway`), `go vet ./...`, `golangci-lint` (`.golangci.yml`), dan `govulncheck ./...`.
- **GoMock Sync Check**: menjalankan `make verify-mocks` untuk memastikan file di `backend/*/mocks/` sinkron dengan interface.
- **Docker Compose Config Check**: menjalankan `make setup-env` dan `docker compose config -q`.
- **Vue 3 Frontend**: `npm ci`, `npm test`, dan `npm run build`.
### Integration Test

Integration test dipisahkan menggunakan build tag `//go:build integration` agar `make test-backend` tetap cepat tanpa dependensi eksternal:
- **Gateway Integration Test (tanpa Docker)**: menguji `/health`, JWT auth, sanitasi header internal (`X-User-*`, `X-Gateway`, `X-Internal-Request`), dan forwarding webhook Midtrans ke `httptest.Server`:
  ```bash
  make test-integration-gateway
  ```
- **E2E Multi-Service Integration Test (membutuhkan Docker stack aktif via `make up`)**: menguji alur Auth → Produk & Gudang → Distribusi ke Merchant (RabbitMQ) → Webhook Midtrans:
  ```bash
  make test-integration
  ```
  Opsional: atur `API_GATEWAY_URL` (default `http://localhost:8080`) atau `MIDTRANS_SERVER_KEY` untuk ikut memverifikasi signature callback Midtrans.

### Observability & Load Testing (Per Branch)

Detail konfigurasi lengkap tersedia di [`docs/05-observability-dan-pengujian.md`](docs/05-observability-dan-pengujian.md).

- **Metrics Prometheus (`feat/metrics-prometheus`)**:
  ```bash
  git checkout feat/metrics-prometheus
  make up                                               # Prometheus aktif di http://localhost:9090
  cd backend && docker compose --profile grafana up -d  # Opsional: Grafana di http://localhost:3001 (admin/admin)
  ```
- **Distributed Tracing OpenTelemetry + Jaeger (`feat/distributed-tracing`)**:
  ```bash
  git checkout feat/distributed-tracing
  make up   # Jaeger UI aktif di http://localhost:16686 (OTLP HTTP :4318)
  ```
- **Load Testing k6 (`feat/load-testing`)**:
  ```bash
  git checkout feat/load-testing
  make up && make load-test
  ```
