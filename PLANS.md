# Dokumentasi Final: Marketplace Bahan Bangunan

## Daftar Isi
1. [Ringkasan Proyek](#1-ringkasan-proyek)
2. [Tech Stack](#2-tech-stack)
3. [Arsitektur Sistem](#3-arsitektur-sistem)
4. [Peran & Fitur](#4-peran--fitur)
5. [Alur / Workflow Aplikasi](#5-alur--workflow-aplikasi)
6. [Struktur Database (ERD)](#6-struktur-database-erd)
7. [Detail Tabel Database](#7-detail-tabel-database)
8. [Hal yang Perlu Didiskusikan Tim](#8-hal-yang-perlu-didiskusikan-tim)

---

## 1. Ringkasan Proyek
- **Topik**: Industri bangunan (no. 9)
- **Model**: Semi-marketplace multi-supplier dengan 3 role: **Admin**, **Supplier**, **User**
- **Arsitektur**: Monolith (Core Service dengan RBAC) + service terpisah untuk Payment & Notification, komunikasi antar service pakai gRPC
- **Fitur user**: marketplace standar — cari produk, keranjang, checkout, bayar, lacak pengiriman, review

---

## 2. Tech Stack

| Layer | Tools |
|---|---|
| Backend framework | Golang + Echo |
| Config management | Go-Viper |
| Database | PostgreSQL |
| Cache | Redis |
| Background job / async worker | Go Asynq (broker: Redis) |
| Inter-service communication | gRPC |
| Payment | Xendit / simulasi |
| Shipping | RajaOngkir |
| Email notifikasi | Mailjet |

---

## 3. Arsitektur Sistem

**Core Service (Monolith + RBAC)** — 1 backend untuk 3 role (admin, supplier, user), dibedakan lewat role di tabel user:
- Auth & User Management
- Product & Supplier Management
- Cart & Order Management
- Shipping (integrasi RajaOngkir)

> **Aturan RBAC**: satu akun hanya boleh punya satu role (`role_id` langsung sebagai kolom di tabel `users`, bukan tabel relasi many-to-many). Login credentials (username, password, email) selalu tersimpan di satu tempat yaitu tabel `users`, sehingga tidak ada duplikasi identitas antara user, supplier, dan admin — mencegah konflik saat satu email/username dipakai untuk daftar ulang dengan role berbeda.

**Service Terpisah** (komunikasi ke Core via gRPC):
- **Payment Service** — integrasi Xendit / simulasi
- **Notification Service** — email via Mailjet

**Penggunaan Redis:**
- Cache listing/detail produk
- Session/token management
- Rate limiting
- Broker untuk background job via Go Asynq (mis. kirim email async, generate invoice, sync status ongkir)

```mermaid
flowchart TB
    A["Web / Mobile Frontend"] -->|REST/HTTP| Core

    subgraph Core["Core Service (Monolith + RBAC)"]
        Auth["Auth & User Mgmt"]
        Product["Product & Supplier Mgmt"]
        CartOrder["Cart & Order Mgmt"]
        Ship["Shipping Handler"]
    end

    subgraph PaymentSvc["Payment Service"]
        Pay["Payment Handler"]
    end

    subgraph NotifSvc["Notification Service"]
        Notif["Email Handler"]
    end

    DB[("PostgreSQL")]
    Redis[("Redis")]
    Asynq["Go Asynq Worker"]
    Xendit[["Xendit API"]]
    RajaOngkir[["RajaOngkir API"]]
    Mailjet[["Mailjet API"]]

    Core -->|gRPC| PaymentSvc
    Core -->|gRPC| NotifSvc
    Core --> DB
    Core --> Redis
    Redis --> Asynq
    Asynq -->|enqueue job: kirim email| NotifSvc
    Pay --> Xendit
    Ship --> RajaOngkir
    Notif --> Mailjet
```

---

## 4. Peran & Fitur

**Admin**
- Monitoring transaksi & aktivitas user/supplier
- Approve/reject pendaftaran supplier
- Pencatatan & laporan (penjualan, stok, dll)

**Supplier**
- Tambah, ubah, hapus produk
- Atur harga & diskon produk
- Atur stok barang
- Lihat & proses pesanan masuk

**User**
- Cari & lihat produk
- Keranjang & checkout
- Bayar via Xendit/simulasi
- Lacak status pengiriman
- Kasih review & rating produk

---

## 5. Alur / Workflow Aplikasi

### 5.1 Alur Registrasi & Approval Supplier
```mermaid
flowchart TD
    R["Supplier mendaftar"] --> Pending["status: pending"]
    Pending --> Review["Admin review pendaftaran"]
    Review -->|Approve| Approved["status: approved"]
    Review -->|Reject| Rejected["status: rejected"]
    Approved --> CanSell["Supplier bisa mulai jualan"]
```

### 5.2 Alur Checkout & Pembayaran
```mermaid
sequenceDiagram
    actor User
    participant Core as Core Service
    participant DB as PostgreSQL
    participant Pay as Payment Service
    participant Xendit
    participant Asynq as Asynq Worker
    participant Notif as Notification Service
    participant Mailjet

    User->>Core: Checkout item di keranjang
    Core->>Core: Hitung ongkir (RajaOngkir)
    Core->>DB: Simpan order (status: pending)
    Core->>Pay: gRPC - request payment link
    Pay->>Xendit: Create invoice
    Xendit-->>Pay: payment_link, payment_ref
    Pay-->>Core: payment_link, payment_ref
    Core->>DB: Simpan data payment
    Core-->>User: Redirect ke payment_link

    Xendit-->>Pay: Webhook - pembayaran berhasil
    Pay->>DB: Update payment status: paid
    Pay->>Core: gRPC - notify order paid
    Core->>DB: Update order status: paid
    Core->>Asynq: Enqueue job - kirim email konfirmasi
    Asynq->>Notif: gRPC - send email job
    Notif->>Mailjet: Kirim email konfirmasi
    Notif-->>User: Email diterima
```

### 5.3 Alur Status Order
```mermaid
flowchart LR
    A["pending"] -->|Pembayaran berhasil| B["paid"]
    B --> C["processing"]
    C --> D["shipped"]
    D --> E["completed"]
    A -->|Dibatalkan / timeout| F["cancelled"]
    B -->|Dibatalkan| F
```

### 5.4 Alur Supplier Kelola Produk
```mermaid
flowchart TD
    S["Supplier login"] --> P{"Pilih aksi"}
    P -->|Tambah produk| T1["Input data produk & gambar"]
    P -->|Ubah produk| T2["Edit harga / stok / diskon"]
    P -->|Lihat pesanan| T3["Lihat order masuk"]
    T1 --> Save[("Simpan ke DB")]
    T2 --> Save
    T3 --> Update["Update status order"]
```

---

## 6. Struktur Database (ERD)

```mermaid
erDiagram
    ROLES ||--o{ USERS : has
    USERS ||--o{ ADDRESSES : has
    USERS ||--o| SUPPLIERS : "can be"
    SUPPLIERS ||--o{ PRODUCTS : sells
    CATEGORIES ||--o{ PRODUCTS : classifies
    PRODUCTS ||--o{ PRODUCT_IMAGES : has
    USERS ||--o{ CARTS : has
    PRODUCTS ||--o{ CARTS : "added in"
    USERS ||--o{ ORDERS : places
    ORDERS ||--o{ ORDER_ITEMS : contains
    PRODUCTS ||--o{ ORDER_ITEMS : "ordered as"
    ORDERS ||--|| PAYMENTS : has
    ORDERS ||--|| SHIPMENTS : has
    USERS ||--o{ REVIEWS : writes
    PRODUCTS ||--o{ REVIEWS : receives

    USERS {
        int id PK
        string full_name
        string username
        string password_hash
        string email
        string status
        int role_id FK
        timestamp created_at
        timestamp updated_at
    }
    ROLES {
        int id PK
        string role_name
        string role_slug
        timestamp created_at
        timestamp updated_at
    }
    SUPPLIERS {
        int id PK
        int user_id FK
        string store_name
        string supplier_slug
        string address
        string status
        timestamp created_at
        timestamp updated_at
    }
    ADDRESSES {
        int id PK
        int user_id FK
        string label
        string full_address
        string city
        string district
        string postal_code
        bool is_primary
        timestamp created_at
        timestamp updated_at
    }
    CATEGORIES {
        int id PK
        string category_name
        string category_slug
        timestamp created_at
        timestamp updated_at
    }
    PRODUCTS {
        int id PK
        string product_name
        string product_slug
        int category_id FK
        string unit
        int stock
        int supplier_id FK
        decimal price
        string description
        string discount_type
        decimal discount_amount
        string status
        timestamp created_at
        timestamp updated_at
    }
    PRODUCT_IMAGES {
        int id PK
        int product_id FK
        string image_url
        timestamp created_at
        timestamp updated_at
    }
    CARTS {
        int user_id FK
        int product_id FK
        int qty
        timestamp created_at
        timestamp updated_at
    }
    ORDERS {
        int id PK
        int user_id FK
        decimal total_price
        decimal discount
        int total_items
        decimal final_price
        string status
        timestamp created_at
        timestamp updated_at
    }
    ORDER_ITEMS {
        int id PK
        int order_id FK
        int product_id FK
        decimal price
        int qty
        decimal subtotal
        timestamp created_at
        timestamp updated_at
    }
    PAYMENTS {
        int id PK
        int order_id FK
        decimal amount
        decimal tax
        string status
        string payment_link
        string payment_reference
        decimal shipping_cost_estimate
        timestamp created_at
        timestamp updated_at
    }
    SHIPMENTS {
        int id PK
        int order_id FK
        string shipping_id
        string status
        string tracking_number
        string courier
        decimal actual_shipping_cost
        timestamp created_at
        timestamp updated_at
    }
    REVIEWS {
        int id PK
        int user_id FK
        int product_id FK
        int rating
        string comment
        timestamp created_at
        timestamp updated_at
    }
```

---

## 7. Detail Tabel Database

1. **users** — id, full_name, username, password_hash, email, status, role_id (FK ke roles) → satu akun cuma bisa punya satu role, jadi login credentials (username/password/email) selalu ada di satu tempat, gak digandakan ke tabel lain
2. **roles** — id, role_name, role_slug → role_slug jadi key stabil di middleware RBAC, gak ikut berubah kalau role_name diganti jadi label display
3. **suppliers** — id, user_id (FK, karena supplier juga login lewat `users`), store_name, supplier_slug, address, status
4. **addresses** — id, user_id, label, full_address, city, district, postal_code, is_primary → dibutuhkan buat hitung ongkir RajaOngkir
5. **categories** — id, category_name, category_slug → lebih baik daripada free text field, biar konsisten & gampang di-filter
6. **products** — id, product_name, product_slug, category_id (FK), unit, stock, supplier_id, price, description, discount_type, discount_amount, status
7. **product_images** — id, product_id, image_url → biar bisa multi-foto per produk
8. **carts** — user_id, product_id, qty
9. **orders** — id, user_id, total_price, discount, total_items, final_price, status
10. **order_items** — id, order_id, product_id, price, qty, subtotal → price disalin dari products.price saat order dibuat, biar histori order gak berubah kalau supplier ganti harga produk belakangan
11. **payments** — id, order_id, amount, tax, status, payment_link, payment_reference, shipping_cost_estimate
12. **shipments** — id, order_id, shipping_id, status, tracking_number, courier, actual_shipping_cost
13. **reviews** — id, user_id, product_id, rating, comment

> Kolom `*_slug` (roles, suppliers, categories, products) unique dan dipakai untuk URL/identitas publik yang SEO-friendly. `product_slug` khususnya digenerate app layer dengan suffix (mis. id produk atau random short code), karena `product_name` gak dijamin unique antar supplier.

> Fitur reward/voucher untuk sementara di-drop dari scope (masih brainstorming). Kalau nanti mau dilanjutkan, rencananya jadi microservice terpisah — jadi nggak masuk skema database ini.

---

## 8. Hal yang Perlu Didiskusikan Tim
- Job apa saja yang perlu dijadikan async task via Asynq (mis. kirim email, generate invoice, sync status ongkir)
- Skema gRPC contract antara Core Service ↔ Payment Service ↔ Notification Service
- Kapan fitur voucher/reward mulai dikerjakan sebagai microservice terpisah
