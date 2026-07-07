create table addresses (
    id bigint generated always as identity primary key,
    user_id bigint not null references users (id) on delete cascade,
    label text not null,
    full_address text not null,
    city text not null,
    district text not null,
    postal_code text not null,
    is_primary boolean not null default false,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_addresses_user_id on addresses (user_id);

-- constraint native: cegah lebih dari satu alamat utama per user tanpa app code
create unique index uq_addresses_one_primary_per_user
    on addresses (user_id)
    where is_primary;