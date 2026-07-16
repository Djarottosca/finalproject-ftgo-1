-- Dev Seed Data
-- Data contoh untuk DEVELOPMENT & DEMO. BUKAN migration, jangan dijalankan di
-- production. Aman dijalankan berkali-kali (pakai ON CONFLICT DO NOTHING).
--
-- Cara jalanin (setelah migrate up):
--   psql "postgres://USER:PASS@localhost:5432/proyekin_db" -f seeds/dev_seed.sql
--
-- Prasyarat: tabel roles sudah terisi dari migration 000001.
-- password_hash semua akun pakai bcrypt hash yang sama, plaintext-nya "password123"
-- — bisa dipakai buat tes login lewat /user/login, /supplier/login, /admin/login.
BEGIN;

-- 1. Admin user (buat testing /admin/* routes)
INSERT INTO
    users (full_name, username, email, password_hash, role_id)
VALUES
    (
        'Admin Utama',
        'admin',
        'admin@test.com',
        '$2a$10$JpWJDUVKZog6h3YzfytcyeW3HUNpn84l5Hs5ypfIFqo8Zo7uI88Am',
        (SELECT id FROM roles WHERE role_slug = 'admin')
    ) ON CONFLICT (username) DO NOTHING;

-- 2. Regular users (buat testing cart/checkout/order/review dari sisi user)
INSERT INTO
    users (full_name, username, email, password_hash, role_id)
VALUES
    (
        'Budi Santoso',
        'budi',
        'budi@test.com',
        '$2a$10$JpWJDUVKZog6h3YzfytcyeW3HUNpn84l5Hs5ypfIFqo8Zo7uI88Am',
        (SELECT id FROM roles WHERE role_slug = 'user')
    ),
    (
        'Siti Aminah',
        'siti',
        'siti@test.com',
        '$2a$10$JpWJDUVKZog6h3YzfytcyeW3HUNpn84l5Hs5ypfIFqo8Zo7uI88Am',
        (SELECT id FROM roles WHERE role_slug = 'user')
    ) ON CONFLICT (username) DO NOTHING;

-- 3. Addresses buat kedua regular user
INSERT INTO
    addresses (
        user_id, label, full_address, city, district, postal_code, is_primary
    )
VALUES
    (
        (SELECT id FROM users WHERE username = 'budi'),
        'Rumah', 'Jl. Kenanga No. 10', 'Jakarta Selatan', 'Kebayoran Baru', '12110', true
    ),
    (
        (SELECT id FROM users WHERE username = 'siti'),
        'Rumah', 'Jl. Melati No. 5', 'Bandung', 'Coblong', '40132', true
    ) ON CONFLICT DO NOTHING;

-- 4. Supplier users (2 approved biar produknya tampil, 1 pending buat testing
--    alur /admin/suppliers/:id/review)
INSERT INTO
    users (full_name, username, email, password_hash, role_id)
VALUES
    (
        'Toko Bangunan Jaya', 'toko_bangunan', 'toko@test.com', '$2a$10$JpWJDUVKZog6h3YzfytcyeW3HUNpn84l5Hs5ypfIFqo8Zo7uI88Am',
        (SELECT id FROM roles WHERE role_slug = 'supplier')
    ),
    (
        'Elektronik Makmur', 'toko_elektronik', 'elektronik@test.com', '$2a$10$JpWJDUVKZog6h3YzfytcyeW3HUNpn84l5Hs5ypfIFqo8Zo7uI88Am',
        (SELECT id FROM roles WHERE role_slug = 'supplier')
    ),
    (
        'Toko Baru Daftar', 'toko_baru', 'tokobaru@test.com', '$2a$10$JpWJDUVKZog6h3YzfytcyeW3HUNpn84l5Hs5ypfIFqo8Zo7uI88Am',
        (SELECT id FROM roles WHERE role_slug = 'supplier')
    ) ON CONFLICT (username) DO NOTHING;

-- 5. Supplier profiles
INSERT INTO
    suppliers (user_id, store_name, supplier_slug, address, status)
VALUES
    (
        (SELECT id FROM users WHERE username = 'toko_bangunan'),
        'Toko Bangunan Jaya', 'toko-bangunan-jaya', 'Jl. Merdeka No. 1', 'approved'
    ),
    (
        (SELECT id FROM users WHERE username = 'toko_elektronik'),
        'Elektronik Makmur', 'elektronik-makmur', 'Jl. Sudirman No. 88', 'approved'
    ),
    (
        (SELECT id FROM users WHERE username = 'toko_baru'),
        'Toko Baru Daftar', 'toko-baru-daftar', 'Jl. Diponegoro No. 3', 'pending'
    ) ON CONFLICT (supplier_slug) DO NOTHING;

-- 6. Categories
INSERT INTO
    categories (category_name, category_slug)
VALUES
    ('Material', 'material'),
    ('Elektronik', 'elektronik'),
    ('Perkakas', 'perkakas') ON CONFLICT (category_slug) DO NOTHING;

-- 7. Products. Stok & status sengaja bervariasi biar laporan admin (stok
--    rendah/habis) dan diskon supplier keliatan hidup:
--    Cat Tembok = 0 (out of stock), Semen = 3 & Paku = 8 (low stock < 10),
--    Bata = 500 (aman). Kabel & Bor pakai discount_type/discount_amount.
INSERT INTO
    products (
        product_name, product_slug, category_id, supplier_id, unit, stock,
        price, description, status, discount_type, discount_amount
    )
VALUES
    (
        'Semen Gresik 40kg', 'semen-gresik-40kg',
        (SELECT id FROM categories WHERE category_slug = 'material'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'toko-bangunan-jaya'),
        'sak', 3, 65000, 'Semen abu-abu', 'active', NULL, NULL
    ),
    (
        'Bata Merah', 'bata-merah',
        (SELECT id FROM categories WHERE category_slug = 'material'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'toko-bangunan-jaya'),
        'pcs', 500, 800, 'Bata merah press', 'active', NULL, NULL
    ),
    (
        'Cat Tembok 5kg', 'cat-tembok-5kg',
        (SELECT id FROM categories WHERE category_slug = 'material'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'toko-bangunan-jaya'),
        'kaleng', 0, 120000, 'Cat putih', 'active', NULL, NULL
    ),
    (
        'Paku 5cm 1kg', 'paku-5cm-1kg',
        (SELECT id FROM categories WHERE category_slug = 'material'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'toko-bangunan-jaya'),
        'kg', 8, 18000, 'Paku besi', 'active', NULL, NULL
    ),
    (
        'Pasir Beton per m3', 'pasir-beton-per-m3',
        (SELECT id FROM categories WHERE category_slug = 'material'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'toko-bangunan-jaya'),
        'm3', 50, 250000, 'Pasir cor kualitas bagus', 'active', NULL, NULL
    ),
    (
        'Kabel Listrik NYA 1.5mm', 'kabel-listrik-nya-15mm',
        (SELECT id FROM categories WHERE category_slug = 'elektronik'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'elektronik-makmur'),
        'roll', 30, 350000, 'Kabel listrik SNI 1 roll = 50m', 'active', 'percentage', 10
    ),
    (
        'Lampu LED 12W', 'lampu-led-12w',
        (SELECT id FROM categories WHERE category_slug = 'elektronik'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'elektronik-makmur'),
        'pcs', 5, 35000, 'Lampu LED hemat energi', 'active', NULL, NULL
    ),
    (
        'Saklar Ganda', 'saklar-ganda',
        (SELECT id FROM categories WHERE category_slug = 'elektronik'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'elektronik-makmur'),
        'pcs', 60, 25000, 'Saklar lampu 2 mata', 'active', NULL, NULL
    ),
    (
        'Bor Listrik 500W', 'bor-listrik-500w',
        (SELECT id FROM categories WHERE category_slug = 'perkakas'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'elektronik-makmur'),
        'pcs', 12, 450000, 'Bor listrik serbaguna', 'active', 'fixed', 25000
    ),
    (
        'Obeng Set 6 in 1', 'obeng-set-6-in-1',
        (SELECT id FROM categories WHERE category_slug = 'perkakas'),
        (SELECT id FROM suppliers WHERE supplier_slug = 'elektronik-makmur'),
        'set', 40, 75000, 'Obeng set plus minus', 'active', NULL, NULL
    ) ON CONFLICT (product_slug) DO NOTHING;

-- 8. Product images (1-2 per produk, url dummy)
INSERT INTO
    product_images (product_id, image_url)
SELECT
    p.id, img.url
FROM
    products p
    JOIN (
        VALUES
            ('semen-gresik-40kg', 'https://picsum.photos/seed/semen/400'),
            ('bata-merah', 'https://picsum.photos/seed/bata/400'),
            ('cat-tembok-5kg', 'https://picsum.photos/seed/cat/400'),
            ('kabel-listrik-nya-15mm', 'https://picsum.photos/seed/kabel/400'),
            ('bor-listrik-500w', 'https://picsum.photos/seed/bor/400')
    ) AS img (slug, url) ON img.slug = p.product_slug
WHERE
    NOT EXISTS (
        SELECT 1 FROM product_images pi WHERE pi.product_id = p.id
    );

-- 9. Reviews (biar GET /products/:id/reviews ada isinya)
INSERT INTO
    reviews (user_id, product_id, rating, comment)
VALUES
    (
        (SELECT id FROM users WHERE username = 'budi'),
        (SELECT id FROM products WHERE product_slug = 'semen-gresik-40kg'),
        5, 'Kualitas bagus, cepat kering.'
    ),
    (
        (SELECT id FROM users WHERE username = 'siti'),
        (SELECT id FROM products WHERE product_slug = 'bor-listrik-500w'),
        4, 'Tenaganya lumayan buat pemakaian rumahan.'
    ) ON CONFLICT (user_id, product_id) DO NOTHING;

COMMIT;
