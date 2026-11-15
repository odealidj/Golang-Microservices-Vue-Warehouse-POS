# Dokumentasi Sistem — FMCG Multi-Warehouse & Retail POS Platform

Folder ini berisi dokumentasi arsitektur dan teknis sistem yang disusun secara terstruktur agar mudah dipahami oleh developer, arsitek, QA, maupun stakeholder bisnis.

---

## Daftar Dokumen

| No | Dokumen | Fokus Pembahasan |
|---|---|---|
| 01 | [Gambaran Umum Sistem](01-gambaran-umum-sistem.md) | Domain bisnis, aktor, arsitektur microservices, katalog service/port/database, dan prinsip engineering |
| 02 | [Diagram Proses Bisnis](02-proses-bisnis.md) | Alur supply chain end-to-end (Gudang → Outlet → POS), matriks RBAC (`Manager` vs `Keeper`), dan siklus status pembayaran |
| 03 | [Diagram Proses Teknis](03-proses-teknis.md) | Sequence diagram autentikasi & sanitasi header, event-driven RabbitMQ, checkout POS + Midtrans webhook, serta peta routing gateway |
| 04 | [ERD & Model Data](04-erd-dan-data-model.md) | Entity Relationship Diagram (Database-per-Service) pada 6 database PostgreSQL, relasi logis lintas service, dan kamus data |
| - | [Postman Collection](postman_collection.json) | Koleksi request API siap impor untuk pengujian endpoint manual |

---

## Urutan Membaca yang Disarankan

1. **Baru mengenal project?** Mulai dari [01 - Gambaran Umum Sistem](01-gambaran-umum-sistem.md) lalu [02 - Diagram Proses Bisnis](02-proses-bisnis.md).
2. **Ingin integrasi API atau debugging alur transaksi/stok?** Lanjut ke [03 - Diagram Proses Teknis](03-proses-teknis.md) dan impor [postman_collection.json](postman_collection.json).
3. **Ingin memahami skema database atau menambah fitur baru?** Rujuk [04 - ERD & Model Data](04-erd-dan-data-model.md).
