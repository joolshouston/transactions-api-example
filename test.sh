
ACCOUNT_RESPONSE=$(curl -s -X POST http://localhost:8080/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"document_number": "88888"}')

echo "Account created: $ACCOUNT_RESPONSE"

ACCOUNT_ID=$(echo $ACCOUNT_RESPONSE | jq -r '.account_id')

# Create three debt transactions (operation type 1 - purchases)
echo "Creating first purchase transaction..."
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-idempotency-Key: purchase-$(date +%s)" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"operation_type_id\": 1,
    \"amount\": -50.0
  }"

sleep 1

echo "Creating second purchase transaction..."
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-idempotency-Key: purchase-$(date +%s)" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"operation_type_id\": 1,
    \"amount\": -23.5
  }"

sleep 1

echo "Creating third purchase transaction..."
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-idempotency-Key: purchase-$(date +%s)" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"operation_type_id\": 1,
    \"amount\": -18.7
  }"

sleep 1

echo "Creating payment transaction to discharge debts..."
curl -X POST http://localhost:8080/v1/transactions \
  -H "Content-Type: application/json" \
  -H "X-idempotency-Key: payment-$(date +%s)" \
  -d "{
    \"account_id\": \"$ACCOUNT_ID\",
    \"operation_type_id\": 4,
    \"amount\": 60.0
  }"

  sleep 1

  echo "Creating payment transaction to discharge debts..."
  curl -X POST http://localhost:8080/v1/transactions \
    -H "Content-Type: application/json" \
    -H "X-idempotency-Key: payment-$(date +%s)" \
    -d "{
      \"account_id\": \"$ACCOUNT_ID\",
      \"operation_type_id\": 4,
      \"amount\": 100.0
    }"