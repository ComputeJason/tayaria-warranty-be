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

-- Create warranties table (use user_id UUID from Supabase auth.users)
CREATE TABLE IF NOT EXISTS warranties (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    shop_id UUID NOT NULL REFERENCES shops(id),
    pattern VARCHAR(100) NOT NULL,
    size VARCHAR(50) NOT NULL,
    date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    serial_number VARCHAR(100) UNIQUE NOT NULL,
    retailer VARCHAR(100),
    purchase_date DATE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create claims table (use user_id UUID from Supabase auth.users)
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    warranty_id UUID NOT NULL REFERENCES warranties(id),
    shop_id UUID NOT NULL REFERENCES shops(id),
    date TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
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

-- Insert test data for warranties (replace UUIDs with real Supabase user IDs for real data)
INSERT INTO warranties (user_id, shop_id, pattern, size, serial_number, retailer, purchase_date) VALUES
    ('58577453-ed0c-4cc8-8f80-5cec0703087f', (SELECT id FROM shops WHERE shop_name = 'Tyre Shop KL'), 'Michelin Pilot Sport 4', '225/45R17', 'MS4-2024-001', 'Tyre Shop KL', '2024-01-15'),
    ('58577453-ed0c-4cc8-8f80-5cec0703087f', (SELECT id FROM shops WHERE shop_name = 'Tyre Shop Penang'), 'Bridgestone Turanza T005', '205/55R16', 'BT5-2024-002', 'Tyre Shop Penang', '2024-02-01'),
    ('58577453-ed0c-4cc8-8f80-5cec0703087f', (SELECT id FROM shops WHERE shop_name = 'Tyre Shop JB'), 'Goodyear Eagle F1', '235/40R18', 'GE1-2024-003', 'Tyre Shop JB', '2024-03-10');

-- Insert test data for claims (replace UUIDs with real Supabase user IDs for real data)
INSERT INTO claims (user_id, warranty_id, shop_id, description, status, admin_comment) VALUES
    ('58577453-ed0c-4cc8-8f80-5cec0703087f', (SELECT id FROM warranties WHERE serial_number = 'MS4-2024-001'), (SELECT id FROM shops WHERE shop_name = 'Tyre Shop KL'), 'Tyre sidewall damage', 'pending', NULL),
    ('58577453-ed0c-4cc8-8f80-5cec0703087f', (SELECT id FROM warranties WHERE serial_number = 'BT5-2024-002'), (SELECT id FROM shops WHERE shop_name = 'Tyre Shop Penang'), 'Premature tread wear', 'approved', 'Approved after inspection'),
    ('58577453-ed0c-4cc8-8f80-5cec0703087f', (SELECT id FROM warranties WHERE serial_number = 'GE1-2024-003'), (SELECT id FROM shops WHERE shop_name = 'Tyre Shop JB'), 'Manufacturing defect', 'rejected', 'Not covered under warranty terms');

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_warranties_user_id ON warranties(user_id);
CREATE INDEX IF NOT EXISTS idx_warranties_serial ON warranties(serial_number);
CREATE INDEX IF NOT EXISTS idx_warranties_shop ON warranties(shop_id);
CREATE INDEX IF NOT EXISTS idx_claims_user_id ON claims(user_id);
CREATE INDEX IF NOT EXISTS idx_claims_warranty ON claims(warranty_id);
CREATE INDEX IF NOT EXISTS idx_claims_shop ON claims(shop_id);
CREATE INDEX IF NOT EXISTS idx_claims_status ON claims(status);
CREATE INDEX IF NOT EXISTS idx_admins_shop ON admins(shop_id);
