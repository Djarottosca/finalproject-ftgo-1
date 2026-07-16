alter table product_images add column updated_at timestamptz not null default now();
alter table order_items add column updated_at timestamptz not null default now();
