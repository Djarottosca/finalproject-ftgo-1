-- gak ada id sendiri di ERD: satu baris per (user, product), nambah produk
-- yang sama di keranjang tinggal update qty lewat upsert on conflict
create table carts (
    user_id bigint not null references users (id) on delete cascade,
    product_id bigint not null references products (id) on delete cascade,
    qty integer not null check (qty > 0),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    primary key (user_id, product_id)
);