-- product_name gak dijamin unique antar supplier (mis. dua toko sama-sama
-- jual "Semen Tiga Roda 40kg"), jadi slug digenerate app layer dengan suffix
-- (mis. id produk atau random short code) biar gak collision
create table products (
    id bigint generated always as identity primary key,
    product_name text not null,
    product_slug text not null unique,
    category_id bigint not null references categories (id),
    unit text not null,
    stock integer not null default 0 check (stock >= 0),
    supplier_id bigint not null references suppliers (id) on delete cascade,
    price numeric(12, 2) not null check (price >= 0),
    description text,
    discount_type text check (discount_type in ('percentage', 'fixed')),
    discount_amount numeric(12, 2) check (discount_amount >= 0),
    status text not null default 'active' check (status in ('active', 'inactive')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),
    -- diskon persen gak boleh lebih dari 100%; diskon fixed gak dibatasi di sini
    -- karena batasnya relatif ke price, biar app layer yang validasi
    constraint chk_products_discount_percentage
        check (discount_type <> 'percentage' or discount_amount <= 100)
);

create index idx_products_category_id on products (category_id);
create index idx_products_supplier_id on products (supplier_id);
create index idx_products_status on products (status);