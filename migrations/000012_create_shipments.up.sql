create table shipments (
    id bigint generated always as identity primary key,
    order_id bigint not null unique references orders (id) on delete cascade,
    shipping_id text,
    status text not null default 'pending',
    tracking_number text,
    courier text,
    actual_shipping_cost numeric(12, 2) check (actual_shipping_cost >= 0),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);