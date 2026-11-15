# 04. ERD & Model Data

Sistem menerapkan pola **Database-per-Service** dengan **6 instance PostgreSQL terpisah**. Foreign key relasional SQL hanya berlaku di dalam database masing-masing service, sedangkan keterkaitan antar-service dihubungkan melalui **ID referensi logis** (`product_id`, `warehouse_id`, `merchant_id`, `keeper_id`) yang divalidasi via API internal maupun event RabbitMQ.

---

## 4.1 Diagram ERD per Bounded Context

### A. User Service (`warehouse_user_db` — Port `5432`)

```mermaid
erDiagram
    USERS {
        uint id PK
        string name
        string email UK
        string password
        string photo
        string phone
        timestamp created_at
        timestamp updated_at
    }

    ROLES {
        uint id PK
        string name
        timestamp created_at
        timestamp updated_at
    }

    USER_ROLE {
        uint id PK
        uint user_id FK
        uint role_id FK
        timestamp created_at
        timestamp updated_at
    }

    USERS ||--o{ USER_ROLE : "memiliki"
    ROLES ||--o{ USER_ROLE : "diberikan ke"
```

### B. Product Service (`warehouse_product_db` — Port `5433`)

```mermaid
erDiagram
    CATEGORIES {
        uint id PK
        varchar name
        varchar tagline UK
        text photo
        timestamp created_at
        timestamp updated_at
    }

    PRODUCTS {
        uint id PK
        varchar name
        varchar barcode UK
        uint category_id FK
        string thumbnail
        text about
        float price
        bool is_popular
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    CATEGORIES ||--o{ PRODUCTS : "mengelompokkan"
```

### C. Warehouse Service (`warehouse_warehouse_db` — Port `5437`)

```mermaid
erDiagram
    WAREHOUSES {
        uint id PK
        varchar name
        text address
        text photo
        varchar phone
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    WAREHOUSE_PRODUCTS {
        uint id PK
        uint warehouse_id FK
        uint product_id "Logical ref -> products.id"
        int stock
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    WAREHOUSES ||--o{ WAREHOUSE_PRODUCTS : "menyimpan"
```

### D. Merchant Service (`warehouse_merchant_db` — Port `5435`)

```mermaid
erDiagram
    MERCHANTS {
        uint id PK
        varchar name
        text address
        string photo
        string phone
        uint keeper_id "Logical ref -> users.id"
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    MERCHANT_PRODUCTS {
        uint id PK
        uint merchant_id FK
        uint product_id "Logical ref -> products.id"
        uint warehouse_id "Logical ref -> warehouses.id"
        int stock
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    MERCHANTS ||--o{ MERCHANT_PRODUCTS : "memiliki stok"
```

### E. Transaction Service (`warehouse_transaction_db` — Port `5434`)

```mermaid
erDiagram
    TRANSACTIONS {
        uint id PK
        varchar name
        varchar phone
        varchar email
        text address
        bigint sub_total
        bigint tax_total
        bigint grand_total
        bigint merchant_id "Logical ref -> merchants.id"
        varchar payment_status
        varchar payment_method
        varchar payment_code
        varchar order_id UK
        varchar transaction_code
        text payment_token
        text callback_url
        timestamp expired_at
        text notes
        varchar currency
        varchar fraud_status
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    TRANSACTION_PRODUCTS {
        uint id PK
        bigint product_id "Logical ref -> products.id"
        bigint quantity
        bigint price
        bigint sub_total
        bigint transaction_id FK
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    TRANSACTIONS ||--|{ TRANSACTION_PRODUCTS : "berisi item"
```

> **Catatan `notification-service` (`warehouse_notification_db` — Port `5436`)**: Database disiapkan pada infrastruktur Docker untuk isolasi service, sementara pemrosesan email saat ini berjalan secara *stateless* langsung dari antrean RabbitMQ (`email_queue`) ke SMTP.

---

## 4.2 Peta Relasi Logis Lintas Database

| Sumber Tabel & Kolom | Target Service & Tabel | Mekanisme Validasi / Sinkronisasi |
|---|---|---|
| `warehouse_products.product_id` | `product-service` → `products.id` | HTTP lookup saat pengayaan detail gudang & produk |
| `merchants.keeper_id` | `user-service` → `users.id` | Validasi kepemilikan outlet untuk role `Keeper` |
| `merchant_products.product_id` | `product-service` → `products.id` | HTTP lookup katalog & barcode produk |
| `merchant_products.warehouse_id` | `warehouse-service` → `warehouses.id` | Cek stok gudang asal via HTTP + kurangi stok via RabbitMQ `stock.reduction` |
| `transactions.merchant_id` | `merchant-service` → `merchants.id` | Validasi outlet & pengurangan stok outlet via RabbitMQ `merchant.stock.reduced` |
| `transaction_products.product_id` | `product-service` → `products.id` | Ambil harga otoritatif server-side + validasi stok ke `merchant-service` |

---

## 4.3 Indeks, Constraint & Soft Delete

* **Unique Index / Constraint**:
  * `users.email` (mencegah duplikasi akun pengguna)
  * `categories.tagline` (mencegah tagline kategori ganda)
  * `products.barcode` (menjamin 1 barcode hanya menunjuk 1 master produk untuk scan kamera POS)
  * `transactions.order_id` (idempotensi dan lookup webhook Midtrans)
* **Soft Delete (`deleted_at`)**:
  Diterapkan pada `products`, `warehouses`, `warehouse_products`, `merchants`, `merchant_products`, `transactions`, dan `transaction_products` agar histori transaksi serta audit stok tetap terjaga.
