-- Drop existing tables if they exist (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS tyre_details CASCADE;
DROP TABLE IF EXISTS claim_tyre_details CASCADE;
DROP TABLE IF EXISTS claims CASCADE;
DROP TABLE IF EXISTS warranties CASCADE;
DROP TABLE IF EXISTS shops CASCADE;

-- Drop any existing functions and triggers
DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create shops table (now includes admin credentials)
CREATE TABLE IF NOT EXISTS shops (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_name VARCHAR(255) NOT NULL,
    address TEXT NOT NULL,
    contact VARCHAR(50),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'master')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE OR REPLACE TRIGGER update_shops_updated_at
    BEFORE UPDATE ON shops
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create warranties table
CREATE TABLE IF NOT EXISTS warranties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    purchase_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    car_plate VARCHAR(20) NOT NULL,
    brand VARCHAR(100) NOT NULL,
    receipt VARCHAR(500) NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create claims table
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES shops(id),
    warranty_id UUID,
    customer_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    car_plate VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'unacknowledged' CHECK (status IN ('unacknowledged', 'pending', 'approved', 'rejected', 'closed')),
    rejection_reason TEXT,
    supporting_doc_url TEXT,
    brand VARCHAR(100),
    quantity INT,
    date_settled TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    date_closed TIMESTAMP WITH TIME ZONE
);

CREATE OR REPLACE TRIGGER update_claims_updated_at
    BEFORE UPDATE ON claims
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Create function to check max tyres per claim
CREATE OR REPLACE FUNCTION check_max_tyres_per_claim()
RETURNS TRIGGER AS $$
BEGIN
    IF (
        SELECT COUNT(*)
        FROM tyre_details
        WHERE claim_id = NEW.claim_id
    ) > 3  -- We check for > 3 because the current insert hasn't completed yet
    THEN
        RAISE EXCEPTION 'Maximum of 4 tyres allowed per claim';
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create tyre_details table
CREATE TABLE IF NOT EXISTS tyre_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id UUID REFERENCES claims(id),
    brand VARCHAR(100) NOT NULL,
    size VARCHAR(50) NOT NULL,
    tread_pattern VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT fk_claim
        FOREIGN KEY (claim_id)
        REFERENCES claims(id)
        ON DELETE CASCADE
);

-- Create trigger to enforce max tyres
CREATE TRIGGER enforce_max_tyres_per_claim
    BEFORE INSERT ON tyre_details
    FOR EACH ROW
    EXECUTE FUNCTION check_max_tyres_per_claim();

-- Create function to check max tyres requested per claim (dealer-entered, pre-acceptance)
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

-- Create claim_tyre_details table: per-tyre size + tread pattern entered by the dealer at claim
-- creation time. Brand is shared across all tyres in a claim and lives on claims.brand instead.
-- This is distinct from tyre_details, which holds the finalized tyres once a claim is accepted/rejected.
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

-- Create trigger to enforce max tyres
CREATE TRIGGER enforce_max_claim_tyre_details
    BEFORE INSERT ON claim_tyre_details
    FOR EACH ROW
    EXECUTE FUNCTION check_max_claim_tyre_details();

-- Insert test data
INSERT INTO shops (shop_name, address, contact, username, password, role) VALUES
('Master Admin', 'Corporate Office', '+60123456792', 'master', 'master', 'master'),
('Test Shop 1', '123 Test Street', '+60123456789', 'testshop1', 'password123', 'admin'),
('Test Shop 2', '456 Test Avenue', '+60123456790', 'testshop2', 'password456', 'admin'),
('Test Shop 3', '789 Test Road', '+60123456791', 'testshop3', 'password789', 'admin'),
('asd', 'asd', '+60123456791', 'asd', 'asd', 'admin');

