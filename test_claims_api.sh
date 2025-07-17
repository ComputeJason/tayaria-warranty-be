#!/bin/bash

# Test script for claims API endpoints
# Make sure the server is running on localhost:8080

BASE_URL="http://localhost:8080"

echo "Testing Claims API Endpoints..."
echo "================================"

# Test 1: Create a new claim
echo "1. Creating a new claim..."
CREATE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/admin/claim" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzaG9wX2lkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEiLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MzU2ODgwMDB9.example" \
  -d '{
    "customer_name": "John Doe",
    "phone_number": "+60123456789",
    "email": "john@example.com",
    "car_plate": "ABC1234",
    "description": "Engine overheating issue"
  }')

echo "Create Claim Response: $CREATE_RESPONSE"
CLAIM_ID=$(echo $CREATE_RESPONSE | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
echo "Claim ID: $CLAIM_ID"
echo ""

# Test 2: Get shop claims
echo "2. Getting shop claims..."
curl -s -X GET "$BASE_URL/api/admin/claims" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzaG9wX2lkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEiLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MzU2ODgwMDB9.example"
echo ""
echo ""

# Test 3: Accept a claim (if we have a claim ID)
if [ ! -z "$CLAIM_ID" ]; then
    echo "3. Accepting claim $CLAIM_ID..."
    ACCEPT_RESPONSE=$(curl -s -X POST "$BASE_URL/api/master/claim/$CLAIM_ID/accept" \
      -H "Content-Type: application/json" \
      -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzaG9wX2lkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEiLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MzU2ODgwMDB9.example" \
      -d '{
        "tyre_details": [
          {"brand": "Kumho", "size": "205/55R16", "tread_pattern": "Ecowing ES31"},
          {"brand": "Kumho", "size": "205/55R16", "tread_pattern": "Ecowing ES31"}
        ]
      }')
    
    echo "Accept Claim Response: $ACCEPT_RESPONSE"
    echo ""
else
    echo "3. Skipping accept claim test (no claim ID available)"
    echo ""
fi

# Test 4: Get claim info by ID
if [ ! -z "$CLAIM_ID" ]; then
    echo "4. Getting claim info for $CLAIM_ID..."
    curl -s -X GET "$BASE_URL/api/master/claim/$CLAIM_ID" \
      -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzaG9wX2lkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEiLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MzU2ODgwMDB9.example"
    echo ""
    echo ""
else
    echo "4. Skipping get claim info test (no claim ID available)"
    echo ""
fi

# Test 5: Get all claims by status
echo "5. Getting all pending claims..."
curl -s -X GET "$BASE_URL/api/master/claims?status=pending" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzaG9wX2lkIjoiMTIzNDU2NzgtMTIzNC0xMjM0LTEyMzQtMTIzNDU2Nzg5MDEiLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3MzU2ODgwMDB9.example"
echo ""
echo ""

echo "Tests completed!" 