# 05. Observability & Strategi Pengujian

Dokumen ini menjelaskan strategi pengujian berlapis serta arsitektur **Observability** (Metrics, Distributed Tracing) dan **Load Testing** pada platform FMCG Multi-Warehouse & Retail POS.

> [!NOTE]
> Seluruh fitur **Metrics (Prometheus)**, **Distributed Tracing (OpenTelemetry + Jaeger)**, **Load Testing (k6)**, dan **Go Quality & CI** sudah terintegrasi penuh di `main`, sementara masing-masing branch fitur tetap dipertahankan utuh secara terisolasi agar mudah ditinjau per fitur.

---

## 5.1 Peta Branch Fitur

| Branch | Fokus Fitur | Artefak Utama |
|---|---|---|
| `main` | Core microservices, Vue 3 SPA, unit & integration test, Docker Compose, GitHub Actions CI | `backend/*`, `frontend/`, `backend/tests/integration/`, `docs/` |
| `feat/metrics-prometheus` | Metrics Prometheus + Grafana opsional | `backend/pkg/observability/metrics.go`, `backend/observability/{prometheus.yml,grafana-datasources.yml}` |
| `feat/distributed-tracing` | Distributed tracing OpenTelemetry + Jaeger | `backend/pkg/observability/tracing.go`, propagasi `c.UserContext()` & `NewTracedHTTPClient` |
| `feat/load-testing` | Load testing skenario API dengan k6 | `backend/tests/load/smoke_and_pos.js`, target `make load-test` |
| `feat/go-quality-and-testing` | Quality gate Go (`golangci-lint`, `govulncheck`, `-race`) & GoMock | `.golangci.yml`, `.github/workflows/ci.yml`, `Makefile`, `backend/*/mocks/` |

---

## 5.2 Piramida Pengujian (Testing Strategy)

| Lapisan Test | Perintah | Dependensi Eksternal | Cakupan Utama |
|---|---|---|---|
| **Go Unit Test (`-race`) & GoMock** | `make test-backend` / `make generate-mocks` | Tidak ada (`go.uber.org/mock/gomock`) | Auth/RBAC middleware, server-side pricing + PPN 11%, sanitasi payload event, pagination, lookup barcode (seluruh test usecase memakai mock `mockgen` + detector `-race`) |
| **Static Check, Lint & Security** | `make vet-backend`, `make lint-backend`, `make vulncheck-backend` | Tidak ada (`golangci-lint`), vuln DB (`govulncheck`) | `go vet`, `golangci-lint` (`.golangci.yml`: `errcheck`, `bodyclose`, `govet`, `ineffassign`, `staticcheck`, `unused`, `misspell`), serta pemindaian CVE dependensi dengan `govulncheck` |
| **Frontend Unit Test & Build** | `make test-frontend` / `make build-frontend` | Node.js `20+` | Pinia store transaksi (`node --test`) dan validasi bundle produksi Vite |
| **API Gateway Integration Test** | `make test-integration-gateway` | Tidak ada (`httptest.Server`) | `GET /health`, `GET /metrics`, verifikasi JWT, sanitasi header internal (`X-User-*`, `X-Gateway`, `X-Internal-Request`), forwarding webhook Midtrans, serta `Traceparent`/`X-Trace-ID` |
| **E2E Multi-Service Integration** | `make test-integration` | Docker stack (`make up`) | Alur Auth → Produk & Gudang → Distribusi ke Merchant (RabbitMQ) → Webhook Midtrans |
| **Load Testing (k6)** | `make load-test` | Docker stack + `k6` | Uji beban bertahap ke API Gateway & endpoint katalog/operasional |

Menjalankan verifikasi cepat lokal & CI (unit test Go `-race`, `go vet`, `golangci-lint`, unit test frontend, dan production build frontend):
```bash
make test
```

---

## 5.3 Metrics — Prometheus & Grafana Opsional (`main` / `feat/metrics-prometheus`)

Setiap microservice Go mengimpor modul lokal `micro-warehouse/observability` dan memanggil `observability.RegisterMetrics(app, "<service-name>")`.

### Metrik yang Diekspos (`GET /metrics`)

| Nama Metrik | Tipe | Labels | Deskripsi |
|---|---|---|---|
| `http_requests_total` | Counter | `service`, `method`, `path`, `status` | Total HTTP request yang diproses oleh service berdasarkan route Fiber dan status kode |
| `http_request_duration_seconds` | Histogram | `service`, `method`, `path` | Distribusi latensi request HTTP dalam detik (`prometheus.DefBuckets`) |

> Route `/metrics` sendiri dikecualikan dari pencatatan middleware agar tidak mengotori angka trafik API bisnis.

