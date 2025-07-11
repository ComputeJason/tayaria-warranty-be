-- Drop existing tables if they exist (in correct order due to foreign key constraints)
DROP TABLE IF EXISTS claims;
DROP TABLE IF EXISTS warranties;
DROP TABLE IF EXISTS admins;
DROP TABLE IF EXISTS shops;

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create shops table
CREATE TABLE IF NOT EXISTS shops (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    shop_name VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create admins table
CREATE TABLE IF NOT EXISTS admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL CHECK (role IN ('admin', 'super_user')),
    shop_id UUID REFERENCES shops(id),
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
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    admin_comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Insert test data for shops
INSERT INTO shops (shop_name, address) VALUES
    ('Tyre Shop KL', '123 Jalan Ampang, Kuala Lumpur'),
    ('Tyre Shop Penang', '456 Jalan Burma, Penang'),
    ('Tyre Shop JB', '789 Jalan Tebrau, Johor Bahru');

-- Insert test data for admins (plain text passwords for testing)
INSERT INTO admins (username, password, role, shop_id) VALUES
    ('admin1', 'adminpass1', 'admin', (SELECT id FROM shops WHERE shop_name = 'Tyre Shop KL')),
    ('admin2', 'adminpass2', 'admin', (SELECT id FROM shops WHERE shop_name = 'Tyre Shop Penang')),
    ('admin3', 'adminpass3', 'admin', (SELECT id FROM shops WHERE shop_name = 'Tyre Shop JB')),
    ('superadmin', 'superpass', 'super_user', NULL);

-- Insert test data for warranties
INSERT INTO warranties (name, phone_number, email, purchase_date, expiry_date, car_plate, receipt) VALUES
    ('John Doe', '+60123456789', 'john.doe@email.com', '2024-01-15', '2024-07-15', 'ABC1234', 'https://example.com/receipt1.pdf'),
    ('Jane Smith', '+60123456790', 'jane.smith@email.com', '2024-02-01', '2024-08-01', 'XYZ5678', 'https://example.com/receipt2.pdf'),
    ('Bob Johnson', '+60123456791', NULL, '2024-03-10', '2024-09-10', 'DEF9012', 'https://example.com/receipt3.pdf');

-- Insert test data for claims
INSERT INTO claims (warranty_id, description, status, admin_comment) VALUES
    ((SELECT id FROM warranties WHERE car_plate = 'ABC1234'), 'Tyre sidewall damage', 'pending', NULL),
    ((SELECT id FROM warranties WHERE car_plate = 'XYZ5678'), 'Premature tread wear', 'approved', 'Approved after inspection'),
    ((SELECT id FROM warranties WHERE car_plate = 'DEF9012'), 'Manufacturing defect', 'rejected', 'Not covered under warranty terms');

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_warranties_car_plate ON warranties(car_plate);
CREATE INDEX IF NOT EXISTS idx_warranties_phone_number ON warranties(phone_number);
CREATE INDEX IF NOT EXISTS idx_warranties_expiry_date ON warranties(expiry_date);
CREATE INDEX IF NOT EXISTS idx_claims_warranty ON claims(warranty_id);
CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status);
CREATE INDEX IF NOT EXISTS idx_admins_shop ON admins(shop_id);
