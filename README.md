# Tayaria Warranty Backend

A Go-based backend API for warranty and claim management system built with Gin, PostgreSQL (Supabase), and deployed on Render.

## Changelog (2024-06)
- **Claims Accept/Reject endpoints**: `/api/master/claim/:id/accept` and `/api/master/claim/:id/reject` for master admins, with clear status transitions and error handling.
- **snake_case JSON for claims**: All claim-related API payloads and DB structs now use snake_case for consistency with frontend.
- **Test data improvements**: All claims in test data have at least one valid, non-expired warranty for their car plate. Pending/unacknowledged claims do not have a warranty tagged. At least one claim has two valid warranties to choose from. Approved claims have tyre details (1-4 per claim).
- **Claim info endpoint**: `/api/master/claim/:id` returns full claim info, tagged warranty (if any), and all tyre details.
- **Nullable email handling**: Warranty email fields are nullable in DB and handled with `pgtype.Text` in Go.
- **Close claim endpoint**: `/api/admin/claim/:id/close` only sets `date_closed` for approved/rejected claims, returns updated claim.
- **Date settled**: All approved/rejected claims in test data have a `date_settled` value.
- **Frontend integration**: See below for migration and integration notes for frontend teams.
- **API docs**: All new/changed endpoints are documented below.

## Recent Changes - Database Integration & API Updates

### Database Integration
- **Full database integration** completed - no more mock data
- **PostgreSQL with Supabase** for production database
- **Proper date handling** with `time.Time` in Go, compatible with PostgreSQL `DATE` type
- **Transaction support** for atomic operations (e.g., warranty tagging)

### Authentication & Roles
- **JWT-based authentication** is required for all `/api/admin/*` and `/api/master/*` routes.
- Only **two roles** exist:  
  - `admin`: Regular shop admin  
  - `master`: Master admin (can manage all shops/claims)
- JWT tokens are issued on login:
  - `/api/admin/login` (admin, no expiry)
  - `/api/master/login` (master, 1 month expiry)
- **Include JWT in all protected requests**:
  ```
  Authorization: Bearer <jwt_token>
  ```

### Warranty API Endpoints (Public)

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
**Response**: Returns created warranty with `id`, `purchase_date`, `expiry_date`, and `is_used` (default: `false`)

#### Get Warranties by Car Plate
```
GET /api/user/warranties/car-plate/{car_plate}
```
**Response**: Array of warranties for the car plate

#### Check Valid Warranty
```
GET /api/user/warranties/valid/{car_plate}
```
**Response**: Returns valid warranty if exists and not expired

#### Get Warranty Receipt
```
GET /api/user/warranty/receipt/{id}
```

### Claims API Endpoints (Admin, JWT required)

#### Create Claim
```
POST /api/admin/claim
Content-Type: application/json
Authorization: Bearer <jwt_token>
{
  "customer_name": "John Doe",
  "phone_number": "+60123456789",
  "email": "john@example.com", // optional
  "car_plate": "ABC123",
  "description": "Engine overheating issue"
}
```
- `shop_id` is automatically taken from the JWT token
- `status` defaults to `pending`
- `warranty_id` can be tagged later using `UpdateClaimWarrantyID`
- All fields use snake_case in JSON

#### Get Shop Claims
```
GET /api/admin/claims
Authorization: Bearer <jwt_token>
```
- Returns all claims for the authenticated shop (empty array if none)
- Includes claim details, warranty info, and status

#### Tag Warranty to Claim
```
POST /api/admin/claim/:id/warranty/:warranty_id
Authorization: Bearer <jwt_token>
```
- Links a warranty to a claim
- **Automatically sets `is_used = true`** on the warranty
- Uses database transaction for atomicity
- Validates warranty exists and belongs to the car plate

#### Close Claim
```
POST /api/admin/claim/:id/close
Authorization: Bearer <jwt_token>
```
- Sets `date_closed` for approved/rejected claims only
- Returns updated claim

### Claims API Endpoints (Master, JWT required)

#### Get All Claims
```
GET /api/master/claims
Authorization: Bearer <jwt_token>
```
- Returns all claims across all shops

#### Get Claim Info by ID
```
GET /api/master/claim/:id
Authorization: Bearer <jwt_token>
```
- Returns all claim info, tagged warranty (if any), and all tyre details

#### Accept/Reject Claim
```
POST /api/master/claim/:id/accept
POST /api/master/claim/:id/reject
Authorization: Bearer <jwt_token>
```
- Approves or rejects a claim, sets status and `date_settled`
- Returns updated claim

### Middleware
- **AdminMiddleware**: Checks for a valid JWT and that `role` is `admin`.
- **MasterMiddleware**: Checks for `role` `master` for `/api/master/*` routes.

### Database Schema

#### Shops Table
```sql
CREATE TABLE shops (
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
```