-- Insert test data for warranties
INSERT INTO warranties (name, phone_number, email, purchase_date, expiry_date, car_plate, brand, receipt) VALUES
    -- Two valid warranties for ABC1234 (same customer, different brands - useful for testing the
    -- "verify car plate" unique-brands lookup). The more recent purchase (Pirelli) should sort first.
    ('John Doe', '+60123456789', 'john.doe@email.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'ABC1234', 'Pirelli', 'https://example.com/receipt1.pdf'),
    ('John Doe', '+60123456789', 'john.doe@email.com',
     CURRENT_DATE - INTERVAL '2 months',
     CURRENT_DATE + INTERVAL '4 months',
     'ABC1234', 'GT Radial', 'https://example.com/receipt2.pdf'),

    -- Valid warranty for XYZ5678
    ('Jane Smith', '+60123456790', 'jane.smith@email.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'XYZ5678', 'Kumho', 'https://example.com/receipt3.pdf'),

    -- Valid warranty for DEF9012
    ('Bob Johnson', '+60123456791', NULL,
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'DEF9012', 'Kumho', 'https://example.com/receipt4.pdf'),

    -- Valid warranty for JKL202
    ('Tom Brown', '+60123456703', 'tom@example.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'JKL202', 'RoadX', 'https://example.com/receipt5.pdf'),

    -- Valid warranty for MNO303
    ('Lisa Wong', '+60123456704', 'lisa@example.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'MNO303', 'Dunlop', 'https://example.com/receipt6.pdf'),

    -- Valid warranty for PQR404
    ('Emma Davis', '+60123456705', 'emma@example.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'PQR404', 'Kumho', 'https://example.com/receipt7.pdf'),

    -- Valid warranty for STU505
    ('Alex Tan', '+60123456706', 'alex@example.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'STU505', 'Kumho', 'https://example.com/receipt8.pdf'),

    -- BACKWARDS-COMPATIBILITY TEST: simulates a warranty registered before the "brand" field
    -- existed, backfilled as 'Unknown' by the production migration (see
    -- migrations/2026-08-09_add_brand_and_tyre_fields.sql). Try car plate OLD1234 on the dealer
    -- "Create Claim" -> Verify flow: it should still verify successfully (a valid warranty
    -- exists), but the Tyre Brand dropdown should NOT auto-select anything, since 'Unknown' isn't
    -- a real brand - you should have to manually pick one from the full list yourself.
    ('Legacy Customer', '+60123456799', 'legacy@example.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'OLD1234', 'Unknown', 'https://example.com/receipt_legacy.pdf'),

    -- Warranty backing the master-admin autofill test claim below (car plate NEW5678)
    ('Autofill Customer', '+60123456798', 'autofill@example.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'NEW5678', 'Continental', 'https://example.com/receipt_autofill.pdf'),

    -- Warranty backing the master-admin backwards-compatibility test claim below (car plate
    -- OLD5678). Master's Accept action is only enabled once a warranty is tagged to the claim
    -- (a separate, pre-existing product requirement, not something this feature changed), so
    -- this test case needs a real taggable warranty to exist even though the claim itself has
    -- no brand/quantity/tyre data.
    ('Old Claim Test', '+60123456797', 'oldclaim@example.com',
     CURRENT_DATE - INTERVAL '1 month',
     CURRENT_DATE + INTERVAL '5 months',
     'OLD5678', 'Kumho', 'https://example.com/receipt_oldclaim.pdf');

-- Insert test claims
-- brand/quantity now represent what the dealer enters on claim creation (see
-- models.CreateClaimRequest); the per-tyre size/pattern breakdown lives in claim_tyre_details
-- below. They are independent of warranty_id tagging.
INSERT INTO claims (shop_id, warranty_id, customer_name, phone_number, email, car_plate, status, rejection_reason, supporting_doc_url, date_settled, brand, quantity) VALUES
-- Unacknowledged claims (no warranty_id tagged yet)
((SELECT id FROM shops WHERE username = 'testshop1'), NULL, 'John Doe', '+60123456789', 'john@example.com', 'ABC1234', 'unacknowledged', NULL, NULL, NULL, 'Pirelli', 2),
((SELECT id FROM shops WHERE username = 'testshop2'), NULL, 'Tom Brown', '+60123456703', 'tom@example.com', 'JKL202', 'unacknowledged', NULL, NULL, NULL, 'RoadX', 1),
((SELECT id FROM shops WHERE username = 'testshop3'), NULL, 'Lisa Wong', '+60123456704', 'lisa@example.com', 'MNO303', 'unacknowledged', NULL, NULL, NULL, 'Dunlop', 4),

-- Pending claims (no warranty_id tagged yet)
((SELECT id FROM shops WHERE username = 'testshop1'), NULL, 'Jane Smith', '+60123456790', 'jane@example.com', 'XYZ5678', 'pending', NULL, NULL, NULL, 'Kumho', 2),
((SELECT id FROM shops WHERE username = 'testshop2'), NULL, 'Tom Brown', '+60123456703', 'tom@example.com', 'JKL202', 'pending', NULL, NULL, NULL, 'RoadX', 1),
((SELECT id FROM shops WHERE username = 'testshop3'), NULL, 'Lisa Wong', '+60123456704', 'lisa@example.com', 'MNO303', 'pending', NULL, NULL, NULL, 'Dunlop', 3),

-- BACKWARDS-COMPATIBILITY TEST: a pending claim with NO brand/quantity/claim_tyre_details at
-- all (simulates a claim created before this feature existed). Go to Master > Manage Claims >
-- Pending tab > "Tag Warranty" on this claim (customer "Old Claim Test", car plate OLD5678),
-- pick the OLD5678 warranty seeded above, then click Accept - it should NOT crash or error, and
-- should fall back to the same default blank form master always had (1 tyre, brand defaulted to
-- Pirelli, empty size, first tread pattern) since there's nothing to prefill from.
((SELECT id FROM shops WHERE username = 'testshop2'), NULL, 'Old Claim Test', '+60123456797', 'oldclaim@example.com', 'OLD5678', 'pending', NULL, NULL, NULL, NULL, NULL),

-- MASTER-ADMIN AUTOFILL TEST: a pending claim with 3 distinct tyres already entered by the
-- dealer. Go to Master > Manage Claims > Pending tab > Accept on this claim (customer
-- "Autofill Test", car plate NEW5678) - Brand/Quantity/Size/Tread Pattern should all prefill
-- automatically from the claim_tyre_details rows seeded below, and every field should still be
-- freely editable/overridable before you confirm.
((SELECT id FROM shops WHERE username = 'testshop1'), NULL, 'Autofill Test', '+60123456798', 'autofill@example.com', 'NEW5678', 'pending', NULL, NULL, NULL, 'Continental', 3),

-- Approved claims (must have warranty_id tagged) - brand/size/pattern match the tyre_details seeded below
((SELECT id FROM shops WHERE username = 'testshop1'), (SELECT id FROM warranties WHERE car_plate = 'DEF9012'), 'Bob Johnson', '+60123456791', 'bob@example.com', 'DEF9012', 'approved', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '5 days', 'Kumho', 2),
((SELECT id FROM shops WHERE username = 'testshop2'), (SELECT id FROM warranties WHERE car_plate = 'PQR404'), 'Emma Davis', '+60123456705', 'emma@example.com', 'PQR404', 'approved', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '3 days', 'Kumho', 4),
((SELECT id FROM shops WHERE username = 'testshop3'), (SELECT id FROM warranties WHERE car_plate = 'STU505'), 'Alex Tan', '+60123456706', 'alex@example.com', 'STU505', 'approved', NULL, NULL, CURRENT_TIMESTAMP - INTERVAL '1 day', 'Kumho', 1),

-- Rejected claims (no warranty_id needed)
((SELECT id FROM shops WHERE username = 'testshop1'), NULL, 'David Wilson', '+60123456707', 'david@example.com', 'ABC1234', 'rejected', 'Mechanical failure', NULL, CURRENT_TIMESTAMP - INTERVAL '7 days', 'Pirelli', 1),
((SELECT id FROM shops WHERE username = 'testshop2'), NULL, 'Grace Lee', '+60123456708', 'grace@example.com', 'XYZ5678', 'rejected', 'Mechanical failure', NULL, CURRENT_TIMESTAMP - INTERVAL '6 days', 'Kumho', 2),
((SELECT id FROM shops WHERE username = 'testshop3'), NULL, 'Ryan Lim', '+60123456709', 'ryan@example.com', 'DEF9012', 'rejected', 'Mechanical failure', NULL, CURRENT_TIMESTAMP - INTERVAL '4 days', 'Kumho', 1);

-- Add the per-tyre size/pattern breakdown the dealer entered for each claim above.
-- Matched 1:1 with each claim's `quantity`.
INSERT INTO claim_tyre_details (claim_id, size, tread_pattern) VALUES
    -- ABC1234 unacknowledged (John Doe, Pirelli, qty 2)
    ((SELECT id FROM claims WHERE status = 'unacknowledged' AND car_plate = 'ABC1234'), '205/55R16', 'Cinturato P7'),
    ((SELECT id FROM claims WHERE status = 'unacknowledged' AND car_plate = 'ABC1234'), '205/55R16', 'P Zero'),

    -- JKL202 unacknowledged (Tom Brown, RoadX, qty 1)
    ((SELECT id FROM claims WHERE status = 'unacknowledged' AND car_plate = 'JKL202'), '215/60R17', 'MX440'),

    -- MNO303 unacknowledged (Lisa Wong, Dunlop, qty 4)
    ((SELECT id FROM claims WHERE status = 'unacknowledged' AND car_plate = 'MNO303'), '185/60R14', 'SP Sport LM704'),
    ((SELECT id FROM claims WHERE status = 'unacknowledged' AND car_plate = 'MNO303'), '185/60R14', 'SP Sport J5'),
    ((SELECT id FROM claims WHERE status = 'unacknowledged' AND car_plate = 'MNO303'), '185/65R14', 'Enasave EC300'),
    ((SELECT id FROM claims WHERE status = 'unacknowledged' AND car_plate = 'MNO303'), '185/65R14', 'Direzza DZ102'),

    -- XYZ5678 pending (Jane Smith, Kumho, qty 2)
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'XYZ5678'), '195/65R15', 'Ecowing ES31'),
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'XYZ5678'), '195/65R15', 'Ecsta PS31'),

    -- JKL202 pending (Tom Brown, RoadX, qty 1)
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'JKL202'), '215/60R17', 'U01'),

    -- MNO303 pending (Lisa Wong, Dunlop, qty 3)
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'MNO303'), '185/60R14', 'SP Sport J5'),
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'MNO303'), '185/60R14', 'Enasave EC300'),
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'MNO303'), '185/65R14', 'Direzza DZ102'),

    -- NEW5678 pending (Autofill Test, Continental, qty 3) - MASTER-ADMIN AUTOFILL TEST, see note above
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'NEW5678'), '205/55R16', 'UltraContact UC6'),
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'NEW5678'), '205/55R16', 'EcoContact 6'),
    ((SELECT id FROM claims WHERE status = 'pending' AND car_plate = 'NEW5678'), '215/60R16', 'PremiumContact C'),

    -- DEF9012 approved (Bob Johnson, Kumho, qty 2) - matches the accepted tyre_details below
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'DEF9012'), '205/55R16', 'Ecowing ES31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'DEF9012'), '205/55R16', 'Ecowing ES31'),

    -- PQR404 approved (Emma Davis, Kumho, qty 4) - matches the accepted tyre_details below
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'), '215/45R17', 'Ecsta PS31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'), '215/45R17', 'Ecsta PS31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'), '215/45R17', 'Ecsta PS31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'), '215/45R17', 'Ecsta PS31'),

    -- STU505 approved (Alex Tan, Kumho, qty 1) - matches the accepted tyre_details below
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'STU505'), '225/40R18', 'Ecsta HS52'),

    -- ABC1234 rejected (David Wilson, Pirelli, qty 1)
    ((SELECT id FROM claims WHERE status = 'rejected' AND car_plate = 'ABC1234'), '205/55R16', 'P Zero'),

    -- XYZ5678 rejected (Grace Lee, Kumho, qty 2)
    ((SELECT id FROM claims WHERE status = 'rejected' AND car_plate = 'XYZ5678'), '195/65R15', 'Ecsta PS31'),
    ((SELECT id FROM claims WHERE status = 'rejected' AND car_plate = 'XYZ5678'), '195/65R15', 'Ecsta HS52'),

    -- DEF9012 rejected (Ryan Lim, Kumho, qty 1)
    ((SELECT id FROM claims WHERE status = 'rejected' AND car_plate = 'DEF9012'), '205/55R16', 'Ecowing ES31');

