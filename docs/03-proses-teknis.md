# 03. Diagram Proses Teknis

Dokumen ini memetakan interaksi teknis antar-komponen (HTTP sinkron, Redis, RabbitMQ asinkron, dan integrasi pihak ketiga).

---

## 3.1 Autentikasi, Sanitasi Header & Proteksi Gateway

```mermaid
sequenceDiagram
    autonumber
    actor Client as "Vue 3 SPA / API Client"
    participant GW as "API Gateway (:8080)"
    participant RL as "Redis / In-Memory Limiter"
    participant US as "user-service (:8081)"
    participant PS as "product-service (:8082)"

    Client->>GW: "POST /api/v1/auth/login (email, password)"
    GW->>RL: "Cek Auth Rate Limit (maks 10 req/menit)"
    RL-->>GW: "Diizinkan"
    GW->>US: "Forward login request"
    US-->>GW: "200 OK (data user + role)"
    GW-->>Client: "200 OK + Signed JWT (HS256)"

    Client->>GW: "GET /api/v1/products (Bearer JWT + Header Spoofing)"
    GW->>GW: "Verifikasi JWT & ekstrak user_id, email, roles"
    GW->>RL: "Cek API Rate Limit (maks 100 req/menit)"
    GW->>GW: "Strip X-Internal-Request, X-Gateway, X-User-* dari caller"
    GW->>PS: "Proxy request + X-Gateway + X-User-ID/Email/Roles resmi"
    PS-->>GW: "200 OK (Daftar Produk)"
    GW-->>Client: "200 OK (Daftar Produk)"
```

---

## 3.2 User Onboarding Asinkron (RabbitMQ)

```mermaid
sequenceDiagram
    autonumber
    participant GW as "API Gateway"
    participant US as "user-service"
    participant UDB as "PostgreSQL User"
    participant RMQ as "RabbitMQ (email_queue)"
    participant NS as "notification-service"
    participant SMTP as "SMTP Server"

    GW->>US: "POST /api/v1/users (name, email, password)"
    US->>US: "Hash password dengan bcrypt"
    US->>UDB: "INSERT INTO users"
    UDB-->>US: "User tersimpan (id)"
    US-->>GW: "201 Created"
    par "Publish asinkron tanpa plaintext password"
        US->>RMQ: "Publish ke queue email_queue (type: welcome_email)"
    end
    RMQ-->>NS: "Consume pesan email_queue"
    NS->>SMTP: "Kirim email selamat datang"
```

---

## 3.3 Distribusi Stok Gudang ke Outlet Merchant

```mermaid
sequenceDiagram
    autonumber
    participant GW as "API Gateway"
    participant MS as "merchant-service"
    participant WS as "warehouse-service"
    participant MDB as "PostgreSQL Merchant"
    participant RMQ as "RabbitMQ (warehouse_events)"
    participant WDB as "PostgreSQL Warehouse"

    GW->>MS: "POST /api/v1/merchant-products (warehouse_id, merchant_id, product_id, stock)"
    MS->>WS: "GET stok gudang (via HTTP internal + token sistem)"
    WS-->>MS: "Stok gudang tersedia"
    MS->>MS: "Validasi warehouse_stock >= requested_stock"
    MS->>MDB: "INSERT INTO merchant_products"
    MDB-->>MS: "Tersimpan"
    MS->>RMQ: "Publish StockReductionEvent (exchange: warehouse_events, key: stock.reduction)"
    MS-->>GW: "201 Created"

    RMQ-->>WS: "Consume dari queue stock_reduction_queue"
    WS->>WDB: "Kurangi stok pada warehouse_products"
```

---

## 3.4 Checkout POS, Server-Side Pricing & Webhook Midtrans

```mermaid
sequenceDiagram
    autonumber
    participant UI as "Vue 3 POS (QuaggaJS + Snap)"
    participant GW as "API Gateway"
    participant TS as "transaction-service"
    participant MS as "merchant-service"
    participant PS as "product-service"
    participant MID as "Midtrans Snap API"
    participant TDB as "PostgreSQL Transaction"
    participant RMQ as "RabbitMQ (business_events)"

    UI->>GW: "POST /api/v1/transactions (customer, merchant_id, products[id, qty])"
    GW->>TS: "Forward request terautentikasi"
    TS->>MS: "Validasi stok tiap produk di merchant"
    MS-->>TS: "Stok mencukupi"
    TS->>PS: "Ambil harga otoritatif per product_id"
    PS-->>TS: "Harga master produk"
    TS->>TS: "Hitung SubTotal, PPN 11% (TaxTotal), GrandTotal"
    TS->>MID: "Request Snap Payment Token"
    MID-->>TS: "Snap Token + Redirect URL"
    TS->>TDB: "Simpan transactions (pending) + transaction_products"
    TS-->>GW: "201 Created (snap_token, order_id)"
    GW-->>UI: "Tampilkan popup Midtrans Snap"

    par "Sinkronisasi stok outlet asinkron"
        TS->>RMQ: "Publish StockReducedEvent (exchange: business_events, key: merchant.stock.reduced)"
        RMQ-->>MS: "Consume dari queue merchant_stock_events"
        MS->>MS: "Kurangi stok pada merchant_products"
    end

    MID->>GW: "POST /api/v1/midtrans/callback (public webhook)"
    GW->>TS: "Forward webhook payload"
    TS->>TS: "Verifikasi SHA512(order_id + status_code + gross_amount + server_key) constant-time"
    TS->>TDB: "Update payment_status, payment_method, fraud_status"
    TS-->>GW: "200 OK"
```