#### Warranties Table
```sql
CREATE TABLE warranties (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(255), -- nullable
    purchase_date DATE NOT NULL,
    expiry_date DATE NOT NULL,
    car_plate VARCHAR(20) NOT NULL,
    receipt VARCHAR(500),
    is_used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

#### Claims Table
```sql
CREATE TABLE claims (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    shop_id UUID NOT NULL REFERENCES shops(id),
    warranty_id UUID REFERENCES warranties(id),
    customer_name VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    email VARCHAR(255), -- nullable
    car_plate VARCHAR(20) NOT NULL,
    description TEXT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'closed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    date_settled TIMESTAMP WITH TIME ZONE,
    date_closed TIMESTAMP WITH TIME ZONE
);
```

#### Tyre Details Table
```sql
CREATE TABLE tyre_details (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    claim_id UUID NOT NULL REFERENCES claims(id),
    brand VARCHAR(100),
    model VARCHAR(100),
    size VARCHAR(50),
    serial_number VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Key Features

#### Warranty Tracking
- **`is_used` field**: Tracks if warranty has been tagged to a claim
- **Automatic expiry calculation**: Based on purchase date + warranty period
- **Car plate validation**: Ensures warranty matches claim car plate
- **Receipt storage**: URL-based receipt storage (ready for S3/Supabase integration)
- **Nullable email**: Email is optional for warranty registration

#### Claim Management
- **Status workflow**: `pending` → `approved`/`rejected` (with `date_settled`) → `closed` (with `date_closed`)
- **Warranty linking**: Claims can be linked to existing warranties
- **Shop isolation**: Claims are automatically associated with shop from JWT
- **Transaction safety**: Warranty tagging uses database transactions
- **Tyre details**: Approved claims have 1-4 tyre details

#### Data Integrity
- **Foreign key constraints**: Proper relationships between tables
- **Check constraints**: Valid status values and role types
- **Timestamps**: Automatic `created_at` and `updated_at` tracking
- **UUID primary keys**: Secure, non-sequential identifiers
- **Test data**: All claims have valid, non-expired warranties; only approved claims have warranty tagged; at least one claim has two valid warranties

### Local Development Setup
1. **Environment Variables**: Create a `.env.development` file with:
   ```
   DATABASE_URL=postgresql://username:password@localhost:5432/database_name
   JWT_SECRET=your_jwt_secret_key
   PORT=8080
   ```
2. **Database Setup**: Run the SQL setup script:
   ```bash
   psql -d your_database -f setup.sql
   ```
3. **Run the Application**:
   ```bash
   go run main.go
   ```

### Testing the APIs
- Use the JWT token from login for all `/api/admin/*` and `/api/master/*` requests.
- All claim and shop queries use the `shop_id` from the JWT context.
- Empty claim lists always return `[]` (never `null`).
- Test warranty tagging with valid car plate matches.
- Test claim approval/rejection and closing flows.
- Test nullable email and tyre details handling.

### Error Handling
- `401 Unauthorized`: Missing/invalid JWT
- `403 Forbidden`: Wrong role
- `404 Not Found`: Warranty/claim not found
- `400 Bad Request`: Invalid data or business logic violations
- `200 OK` with `[]`: No claims found

### API Response Examples

#### Warranty Creation
```json
{
  "id": "uuid",
  "name": "John Doe",
  "phone_number": "+60123456789",
  "email": "john@example.com",
  "purchase_date": "2024-01-15T00:00:00Z",
  "expiry_date": "2025-01-15T00:00:00Z",
  "car_plate": "ABC1234",
  "receipt": "https://example.com/receipt.pdf",
  "is_used": false,
  "created_at": "2024-01-15T10:30:00Z"
}
```

#### Claim with Warranty and Tyre Details
```json
{
  "id": "uuid",
  "shop_id": "uuid",
  "warranty_id": "uuid",
  "customer_name": "John Doe",
  "phone_number": "+60123456789",
  "email": "john@example.com",
  "car_plate": "ABC1234",
  "description": "Engine overheating issue",
  "status": "approved",
  "created_at": "2024-01-15T10:30:00Z",
  "date_settled": "2024-01-20T10:30:00Z",
  "date_closed": null,
  "warranty": {
    "id": "uuid",
    "name": "John Doe",
    "purchase_date": "2024-01-15T00:00:00Z",
    "expiry_date": "2025-01-15T00:00:00Z",
    "is_used": true
  },
  "tyre_details": [
    {
      "id": "uuid",
      "brand": "Michelin",
      "model": "Primacy 4",
      "size": "205/55R16",
      "serial_number": "SN123456"
    }
  ]
}
```

### Frontend Integration & Migration Notes
- All claim-related payloads and types must use snake_case.
- Email is optional for claims and warranties.
- When displaying claim info, show warranty and tyre details if present.
- For claim approval/rejection, use the new endpoints and handle error responses as documented.
- See migration guide in project docs for more details.

### TODO Items
- [ ] Implement file upload to S3 or Supabase storage for receipts
- [ ] Add input validation for phone number format (Malaysian format)
- [ ] Add input validation for car plate format (Malaysian format)
- [ ] Implement warranty status management (active/expired/used)
- [ ] Add warranty search and filtering capabilities
- [ ] Add claim approval/rejection workflow (DONE)
- [ ] Implement warranty period configuration
- [ ] Improve test data for edge cases (DONE for basic flows)
- [ ] Add more API integration tests

### Deployment
The application is configured for deployment on Render:
- **Environment variables** set through Render dashboard
- **PostgreSQL add-on** for production database
- **Automatic deployments** from Git repository
- **Health checks** and monitoring

### API Documentation
Full API documentation is available in `openapi.yaml` and can be viewed with Swagger UI or similar tools. The specification includes:
- All endpoint definitions
- Request/response schemas
- Authentication requirements
- Error codes and messages 