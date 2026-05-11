# Rest User Aggregator

REST API сервис для агрегации данных онлайн подписок пользователей.

## Стек технологий

- Версия Go: 1.25
- Версия PostgreSQL: 15-alpine
- Docker / Docker Compose
- Swagger
- Логирование с уровнями (DEBUG/INFO/WARN/ERROR/FATAL)
- Graceful shutdown
- Интерфейсы для репозитория

## Функциональность

- CRUDL операции с подписками
- Подсчёт суммарной стоимости подписок за период с фильтрацией:
  - по ID пользователя 
  - по названию сервиса
- Валидация входных данных:
  - UUID пользователя (наличие обязательно, диагностируется ошибка в базе данных)
  - Дата в формате MM-YYYY
  - Цена подписки ≥ 0

## Логирование

Поддерживаются уровни логирования:
- `DEBUG` — для отладки (не используется в продакшене)
- `INFO` — нормальные события (запуск, остановка, HTTP запросы)
- `WARN` — проблемы, не требующие остановки
- `ERROR` — сбои, требующие внимания
- `FATAL` — критические ошибки, сервер падает

Уровень задаётся переменной `LOG_LEVEL` в `.env`

## Graceful Shutdown

При получении сигналов SIGINT (Ctrl+C) или SIGTERM сервер:
1. Перестаёт принимать новые соединения
2. Завершает обработку текущих запросов
3. Закрывает соединение с БД
4. Завершает работу с кодом 0

## Запуск

### Через Docker Compose (рекомендуется)

```bash
docker-compose up --build
Сервер будет доступен по адресу: http://localhost:8080

Локальный запуск (без Docker)
Установите PostgreSQL и создайте базу данных subscriptions

Создайте файл .env в корне проекта (скопируйте из .env.example):

env
DB_PATH=postgres://postgres:mysecret@localhost:5432/subscriptions?sslmode=disable
SERVER_PORT=8080
POSTGRES_PASSWORD=mysecret
POSTGRES_DB=subscriptions
LOG_LEVEL=info
Запустите сервер:

bash
go run cmd/api/main.go
API Endpoints
Метод	Endpoint	Описание
POST	/api/subscriptions	Создать подписку
GET	/api/subscriptions/{id}	Получить подписку по ID
PUT	/api/subscriptions/{id}	Обновить подписку
DELETE	/api/subscriptions/{id}	Удалить подписку
GET	/api/subscriptions	Список подписок (с пагинацией)
GET	/api/subscriptions/total-cost	Суммарная стоимость подписок
Параметры фильтрации для /api/subscriptions/total-cost
Параметр	Тип	Описание
user_id	UUID	ID пользователя
service_name	string	Название сервиса
start_date	string	Дата начала (MM-YYYY)
end_date	string	Дата окончания (MM-YYYY)
Примеры запросов
Создание подписки
json
POST /api/subscriptions
{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
}
Ответ:

json
{
    "id": 1
}
Получение суммарной стоимости
json
GET /api/subscriptions/total-cost?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&start_date=01-2025&end_date=12-2025

{
    "total": 1500
}
Документация Swagger
После запуска сервера документация доступна по адресу:

text
http://localhost:8080/swagger/index.html
Тестирование
bash
# Запуск всех тестов
go test ./... -v

# Запуск тестов с покрытием
go test ./... -cover

# Результат: handlers: ~68%, logger: ~82%
Миграции
Миграции применяются автоматически при запуске сервера.

Up migration: migrations/000001_create_subscriptions_table.up.sql

Down migration: migrations/000001_create_subscriptions_table.down.sql

Откат миграций
Для отката миграций используйте флаг -down:

bash
go run cmd/api/main.go -down
Структура проекта
text
Rest-user-aggregator/
├── cmd/api/                 # Точка входа
├── internal/
│   ├── database/            # Инициализация БД + CRUDL + PostgresRepo
│   ├── handlers/            # HTTP хэндлеры + хелперы + тесты
│   ├── models/              # Модели данных
│   ├── repository/          # Интерфейс репозитория
├── migrations/              # SQL миграции
├── docs/                    # Swagger документация
├── pkg/logger/              # Логирование с уровнями
├── compose.yaml             # Docker Compose
├── .env.example             # Пример конфигурации
└── go.mod                   # Зависимости
Переменные окружения
Переменная	                      Описание	                              Значение по умолчанию
DB_PATH                          	Строка подключения к PostgreSQL	        postgres://postgres:mysecret@db:5432/subscriptions?sslmode=disable
SERVER_PORT                      	Порт сервера	                          8080
POSTGRES_PASSWORD                	Пароль PostgreSQL	                      mysecret
POSTGRES_DB	                      Имя базы данных	                        subscriptions
LOG_LEVEL	                        Уровень логирования	                    info
Архитектура
Проект                            построен на принципах:

Инкапсуляция — БД приватная (db), доступ только через методы

Интерфейсы — SubscriptionRepository отделяет бизнес-логику от работы с БД

Внедрение зависимостей — хендлеры получают репозиторий через конструктор

Слабая связность — легко подменить реализацию БД или мокировать в тестах

Возможные              ошибки                              и их решение
Ошибка	Решение
user_id                must always be a valid UUID	       Проверьте, что переданный user_id соответствует формату UUID
start_date             must be in format MM-YYYY	         Используйте формат: месяц (01-12) и год (1900-2100) через дефис
price                  must not be negative	               Цена подписки должна быть ≥ 0
Database               error	                             Проверьте подключение к PostgreSQL и выполнение миграций