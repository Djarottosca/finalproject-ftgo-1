-- Dev Seed Data
-- Data contoh untuk DEVELOPMENT & DEMO. BUKAN migration, jangan dijalankan di
-- production. Aman dijalankan berkali-kali (pakai ON CONFLICT DO NOTHING).
--
-- Cara jalanin (setelah migrate up):
--   psql "postgres://USER:PASS@localhost:5432/proyekin_db" -f seeds/dev_seed.sql
--
-- Prasyarat: tabel roles sudah terisi dari migration 000001.
BEGIN;

-- 1. Supplier user (akar rantai foreign key)
INSERT INTO
    users (
        full_name,
        username,
        email,
        password_hash,
        role_id
    )
VALUES
    (
        'Toko Bangunan Jaya',
        'toko_bangunan',
        'toko@test.com',
        'hashed_dummy',
        (
            SELECT
                id
            FROM
                roles
            WHERE
                role_slug = 'supplier'
        )
    ) ON CONFLICT (username) DO NOTHING;

-- 2. Supplier (status approved biar produknya langsung tampil)
INSERT INTO
    suppliers (
        user_id,
        store_name,
        supplier_slug,
        address,
        status
    )
VALUES
    (
        (
            SELECT
                id
            FROM
                users
            WHERE
                username = 'toko_bangunan'
        ),
        'Toko Bangunan Jaya',
        'toko-bangunan-jaya',
        'Jl. Merdeka No. 1',
        'approved'
    ) ON CONFLICT (supplier_slug) DO NOTHING;

-- 3. Category
INSERT INTO
    categories (category_name, category_slug)
VALUES
    ('Material', 'material') ON CONFLICT (category_slug) DO NOTHING;

-- 4. Products. Stok sengaja bervariasi biar laporan stok admin keliatan hidup:
--    Cat Tembok = 0 (out of stock), Semen = 3 & Paku = 8 (low stock < 10),
--    Bata = 500 (aman).
INSERT INTO
    products (
        product_name,
        product_slug,
        category_id,
        supplier_id,
        unit,
        stock,
        price,
        description,
        status
    )
VALUES
    (
        'Semen Gresik 40kg',
        'semen-gresik-40kg',
        (
            SELECT
                id
            FROM
                categories
            WHERE
                category_slug = 'material'
        ),
        (
            SELECT
                id
            FROM
                suppliers
            WHERE
                supplier_slug = 'toko-bangunan-jaya'
        ),
        'sak',
        3,
        65000,
        'Semen abu-abu',
        'active'
    ),
    (
        'Bata Merah',
        'bata-merah',
        (
            SELECT
                id
            FROM
                categories
            WHERE
                category_slug = 'material'
        ),
        (
            SELECT
                id
            FROM
                suppliers
            WHERE
                supplier_slug = 'toko-bangunan-jaya'
        ),
        'pcs',
        500,
        800,
        'Bata merah press',
        'active'
    ),
    (
        'Cat Tembok 5kg',
        'cat-tembok-5kg',
        (
            SELECT
                id
            FROM
                categories
            WHERE
                category_slug = 'material'
        ),
        (
            SELECT
                id
            FROM
                suppliers
            WHERE
                supplier_slug = 'toko-bangunan-jaya'
        ),
        'kaleng',
        0,
        120000,
        'Cat putih',
        'active'
    ),
    (
        'Paku 5cm 1kg',
        'paku-5cm-1kg',
        (
            SELECT
                id
            FROM
                categories
            WHERE
                category_slug = 'material'
        ),
        (
            SELECT
                id
            FROM
                suppliers
            WHERE
                supplier_slug = 'toko-bangunan-jaya'
        ),
        'kg',
        8,
        18000,
        'Paku besi',
        'active'
    ) ON CONFLICT (product_slug) DO NOTHING;

COMMIT;