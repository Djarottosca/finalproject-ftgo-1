-- price disalin dari products.price saat order dibuat, biar histori order
-- gak berubah kalau supplier ganti harga produk belakangan
create table order_items (
    id bigint generated always as identity primary key,
    order_id bigint not null references orders (id) on delete cascade,
    product_id bigint not null references products (id),
    price numeric(12, 2) not null check (price >= 0),
    qty integer not null check (qty > 0),
    subtotal numeric(12, 2) not null check (subtotal >= 0),
    created_at timestamptz not null default now()
);

create index idx_order_items_order_id on order_items (order_id);
create index idx_order_items_product_id on order_items (product_id);
