create table users (
    id bigint generated always as identity primary key,
    full_name text not null,
    username text not null unique,
    password_hash text not null,
    email text not null unique,
    status text not null default 'active' check (status in ('active', 'inactive', 'banned')),
    role_id bigint not null references roles (id),
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

create index idx_users_role_id on users (role_id);