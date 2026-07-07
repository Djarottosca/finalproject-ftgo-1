create table product_images (
    id bigint generated always as identity primary key,
    product_id bigint not null references products (id) on delete cascade,
    image_url text not null,
    created_at timestamptz not null default now()
);

create index idx_product_images_product_id on product_images (product_id);
