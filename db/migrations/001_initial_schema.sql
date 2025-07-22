-- Create claims table
CREATE TABLE IF NOT EXISTS claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_name VARCHAR(255) NOT NULL,
    contact VARCHAR(255) NOT NULL,
    car_plate VARCHAR(20) NOT NULL,
    tyre_details JSONB NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    warranty_id UUID REFERENCES warranties(id),
    supporting_doc_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
); 