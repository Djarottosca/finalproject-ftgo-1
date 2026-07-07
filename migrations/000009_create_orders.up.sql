-- status enum ikut alur 5.3 di PLANS.md: pending -> paid -> processing -> shipped -> completed,
-- dengan cancelled sebagai exit dari pending/paid
create table orders (
    id bigint generated always as identity primary key,
    user_id bigint not null references users (id),
    total_price numeric(12, 2) not null check (total_price >= 0),
    discount numeric(12, 2) not null default 0 check (discount >= 0),
    total_items integer not null check (total_items >= 0),
    final_price numeric(12, 2) not null check (final_price >= 0),
    status text not null default 'pending'
        check (status in ('pending', 'paid', 'processing', 'shipped', 'completed', 'cancelled')),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_orders_user_id on orders (user_id);
create index idx_orders_status on orders (status);