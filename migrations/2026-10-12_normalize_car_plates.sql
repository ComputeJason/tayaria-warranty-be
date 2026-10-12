-- Production-safe migration: normalize car plates (strip all whitespace, uppercase) on both
-- existing data and going forward.
--
-- Context: the app code (db/carplate.go's normalizeCarPlate, and its FE mirror in
-- src/lib/utils.ts) now strips every space - leading, trailing, and internal - and uppercases
-- car plates on every write and every lookup. That code change alone only fixes *future*
-- writes and searches; rows already stored with stray spaces (e.g. "ABC 1234", " abc1234 ")
-- stay dirty until backfilled here, which is why this migration exists: customers who
-- registered with spaces couldn't find their own warranty, and dealers couldn't verify claims
-- against those same plates. This migration is the one-time cleanup pass for that backlog.
--
-- Safe to run more than once (idempotent: already-clean rows are unaffected, the WHERE clause
-- just matches zero rows on a second run). Does not drop or delete anything.
--
-- Deployment order: run this anytime - it's independent of the brand/tyre-details migration
-- and needs no corresponding backend/frontend deploy to be safe on its own (old code never
-- reads a car plate differently based on whether it has spaces, it just fails to *match* one
-- that does - so cleaning the data is harmless even before the new normalizing code is live).
-- That said, it's most useful run together with this feature's BE/FE deploy, since that's what
-- actually stops the problem from recurring for new records.

BEGIN;

-- Match app's normalizeCarPlate semantics exactly: collapse (remove) every run of whitespace,
-- not just single spaces, then uppercase. '\s+' covers spaces/tabs/newlines the same way Go's
-- strings.Fields does.
UPDATE warranties
SET car_plate = UPPER(REGEXP_REPLACE(car_plate, '\s+', '', 'g'))
WHERE car_plate <> UPPER(REGEXP_REPLACE(car_plate, '\s+', '', 'g'));

UPDATE claims
SET car_plate = UPPER(REGEXP_REPLACE(car_plate, '\s+', '', 'g'))
WHERE car_plate <> UPPER(REGEXP_REPLACE(car_plate, '\s+', '', 'g'));

COMMIT;
