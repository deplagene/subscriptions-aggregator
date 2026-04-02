# deplagene/subscriptions-aggregator

Тестовое задание для Effective Mobile.

Это REST-сервис для агрегации онлайн-подписок пользователей. Сервис позволяет:

- создавать подписки;
- получать подписку по `id`;
- получать список подписок с фильтрами и пагинацией;
- обновлять подписки;
- удалять подписки;
- считать суммарную стоимость подписок за выбранный период.

## Что используется

- Go 1.26.1
- `chi` для HTTP-роутинга
- PostgreSQL 17
- `pgx/v5` для работы с PostgreSQL
- `sqlc` для генерации кода по SQL-запросам
- `goose` для миграций
- `swaggo` для Swagger-документации
- Docker + Docker Compose для запуска проекта

## Что важно по предметной области

- Стоимость подписки хранится как целое число рублей.
- Формат дат в API: `MM-YYYY`.
- `start_date` обязателен, `end_date` опционален.
- При расчёте total я считаю месяцы включительно.
- Частичные дни не учитываю, потому что по ТЗ в запросах передаются только месяц и год.

Пример:

- подписка `400` рублей;
- период подписки `03-2025` -> `05-2025`;
- total за период `03-2025` -> `05-2025` будет `1200`.

## Как запустить

Я закладывал основной сценарий запуска через Docker Compose.

### 1. Подготовить `.env`

```bash
cp .env.example .env
```

При необходимости можно поменять порт приложения или параметры БД в `.env`.

### 2. Поднять проект

```bash
docker compose up -d --build
```

После старта будет доступно:

- API: `http://localhost:8080`
- Swagger UI: `http://localhost:8080/swagger/index.html`

## Swagger

Swagger-документация уже подключена в проект.

Если я меняю аннотации в коде, документацию можно пересобрать командой:

```bash
task swagger-generate
```

или

```bash
task swag-init
```

## Основные ручки

- `POST /api/v1/subscriptions`
- `GET /api/v1/subscriptions`
- `GET /api/v1/subscriptions/{id}`
- `PUT /api/v1/subscriptions/{id}`
- `PATCH /api/v1/subscriptions/{id}`
- `DELETE /api/v1/subscriptions/{id}`
- `GET /api/v1/subscriptions/total`
- `GET /healthz`

## Примеры запросов

### Создание подписки

```bash
curl -X POST http://localhost:8080/api/v1/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 400,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025"
  }'
```

Пример ответа:

```json
{
  "id": "2df8d0d4-915f-43df-b4d2-c4b4e2f410d2"
}
```

### Получение списка подписок

```bash
curl "http://localhost:8080/api/v1/subscriptions?limit=10&offset=0&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba"
```

### Получение подписки по `id`

```bash
curl "http://localhost:8080/api/v1/subscriptions/2df8d0d4-915f-43df-b4d2-c4b4e2f410d2"
```

### Обновление подписки

```bash
curl -X PUT http://localhost:8080/api/v1/subscriptions/2df8d0d4-915f-43df-b4d2-c4b4e2f410d2 \
  -H "Content-Type: application/json" \
  -d '{
    "service_name": "Yandex Plus",
    "price": 500,
    "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
    "start_date": "07-2025",
    "end_date": "09-2025"
  }'
```

### Удаление подписки

```bash
curl -X DELETE http://localhost:8080/api/v1/subscriptions/2df8d0d4-915f-43df-b4d2-c4b4e2f410d2
```

### Расчёт total

```bash
curl "http://localhost:8080/api/v1/subscriptions/total?from=07-2025&to=09-2025&user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus"
```

Пример ответа:

```json
{
  "total": 1500
}
```

## Полезные команды для разработки

Проверка тестов:

```bash
task test
```

Проверка покрытия:

```bash
task test-coverage
```

Генерация SQL-кода:

```bash
task sqlc-generate
```

Проверка SQL-запросов:

```bash
task sqlc-vet
```

## Что внутри проекта

- `cmd/subaggregator` — точка входа приложения
- `internal/httpapi` — HTTP-слой
- `internal/subscriptions` — сервисный и доменный слой
- `internal/postgres` — работа с PostgreSQL
- `migrations` — SQL-миграции
- `docs` — swagger-документация

## Примечание

Проверку существования пользователя я не делал, потому что это явно вынесено за рамки сервиса в самом ТЗ.
