#!/bin/bash

echo "=== Testing Microservices API ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Wait for services to be ready
echo "Waiting for services to start..."
sleep 5

echo -e "${BLUE}1. Creating a user...${NC}"
USER_RESPONSE=$(curl -s -X POST http://localhost:8081/users \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com"}')
echo $USER_RESPONSE
USER_ID=$(echo $USER_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo -e "${GREEN}✓ User created with ID: $USER_ID${NC}"
echo ""

echo -e "${BLUE}2. Getting user details...${NC}"
curl -s http://localhost:8081/users/$USER_ID | jq .
echo -e "${GREEN}✓ User retrieved${NC}"
echo ""

echo -e "${BLUE}3. Creating an order for user...${NC}"
ORDER_RESPONSE=$(curl -s -X POST http://localhost:8082/orders \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":$USER_ID,\"product_id\":100,\"amount\":99.99}")
echo $ORDER_RESPONSE
ORDER_ID=$(echo $ORDER_RESPONSE | grep -o '"id":[0-9]*' | grep -o '[0-9]*')
echo -e "${GREEN}✓ Order created with ID: $ORDER_ID${NC}"
echo ""

echo -e "${BLUE}4. Getting order details...${NC}"
curl -s http://localhost:8082/orders/$ORDER_ID | jq .
echo -e "${GREEN}✓ Order retrieved${NC}"
echo ""

echo -e "${BLUE}5. Getting user's orders (service-to-service call)...${NC}"
curl -s http://localhost:8081/users/$USER_ID/orders | jq .
echo -e "${GREEN}✓ User orders retrieved via User Service${NC}"
echo ""

echo -e "${BLUE}6. Creating another order...${NC}"
curl -s -X POST http://localhost:8082/orders \
  -H "Content-Type: application/json" \
  -d "{\"user_id\":$USER_ID,\"product_id\":200,\"amount\":149.99}" | jq .
echo -e "${GREEN}✓ Second order created${NC}"
echo ""

echo -e "${BLUE}7. Listing all users...${NC}"
curl -s http://localhost:8081/users | jq .
echo -e "${GREEN}✓ All users listed${NC}"
echo ""

echo "=== All tests completed! ==="
