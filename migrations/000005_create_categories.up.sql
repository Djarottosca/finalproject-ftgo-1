create table categories (
    id bigint generated always as identity primary key,
    category_name text not null unique,
    category_slug text not null unique,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);