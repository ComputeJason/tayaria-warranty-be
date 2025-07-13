-- Drop existing tables if they exist (in correct order due to foreign key constraints)
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

-- Insert test data
INSERT INTO shops (shop_name, address, contact, username, password, role) VALUES
('Master Admin', 'Corporate Office', '+60123456792', 'master', 'master', 'master'),
('Test Shop 1', '123 Test Street', '+60123456789', 'testshop1', 'password123', 'admin'),
('Test Shop 2', '456 Test Avenue', '+60123456790', 'testshop2', 'password456', 'admin'),
('Test Shop 3', '789 Test Road', '+60123456791', 'testshop3', 'password789', 'admin'),
('asd', 'asd', '+60123456791', 'asd', 'asd', 'admin');

-- Insert test data for warranties
INSERT INTO warranties (name, phone_number, email, purchase_date, expiry_date, car_plate, receipt) VALUES
    ('John Doe', '+60123456789', 'john.doe@email.com', '2025-07-07', '2026-01-07', 'ABC1234', 'https://example.com/receipt1.pdf'),
    ('Jane Smith', '+60123456790', 'jane.smith@email.com', '2024-02-01', '2024-08-01', 'XYZ5678', 'https://example.com/receipt2.pdf'),
    ('07ob Johnson', '+60123456791', NULL, '2024-03-10', '2024-09-10', 'DEF9012', 'https://example.com/receipt3.pdf');

-- Insert test claims
INSERT INTO claims (shop_id, customer_name, phone_number, email, car_plate, status, warranty_id) VALUES
-- Unacknowledged claims
((SELECT id FROM shops WHERE username = 'testshop1'), 'John Doe', '+60123456789', 'john@example.com', 'ABC1234', 'unacknowledged', NULL),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Sarah Lee', '+60123456701', 'sarah@example.com', 'DEF789', 'unacknowledged', NULL),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Mike Chen', '+60123456702', 'mike@example.com', 'GHI101', 'unacknowledged', NULL),

-- Pending claims
((SELECT id FROM shops WHERE username = 'testshop1'), 'Jane Smith', '+60123456790', 'jane@example.com', 'XYZ5678', 'pending', (SELECT id FROM warranties WHERE car_plate = 'XYZ5678')),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Tom Brown', '+60123456703', 'tom@example.com', 'JKL202', 'pending', NULL),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Lisa Wong', '+60123456704', 'lisa@example.com', 'MNO303', 'pending', NULL),

-- Approved claims
((SELECT id FROM shops WHERE username = 'testshop1'), 'Bob Johnson', '+60123456791', 'bob@example.com', 'DEF9012', 'approved', (SELECT id FROM warranties WHERE car_plate = 'DEF9012')),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Emma Davis', '+60123456705', 'emma@example.com', 'PQR404', 'approved', NULL),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Alex Tan', '+60123456706', 'alex@example.com', 'STU505', 'approved', NULL),

-- Rejected claims
((SELECT id FROM shops WHERE username = 'testshop1'), 'David Wilson', '+60123456707', 'david@example.com', 'VWX606', 'rejected', NULL),
((SELECT id FROM shops WHERE username = 'testshop2'), 'Grace Lee', '+60123456708', 'grace@example.com', 'YZA707', 'rejected', NULL),
((SELECT id FROM shops WHERE username = 'testshop3'), 'Ryan Lim', '+60123456709', 'ryan@example.com', 'BCD808', 'rejected', NULL);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_warranties_car_plate ON warranties(car_plate);
CREATE INDEX IF NOT EXISTS idx_warranties_phone_number ON warranties(phone_number);
CREATE INDEX IF NOT EXISTS idx_warranties_expiry_date ON warranties(expiry_date);
CREATE INDEX IF NOT EXISTS idx_claims_warranty ON claims(warranty_id);
CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status);
CREATE INDEX IF NOT EXISTS idx_claims_shop ON claims(shop_id);
CREATE INDEX IF NOT EXISTS idx_shops_username ON shops(username);
CREATE INDEX IF NOT EXISTS idx_shops_role ON shops(role);
