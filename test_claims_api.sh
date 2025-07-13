#!/bin/bash

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzaG9wX2lkIjoiYWRkMWQwYjAtMTI5Zi00MDZkLWJjYzctMDZjNWVhYTFmNGJlIiwidXNlcm5hbWUiOiJtYXN0ZXIiLCJyb2xlIjoibWFzdGVyIiwiZXhwIjoxNzU0OTg4NDIzLCJuYmYiOjE3NTIzOTY0MjMsImlhdCI6MTc1MjM5NjQyM30.E7Ey0qKnZNN3zVtTZeCr6_cyreRFVW8TAa0Nl1zec3s"
BASE_URL="http://localhost:8080"

echo "1. Getting unacknowledged claims:"
curl -X GET "$BASE_URL/api/master/claims?status=unacknowledged" \
-H "Authorization: Bearer $TOKEN"
echo -e "\n"

echo "2. Getting pending claims:"
curl -X GET "$BASE_URL/api/master/claims?status=pending" \
-H "Authorization: Bearer $TOKEN"
echo -e "\n"

# Use a claim ID from the previous response
CLAIM_ID="YOUR_CLAIM_ID"  # Replace this with an actual claim ID from the response

echo "3. Change claim to pending:"
curl -X POST "$BASE_URL/api/master/claim/$CLAIM_ID/pending" \
-H "Authorization: Bearer $TOKEN"
echo -e "\n"

echo "4. Accept claim with tyre details:"
curl -X POST "$BASE_URL/api/master/claim/$CLAIM_ID/accept" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "tyre_details": [
    {
      "brand": "Michelin",
      "size": "205/55R16",
      "cost": 450.00
    },
    {
      "brand": "Michelin",
      "size": "205/55R16",
      "cost": 450.00
    }
  ]
}'
echo -e "\n"

echo "5. Reject claim with reason:"
curl -X POST "$BASE_URL/api/master/claim/$CLAIM_ID/reject" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "rejection_reason": "Invalid warranty claim"
}'
echo -e "\n"

echo "6. Try to accept a rejected claim (should fail):"
curl -X POST "$BASE_URL/api/master/claim/$CLAIM_ID/accept" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "tyre_details": [
    {
      "brand": "Michelin",
      "size": "205/55R16",
      "cost": 450.00
    }
  ]
}'
echo -e "\n"

echo "7. Try to add more than 4 tyres (should fail):"
curl -X POST "$BASE_URL/api/master/claim/$CLAIM_ID/accept" \
-H "Authorization: Bearer $TOKEN" \
-H "Content-Type: application/json" \
-d '{
  "tyre_details": [
    {"brand": "Michelin", "size": "205/55R16", "cost": 450.00},
    {"brand": "Michelin", "size": "205/55R16", "cost": 450.00},
    {"brand": "Michelin", "size": "205/55R16", "cost": 450.00},
    {"brand": "Michelin", "size": "205/55R16", "cost": 450.00},
    {"brand": "Michelin", "size": "205/55R16", "cost": 450.00}
  ]
}' 