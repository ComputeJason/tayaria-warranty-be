# Claims API Integration Guide

## Base URL
```
http://localhost:8080/api/user
```

## Authentication
Currently, these endpoints are public (no authentication required).

---

## 1. Create Claim API

### Endpoint
```
POST /api/user/claim
```

### Purpose
Creates a new claim for a customer. Validates that the car plate has a valid (non-expired) warranty before creating the claim.

### Request Headers
```
Content-Type: application/json
```

### Request Body
```typescript
interface CreateClaimRequest {
  customer_name: string;     // Required: Customer's full name
  phone_number: string;      // Required: Customer's phone number
  email?: string;            // Optional: Customer's email
  car_plate: string;         // Required: Vehicle registration number
  shop_id: string;           // Required: UUID of the shop handling the claim
}
```

### Request Example
```json
{
  "customer_name": "John Doe",
  "phone_number": "+60123456789",
  "email": "john@example.com",
  "car_plate": "ABC1234",
  "shop_id": "your-shop-uuid-here"
}
```

### Success Response (201 Created)
```typescript
interface ClaimResponse {
  id: string;                // UUID of the created claim
  warranty_id: string;       // UUID of the associated warranty
  shop_id: string;           // UUID of the shop handling the claim
  status: "pending";         // Always "pending" for new claims
  rejectionReason: string;   // Empty string for new claims
  dateSettled: null;         // null for new claims
  dateClosed: null;          // null for new claims
  customerName: string;      // Customer's name
  phoneNumber: string;       // Customer's phone number
  email: string;             // Customer's email (can be empty)
  carPlate: string;          // Vehicle registration number
  createdAt: string;         // ISO timestamp
  updatedAt: string;         // ISO timestamp
}
```

### Success Response Example
```json
{
  "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
  "warranty_id": "550e8400-e29b-41d4-a716-446655440000",
  "shop_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "status": "pending",
  "rejectionReason": "",
  "dateSettled": null,
  "dateClosed": null,
  "customerName": "John Doe",
  "phoneNumber": "+60123456789",
  "email": "john@example.com",
  "carPlate": "ABC1234",
  "createdAt": "2024-07-12T10:30:00Z",
  "updatedAt": "2024-07-12T10:30:00Z"
}
```

### Error Responses

#### 400 Bad Request - Validation Errors
```json
{
  "error": "Key: 'CreateClaimRequest.customer_name' Error:Field validation for 'customer_name' failed on the 'required' tag"
}
```

#### 404 Not Found - No Valid Warranty
```json
{
  "error": "No valid warranty found for this car plate"
}
```

#### 500 Internal Server Error - Database/Server Issues
```json
{
  "error": "failed to create claim: [detailed error message]"
}
```

### Frontend Integration Example
```javascript
// Create Claim Function
async function createClaim(claimData) {
  try {
    const response = await fetch('http://localhost:8080/api/user/claim', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(claimData)
    });

    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || 'Failed to create claim');
    }

    return {
      success: true,
      data: data
    };
  } catch (error) {
    return {
      success: false,
      error: error.message
    };
  }
}

// Usage Example
const claimData = {
  customer_name: "John Doe",
  phone_number: "+60123456789",
  email: "john@example.com",
  car_plate: "ABC1234",
  shop_id: "your-shop-uuid-here"
};

const result = await createClaim(claimData);
if (result.success) {
  console.log('Claim created:', result.data);
} else {
  console.error('Error:', result.error);
}
```

---

## 2. Get Shop Claims API

### Endpoint
```
GET /api/user/claims/{shop_id}
```

### Purpose
Retrieves all claims for a specific shop, ordered by creation date (newest first).

### Request Parameters
- `shop_id` (path parameter): UUID of the shop

### Request Example
```
GET /api/user/claims/6ba7b810-9dad-11d1-80b4-00c04fd430c8
```

### Success Response (200 OK)
```typescript
interface ClaimResponse {
  id: string;
  warranty_id: string;
  shop_id: string;
  status: "pending" | "approved" | "rejected";
  rejectionReason: string;       // Only populated when status is "rejected"
  dateSettled: string | null;    // ISO timestamp when status changed to approved/rejected
  dateClosed: string | null;     // ISO timestamp when claim was closed
  customerName: string;
  phoneNumber: string;
  email: string;
  carPlate: string;
  createdAt: string;             // ISO timestamp
  updatedAt: string;             // ISO timestamp
}

type GetShopClaimsResponse = ClaimResponse[];
```

