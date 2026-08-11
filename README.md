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
- JWT (авторизация)
- bcrypt (хеширование паролей)
- Middleware (CORS, логирование, авторизация)
- Тестирование (юнит-тесты + интеграционные)
- GitHub Actions (CI/CD)
- Prometheus + Grafana (мониторинг)
- vegeta (нагрузочное тестирование)
- HTML + CSS + JS (чистый) фронтенд (демо)

# Функциональность

- CRUDL операции с подписками
- Подсчёт суммарной стоимости подписок за период с фильтрацией:
  - по ID пользователя 
  - по названию сервиса
- Валидация входных данных:
  - UUID пользователя (наличие обязательно, диагностируется ошибка в базе данных)
  - Дата в формате MM-YYYY
  - Цена подписки ≥ 0
- Регистрация и авторизация пользователей (JWT)
- Роли: пользователь и администратор
- Защита API-эндпоинтов через middleware
- Фронтенд на чистом JS (HTML + CSS)
- Динамическая подгрузка конфигурации через `/api/config`

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

Сервер будет доступен по адресу: http://localhost:8087

### Локальный запуск (без Docker)

Установите PostgreSQL и создайте базу данных `subscriptions`.

Создайте файл `.env` в корне проекта (скопируйте из `.env.example`):

```env
DB_PATH=postgres://postgres:mysecret@localhost:5432/subscriptions?sslmode=disable
SERVER_PORT=8087
POSTGRES_PASSWORD=mysecret
POSTGRES_DB=subscriptions
LOG_LEVEL=info
LOG_PATH=./logs/app.log
JWT_SECRET=your-secret-key-here
POSTGRES_USER=postgres
FRONTEND_URL=http://localhost:8087
```
Запустите сервер:
make run
Или вручную:
```bash
go run cmd/api/main.go
```

## API Endpoints

| Метод   | Эндпоинт                        | Защита | Описание                             |
|---------|---------------------------------|--------|--------------------------------------|
| POST    | `/api/register`                 | ❌ Нет | Регистрация пользователя             |
| POST    | `/api/login`                    | ❌ Нет | Вход, получение JWT                  |
| POST    | `/api/subscriptions`            | ✅ Да  | Создать подписку                     |
| GET     | `/api/subscriptions`            | ✅ Да  | Список подписок                      |
| GET     | `/api/subscriptions/{id}`       | ✅ Да  | Получить подписку по ID              |
| PUT     | `/api/subscriptions/{id}`       | ✅ Да  | Обновить подписку                    |
| DELETE  | `/api/subscriptions/{id}`       | ✅ Да  | Удалить подписку                     |
| GET     | `/api/subscriptions/total-cost` | ✅ Да  | Суммарная стоимость                  |
| GET     | `/health`                       | ❌ Нет | Проверка работоспособности           |
| GET     | `/api/config`                   | ❌ Нет | Адрес бэкенда для фронтенда          |

## 🔐 Авторизация

Реализована JWT-авторизация с ролями:
- `user` — обычный пользователь
- `admin` — администратор (создаётся только через БД)

### Заголовок для защищённых запросов:
```
Authorization: Bearer <jwt_token>
```

### Эндпоинты авторизации:

| Метод | Эндпоинт        | Описание                        |
|-------|-----------------|---------------------------------|
| POST  | `/api/register` | Регистрация нового пользователя |
| POST  | `/api/login`    | Вход, получение JWT-токена      |

### Пример регистрации:
```json
{
    "email": "user@example.com",
    "password": "123456",
    "role": "user"
}
```

### Пример логина:
```json
{
    "email": "user@example.com",
    "password": "123456"
}
```

### Ответ логина:
```json
{
    "token": "eyJhbGciOiJIUzI1NiIs...",
    "email": "user@example.com",
    "role": "user"
}
```
## 🖥️ Фронтенд

- Интерфейс на чистом JS (HTML + CSS)
- Раздаётся бэкендом через `/css/`, `/js/`, `/`
- Эндпоинт `/api/config` для динамического получения адреса бэкенда
- Включает: регистрацию, логин, CRUDL подписок, фильтры, пагинацию, total-cost

## 🌐 CORS

- **Разработка:** разрешён `http://localhost:8087` (порт из `SERVER_PORT`)
- **Продакшен:** если фронтенд и бэкенд на одном домене — CORS не требуется.  
  При разделении — добавить `FRONTEND_URL` в `.env`.

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

## Нагрузочное тестирование

### Инструменты
- **vegeta** — генерация нагрузки
- **Prometheus** — сбор метрик
- **Grafana** — визуализация

### Установка vegeta

Для запуска нагрузочного тестирования потребуется утилита **vegeta**:

**macOS / Linux:**
```bash
brew install vegeta
```

**Windows (Git Bash):**
```bash
curl -L -o vegeta.zip https://github.com/tsenart/vegeta/releases/download/v12.11.1/vegeta_12.11.1_windows_amd64.zip
unzip vegeta.zip
mv vegeta.exe ~/bin/
```

**Проверка установки:**
```bash
vegeta -version
```

### Запуск нагрузочного тестирования

Пример команды для GET-запроса:

```bash
echo "GET http://localhost:8087/api/subscriptions" | vegeta attack -duration=30s -rate=50 -header "Authorization: Bearer $TOKEN" | vegeta report
```

### Сценарий тестирования
- Длительность каждого теста: **30 секунд**
- Тестирование проводилось на **локальной машине (Windows 10)** через Docker-контейнер с сервисом
- Использовался **один и тот же JWT-токен** для всех запросов
- Тесты запускались **последовательно**, без перезапуска сервера между ними

### Тестируемые эндпоинты

