-- satu user maksimal satu baris supplier (relasi "can be" di ERD),
-- login credentials tetap cuma ada di users, gak digandakan ke sini
create table suppliers (
    id bigint generated always as identity primary key,
    user_id bigint not null unique references users (id) on delete cascade,
    store_name text not null,
    supplier_slug text not null unique,
    address text not null,
    status text not null default 'pending' check (status in ('pending', 'approved', 'rejected')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_suppliers_status on suppliers (status);