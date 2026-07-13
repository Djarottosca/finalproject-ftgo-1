# TODO — Fitur yang Belum Dibuat

Snapshot progres berdasarkan `PLANS.md` vs kode saat ini. Update manual, bukan otomatis.

## Sudah ada
- Struktur monorepo 3 service + migrations (13 tabel)
- `core-service`: config, postgres (gorm), redis, bootstrap, echo server/worker entrypoint
- `core-service/modules/auth`: login dengan static credentials (belum ke DB)
- `core-service`: auth middleware (JWT verify) + request logger middleware
- `payment-service`: provider interface (Xendit + simulasi), factory, webhook handler, gRPC server skeleton, proto `payment/v1` (generated)
- `notification-service`: config, folder grpcserver/provider (masih kosong)
- Makefile (run/build/lint/migrate/proto)

## Belum dibuat — Core Service

### Auth & User Management
- [ ] `modules/user`: CRUD user, register, repository (GORM) — auth masih hardcoded, belum baca tabel `users`
- [ ] Swap `auth.Service` dari `staticUsers` map ke `user.Repository` + bcrypt compare
- [ ] Role-based route guard (middleware cek `role` dari JWT claims, bukan cuma auth check)

### Supplier Management
- [ ] `modules/supplier`: registrasi supplier, approve/reject oleh admin (status pending → approved/rejected)
- [ ] Endpoint admin untuk list & review pendaftaran supplier

### Product Management
- [ ] `modules/product`: CRUD produk (create/update/delete oleh supplier)
- [ ] Kelola stok & diskon
- [ ] Kategori produk (relasi ke `categories`)
- [ ] Multi-image upload (`product_images`)
- [ ] Endpoint publik: search & detail produk (dengan cache Redis)

### Cart & Order Management
- [ ] `modules/cart`: add/update/remove item keranjang
- [ ] `modules/order`: checkout flow — hitung total, buat order + order_items
- [ ] Order status transition (pending → paid → processing → shipped → completed / cancelled)
- [ ] Endpoint admin: laporan penjualan & stok

### Shipping
- [ ] `modules/shipping`: integrasi RajaOngkir (hitung ongkir saat checkout)
- [ ] Sync status pengiriman → update `shipments` (tracking_number, courier, status)
- [ ] `modules/shipping` folder masih kosong (`.gitkeep` doang)

### Review
- [ ] Modul review belum ada folder sama sekali — user kasih rating & komentar produk

### Infra pendukung core
- [ ] `grpcclient`: client stub untuk manggil payment-service & notification-service dari core (folder masih kosong)
- [ ] `task` (Asynq): job async — kirim email konfirmasi, generate invoice, sync status ongkir (folder masih kosong)
- [ ] Redis cache dipakai nyata (listing produk, rate limiting) — `cache/redis.go` baru nyambung, belum dipakai di modul manapun

## Belum dibuat — Payment Service
- [ ] gRPC server method beneran (`grpcserver/server.go` ada tapi perlu dicek — kemungkinan skeleton doang, belum connect ke provider penuh)
- [ ] Callback ke core-service pas payment sukses (di `httpserver/server.go` ada `TODO(callback)` eksplisit — gRPC-back ke core belum diputuskan/dibuat)
- [ ] Simpan payment record ke DB (payment-service statless sekarang, gak ada koneksi DB — perlu didiskusikan apa payment nyimpen sendiri atau titip ke core)

## Belum dibuat — Notification Service
- [ ] Provider Mailjet (folder `provider/` kosong)
- [ ] gRPC server implementation (folder `grpcserver/` kosong)
- [ ] Proto `notification/v1` — file `.proto` belum ditulis, folder cuma `.gitkeep`
- [ ] Handler buat terima job dari Asynq worker core, lalu kirim email

## Belum diputuskan (dari PLANS.md §8)
- [ ] Job apa aja yang jadi async task via Asynq
- [ ] Skema gRPC contract lengkap Core ↔ Payment ↔ Notification (baru payment yang ada proto)
- [ ] Kapan voucher/reward digarap (out of scope untuk sekarang, rencana microservice terpisah)
