-- status ikut alur 5.2: pending -> paid, plus failed/expired buat webhook Xendit yang gagal/kadaluarsa
create table payments (
    id bigint generated always as identity primary key,
    order_id bigint not null unique references orders (id) on delete cascade,
    amount numeric(12, 2) not null check (amount >= 0),
    tax numeric(12, 2) not null default 0 check (tax >= 0),
    status text not null default 'pending'
        check (status in ('pending', 'paid', 'failed', 'expired')),
    payment_link text,
    payment_reference text,
    shipping_cost_estimate numeric(12, 2) check (shipping_cost_estimate >= 0),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_payments_status on payments (status);