### Success Response Example
```json
[
  {
    "id": "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    "warranty_id": "550e8400-e29b-41d4-a716-446655440000",
    "shop_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "status": "pending",
    "rejectionReason": "",
    "dateSettled": null,
    "dateClosed": null,
    "customerName": "John Doe",
    "phoneNumber": "+60123456789",
    "email": "john@example.com",
    "carPlate": "ABC1234",
    "createdAt": "2024-07-12T10:30:00Z",
    "updatedAt": "2024-07-12T10:30:00Z"
  },
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "warranty_id": "550e8400-e29b-41d4-a716-446655440001",
    "shop_id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
    "status": "rejected",
    "rejectionReason": "Warranty expired and customer provided insufficient documentation",
    "dateSettled": "2024-07-11T14:30:00Z",
    "dateClosed": "2024-07-11T16:00:00Z",
    "customerName": "Jane Smith",
    "phoneNumber": "+60123456790",
    "email": "jane@example.com",
    "carPlate": "XYZ5678",
    "createdAt": "2024-07-11T09:15:00Z",
    "updatedAt": "2024-07-11T16:00:00Z"
  }
]
```

### Error Responses

#### 500 Internal Server Error - Database/Server Issues
```json
{
  "error": "failed to query claims: [detailed error message]"
}
```

### Frontend Integration Example
```javascript
// Get Shop Claims Function
async function getShopClaims(shopId) {
  try {
    const response = await fetch(`http://localhost:8080/api/user/claims/${shopId}`);
    const data = await response.json();

    if (!response.ok) {
      throw new Error(data.error || 'Failed to fetch claims');
    }

    return {
      success: true,
      data: data
    };
  } catch (error) {
    return {
      success: false,
      error: error.message
    };
  }
}

// Usage Example
const shopId = "6ba7b810-9dad-11d1-80b4-00c04fd430c8";
const result = await getShopClaims(shopId);

if (result.success) {
  console.log('Claims retrieved:', result.data);
  // result.data is an array of claims
} else {
  console.error('Error:', result.error);
}
```

---

## Error Handling Best Practices

### 1. Network Errors
```javascript
// Handle network/connection errors
try {
  const response = await fetch(url);
  // ... handle response
} catch (error) {
  if (error.name === 'TypeError') {
    // Network error (server down, no internet, etc.)
    console.error('Network error:', error.message);
  }
}
```

### 2. API Error Responses
```javascript
// Handle API error responses
if (!response.ok) {
  const errorData = await response.json();
  
  switch (response.status) {
    case 400:
      // Validation errors
      console.error('Validation error:', errorData.error);
      break;
    case 404:
      // Not found errors (no valid warranty, etc.)
      console.error('Not found:', errorData.error);
      break;
    case 500:
      // Server errors
      console.error('Server error:', errorData.error);
      break;
    default:
      console.error('Unknown error:', errorData.error);
  }
}
```

---

## Testing & Setup

### 1. Test Car Plates (Valid Warranties)
Use these car plates for testing (they have valid warranties in the test data):
- `ABC1234`
- `XYZ5678`
- `DEF9012`

### 2. Getting Shop IDs
To get valid shop IDs for testing, you can:

#### Option A: Use Database Query
```sql
SELECT id, shop_name FROM shops;
```

#### Option B: Temporarily add a test endpoint
```javascript
// Add this to your backend for testing
GET /api/test/shops
```

### 3. Environment Variables
Make sure your backend is running with:
```bash
APP_ENV=development go run main.go
```

---

## Frontend State Management Examples

### React State Example
```javascript
// React component example
const [claims, setClaims] = useState([]);
const [loading, setLoading] = useState(false);
const [error, setError] = useState(null);

const fetchClaims = async (shopId) => {
  setLoading(true);
  setError(null);
  
  const result = await getShopClaims(shopId);
  
  if (result.success) {
    setClaims(result.data);
  } else {
    setError(result.error);
  }
  
  setLoading(false);
};

const handleCreateClaim = async (claimData) => {
  const result = await createClaim(claimData);
  
  if (result.success) {
    // Refresh claims list
    await fetchClaims(shopId);
    // Show success message
  } else {
    setError(result.error);
  }
};
```

---

## Migration from Mock Data

### Your Current Mock Data Format
```javascript
// Your existing mock data structure
{
  id: '3',
  customerName: 'Michael Brown',
  phoneNumber: '+60123456790',
  email: 'michael@example.com',
  carPlate: 'DEF456',
  status: 'approved',
  createdAt: '2024-03-18T09:15:00Z',
  dateSettled: '2024-03-19T14:30:00Z',
  dateClosed: '2024-03-20T10:00:00Z',
  rejectionReason: 'Warranty expired...'
}
```

### API Response Format
The API returns the same structure with these key differences:
- ✅ Field names match exactly (camelCase)
- ✅ Date fields are ISO strings or null
- ✅ All required fields are present
- ✅ Status values match: "pending", "approved", "rejected"

### Easy Migration
You can replace your mock data calls with API calls directly since the response format matches your existing structure!

```javascript
// Replace this:
const mockClaims = [...];

// With this:
const { data: claims } = await getShopClaims(shopId);
``` 