### Infrastruktur Monitoring
- **Prometheus (`:9090`)**: Menarik `/metrics` dari seluruh 7 target (`api-gateway:8080`, `user-service:8081`, `product-service:8082`, `warehouse-service:8083`, `merchant-service:8084`, `transaction-service:8085`, `notification-service:8086`) setiap `15s` sesuai `backend/observability/prometheus.yml`.
- **Grafana (`:3001`, opsional)**: Menggunakan Docker Compose profile `grafana` dengan datasource Prometheus terkonfigurasi otomatis melalui `backend/observability/grafana-datasources.yml`.

```bash
make setup-env && make build
make up                                            # Backend + Prometheus (:9090) + Jaeger (:16686)
cd backend && docker compose --profile grafana up -d  # Opsional: aktifkan Grafana (:3001, admin/admin)
```

---

## 5.4 Distributed Tracing — OpenTelemetry & Jaeger (`main` / `feat/distributed-tracing`)

Distributed tracing menggunakan **OpenTelemetry Go SDK (`v1.35.0`)** dengan protokol **OTLP HTTP** menuju **Jaeger All-in-One**.

### Komponen Utama (`backend/pkg/observability/tracing.go`)

1. **`InitTracing(ctx, serviceName)`**: Mengatur global W3C propagator (`TraceContext` + `Baggage`). Jika `OTEL_EXPORTER_OTLP_ENDPOINT` di-set (mis. `http://jaeger:4318`), menginisialisasi OTLP HTTP batch exporter; jika kosong (seperti saat unit test lokal), mengembalikan no-op shutdown tanpa error.
2. **`TracingMiddleware(serviceName)`**: Mengekstrak header W3C `traceparent`/`tracestate` dari request masuk, membuka server span (`SpanKindServer`), menyimpan span context ke `c.UserContext()` Fiber, dan mengembalikan header `X-Trace-ID` ke client.
3. **`NewTracedHTTPClient(timeout)` / `TracedRoundTripper`**: Membungkus HTTP client internal antar-service (`api-gateway` → `user-service`, serta 9 `httpclient` di `merchant-service`, `product-service`, `transaction-service`, `warehouse-service`) untuk membuat client span dan menyuntikkan header `traceparent` ke downstream.
4. **Propagasi Konteks Fiber v2**: Controller meneruskan `c.UserContext()` (bukan `c.Context()` FastHTTP) ke usecase dan HTTP client agar parent span tetap tersambung lintas service.

```bash
make setup-env && make build && make up
# Buka Jaeger UI: http://localhost:16686 (OTLP collector: :4317 gRPC, :4318 HTTP)
```

---

## 5.5 Load Testing — k6 (`main` / `feat/load-testing`)

Skrip load test `backend/tests/load/smoke_and_pos.js` menguji kesiapan gateway, login JWT pada fase `setup()`, serta beban paralel ke endpoint baca utama (`/health`, `/api/v1/products`, `/api/v1/categories`, `/api/v1/warehouses`, `/api/v1/merchants`).

### Konfigurasi Stages & Thresholds

- **Stages**:
  1. Ramp-up `10s` → `5` Virtual Users (VUs)
  2. Sustained load `20s` → `15` VUs
  3. Ramp-down `10s` → `0` VUs
- **Thresholds (SLA Gate)**:
  - `http_req_failed`: `rate < 0.05` (gagal HTTP < 5%)
  - `http_req_duration`: `p(95) < 500` (persentil ke-95 latensi < 500 ms)
  - `errors` (custom `Rate` metric): `rate < 0.05`

### Cara Menjalankan
```bash
make up        # Pastikan stack Docker aktif
make load-test # Menjalankan: k6 run backend/tests/load/smoke_and_pos.js
```
Variabel lingkungan opsional yang didukung skrip k6: `API_GATEWAY_URL` (default `http://localhost:8080`), `K6_EMAIL` (default `manager@mail.com`), dan `K6_PASSWORD` (default `manager123`).

---

## 5.6 Integrasi GitHub Actions CI (`main` / `feat/go-quality-and-testing`)

Workflow [`.github/workflows/ci.yml`](../.github/workflows/ci.yml) menjalankan pengecekan otomatis pada `push` (`main`, `master`, `feat/**`), `pull_request` (`main`, `master`), serta `workflow_dispatch`:

| Job | Target | Pengecekan yang Dijalankan |
|---|---|---|
| `backend` | Matrix 7 microservices Go + `pkg/observability` (`1.26.x`) | `go test -race -v ./...`, `go test -tags=integration -race -v ./...` (`api-gateway`), `go vet ./...`, `golangci-lint` (`v1.64.8`), dan `govulncheck ./...` |
| `gomock` | 5 service dengan interface mock | `go install go.uber.org/mock/mockgen@v0.6.0` & `make verify-mocks` (`git diff --exit-code -- 'backend/*/mocks'`) |
| `compose` | `backend/docker-compose.yml` | `make setup-env` & `docker compose config -q` |
| `frontend` | `frontend/` (Node `22`) | `npm ci`, `npm test`, dan `npm run build` |
