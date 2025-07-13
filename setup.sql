-- Drop existing tables if they exist (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS claims;
DROP TABLE IF EXISTS warranties;
DROP TABLE IF EXISTS shops;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create shops table (now includes admin credentials)
CREATE TABLE IF NOT EXISTS shops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_name VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    contact VARCHAR(20),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'admin' CHECK (role IN ('admin', 'super_user')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

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
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    warranty_id UUID NOT NULL REFERENCES warranties(id),
    shop_id UUID NOT NULL REFERENCES shops(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    rejection_reason TEXT,
    date_settled TIMESTAMP WITH TIME ZONE,
    date_closed TIMESTAMP WITH TIME ZONE,
    -- Customer info (duplicated from form)
    customer_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(100),
    car_plate VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert test data for shops (now includes admin credentials)
INSERT INTO shops (shop_name, address, contact, username, password, role) VALUES
    ('Tyre Shop KL', '123 Jalan Ampang, Kuala Lumpur', '+60123456789', 'admin1', 'adminpass1', 'admin'),
    ('Tyre Shop Penang', '456 Jalan Burma, Penang', '+60123456790', 'admin2', 'adminpass2', 'admin'),
    ('Tyre Shop JB', '789 Jalan Tebrau, Johor Bahru', '+60123456791', 'admin3', 'adminpass3', 'admin'),
    ('Master Admin', 'Corporate Office', '+60123456792', 'superadmin', 'superpass', 'super_user');

-- Insert test data for warranties
INSERT INTO warranties (name, phone_number, email, purchase_date, expiry_date, car_plate, receipt) VALUES
    ('John Doe', '+60123456789', 'john.doe@email.com', '2025-07-07', '2026-01-07', 'ABC1234', 'https://example.com/receipt1.pdf'),
    ('Jane Smith', '+60123456790', 'jane.smith@email.com', '2024-02-01', '2024-08-01', 'XYZ5678', 'https://example.com/receipt2.pdf'),
    ('Bob Johnson', '+60123456791', NULL, '2024-03-10', '2024-09-10', 'DEF9012', 'https://example.com/receipt3.pdf');

-- Insert test data for claims
INSERT INTO claims (warranty_id, shop_id, status, rejection_reason, customer_name, phone_number, email, car_plate) VALUES
    ((SELECT id FROM warranties WHERE car_plate = 'ABC1234'), (SELECT id FROM shops WHERE username = 'admin1'), 'pending', NULL, 'John Doe', '+60123456789', 'john.doe@email.com', 'ABC1234'),
    ((SELECT id FROM warranties WHERE car_plate = 'XYZ5678'), (SELECT id FROM shops WHERE username = 'admin2'), 'approved', NULL, 'Jane Smith', '+60123456790', 'jane.smith@email.com', 'XYZ5678'),
    ((SELECT id FROM warranties WHERE car_plate = 'DEF9012'), (SELECT id FROM shops WHERE username = 'admin3'), 'rejected', 'Not covered under warranty terms', 'Bob Johnson', '+60123456791', NULL, 'DEF9012');

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_warranties_car_plate ON warranties(car_plate);
CREATE INDEX IF NOT EXISTS idx_warranties_phone_number ON warranties(phone_number);
CREATE INDEX IF NOT EXISTS idx_warranties_expiry_date ON warranties(expiry_date);
CREATE INDEX IF NOT EXISTS idx_claims_warranty ON claims(warranty_id);
CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status);
CREATE INDEX IF NOT EXISTS idx_claims_shop ON claims(shop_id);
CREATE INDEX IF NOT EXISTS idx_shops_username ON shops(username);
CREATE INDEX IF NOT EXISTS idx_shops_role ON shops(role);
