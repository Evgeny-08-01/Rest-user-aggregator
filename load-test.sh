#!/bin/bash

# 1. Переменные
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiMGEyNDA1NjUtYWYzOC00OWVjLThkYmItODIxNWQzMDExMDE1IiwiZW1haWwiOiJhZG1pbjFAbWFpbC5jb20iLCJyb2xlIjoiYWRtaW4iLCJleHAiOjE3ODY1Mzg5NjgsImlhdCI6MTc4NjQ1MjU2OH0.JE7oZvPlP5HhrLFmpU9DPW4QiVny4J_soHJYC-QqeZA"
mkdir -p load-tests

# 2. GET
echo "=== GET /api/subscriptions ==="
echo "GET http://localhost:8087/api/subscriptions" | vegeta attack -duration=30s -rate=200 -header "Authorization: Bearer $TOKEN" > load-tests/get_200.bin
vegeta report load-tests/get_200.bin > load-tests/get_200.txt
vegeta report -type=text load-tests/get_200.bin

# 3. POST
echo "=== POST /api/subscriptions ==="
echo "POST http://localhost:8087/api/subscriptions" | vegeta attack -duration=30s -rate=100 -header "Authorization: Bearer $TOKEN" -header "Content-Type: application/json" -body '{"service_name":"LoadTest","price":100,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}' > load-tests/post_100.bin
vegeta report load-tests/post_100.bin > load-tests/post_100.txt
vegeta report -type=text load-tests/post_100.bin

# 4. PUT
echo "=== PUT /api/subscriptions/1 ==="
echo "PUT http://localhost:8087/api/subscriptions/1" | vegeta attack -duration=30s -rate=100 -header "Authorization: Bearer $TOKEN" -header "Content-Type: application/json" -body '{"service_name":"LoadTestUpdated","price":150,"user_id":"550e8400-e29b-41d4-a716-446655440000","start_date":"07-2025"}' > load-tests/put_100.bin
vegeta report load-tests/put_100.bin > load-tests/put_100.txt
vegeta report -type=text load-tests/put_100.bin

# 5. DELETE
echo "=== DELETE /api/subscriptions/1 ==="
echo "DELETE http://localhost:8087/api/subscriptions/1" | vegeta attack -duration=30s -rate=100 -header "Authorization: Bearer $TOKEN" > load-tests/delete_100.bin
vegeta report load-tests/delete_100.bin > load-tests/delete_100.txt
vegeta report -type=text load-tests/delete_100.bin

# 6. Все вместе
echo "=== ALL ENDPOINTS TOGETHER ==="
cat > load-tests/targets.txt << 'EOF'
GET http://localhost:8087/api/subscriptions
GET http://localhost:8087/health
GET http://localhost:8087/api/config
POST http://localhost:8087/api/subscriptions
PUT http://localhost:8087/api/subscriptions/1
DELETE http://localhost:8087/api/subscriptions/1
EOF
vegeta attack -duration=30s -rate=100 -header "Authorization: Bearer $TOKEN" -targets=load-tests/targets.txt > load-tests/all_100.bin
vegeta report load-tests/all_100.bin > load-tests/all_100.txt
vegeta report -type=text load-tests/all_100.bin