# Master API Documentation

## Authentication

All master APIs require a valid JWT token obtained through the master login endpoint. The token must be included in the `Authorization` header of all requests in the format: `Bearer <token>`.

### Master Login

Authenticates a master user and returns a JWT token.

```
POST /api/master/login
```

#### Request Body
```json
{
  "username": "string",
  "password": "string"
}
```

#### Response (200 OK)
```json
{
  "token": "string",
  "shop": {
    "id": "string",
    "shopName": "string",
    "address": "string",
    "contact": "string",
    "username": "string",
    "role": "master",
    "createdAt": "string"
  }
}
```

#### Error Responses
- `401 Unauthorized`: Invalid credentials or non-master account
- `400 Bad Request`: Invalid request body
- `500 Internal Server Error`: Server error

## Retail Account Management APIs

### Create Retail Account

Creates a new retail shop account in the system.

```
POST /api/master/account
```

#### Request Body
```json
{
  "shopName": "string",
  "address": "string",
  "contact": "string",
  "username": "string",
  "password": "string"
}
```

#### Response (201 Created)
```json
{
  "id": "string",
  "shopName": "string",
  "address": "string",
  "contact": "string",
  "username": "string",
  "role": "admin",
  "createdAt": "string"
}
```

#### Error Responses
- `401 Unauthorized`: Invalid or missing token
- `403 Forbidden`: Non-master user
- `409 Conflict`: Username already exists
- `400 Bad Request`: Invalid request body
- `500 Internal Server Error`: Server error

### Get Retail Accounts

Retrieves a list of all retail shop accounts.

```
GET /api/master/account
```

#### Response (200 OK)
```json
[
  {
    "id": "string",
    "shopName": "string",
    "address": "string",
    "contact": "string",
    "username": "string",
    "role": "admin",
    "createdAt": "string"
  }
]
```

#### Error Responses
- `401 Unauthorized`: Invalid or missing token
- `403 Forbidden`: Non-master user
- `500 Internal Server Error`: Server error

## Important Notes

1. **Authentication Header**
   ```
   Authorization: Bearer <your_jwt_token>
   ```

2. **Account Creation**
   - Each retail account must have a unique username
   - Passwords are stored in plain text (Note: This should be enhanced with proper hashing in production)
   - All retail accounts are created with 'admin' role

3. **Token Expiration**
   - Master tokens expire after 1 month
   - You'll need to re-login to get a new token after expiration

4. **Response Security**
   - Password fields are never returned in responses
   - All successful responses include timestamps and unique IDs 