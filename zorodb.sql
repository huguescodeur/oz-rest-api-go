-- =============================================================
-- O'Z — Schéma PostgreSQL
-- Converti depuis MySQL (phpMyAdmin 5.2.3 / MySQL 8.0.44)
-- =============================================================
BEGIN;
-- -------------------------------------------------------------
-- Types ENUM
-- -------------------------------------------------------------
CREATE TYPE user_role AS ENUM ('super', 'admin', 'vendeur');
CREATE TYPE stock_movement_type AS ENUM ('SALE', 'STOCK_IN', 'GIFT', 'LOSS', 'ADJUSTMENT');
-- -------------------------------------------------------------
-- Table : users
-- -------------------------------------------------------------
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    username VARCHAR(50) NOT NULL,
    firstname VARCHAR(100) NOT NULL,
    lastname VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL,
    phone VARCHAR(20),
    password_hash VARCHAR(255) NOT NULL,
    role user_role NOT NULL DEFAULT 'vendeur',
    profile_picture VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    parent_id INTEGER REFERENCES users(id) ON DELETE
    SET NULL,
        CONSTRAINT uq_users_uuid UNIQUE (uuid),
        CONSTRAINT uq_users_username UNIQUE (username),
        CONSTRAINT uq_users_email UNIQUE (email)
);
CREATE INDEX idx_users_parent ON users(parent_id);
-- -------------------------------------------------------------
-- Table : categories
-- -------------------------------------------------------------
CREATE TABLE categories (
    category_id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    category_name VARCHAR(100) NOT NULL,
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users(id),
    CONSTRAINT uq_categories_uuid UNIQUE (uuid),
    CONSTRAINT uq_category_name UNIQUE (category_name)
);
-- -------------------------------------------------------------
-- Table : shops
-- -------------------------------------------------------------
CREATE TABLE shops (
    shop_id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    shop_name VARCHAR(100) NOT NULL,
    shop_address VARCHAR(255) NOT NULL,
    shop_phone VARCHAR(20),
    shop_mail VARCHAR(150),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users(id),
    CONSTRAINT uq_shops_uuid UNIQUE (uuid)
);
CREATE INDEX idx_shops_user ON shops(user_id);
-- -------------------------------------------------------------
-- Table : products
-- -------------------------------------------------------------
CREATE TABLE products (
    product_id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    unit_price INTEGER NOT NULL,
    category_id INTEGER NOT NULL REFERENCES categories(category_id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    user_id INTEGER NOT NULL REFERENCES users(id),
    CONSTRAINT uq_products_uuid UNIQUE (uuid)
);
CREATE INDEX idx_products_user ON products(user_id);
CREATE INDEX idx_products_category ON products(category_id);
-- -------------------------------------------------------------
-- Table : orders
-- -------------------------------------------------------------
CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    uuid UUID NOT NULL,
    shop_id INTEGER REFERENCES shops(shop_id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    total_amount INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT uq_orders_uuid UNIQUE (uuid)
);
CREATE INDEX idx_orders_shop ON orders(shop_id);
CREATE INDEX idx_orders_user ON orders(user_id);
-- -------------------------------------------------------------
-- Table : order_items
-- -------------------------------------------------------------
CREATE TABLE order_items (
    item_id SERIAL PRIMARY KEY,
    order_id INTEGER REFERENCES orders(order_id),
    product_id INTEGER REFERENCES products(product_id),
    quantity INTEGER,
    unit_price INTEGER
);
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_order_items_product ON order_items(product_id);
-- -------------------------------------------------------------
-- Table : stocks  (clé primaire composite product_id + shop_id)
-- -------------------------------------------------------------
CREATE TABLE stocks (
    product_id INTEGER NOT NULL REFERENCES products(product_id) ON DELETE CASCADE,
    shop_id INTEGER NOT NULL REFERENCES shops(shop_id) ON DELETE CASCADE,
    quantity INTEGER DEFAULT 0,
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (product_id, shop_id)
);
CREATE INDEX idx_stocks_shop ON stocks(shop_id);
-- -------------------------------------------------------------
-- Table : stock_movements
-- -------------------------------------------------------------
CREATE TABLE stock_movements (
    movement_id SERIAL PRIMARY KEY,
    product_id INTEGER NOT NULL REFERENCES products(product_id),
    shop_id INTEGER NOT NULL REFERENCES shops(shop_id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    quantity_change INTEGER NOT NULL,
    movement_type stock_movement_type NOT NULL,
    comment VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_movements_product ON stock_movements(product_id);
CREATE INDEX idx_movements_shop ON stock_movements(shop_id);
CREATE INDEX idx_movements_user ON stock_movements(user_id);
-- -------------------------------------------------------------
-- Table : user_shops
-- -------------------------------------------------------------
CREATE TABLE user_shops (
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shop_id INTEGER NOT NULL REFERENCES shops(shop_id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, shop_id)
);
CREATE INDEX idx_user_shops_shop ON user_shops(shop_id);
-- =============================================================
-- Données de test
-- =============================================================
-- users
INSERT INTO users (
        id,
        uuid,
        username,
        firstname,
        lastname,
        email,
        phone,
        password_hash,
        role,
        created_at,
        updated_at
    )
VALUES (
        1,
        '29e2afe5-50b9-46c0-a6ec-b30b6a36b5cc',
        'hgsdev',
        'Hgs',
        'Dev',
        'hgsdev@example.com',
        '0503020385',
        '$2a$10$Q0qY2f7Axq0mX0nMPETurOV176GySWP5NYyMPn50JRyVU/FJjOl3K',
        'admin',
        '2026-04-01 04:35:51+00',
        '2026-04-01 05:57:35+00'
    ),
    (
        2,
        '4fce273d-b5c0-40cf-ad32-d38ffe5bdaf5',
        'zoro_dev',
        'Roronoa',
        'Zoro',
        'zoro@example.com',
        '0123456789',
        '$2a$10$sYAUcLuqOv2UxeAdn5KFTemOe7lKJz4r1ESlj8dLtaYxzVB6jQt2G',
        'vendeur',
        '2026-04-01 06:01:24+00',
        '2026-04-01 06:01:24+00'
    ),
    (
        4,
        'e072fe98-824b-49cf-be80-1c0b32c0d9e1',
        'gnabacharles',
        'Charles',
        'Gnaba',
        'gc@example.com',
        '0154632454',
        '$2a$10$8vVeyo0fZ5eFAbr8cwUaBOUI57CmNf7pk8q..1BQi.3LJ6xlw8/Pm',
        'vendeur',
        '2026-04-01 07:06:17+00',
        '2026-04-03 13:23:54+00'
    ),
    (
        5,
        '9028bcae-2ce0-4289-a095-686ab41c1b9f',
        'kpriscille',
        'Priscille',
        'Kouamé',
        'pk@example.com',
        '0700632454',
        '$2a$10$sM9Hou9KiqNMKwYKjULGDOvLBA5R/VgjHEVK7Qqb6a9SxhPKAEc4q',
        'vendeur',
        '2026-04-01 07:14:31+00',
        '2026-04-01 07:14:31+00'
    ),
    (
        6,
        'a9780f12-a77b-4715-8ea1-49e81d57e82f',
        'yannsela',
        'Yann',
        'Sela',
        'ys@example.com',
        '0554632400',
        '$2a$10$SIDRZXNbfdezzuv.28BMketD4HCxRxCD6vP34zzqe4/PGWl0BWkhS',
        'vendeur',
        '2026-04-01 07:30:10+00',
        '2026-04-01 07:30:10+00'
    ),
    (
        7,
        '317e8092-e67f-4e69-bb2c-fa8bca3e6105',
        'admin2',
        'Two',
        'Admin',
        'at@example.com',
        '0554632454',
        '$2a$10$rOewy/APKX16PMbon0IupuH09EcYt.tUS2h3NhLN8P8aciWzGjUYq',
        'admin',
        '2026-04-01 21:27:52+00',
        '2026-04-03 13:32:41+00'
    ),
    (
        8,
        '8f1053fd-0445-46d5-9385-7e3446ccd080',
        'kenpat',
        'Pat',
        'Ken',
        'kp@example.com',
        '0554630051',
        '$2a$10$nzXRRoYFMdn1xg5RMVhdpOp7Slhh/WWuHkGzvJLuXpKjts0KP9o0y',
        'vendeur',
        '2026-04-01 21:30:11+00',
        '2026-04-01 21:30:11+00'
    ),
    (
        9,
        '73467ba3-e6b0-4bc3-803d-f1fc43c9433d',
        'codeclaude',
        'Code',
        'Claude',
        'c@example.com',
        '0554630051',
        '$2a$10$Qq1lRjqKnyRLmXyNKKhBD.YYiHCWc9ots/rWJPYHEHa5FWi1iMV3a',
        'admin',
        '2026-04-01 22:45:18+00',
        '2026-04-03 12:07:12+00'
    ),
    (
        10,
        '1af6a38b-d635-46ec-8a86-bda5ae51727a',
        'angekonan',
        'Ange',
        'Konan',
        'kn@example.com',
        '0554630051',
        '$2a$10$xfMshxj3tkX4TH4yBRlWe.E8T4aA6mOGPSiw5A5Tc4G42NIzgtR6q',
        'vendeur',
        '2026-04-01 22:46:21+00',
        '2026-04-01 22:46:21+00'
    ),
    (
        13,
        '12eaccb1-5f09-41d8-b862-fed6e83272ff',
        'kaderkonan',
        'Kader',
        'Konan',
        'kk@example.com',
        '0554630051',
        '$2a$10$Cc5fvlBNirvbFv6Z3EOsPOBB7vHIK69g356HwZUZnXu/gbvfRiB2y',
        'admin',
        '2026-04-01 22:51:55+00',
        '2026-04-01 22:51:55+00'
    ),
    (
        15,
        'c632d78d-eff6-4738-a816-fee894ffa07b',
        'rpkonan',
        'Prince',
        'Konan',
        'prk@example.com',
        '0554630051',
        '$2a$10$DsW3./XQVlz7Ijf0YMM6OuMHEkrFpF2wCpqRrhFcxGr9n.0f77fni',
        'vendeur',
        '2026-04-01 22:53:01+00',
        '2026-04-02 03:23:19+00'
    ),
    (
        16,
        'eb3c9af3-f4ef-43f5-ac97-dfae690c9e19',
        'ismo',
        'Ismo',
        'Kader',
        'is@example.com',
        '0554630051',
        '$2a$10$AfPrpixCKJJsLgxBGSoLN.QG7ZBcx2ij9eXVqiwLdM9wq8uu17joq',
        'admin',
        '2026-04-02 04:52:33+00',
        '2026-04-02 04:58:04+00'
    ),
    (
        17,
        '71fcca14-a1bc-4b6b-8e9c-7a6769279498',
        'fabriceirie',
        'Fabrice',
        'Irie',
        'if@example.com',
        '0554630051',
        '$2a$10$FdENx5HAKXtoZ8O5whGQuO4rdN4FDTdeqAvcBHtVHgISrWOWthURe',
        'admin',
        '2026-04-02 05:06:15+00',
        '2026-04-02 05:06:15+00'
    ),
    (
        18,
        '98336598-4f3e-46c7-a440-2b7e137fafc2',
        'huguesd7',
        'Hugues',
        'Satchi',
        'sj@example.com',
        '0554630051',
        '$2a$10$wzBhVHPx7bZKiyzgJO2PTu1Ijiz9OArdc8zlXsB0/V.m50InBB7ei',
        'admin',
        '2026-04-02 05:20:14+00',
        '2026-04-02 05:54:37+00'
    ),
    (
        19,
        '0b29d03a-71b5-4a06-a1f0-cac408a85a4f',
        'huguescodeur',
        'Hugues',
        'Codeur',
        'hg@example.com',
        '0503020385',
        '$2a$10$ZXBFE.kU3wsdYymnfksJaeUo/rLffd8tx3aoMpnMb0k.1Wvv9eWea',
        'super',
        '2026-04-02 05:57:18+00',
        '2026-04-02 05:57:18+00'
    ),
    (
        20,
        'f7ae417f-588c-4a7a-b78a-5b3b5870038e',
        'dar_ryl',
        'Daryl',
        'DEGAN',
        'degandaryl@gmail.com',
        '0169349151',
        '$2a$10$nBE4zM/MK6ph5e/TzsaPAu83NGEFR2Dcw4evCaM.rRjZGzRhvEeS.',
        'admin',
        '2026-04-02 13:04:40+00',
        '2026-04-02 13:04:40+00'
    ),
    (
        21,
        '08c3debd-5e1d-444d-89a2-8f4c8bd257fc',
        'kaderkouadio',
        'Kader',
        'Kouadio',
        'kk@gmail.com',
        '0503020385',
        '$2a$10$YFClCztdX1mAJbm5uviQGuGdn4jSvDbf/9dtf/FXYMo49gfdmCOyW',
        'vendeur',
        '2026-04-03 08:58:15+00',
        '2026-04-03 08:58:15+00'
    ),
    (
        22,
        '5f1a7087-8854-401d-82ae-84c0a4698560',
        'gnabacharles09',
        'c&',
        'g&',
        'gc&@example.com',
        '0154632454',
        '$2a$10$QLtEIw1w6Kfee9Fmt4BaQOviyuOJExCJmQUCnooRJiGvLV1mPOfFm',
        'vendeur',
        '2026-04-03 23:22:22+00',
        '2026-04-03 23:22:22+00'
    ),
    (
        23,
        '177b9e47-df9e-4df3-990e-ae7b53346ce6',
        'gc12',
        'c&',
        'g&',
        'gc12@gmail.com',
        '0754632454',
        '$2a$10$6anythJu9FoV2BLIUuk/lOoS4OXe4fcAeiaoCcWcPCBIUha1tFAFi',
        'vendeur',
        '2026-04-03 23:31:49+00',
        '2026-04-03 23:31:49+00'
    ),
    (
        24,
        'be65684a-c44d-44d9-8168-42604b55324a',
        'baz',
        'b',
        'a',
        'baz@gmail.com',
        '0554630051',
        '$2a$10$1lqF.DN83Ct/8Bnio55Rbene2h3/WVCcRHwhyft5KCw2VE4uslTcO',
        'vendeur',
        '2026-04-03 23:39:30+00',
        '2026-04-03 23:39:30+00'
    ),
    (
        26,
        '41eb5d51-4332-4826-a54f-30adea1cb233',
        'adamo2',
        'ad',
        'mo',
        'adm@gmail.com',
        '0154632454',
        '$2a$10$Y9OgSHTqP7727TPxC8XAYuO6FcHjvCVXv6rAhJxdrc0PQNQl/kufi',
        'admin',
        '2026-04-04 00:59:46+00',
        '2026-04-04 00:59:46+00'
    );
-- Mettre à jour parent_id (auto-ref, doit être fait après les inserts)
UPDATE users
SET parent_id = 1
WHERE id IN (2, 4, 5, 6);
UPDATE users
SET parent_id = 7
WHERE id = 8;
UPDATE users
SET parent_id = 9
WHERE id = 10;
UPDATE users
SET parent_id = 13
WHERE id = 15;
-- categories
INSERT INTO categories (category_id, uuid, category_name, user_id)
VALUES (
        1,
        '263a5cd0-2d2f-11f1-a267-4fa3467576ef',
        'Informatique',
        2
    ),
    (
        2,
        '263a5e74-2d2f-11f1-a267-4fa3467576ef',
        'Bureautique',
        1
    ),
    (
        3,
        '263a5ec4-2d2f-11f1-a267-4fa3467576ef',
        'Accessoires',
        2
    );
-- shops
INSERT INTO shops (
        shop_id,
        uuid,
        shop_name,
        shop_address,
        shop_phone,
        shop_mail,
        created_at,
        updated_at,
        user_id
    )
VALUES (
        1,
        '895fe6aa-2d21-11f1-a267-4fa3467576ef',
        'Chic Shop Yop',
        'Yopougon Rond Point Gandhi',
        '0598675460',
        'chicshop@gmail.com',
        '2026-03-30 05:45:09+00',
        '2026-03-31 23:56:46+00',
        2
    ),
    (
        3,
        '217fe77c-2d21-11f1-a267-4fa3467576ef',
        'Chic Shop Yop Terminus 40',
        'Yopougon Terminus 40',
        '0510675460',
        'chicshop@gmail.com',
        '2026-03-30 12:11:21+00',
        '2026-03-31 16:54:42+00',
        1
    ),
    (
        4,
        '3c5c67df-a9a3-4832-bac9-d6b4accfd812',
        'NEXIUM.AI Store',
        'Yopougon Camp-Militaire',
        '6028675460',
        'ai.nexium@gmail.com',
        '2026-03-31 23:59:26+00',
        '2026-03-31 23:59:26+00',
        2
    );
-- products  (les 2 lignes avec user_id=0 d'origine ont été corrigées → user_id=1)
INSERT INTO products (
        product_id,
        uuid,
        product_name,
        unit_price,
        category_id,
        created_at,
        updated_at,
        deleted_at,
        user_id
    )
VALUES (
        1,
        'c0d25081-45b1-42e2-a95b-46c82cfe8112',
        'Clavier Mécanique RGB',
        5700,
        1,
        '2026-03-29 06:37:13+00',
        '2026-03-31 19:10:26+00',
        '2026-03-31 19:10:26+00',
        2
    ),
    (
        2,
        '263908da-2d2f-11f1-a267-4fa3467576ef',
        'Souris Sans Fil',
        4500,
        3,
        '2026-03-29 06:37:13+00',
        '2026-03-31 19:31:41+00',
        NULL,
        2
    ),
    (
        3,
        '2639093e-2d2f-11f1-a267-4fa3467576ef',
        'Bureau Assis-Debout',
        31000,
        2,
        '2026-03-29 06:37:13+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        4,
        '2639097a-2d2f-11f1-a267-4fa3467576ef',
        'Écran 27 pouces 4K',
        29900,
        1,
        '2026-03-29 06:37:13+00',
        '2026-03-31 18:26:35+00',
        NULL,
        2
    ),
    (
        5,
        '26390a06-2d2f-11f1-a267-4fa3467576ef',
        'Tapis de souris XL',
        1500,
        3,
        '2026-03-29 06:37:13+00',
        '2026-03-31 18:26:35+00',
        NULL,
        2
    ),
    (
        9,
        '26390a42-2d2f-11f1-a267-4fa3467576ef',
        'Clavier Lumineux',
        1800,
        1,
        '2026-03-29 17:52:30+00',
        '2026-03-31 18:26:35+00',
        NULL,
        2
    ),
    (
        11,
        '26390a74-2d2f-11f1-a267-4fa3467576ef',
        'Clavier Lumineux',
        2200,
        3,
        '2026-03-30 12:35:31+00',
        '2026-03-31 18:26:35+00',
        NULL,
        2
    ),
    (
        12,
        '26390a9c-2d2f-11f1-a267-4fa3467576ef',
        'Clavier Lumineux',
        2800,
        3,
        '2026-03-30 12:41:38+00',
        '2026-03-31 18:26:35+00',
        NULL,
        2
    ),
    (
        13,
        '26390ace-2d2f-11f1-a267-4fa3467576ef',
        'Chargeur PC HP',
        3200,
        2,
        '2026-03-31 01:53:03+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        14,
        '26390af6-2d2f-11f1-a267-4fa3467576ef',
        'Chargeur PC DELL',
        15500,
        2,
        '2026-03-31 02:03:59+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        15,
        '26390b1e-2d2f-11f1-a267-4fa3467576ef',
        'Chargeur PC LENOVO',
        15500,
        2,
        '2026-03-31 02:08:49+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        16,
        '26390b46-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer R',
        35500,
        2,
        '2026-03-31 02:12:33+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        17,
        '26390b6e-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer R',
        37500,
        2,
        '2026-03-31 02:17:07+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        18,
        '26390b96-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer X',
        38600,
        2,
        '2026-03-31 02:19:54+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        19,
        '26390bc8-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer X',
        38600,
        3,
        '2026-03-31 02:49:05+00',
        '2026-03-31 18:26:35+00',
        NULL,
        2
    ),
    (
        20,
        '26390be6-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer LT',
        40600,
        2,
        '2026-03-31 02:49:24+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        21,
        '26390c0e-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer LT',
        40600,
        2,
        '2026-03-31 14:45:41+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        22,
        '26390c36-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer LT',
        0,
        2,
        '2026-03-31 14:58:26+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        23,
        '26390c5e-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer LT',
        234,
        2,
        '2026-03-31 15:10:10+00',
        '2026-03-31 18:26:35+00',
        NULL,
        1
    ),
    (
        24,
        '26390c86-2d2f-11f1-a267-4fa3467576ef',
        'Casque Gamer LT',
        2700,
        2,
        '2026-03-31 15:12:02+00',
        '2026-04-01 07:17:27+00',
        NULL,
        1
    ),
    (
        25,
        '1a3a93b1-11f6-41ac-9f8e-bd890e876ab2',
        'Souris avec fil',
        1200,
        3,
        '2026-03-31 20:18:04+00',
        '2026-03-31 20:29:04+00',
        NULL,
        2
    ),
    (
        26,
        'cd165466-64e4-4e1d-a4e8-84519a623eaf',
        'Maths',
        1000,
        1,
        '2026-04-02 14:42:04+00',
        '2026-04-02 14:42:04+00',
        NULL,
        20
    );
-- user_shops
INSERT INTO user_shops (user_id, shop_id, assigned_at)
VALUES (2, 3, '2026-04-03 02:46:38+00');
-- =============================================================
-- Resynchronisation des séquences SERIAL
-- =============================================================
SELECT setval(
        'users_id_seq',
        (
            SELECT MAX(id)
            FROM users
        )
    );
SELECT setval(
        'categories_category_id_seq',
        (
            SELECT MAX(category_id)
            FROM categories
        )
    );
SELECT setval(
        'shops_shop_id_seq',
        (
            SELECT MAX(shop_id)
            FROM shops
        )
    );
SELECT setval(
        'products_product_id_seq',
        (
            SELECT MAX(product_id)
            FROM products
        )
    );
COMMIT;