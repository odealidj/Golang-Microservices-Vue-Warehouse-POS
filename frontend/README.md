# Warehouse & POS Frontend — Vue.js 3

Frontend SPA untuk sistem distribusi multi-gudang dan kasir (POS) ritel FMCG / sembako kemasan berbasis **Vue 3**, **Pinia**, **Vue Router**, dan **Tailwind CSS**.

## Fitur Utama

- **Role-Based Access Control (RBAC)**:
  - **Manager**: Overview statistik global, manajemen Produk, Kategori, Gudang (*Warehouse*), Outlet (*Merchant*), User, dan Role.
  - **Keeper (Kasir/Kepala Outlet)**: Overview statistik merchant, manajemen stok outlet, riwayat transaksi, dan **POS Checkout Wizard**.
- **3-Step POS Transaction Wizard**:
  1. Input detail pelanggan & pemilihan outlet (`AddNewTransaction1.vue`).
  2. Scan barcode produk via kamera menggunakan **QuaggaJS** atau input manual + validasi stok real-time (`AddNewTransaction2.vue`).
  3. Review subtotal, **PPN 11%**, grand total, dan integrasi popup **Midtrans Snap** (`AddNewTransaction3.vue`).

## Menjalankan Secara Lokal

```bash
cp .env.example .env
npm install
npm run dev
```

Build produksi:

```bash
npm run build
```

## Konfigurasi Environment (`.env`)

- `VITE_API_BASE_URL` — URL API Gateway (default: `http://localhost:8080`)
- `VITE_MIDTRANS_CLIENT_KEY` — Client Key Midtrans Snap (Sandbox / Production)
- `VITE_MIDTRANS_IS_PRODUCTION` — `false` untuk Sandbox, `true` untuk Production
