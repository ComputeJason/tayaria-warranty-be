-- Production-safe migration for the "tyre brand" + "dealer claim form" enhancements.
--
-- Unlike setup.sql (which DROPS and recreates tables — dev/local only, never run against prod),
-- this script only ADDS columns/tables. It does not drop, truncate, or delete anything, and it
-- is safe to run more than once (every statement is idempotent).
--
-- Deployment order: run this migration first (safe on its own, old backend code ignores the new
-- columns/table). Then deploy the backend and frontend as close together as possible - the new
-- backend requires a `brand` field on warranty/claim creation that the old frontend never sends,
-- and the new frontend calls a `/admin/warranties/verify/:carPlate` route that doesn't exist on
-- the old backend. Whichever of the two you deploy first has a brief window where warranty
-- registration / claim creation will fail for users until the other one catches up.

BEGIN;

-- 1. warranties.brand
--    New required field going forward, but existing rows won't have a value yet. We add it as
--    nullable first, backfill existing rows with a placeholder, then lock it down to NOT NULL
--    so it matches what the Go code expects (it reads this column as a plain, non-nullable string).
ALTER TABLE warranties ADD COLUMN IF NOT EXISTS brand VARCHAR(100);
UPDATE warranties SET brand = 'Unknown' WHERE brand IS NULL;
ALTER TABLE warranties ALTER COLUMN brand SET NOT NULL;

-- 2. claims: brand + tyre count now captured from the dealer at claim-creation time.
--    Stay nullable on purpose — historic claims created before this feature won't have them,
--    and both the Go code and the frontend already handle that (nil-safe scanning on the
--    backend, and the master Accept form falls back to its old blank-form behavior on the
--    frontend when this data is missing).
ALTER TABLE claims ADD COLUMN IF NOT EXISTS brand VARCHAR(100);
ALTER TABLE claims ADD COLUMN IF NOT EXISTS quantity INT;

-- 3. claim_tyre_details: per-tyre size + tread pattern entered by the dealer at claim creation
--    time (one row per tyre, brand is shared and lives on claims.brand instead). Mirrors the
--    existing tyre_details table, which instead holds the finalized tyres once a claim is
--    accepted/rejected.
CREATE OR REPLACE FUNCTION check_max_claim_tyre_details()
RETURNS TRIGGER AS $$
BEGIN
    IF (
        SELECT COUNT(*)
        FROM claim_tyre_details
        WHERE claim_id = NEW.claim_id
    ) > 3  -- We check for > 3 because the current insert hasn't completed yet
    THEN
        RAISE EXCEPTION 'Maximum of 4 tyres allowed per claim';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS claim_tyre_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id UUID REFERENCES claims(id),
    size VARCHAR(50) NOT NULL,
    tread_pattern VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_claim_tyre_details_claim
        FOREIGN KEY (claim_id)
        REFERENCES claims(id)
        ON DELETE CASCADE
);

DROP TRIGGER IF EXISTS enforce_max_claim_tyre_details ON claim_tyre_details;
CREATE TRIGGER enforce_max_claim_tyre_details
    BEFORE INSERT ON claim_tyre_details
    FOR EACH ROW
    EXECUTE FUNCTION check_max_claim_tyre_details();

CREATE INDEX IF NOT EXISTS idx_claim_tyre_details_claim ON claim_tyre_details(claim_id);

COMMIT;
