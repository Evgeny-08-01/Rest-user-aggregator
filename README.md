# Rest User Agregator

[![CI/CD](https://github.com/Evgeny-08-01/Rest-user-agregator/actions/workflows/workflows.yml/badge.svg)](https://github.com/Evgeny-08-01/Rest-user-agregator/actions)

**📄 Техническое задание:** [Посмотреть ТЗ](./technical%20requirements/technical%20requirements.txt)

REST API сервис для агрегации данных онлайн подписок пользователей.

## Стек технологий

- Версия Go: 1.23
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
4. Завершает работу с кодом 
   - Код 0 — если все запросы успели завершиться.
   - Код 1 — если произошла ошибка при старте или остановке.


## Запуск

### Через Docker Compose (рекомендуется)

```bash
docker-compose up --build
```
### Управление проектом через Makefile

| Команда                    | Описание                                                    |
|----------------------------|-------------------------------------------------------------|
| `make help`                | Показать все команды                                        |
| `make test-u`              | Юнит-тесты (с моками, без БД) — быстро                      |
| `make test-int`            | Интеграционные тесты (с БД, автоматически запускает Docker) |
| `make test-all`            | Сначала юнит, потом интеграционные                          |
| `make build`               | Сборка бинарника                                            |
| `make docker-up`           | Запуск всех контейнеров                                     |
| `make docker-down`         | Остановка контейнеров                                       |
| `make docker-up-db`        | Запуск только PostgreSQL                                    |
| `make docker-logs`         | Просмотр логов всех контейнеров                             |
| `make docker-logs-server`  | Просмотр логов только сервера                               |
| `make clean`               | Очистка артефактов сборки (bin/, coverage/, cache)          |
| `make migrate-up`          | Применить все миграции                                      |
| `make migrate-down`        | Откатить все миграции                                       |
| `make migrate-down-users`  | Откатить только таблицу `users`                             |
| `make migrate-down-subs`   | Откатить только таблицу `subscriptions`                     |

Сервер будет доступен по адресу: http://localhost:8080

### Локальный запуск (без Docker)
Установите PostgreSQL и создайте базу данных subscriptions

Создайте файл .env в корне проекта (скопируйте из .env.example):

```env
DB_PATH=postgres://postgres:mysecret@localhost:5432/subscriptions?sslmode=disable
SERVER_PORT=8080
POSTGRES_PASSWORD=mysecret
POSTGRES_DB=subscriptions
LOG_LEVEL=info
```
Запустите сервер:

```bash
go run cmd/api/main.go
```

## API Endpoints

| Метод     | Endpoint                          | Описание                           |
|-----------|-----------------------------------|------------------------------------|
| POST      | `/api/subscriptions`              | Создать подписку                   |
| GET       | `/api/subscriptions/{id}`         | Получить подписку по ID            |
| PUT       | `/api/subscriptions/{id}`         | Обновить подписку                  |
| DELETE    | `/api/subscriptions/{id}`         | Удалить подписку                   |
| GET       | `/api/subscriptions`              | Список подписок (с пагинацией)     |
| GET       | `/api/subscriptions/total-cost`   | Суммарная стоимость подписок       |
| GET       | `/health`                         | Проверка работоспособности сервиса |

### Параметры фильтрации для `/api/subscriptions/total-cost`

| Параметр        | Тип             | Описание                 |
|-----------------|-----------------|--------------------------|
| `user_id`       | UUID            | ID пользователя          |
| `service_name`  | string          | Название сервиса         |
| `start_date`    | string          | Дата начала (MM-YYYY)    |
| `end_date`      | string          | Дата окончания (MM-YYYY) |

### Healthcheck
Для проверки работоспособности сервиса доступен отдельный эндпоинт:


GET /health

Ответ:

{"status":"ok"}

## Примеры запросов

### Создание подписки


POST /api/subscriptions
{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
}


Ответ:

{
    "id": 1
}

### Получение суммарной стоимости


GET /api/subscriptions/total-cost?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&start_date=01-2025&end_date=12-2025


Ответ:
```json
{
    "total": 1500
}
```
## Документация Swagger

После запуска сервера документация доступна по адресу:


http://localhost:8080/swagger/index.html
## Тестирование

Проект покрыт двумя типами тестов:

- **Юнит-тесты** (с моками) — не требуют БД, быстрые
- **Интеграционные тесты** (с реальной БД) — запускаются через Docker

```bash
# Юнит-тесты (без БД)
make test-u

# Интеграционные тесты (с БД)
make test-int

# Все тесты
make test-all
Результат: `main: ~23%, handlers: ~72%, logger: ~82%`

## Миграции

Миграции применяются автоматически при запуске сервера.

| `migrations/000001_create_subscriptions_table.up.sql`   | Создание таблицы подписок      |
| `migrations/000001_create_subscriptions_table.down.sql` | Удаление таблицы подписок      |
| `migrations/000002_create_users_table.up.sql`           | Создание таблицы пользователей |
| `migrations/000002_create_users_table.down.sql`         | Удаление таблицы пользователей |


### Откат миграций

Для отката миграций используйте флаг `-down`:

```bash
go run cmd/api/main.go -down
```

## Структура проекта

```
Rest-user-agregator/
├── .github/
│   └── workflows/
│       └── workflows.yml                  # GitHub Actions CI/CD
├── cmd/
│   └── api/
│       └── main.go                        # Точка входа
├── internal/
│   ├── authentication/                    # JWT, middleware, тесты
│   │   ├── jwt.go
│   │   ├── jwt_test.go
│   │   ├── middleware.go
│   │   └── middleware_test.go
│   ├── database/                          # Инициализация БД + CRUDL + PostgresRepo
│   │   ├── database.go
│   │   ├── database_CRUDL_func.go
│   │   ├── migrations.go
│   │   └── user_repo.go
│   ├── handlers/                          # HTTP хэндлеры + хелперы + тесты
│   │   ├── auth.go
│   │   ├── handlers_api.go
│   │   ├── handlers_integration_test.go
│   │   ├── handlers_unit_test.go
│   │   └── helpers.go
│   ├── metrics/                           # Prometheus метрики
│   │   └── metrics.go
│   ├── middleware/                        # CORS и метрики middleware
│   │   ├── cors.go
│   │   └── metrics.go
│   ├── models/                            # Модели данных
│   │   ├── subscriptions.go
│   │   └── user.go
│   ├── repository/                        # Интерфейсы репозиториев + моки
│   │   ├── interface.go
│   │   └── mock.go
│   └── service/                           # Бизнес-логика
│       ├── auth_service.go
│       ├── auth_service_test.go
│       └── subscription_service.go
├── migrations/                            # SQL миграции
│   ├── 000001_create_subscriptions_table.up.sql
│   ├── 000001_create_subscriptions_table.down.sql
│   ├── 000002_create_users_table.up.sql
│   └── 000002_create_users_table.down.sql
├── web/                                   # Фронтенд (HTML + CSS + JS)
│   ├── css/
│   │   └── style.css
│   ├── js/
│   │   ├── api.js
│   │   ├── app.js
│   │   ├── auth.js
│   │   ├── components.js
│   │   └── utils.js
│   └── index.html
├── docs/                                  # Swagger документация
│   ├── docs.go
│   ├── swagger.json
│   └── swagger.yaml
├── screenshots/                           # Скриншоты Grafana для README
│   ├── grafana_errors.png
│   ├── grafana_p95.png
│   └── grafana_rps.png
├── technical_requirements/                # Техническое задание
│   └── technical requirements.txt
├── pkg/
│   └── logger/                            # Логирование с уровнями
│       ├── logger.go
│       └── logger_test.go
├── compose.yaml                           # Docker Compose
├── Makefile                               # Управление проектом
├── Dockerfile                             # Docker образ
├── prometheus.yml                         # Конфиг Prometheus
├── .env.example                           # Пример конфигурации
├── .env.test                              # Тестовое окружение
├── .gitignore                             # Игнорируемые файлы
├── .dockerignore                          # Игнорируемые файлы для Docker
├── go.mod                                 # Зависимости
├── go.sum                                 # Контрольные суммы зависимостей
└── README.md                              # Документация проекта
```
## Переменные окружения

| Переменная         | Описание                 | Значение по умолчанию                                                |
|--------------------|--------------------------|----------------------------------------------------------------------|
| `DB_PATH`          | Подключение к PostgreSQL | `postgres://postgres:mysecret@db:5432/subscriptions?sslmode=disable` |
| `SERVER_PORT`      | Порт сервера             | `8080`                                                               |
| `POSTGRES_PASSWORD`| Пароль PostgreSQL        | `mysecret`                                                           |
| `POSTGRES_DB`      | Имя базы данных          | `subscriptions`                                                      |
| `LOG_LEVEL`        | Уровень логирования      | `info`                                                               |
| `LOG_PATH`         | Путь к файлу логов       | `/var/log/app/app.log`                                               |
| `POSTGRES_USER`    | Пользователь PostgreSQL  | `postgres`                                                           |
## Архитектура

Проект построен на принципах **чистой архитектуры** и разделён на 5 слоёв:

| Слой                               | Папка                  | Ответственность                                                        |
|------------------------------------|------------------------|------------------------------------------------------------------------|
| **1. Презентационный слой (HTTP)** | `internal/handlers/`   | Принимает HTTP-запросы, парсит JSON, вызывает сервис, возвращает ответ |
| **2. Бизнес-слой (Service)**       | `internal/service/`    | Содержит бизнес-логику: парсинг дат, валидацию, расчёты                |
| **3. Интерфейс репозитория**       | `internal/repository/` | Определяет контракт для работы с БД (позволяет подменять реализацию)   |
| **4. Слой данных (Repository)**    | `internal/database/`   | Реализует интерфейс репозитория, выполняет SQL-запросы                 |
| **5. База данных**                 | PostgreSQL             | Хранит данные                                                          |

**Цепочка вызовов:**

```text
HTTP-запрос
    ↓
handlers (парсинг JSON)
    ↓
service (бизнес-логика)
    ↓
repository (SQL-запросы)
    ↓
PostgreSQL (хранилище)
```
Принципы, соблюдённые в проекте:

Инкапсуляция — БД приватная (db), доступ только через методы

Интерфейсы — SubscriptionRepository отделяет бизнес-логику от работы с БД

Внедрение зависимостей — хендлеры и сервис получают зависимости через конструкторы

Слабая связность — легко подменить реализацию БД или мокировать в тестах
## Возможные ошибки и их решение

| Ошибка                                 | Решение                                                         |
|----------------------------------------|-----------------------------------------------------------------|
| `user_id` must always be a valid UUID  | Проверьте, что переданный user_id соответствует формату UUID    |
| `start_date` must be in format MM-YYYY | Используйте формат: месяц (01-12) и год (1900-2100) через дефис |
| `price` must not be negative           | Цена подписки должна быть ≥ 0                                   |
| Database error                         | Проверьте подключение к PostgreSQL и выполнение миграций        |
## CI/CD (GitHub Actions)

Проект использует GitHub Actions для автоматического тестирования, проверки качества кода и публикации Docker-образов.

### Workflow

Пайплайн состоит из трёх последовательных джобов:

| Джоб          | Описание                                 | Условие запуска                               |
|---------------|------------------------------------------|-----------------------------------------------|
| Lint          | Проверка качества кода (`golangci-lint`) | Всегда                                        |
| Test          | Сборка и запуск тестов с PostgreSQL      | После успешного Lint                          |
| Publish       | Публикация Docker-образа в Docker Hub    |Только при создании тега `v*` и                |
                |                                          | успешном прохождении всех предыдущих джобов   |

### Что проверяет линтер

`golangci-lint` запускается со стандартным набором линтеров, включённых по умолчанию:

| Линтер                | Что проверяет                       |
|-----------------------|-------------------------------------|
| `errcheck`            | Необработанные ошибки               |
| `govet`               | Подозрительные конструкции          |
| `staticcheck`         | Ошибки в коде                       |
| `unused`              | Неиспользуемые переменные и функции |
| `gosimple`            | Упрощение кода                      |
| `ineffassign`         | Бесполезные присваивания            |


Секреты для GitHub Actions
Для публикации в Docker Hub в настройках репозитория должны быть установлены следующие секреты:

Секрет	Описание
DOCKER_USERNAME	Имя пользователя Docker Hub
DOCKER_ACCESS_TOKEN	Токен доступа к Docker Hub
Настройка: Settings → Secrets and variables → Actions