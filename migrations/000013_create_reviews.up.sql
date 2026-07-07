create table reviews (
    id bigint generated always as identity primary key,
    user_id bigint not null references users (id) on delete cascade,
    product_id bigint not null references products (id) on delete cascade,
    rating integer not null check (rating between 1 and 5),
    comment text,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_reviews_product_id on reviews (product_id);

-- satu user cuma bisa review satu produk sekali
create unique index uq_reviews_user_product on reviews (user_id, product_id);