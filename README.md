# Tayaria Warranty Backend

A Go-based backend API for warranty and claim management system.

## Recent Changes - API Updates

The warranty and claim APIs have been updated to use new data models and database integration:

### New Warranty Model
- **Name** (required): Customer's full name
- **PhoneNumber** (required): Customer's phone number  
- **Email** (optional): Customer's email address
- **PurchaseDate** (required): Date of purchase (ISO 8601 format)
- **ExpiryDate** (auto-calculated): 6 months from purchase date
- **CarPlate** (required): Vehicle registration number
- **Receipt** (required): URL to receipt document

### Updated API Endpoints

#### Register Warranty
```
POST /api/user/warranty
Content-Type: application/json

{
  "name": "John Doe",
  "phone_number": "+60123456789",
  "email": "john.doe@email.com",
  "purchase_date": "2024-01-15T00:00:00Z",
  "car_plate": "ABC1234",
  "receipt": "https://example.com/receipt.pdf"
}
```

#### Get Warranties by Car Plate
```
GET /api/user/warranties/car-plate/{carPlate}
```

#### Check Valid Warranty
```
GET /api/user/warranties/valid/{carPlate}
```

#### Get Warranty Receipt
```
GET /api/user/warranty/receipt/{id}
```

### New Claim Model
- **WarrantyID** (required): ID of the warranty this claim is for
- **Description** (required): Description of the claim issue
- **Status** (auto-set): Claim status (unacknowledged, pending, approved, rejected)
- **AdminComment** (optional): Admin comments on the claim

### Updated Claim API Endpoints

#### Create Claim
```
POST /api/user/claim
Content-Type: application/json

{
  "warranty_id": "uuid-of-warranty",
  "description": "Tyre sidewall damage"
}
```

#### Get Claims
```
GET /api/user/claims?warranty_id={uuid}&status={status}
```

#### Change Claim Status
```
POST /api/user/claim/{id}/change-status
Content-Type: application/json

{
  "status": "approved",
  "admin_comment": "Approved after inspection"
}
```

#### Close Claim
```
POST /api/user/claim/{id}/close
```

## Database Schema Updates

The warranty and claim tables have been updated to match the new models:

```sql
CREATE TABLE warranties (
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

-- Claims table
CREATE TABLE claims (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    warranty_id UUID NOT NULL REFERENCES warranties(id),
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
    admin_comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

## Local Development Setup

1. **Environment Variables**
   Create a `.env.development` file with:
   ```
   APP_ENV=development
   SUPABASE_URL=your_supabase_url
   SUPABASE_KEY=your_supabase_key
   DATABASE_URL=your_database_url
   STORAGE_BUCKET=your_storage_bucket
   ```

2. **Database Setup**
   Run the SQL setup script:
   ```bash
   psql -d your_database -f setup.sql
   ```

3. **Run the Application**
   ```bash
   go run main.go
   ```

## Testing the Warranty APIs

### Test Data
The setup script includes test warranties:
- John Doe: ABC1234 (expires 2024-07-15)
- Jane Smith: XYZ5678 (expires 2024-08-01)  
- Bob Johnson: DEF9012 (expires 2024-09-10)

### Example API Calls

```bash
# Register a new warranty
curl -X POST http://localhost:8080/api/user/warranty \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "phone_number": "+60123456789",
    "email": "test@example.com",
    "purchase_date": "2024-01-15T00:00:00Z",
    "car_plate": "TEST123",
    "receipt": "https://example.com/receipt.pdf"
  }'

# Get warranties by car plate
curl http://localhost:8080/api/user/warranties/car-plate/ABC1234

# Check valid warranty
curl http://localhost:8080/api/user/warranties/valid/ABC1234

# Get warranty receipt
curl http://localhost:8080/api/user/warranty/receipt/{warranty_id}

# Create a claim
curl -X POST http://localhost:8080/api/user/claim \
  -H "Content-Type: application/json" \
  -d '{
    "warranty_id": "uuid-of-warranty",
    "description": "Tyre sidewall damage"
  }'

# Get claims
curl http://localhost:8080/api/user/claims?warranty_id={uuid}

# Change claim status
curl -X POST http://localhost:8080/api/user/claim/{claim_id}/change-status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "approved",
    "admin_comment": "Approved after inspection"
  }'
```

## TODO Items

- [ ] Implement file upload to S3 or Supabase storage for receipts
- [ ] Add input validation for phone number format
- [ ] Add input validation for car plate format
- [ ] Implement warranty status management
- [ ] Add warranty search and filtering capabilities

## Deployment

The application is configured for deployment on Render. Environment variables are set through the Render dashboard.

## API Documentation

Full API documentation is available in `openapi.yaml` and can be viewed with Swagger UI or similar tools. 