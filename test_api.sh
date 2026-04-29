#!/bin/bash
URL="http://localhost:8080/api"
USER="60601fee-2bf1-4721-ae6f-7636e79a0cba"

echo "1. Создаём подписку"
cat > /tmp/sub1.json <<EOF
{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "$USER",
  "start_date": "07-2025"
}
EOF
curl -X POST "$URL/subscriptions" -H "Content-Type: application/json" -d @/tmp/sub1.json

echo -e "\n2. Проверяем созданную подписку"
curl "$URL/subscriptions/1"

echo -e "\n3. Список всех подписок"
curl "$URL/subscriptions"

echo -e "\n4. Обновляем подписку"
cat > /tmp/sub2.json <<EOF
{
  "service_name": "Яндекс Плюс",
  "price": 500,
  "user_id": "$USER",
  "start_date": "08-2025"
}
EOF
curl -X PUT "$URL/subscriptions/1" -H "Content-Type: application/json" -d @/tmp/sub2.json

echo -e "\n5. Удаляем подписку"
curl -X DELETE "$URL/subscriptions/1"

echo -e "\n6. Считаем сумму за год"
curl "$URL/subscriptions/total-cost?user_id=$USER&start_date=01-2025&end_date=12-2025"

echo -e "\n7. Проверяем ошибку — несуществующий ID"
curl "$URL/subscriptions/99999"

echo -e "\n8. Проверяем ошибку — неверный формат даты"
cat > /tmp/sub3.json <<EOF
{
  "service_name": "Test",
  "price": 100,
  "user_id": "$USER",
  "start_date": "2025.07"
}
EOF
curl -X POST "$URL/subscriptions" -H "Content-Type: application/json" -d @/tmp/sub3.json

echo -e "\n9. Проверяем ошибку — пустое название"
cat > /tmp/sub4.json <<EOF
{
  "service_name": "",
  "price": 100,
  "user_id": "$USER",
  "start_date": "07-2025"
}
EOF
curl -X POST "$URL/subscriptions" -H "Content-Type: application/json" -d @/tmp/sub4.json

echo -e "\n10. Проверяем ошибку — отрицательная цена"
cat > /tmp/sub5.json <<EOF
{
  "service_name": "Test",
  "price": -50,
  "user_id": "$USER",
  "start_date": "07-2025"
}
EOF
curl -X POST "$URL/subscriptions" -H "Content-Type: application/json" -d @/tmp/sub5.json

echo ""
rm -f /tmp/sub1.json /tmp/sub2.json /tmp/sub3.json /tmp/sub4.json /tmp/sub5.json 2>/dev/null