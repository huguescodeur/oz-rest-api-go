BEGIN;

-- Statut commande
DO $$ BEGIN
    CREATE TYPE order_status AS ENUM ('PENDING', 'CONFIRMED', 'DELIVERED', 'CANCELLED');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Modes de paiement
DO $$ BEGIN
    CREATE TYPE payment_method AS ENUM ('ESPECES', 'ORANGE_MONEY', 'MTN_MOMO', 'WAVE', 'CREDIT', 'AUTRE');
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

-- Enrichissement de la table orders
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS status          order_status    NOT NULL DEFAULT 'PENDING',
    ADD COLUMN IF NOT EXISTS customer_name   VARCHAR(100),
    ADD COLUMN IF NOT EXISTS customer_phone  VARCHAR(20),
    ADD COLUMN IF NOT EXISTS payment_method  payment_method,
    ADD COLUMN IF NOT EXISTS notes           TEXT,
    ADD COLUMN IF NOT EXISTS updated_at      TIMESTAMPTZ     DEFAULT NOW();

CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

-- Seuil de stock faible
ALTER TABLE stocks
    ADD COLUMN IF NOT EXISTS min_stock INT NOT NULL DEFAULT 0;

COMMIT;

-- Fix contrainte catégorie unique par propriétaire (multi-tenant)
ALTER TABLE categories DROP CONSTRAINT IF EXISTS uq_category_name;
CREATE UNIQUE INDEX IF NOT EXISTS uq_category_name_per_owner ON categories(category_name, user_id) WHERE deleted_at IS NULL;
