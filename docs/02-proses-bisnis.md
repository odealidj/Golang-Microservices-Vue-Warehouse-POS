# 02. Diagram Proses Bisnis

Dokumen ini menjelaskan alur operasional bisnis dari hulu (gudang pusat) hingga hilir (kasir outlet dan pembayaran pelanggan).

---

## 2.1 Alur Bisnis Supply Chain & Retail POS (End-to-End)

```mermaid
flowchart TD
    A["1. Onboarding Pengguna (Manager membuat akun Keeper dan Role)"] --> B["2. Setup Master Data (Kategori & Produk ber-Barcode)"]
    B --> C["3. Pengelolaan Gudang (Manager membuat Gudang & mengisi Stok Awal)"]
    C --> D["4. Distribusi Stok (Alokasi produk dari Gudang ke Outlet Merchant)"]
    D --> E{"Stok Gudang Cukup?"}
    E -- "Tidak" --> F["Proses ditolak, Manager perlu restock gudang"]
    E -- "Ya" --> G["Stok Outlet bertambah & Stok Gudang berkurang otomatis"]
    G --> H["5. Operasional POS oleh Keeper (Step 1: Data Pelanggan & Outlet)"]
    H --> I["Step 2: Scan Barcode QuaggaJS / Pilih Produk + Validasi Stok Outlet"]
    I --> J["Step 3: Review Subtotal + PPN 11% & Bayar via Midtrans Snap"]
    J --> K{"Status Pembayaran Midtrans"}
    K -- "Settlement / Capture" --> L["Transaksi Success & Omzet Outlet Tercatat"]
    K -- "Pending" --> M["Menunggu Pelanggan Menyelesaikan Pembayaran"]
    K -- "Deny / Cancel / Expired" --> N["Transaksi Gagal / Dibatalkan"]
```

---

## 2.2 Matriks Hak Akses & Kapabilitas (RBAC)

| Modul / Fitur | Manager (Back-Office) | Keeper (Front-Office / POS) |
|---|---|---|
| **Login & Profil Pribadi** | Ya | Ya |
| **Dashboard Global** | Omzet, total transaksi, produk terjual seluruh jaringan | Tidak (diarahkan ke dashboard outlet) |
| **Dashboard Outlet (`My Merchants`)** | Lihat daftar merchant | Lihat outlet milik sendiri & statistik omzet outlet |
| **Manajemen User & Role** | CRUD User, Role, dan Assign Role | Tidak |
| **Manajemen Kategori & Produk** | CRUD Kategori, Produk, Barcode, Upload Foto | Hanya membaca katalog produk outlet |
| **Manajemen Gudang & Stok Gudang** | CRUD Gudang dan tambah/ubah stok gudang | Tidak |
| **Distribusi Stok Gudang ke Merchant** | Alokasi dan penyesuaian stok merchant | Lihat stok outlet yang dipegang |
| **Transaksi Kasir (POS)** | Melihat seluruh riwayat transaksi | Membuat transaksi baru (3 langkah) & melihat transaksi outlet |

---

## 2.3 Aturan Bisnis Utama

1. **Pajak Pertambahan Nilai (PPN 11%)**:
   Setiap transaksi POS otomatis dikenakan pajak sebesar `11%` (`ppnRate = 0.11`) dari `SubTotal` yang dibulatkan ke bilangan bulat terdekat (`GrandTotal = SubTotal + TaxTotal`).
2. **Validasi Stok Berlapis**:
   * **Saat distribusi ke outlet**: jumlah stok yang dialokasikan tidak boleh melebihi stok tersedia pada gudang asal (`warehouse_products.stock`).
   * **Saat checkout POS**: kuantitas item keranjang divalidasi langsung terhadap stok outlet (`merchant_products.stock`) sebelum transaksi disimpan.
3. **Pemetaan Status Pembayaran Midtrans**:

```mermaid
stateDiagram-v2
    [*] --> pending: "Transaksi POS dibuat + Snap Token diterbitkan"
    pending --> success: "Webhook: settlement / capture"
    pending --> pending: "Webhook: pending"
    pending --> failed: "Webhook: deny / status tak dikenal"
    pending --> cancel: "Webhook: cancel"
    pending --> expired: "Webhook: expire"
    success --> [*]
    failed --> [*]
    cancel --> [*]
    expired --> [*]
```