---

## 3.5 Topologi Messaging RabbitMQ

| Exchange | Tipe | Routing Key | Queue | Producer | Consumer | Tujuan |
|---|---|---|---|---|---|---|
| *(default)* | direct | `email_queue` | `email_queue` | `user-service` | `notification-service` | Pengiriman email selamat datang (tanpa plaintext password) |
| `warehouse_events` | topic | `stock.reduction` | `stock_reduction_queue` | `merchant-service` | `warehouse-service` | Mengurangi stok gudang saat stok dialokasikan ke merchant |
| `business_events` | topic | `merchant.stock.reduced` (binding `merchant.stock.*`) | `merchant_stock_events` | `transaction-service` | `merchant-service` | Mengurangi stok outlet merchant saat transaksi POS dibuat |

---

## 3.6 Peta Routing API Gateway (`:8080`)

| Prefix Path | Autentikasi | Rate Limit | Service Tujuan |
|---|---|---|---|
| `GET /health` | Publik | Global | Internal API Gateway |
| `GET /metrics` *(branch `feat/metrics-prometheus`)* | Publik | Global | Prometheus Metrics (API Gateway & tiap service) |
| `POST /api/v1/auth/login` | Publik | Auth (`10 req/min`) | `user-service` (`:8081`) |
| `POST /api/v1/midtrans/callback` | Publik (verifikasi SHA512 di service) | Global | `transaction-service` (`:8085`) |
| `/api/v1/users*`, `/api/v1/roles*`, `/api/v1/assign-role*`, `/api/v1/upload*` | JWT Protected | API (`100 req/min`) | `user-service` (`:8081`) |
| `/api/v1/products*`, `/api/v1/categories*`, `/api/v1/upload-product*` | JWT Protected | API (`100 req/min`) | `product-service` (`:8082`) |
| `/api/v1/warehouses*`, `/api/v1/warehouse-products*`, `/api/v1/upload-warehouse*` | JWT Protected | API (`100 req/min`) | `warehouse-service` (`:8083`) |
| `/api/v1/merchants*`, `/api/v1/merchant-products*`, `/api/v1/upload-merchant*` | JWT Protected | API (`100 req/min`) | `merchant-service` (`:8084`) |
| `/api/v1/transactions*`, `/api/v1/dashboard*` | JWT Protected | API (`100 req/min`) | `transaction-service` (`:8085`) |

---

## 3.7 Alur Distributed Tracing Lintas Service *(branch `feat/distributed-tracing`)*

```mermaid
sequenceDiagram
    autonumber
    actor Client as "Vue 3 SPA / API Client"
    participant GW as "API Gateway (:8080)"
    participant MS as "merchant-service (:8084)"
    participant WS as "warehouse-service (:8083)"
    participant JAEGER as "Jaeger OTLP (:4318 / UI :16686)"

    Client->>GW: "HTTP Request (+ traceparent opsional)"
    GW->>GW: "TracingMiddleware: Start Server Span & c.SetUserContext(ctx)"
    GW->>MS: "Proxy + W3C traceparent/tracestate"
    MS->>MS: "TracingMiddleware + controller teruskan c.UserContext()"
    MS->>WS: "NewTracedHTTPClient: Client Span + inject traceparent"
    WS-->>MS: "200 OK"
    MS-->>GW: "200 OK"
    GW-->>Client: "200 OK + Header X-Trace-ID"
    par "Batch Export OTLP HTTP (Asinkron)"
        GW-->>JAEGER: "Spans (api-gateway)"
        MS-->>JAEGER: "Spans (merchant-service)"
        WS-->>JAEGER: "Spans (warehouse-service)"
    end
```

---

## 3.8 Alur Scraping Metrics Prometheus *(branch `feat/metrics-prometheus`)*

```mermaid
flowchart LR
    REQ["HTTP Traffic"] -->|"Middleware"| OBS["RegisterMetrics (Fiber)<br/>http_requests_total &<br/>http_request_duration_seconds"]
    OBS -->|"/metrics (exclude self-metric)"| PROM["Prometheus (:9090)<br/>scrape_interval: 15s"]
    PROM -->|"Datasource (profile: grafana)"| GRAF["Grafana (:3001, Opsional)"]
```
