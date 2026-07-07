# Marketplace Bahan Bangunan

Semi-marketplace multi-supplier untuk industri bangunan. Tiga role — **Admin**, **Supplier**, **User** — dalam satu core service ber-RBAC, dengan Payment dan Notification sebagai service terpisah yang saling terhubung lewat gRPC.

Dokumentasi lengkap (arsitektur, alur bisnis, ERD): [PLANS.md](./PLANS.md)

## Tech Stack

| Layer | Tools |
|---|---|
| Backend framework | Golang + Echo |
| Config management | Go-Viper |
| Database | PostgreSQL |
| Cache | Redis |
| Background job | Go Asynq (broker: Redis) |
| Inter-service communication | gRPC |
| Payment | Xendit / simulasi |
| Shipping | RajaOngkir |
| Email notifikasi | Mailjet |

## Struktur Proyek

Monorepo, satu `go.mod` untuk semua service.

```
proto/                    # kontrak gRPC (payment, notification), shared semua service
migrations/                # semua migration SQL, satu sumber kebenaran
pkg/                        # util shared (logger, response, jwt)

core-service/               # monolith + RBAC: auth, product, cart, order, shipping
  cmd/server/                # entrypoint HTTP (Echo) + gRPC client
  cmd/worker/                # entrypoint Asynq worker
  internal/
    modules/                 # domain modules (auth, user, supplier, product, cart, order, shipping)
    bootstrap/                # dependency wiring
    config/                   # Viper config
    database/                 # koneksi Postgres + Redis
    cache/                     # Redis cache helpers (produk, session, rate limit)
    middleware/                # RBAC middleware
    grpcclient/                 # client ke payment & notification service
    task/                       # definisi + handler Asynq task

payment-service/            # stateless, integrasi Xendit / simulasi
  cmd/server/
  internal/
    config/
    provider/                 # adapter Xendit / simulasi
    grpcserver/                 # implementasi PaymentService

notification-service/       # stateless, integrasi Mailjet
  cmd/server/
  internal/
    config/
    provider/                 # adapter Mailjet
    grpcserver/                 # implementasi NotificationService
```

## Arsitektur Singkat

- **core-service** terima traffic HTTP dari frontend, simpan semua state (users, products, orders, payments, shipments, dst) di satu PostgreSQL.
- **payment-service** dan **notification-service** stateless — cuma jembatan gRPC ke pihak ketiga (Xendit/RajaOngkir lewat core, Mailjet), tidak punya database sendiri.
- Redis dipakai core-service untuk cache produk, session/token, rate limiting, dan sebagai broker Asynq (job async: kirim email, generate invoice, sync status ongkir).

Detail alur (registrasi supplier, checkout & pembayaran, status order) ada di [PLANS.md](./PLANS.md#5-alur--workflow-aplikasi).