-- Add tyre details for approved claims
INSERT INTO tyre_details (claim_id, brand, size, tread_pattern) VALUES
    -- First approved claim (Bob Johnson - DEF9012) gets 2 tyres
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'DEF9012'),
     'Kumho', '205/55R16', 'Ecowing ES31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'DEF9012'),
     'Kumho', '205/55R16', 'Ecowing ES31'),
    
    -- Second approved claim (Emma Davis - PQR404) gets 4 tyres
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'),
     'Kumho', '215/45R17', 'Ecsta PS31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'),
     'Kumho', '215/45R17', 'Ecsta PS31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'),
     'Kumho', '215/45R17', 'Ecsta PS31'),
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'PQR404'),
     'Kumho', '215/45R17', 'Ecsta PS31'),
    
    -- Third approved claim (Alex Tan - STU505) gets 1 tyre
    ((SELECT id FROM claims WHERE status = 'approved' AND car_plate = 'STU505'),
     'Kumho', '225/40R18', 'Ecsta HS52');

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_warranties_car_plate ON warranties(car_plate);
CREATE INDEX IF NOT EXISTS idx_warranties_phone_number ON warranties(phone_number);
CREATE INDEX IF NOT EXISTS idx_warranties_expiry_date ON warranties(expiry_date);
CREATE INDEX IF NOT EXISTS idx_claims_warranty ON claims(warranty_id);
CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status);
CREATE INDEX IF NOT EXISTS idx_claims_shop ON claims(shop_id);
CREATE INDEX IF NOT EXISTS idx_claim_tyre_details_claim ON claim_tyre_details(claim_id);
CREATE INDEX IF NOT EXISTS idx_shops_username ON shops(username);
CREATE INDEX IF NOT EXISTS idx_shops_role ON shops(role);
