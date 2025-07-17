-- Drop existing tables if they exist (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS tyre_details CASCADE;
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

-- Insert test data
INSERT INTO shops (shop_name, address, contact, username, password, role) VALUES
('Master Admin', 'Corporate Office', '+60123456792', 'master', 'master', 'master'),
('Test Shop 1', '123 Test Street', '+60123456789', 'testshop1', 'password123', 'admin'),
('Test Shop 2', '456 Test Avenue', '+60123456790', 'testshop2', 'password456', 'admin'),
('Test Shop 3', '789 Test Road', '+60123456791', 'testshop3', 'password789', 'admin'),
('asd', 'asd', '+60123456791', 'asd', 'asd', 'admin');

-- Insert test data for warranties
INSERT INTO warranties (name, phone_number, email, purchase_date, expiry_date, car_plate, receipt) VALUES
    -- Two valid warranties for ABC1234 (same customer)
    ('John Doe', '+60123456789', 'john.doe@email.com', 
     CURRENT_DATE - INTERVAL '1 month', 
     CURRENT_DATE + INTERVAL '5 months', 
     'ABC1234', 'https://example.com/receipt1.pdf'),
    ('John Doe', '+60123456789', 'john.doe@email.com', 
     CURRENT_DATE - INTERVAL '2 months', 
     CURRENT_DATE + INTERVAL '4 months', 
     'ABC1234', 'https://example.com/receipt2.pdf'),
    
    -- Valid warranty for XYZ5678
    ('Jane Smith', '+60123456790', 'jane.smith@email.com', 
     CURRENT_DATE - INTERVAL '1 month', 
     CURRENT_DATE + INTERVAL '5 months', 
     'XYZ5678', 'https://example.com/receipt3.pdf'),
    
    -- Valid warranty for DEF9012
    ('Bob Johnson', '+60123456791', NULL, 
     CURRENT_DATE - INTERVAL '1 month', 
     CURRENT_DATE + INTERVAL '5 months', 
     'DEF9012', 'https://example.com/receipt4.pdf'),
    
    -- Valid warranty for JKL202
    ('Tom Brown', '+60123456703', 'tom@example.com', 
     CURRENT_DATE - INTERVAL '1 month', 
     CURRENT_DATE + INTERVAL '5 months', 
     'JKL202', 'https://example.com/receipt5.pdf'),
    
    -- Valid warranty for MNO303
    ('Lisa Wong', '+60123456704', 'lisa@example.com', 
     CURRENT_DATE - INTERVAL '1 month', 
     CURRENT_DATE + INTERVAL '5 months', 
     'MNO303', 'https://example.com/receipt6.pdf'),
    
    -- Valid warranty for PQR404
    ('Emma Davis', '+60123456705', 'emma@example.com', 
     CURRENT_DATE - INTERVAL '1 month', 
     CURRENT_DATE + INTERVAL '5 months', 
     'PQR404', 'https://example.com/receipt7.pdf'),
    
    -- Valid warranty for STU505
    ('Alex Tan', '+60123456706', 'alex@example.com', 
     CURRENT_DATE - INTERVAL '1 month', 
     CURRENT_DATE + INTERVAL '5 months', 
     'STU505', 'https://example.com/receipt8.pdf');

-- Insert test claims
INSERT INTO claims (shop_id, customer_name, phone_number, email, car_plate, status, warranty_id, date_settled) VALUES
-- Unacknowledged claims (no warranty_id tagged yet)
((SELECT id FROM shops WHERE username = 'testshop1'), 'John Doe', '+60123456789', 'john@example.com', 'ABC1234', 'unacknowledged', NULL, NULL),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Tom Brown', '+60123456703', 'tom@example.com', 'JKL202', 'unacknowledged', NULL, NULL),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Lisa Wong', '+60123456704', 'lisa@example.com', 'MNO303', 'unacknowledged', NULL, NULL),

-- Pending claims (no warranty_id tagged yet)
((SELECT id FROM shops WHERE username = 'testshop1'), 'Jane Smith', '+60123456790', 'jane@example.com', 'XYZ5678', 'pending', NULL, NULL),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Tom Brown', '+60123456703', 'tom@example.com', 'JKL202', 'pending', NULL, NULL),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Lisa Wong', '+60123456704', 'lisa@example.com', 'MNO303', 'pending', NULL, NULL),

-- Approved claims (must have warranty_id tagged)
((SELECT id FROM shops WHERE username = 'testshop1'), 'Bob Johnson', '+60123456791', 'bob@example.com', 'DEF9012', 'approved', 
 (SELECT id FROM warranties WHERE car_plate = 'DEF9012'), CURRENT_TIMESTAMP - INTERVAL '5 days'),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Emma Davis', '+60123456705', 'emma@example.com', 'PQR404', 'approved',
 (SELECT id FROM warranties WHERE car_plate = 'PQR404'), CURRENT_TIMESTAMP - INTERVAL '3 days'),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Alex Tan', '+60123456706', 'alex@example.com', 'STU505', 'approved',
 (SELECT id FROM warranties WHERE car_plate = 'STU505'), CURRENT_TIMESTAMP - INTERVAL '1 day'),

-- Rejected claims (no warranty_id needed)
((SELECT id FROM shops WHERE username = 'testshop1'), 'David Wilson', '+60123456707', 'david@example.com', 'ABC1234', 'rejected', NULL, CURRENT_TIMESTAMP - INTERVAL '7 days'),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Grace Lee', '+60123456708', 'grace@example.com', 'XYZ5678', 'rejected', NULL, CURRENT_TIMESTAMP - INTERVAL '6 days'),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Ryan Lim', '+60123456709', 'ryan@example.com', 'DEF9012', 'rejected', NULL, CURRENT_TIMESTAMP - INTERVAL '4 days');

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
CREATE INDEX IF NOT EXISTS idx_shops_username ON shops(username);
CREATE INDEX IF NOT EXISTS idx_shops_role ON shops(role);
