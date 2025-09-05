ACCOUNT_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number": "132123123123123"}')

echo "Account created: $ACCOUNT_RESPONSE"

ACCOUNT_ID=$(echo $ACCOUNT_RESPONSE | jq -r '.account_id')

curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-uuuuuu-Key: purchase-$(date +%s)" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"operation_type_id\": 1,
    \"amount\": -50.75
  }"