create table roles (
    id bigint generated always as identity primary key,
    role_name text not null unique,
    role_slug text not null unique,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now()
);

-- seed wajib: role_id di users butuh nilai ini buat RBAC jalan dari awal
-- slug dipakai sebagai key stabil di middleware RBAC (gak ikut berubah kalau role_name diganti)
insert into roles (role_name, role_slug) values ('Admin', 'admin'), ('Supplier', 'supplier'), ('User', 'user');