| Метод  | Эндпоинт                  | Описание                         |
|--------|---------------------------|----------------------------------|
| GET    | `/api/subscriptions`      | Получение списка всех подписок   |
| GET    | `/api/subscriptions/{id}` | Получение подписки по ID         |
| POST   | `/api/subscriptions`      | Создание новой подписки          |
| PUT    | `/api/subscriptions/{id}` | Обновление существующей подписки |
| DELETE | `/api/subscriptions/{id}` | Удаление подписки                |

### Результаты

| Нагрузка (RPS)            | Успешность | Средняя задержка | p95      | Throughput  | Статус                          |
|---------------------------|------------|------------------|----------|-------------|---------------------------------|
| **50**                    | 100%       | 3.49 ms          | 4.21 ms  | 50.02 RPS   | ✅ Стабильно                    |
| **100**                   | 100%       | 3.34 ms          | 4.47 ms  | 100.04 RPS  | ✅ Стабильно                    |
| **150**                   | 100%       | 4.16 ms          | 5.57 ms  | 149.97 RPS  | ✅ Стабильно                    |
| **200**                   | 79.43%     | 5.74 s           | 30.01 s  | 79.43 RPS   | ⚠️ Частичное падение            |
| **300**                   | 10.67%     | 25.45 s          | 30.01 s  | 16.01 RPS   | ❌ Падение (таймауты, EOF, 500) |
| **POST/PUT/DELETE (100)** | 0%         | 13.62 ms         | 26.19 ms | 0.00 RPS    | ❌ Падение (EOF)                |

### Анализ методов записи

| Метод                     | Нагрузка (RPS)   | Успешность | Статус               |
|---------------------------|------------------|------------|----------------------|
| **GET**                   | 150              | 100%       | ✅ Стабильно         |
| **GET**                   | 200              | 79.43%     | ⚠️ Частичное падение |
| **POST / PUT / DELETE**   | 100              | 0%         | ❌ Падают            |

### Выводы

- **GET-запросы** стабильны до **150 RPS** включительно.
- При **200 RPS** начинаются ошибки — сервер перегружается.
- При **300 RPS** — почти полный отказ.
- **Методы записи (POST, PUT, DELETE)** — падают при 100 RPS.
- Тесты проводились **без перезапуска сервера**, что показывает реальное поведение системы при длительной работе.
- Рекомендации: оптимизация SQL-запросов, индексы в БД, внедрение кеширования (Redis).

## Мониторинг (Prometheus + Grafana)

### Стек
- **Prometheus** — сбор и хранение метрик
- **Grafana** — визуализация и дашборды

### Доступные дашборды

В Grafana настроен дашборд **HTTP Monitoring** со следующими панелями:

| Панель                                       | Описание | Пример запроса                                                                         |
|----------------------------------------------|----------|----------------------------------------------------------------------------------------|
| **RPS (запросы в секунду)**                  | Количество запросов, обрабатываемых сервером | `rate(http_requests_total[1m])`                    |
| **p95 задержка по эндпоинтам**               | 95-й процентиль времени ответа               | `histogram_quantile(0.95, sum by(le, method, path) (rate                                                                     (http_request_duration_seconds_bucket[$__rate_interval])))` |
| **Ошибки (4xx, 5xx, Internal Server Error)** | Количество ошибочных запросов                | `rate(http_requests_total{status!="OK"}[1m])`      |

### Скриншоты

![RPS](screenshots/grafana_rps.png)
![p95](screenshots/grafana_p95.png)
![Errors](screenshots/grafana_errors.png)

> 💡 *Скриншоты находятся в папке `screenshots/`*

## Документация Swagger

После запуска сервера документация доступна по адресу:


http://localhost:8087/swagger/index.html
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

Проект покрыт юнит- и интеграционными тестами.

**Общее покрытие: ~65%**

| Пакет            | Покрытие |
|------------------|----------|
| `authentication` | 95.9%    |
| `middleware`     | 95.5%    |
| `logger`         | 82.1%    |
| `service`        | 77.7%    |
| `handlers`       | 62.4%    |
| `database`       | 47.6%    |
| `cmd/api`        | 21.4%    |

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

| Переменная         | Описание                      | Значение по умолчанию                                                |
|--------------------|-------------------------------|----------------------------------------------------------------------|
| `DB_PATH`          | Подключение к PostgreSQL      | `postgres://postgres:mysecret@db:5432/subscriptions?sslmode=disable` |
| `SERVER_PORT`      | Порт сервера                  | `8087`                                                               |
| `POSTGRES_PASSWORD`| Пароль PostgreSQL             | `mysecret`                                                           |
| `POSTGRES_DB`      | Имя базы данных               | `subscriptions`                                                      |
| `LOG_LEVEL`        | Уровень логирования           | `info`                                                               |
| `LOG_PATH`         | Путь к файлу логов            | `/var/log/app/app.log`                                               |
| `POSTGRES_USER`    | Пользователь PostgreSQL       | `postgres`                                                           |
| `JWT_SECRET`       | Секрет для подписи JWT        | Обязательно задать в `.env`                                          |
| `FRONTEND_URL`     | Разрешённый источник для CORS | `http://localhost:8087`                                              |

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

- Инкапсуляция — БД приватная (db), доступ только через методы
- Интерфейсы — SubscriptionRepository отделяет бизнес-логику от работы с БД
- Внедрение зависимостей — хендлеры и сервис получают зависимости через конструкторы
- Слабая связность — легко подменить реализацию БД или мокировать в тестах
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
### Триггеры запуска

| Триггер                        | Что запускается                                 |
|--------------------------------|-------------------------------------------------|
| **Push** в любую ветку         | Линтер + Тесты                                  |
| **Pull Request** в любую ветку | Линтер + Тесты                                  |
| **Создание тега `v*`**         | Линтер + Тесты + Публикация Docker-образа       |


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