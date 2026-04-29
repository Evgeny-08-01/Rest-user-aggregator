@echo off
setlocal enabledelayedexpansion

:: Colors for output (optional, works in some terminals)
set "GREEN=[92m"
set "RED=[91m"
set "NC=[0m"

echo ========================================
echo Testing Subscription API
echo ========================================
echo.

:: Store created ID
set "USER_ID=60601fee-2bf1-4721-ae6f-7636e79a0cba"
set "BASE_URL=http://localhost:8080/api"

:: 1. CREATE TESTS
echo [1] CREATE SUBSCRIPTION TESTS
echo ------------------------------

echo 1.1 Success
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"Yandex Plus\",\"price\":400,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}" > temp.json
set /p created=<temp.json
echo Response: %created%
for /f "tokens=2 delims=:{}" %%a in ("%created%") do set "SUB_ID=%%a"
echo.

echo 1.2 Empty service_name
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"\",\"price\":400,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}"
echo.

echo 1.3 Zero price
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"Zero Price\",\"price\":0,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}"
echo.

echo 1.4 Negative price
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"Negative\",\"price\":-100,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}"
echo.

echo 1.5 Empty user_id
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"No User\",\"price\":100,\"user_id\":\"\",\"start_date\":\"07-2025\"}"
echo.

echo 1.6 Invalid date format
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"Wrong Date\",\"price\":100,\"user_id\":\"%USER_ID%\",\"start_date\":\"2025-07\"}"
echo.

echo 1.7 Invalid JSON
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\": \"Bad JSON\", \"price\": 100, \"user_id\": \"%USER_ID%\", \"start_date\": \"07-2025\""
echo.

:: 2. READ TESTS
echo.
echo [2] READ SUBSCRIPTION TESTS
echo ---------------------------

echo 2.1 Get subscription by ID (using ID from creation)
if defined SUB_ID (
    curl -s "%BASE_URL%/subscriptions/%SUB_ID%"
) else (
    echo No subscription ID available
)
echo.

echo 2.2 Get non-existent ID
curl -s "%BASE_URL%/subscriptions/99999"
echo.

echo 2.3 Get invalid ID (letters)
curl -s "%BASE_URL%/subscriptions/abc"
echo.

:: 3. UPDATE TESTS
echo.
echo [3] UPDATE SUBSCRIPTION TESTS
echo -----------------------------

echo 3.1 Update existing subscription
if defined SUB_ID (
    curl -s -X PUT "%BASE_URL%/subscriptions/%SUB_ID%" -H "Content-Type: application/json" -d "{\"service_name\":\"Yandex Plus Updated\",\"price\":500,\"user_id\":\"%USER_ID%\",\"start_date\":\"08-2025\",\"end_date\":\"12-2025\"}"
) else (
    echo No subscription ID available
)
echo.

echo 3.2 Update non-existent ID
curl -s -X PUT "%BASE_URL%/subscriptions/99999" -H "Content-Type: application/json" -d "{\"service_name\":\"Test\",\"price\":100,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}"
echo.

echo 3.3 Update invalid ID (letters)
curl -s -X PUT "%BASE_URL%/subscriptions/abc" -H "Content-Type: application/json" -d "{\"service_name\":\"Test\",\"price\":100,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}"
echo.

:: 4. DELETE TESTS
echo.
echo [4] DELETE SUBSCRIPTION TESTS
echo -----------------------------

echo 4.1 Delete existing subscription
if defined SUB_ID (
    curl -s -X DELETE "%BASE_URL%/subscriptions/%SUB_ID%"
) else (
    echo No subscription ID available
)
echo.

echo 4.2 Delete already deleted (or non-existent)
curl -s -X DELETE "%BASE_URL%/subscriptions/99999"
echo.

echo 4.3 Delete invalid ID (letters)
curl -s -X DELETE "%BASE_URL%/subscriptions/abc"
echo.

:: 5. LIST TESTS
echo.
echo [5] LIST SUBSCRIPTIONS TESTS
echo ----------------------------

echo 5.1 Get list (without params)
curl -s "%BASE_URL%/subscriptions"
echo.

echo 5.2 Get list with limit=5, offset=0
curl -s "%BASE_URL%/subscriptions?limit=5&offset=0"
echo.

echo 5.3 Get list with offset greater than count
curl -s "%BASE_URL%/subscriptions?limit=10&offset=100"
echo.

echo 5.4 Get list with invalid limit
curl -s "%BASE_URL%/subscriptions?limit=abc"
echo.

:: 6. TOTAL COST TESTS
echo.
echo [6] TOTAL COST TESTS
echo --------------------

echo 6.1 Total cost without filters
curl -s "%BASE_URL%/subscriptions/total-cost"
echo.

echo 6.2 Total cost with user_id filter
curl -s "%BASE_URL%/subscriptions/total-cost?user_id=%USER_ID%"
echo.

echo 6.3 Total cost with service_name filter
curl -s "%BASE_URL%/subscriptions/total-cost?service_name=Yandex%20Plus"
echo.

echo 6.4 Total cost with date period
curl -s "%BASE_URL%/subscriptions/total-cost?start_date=01-2025&end_date=12-2025"
echo.

echo 6.5 Total cost with user_id + period
curl -s "%BASE_URL%/subscriptions/total-cost?user_id=%USER_ID%&start_date=01-2025&end_date=12-2025"
echo.

echo 6.6 Total cost with non-existent user (should return 0)
curl -s "%BASE_URL%/subscriptions/total-cost?user_id=00000000-0000-0000-0000-000000000000"
echo.

:: 7. ADDITIONAL BOUNDARY TESTS
echo.
echo [7] ADDITIONAL BOUNDARY TESTS
echo -----------------------------

echo 7.1 Create with very long service_name (1000 chars)
set "long_name="
for /l %%i in (1,1,100) do set "long_name=!long_name!aaaaaaaaaa"
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"!long_name!\",\"price\":100,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}"
echo.

echo 7.2 Create subscription with end_date (optional)
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"Limited\",\"price\":300,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\",\"end_date\":\"12-2025\"}"
echo.

echo 7.3 Create duplicate (should succeed - no unique constraints)
curl -s -X POST "%BASE_URL%/subscriptions" -H "Content-Type: application/json" -d "{\"service_name\":\"Duplicate\",\"price\":200,\"user_id\":\"%USER_ID%\",\"start_date\":\"07-2025\"}"
echo.

:: Cleanup temp file
del temp.json 2>nul

echo.
echo ========================================
echo Testing completed
echo ========================================
pause