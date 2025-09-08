# WB L0 Orders — демо-сервис заказов (Go + Kafka + PostgreSQL + Cache)

Небольшой микросервис на Go, который:

* потребляет сообщения о заказах из Kafka,
* валидирует и сохраняет их в PostgreSQL (с транзакциями),
* кэширует последние заказы в памяти для быстрого чтения,
* при старте восстанавливает кеш из БД,
* отдает заказ по `order_uid` через HTTP JSON API,
* имеет простой веб-интерфейс для запроса заказа по ID,
* корректно завершает работу (graceful shutdown).



## Быстрый старт

### 1) Конфигурация

Все настройки лежат в `./config/config-local.yml`. Укажите там DSN/параметры PostgreSQL, Kafka (brokers, topic, group), порт HTTP-сервера и уровень логов.

### 2) Запуск в Docker

```bash
docker compose up -d --build
```

Команда поднимет:

* PostgreSQL (с автоприменением миграций),
* Kafka (и, при необходимости, ZooKeeper/контроллер),
* backend-сервис (API + консьюмер),
* web (Nginx со статикой `web/index.html`).

Проверьте логи:

```bash
docker compose logs -f api
```

### 3) Проверка API

Эндпоинт чтения заказа:

```
GET http://localhost:8082/order/<order_uid>
```


Пример:

```bash
curl http://localhost:8082/order/b563feb7b2b84b6test
```

### 4) Веб-интерфейс

Откройте страницу из контейнера `web` (порт смотрите в `docker-compose.yml` и/или `web/nginx.conf`), например:

```
http://localhost:8080
```

Введите `order_uid` и получите данные, которые страница подтягивает из HTTP API.

---

## Отправка тестового сообщения в Kafka

Сервис слушает топик из `config-local.yml`. Отправьте валидный JSON из `model.json`:


## Миграции БД

Миграции применяются автоматически при старте (см. `migrations/embed.go` и логи сервиса).

Ручной откат (нужно установленное CLI `migrate`):

```bash
migrate -path migrations -database "postgres://backend:123123123a@localhost:5432/wb_l0?sslmode=disable" down
```



## HTTP API

* `GET /order/{order_uid}` — вернуть заказ в JSON.

  * Источник: in-memory кеш (горячий путь), при отсутствии — чтение из БД с последующим кэшированием.
  * Ответ соответствует структуре `model.json`:

    * `order` + вложенные `delivery`, `payment`, `items`.

Пример ответа (фрагмент):

```json
{
  "order_uid": "b563feb7b2b84b6test",
  "track_number": "WBILMTESTTRACK",
  "delivery": { "name": "Test Testov", "...": "..." },
  "payment": { "transaction": "b563feb7b2b84b6test", "...": "..." },
  "items": [ { "chrt_id": 9934930, "...": "..." } ],
  "date_created": "2021-11-26T06:22:19Z",
  "locale": "en",
  "...": "..."
}
```



## Архитектура

* **Kafka Consumer** (`internal/kafka`, `internal/orders/consumer.go`): подписка на топик, парсинг JSON, валидация, запись в БД с транзакцией, ACK только после успешного сохранения.
* **БД**: PostgreSQL, репозиторий `internal/orders/repository` с SQL-запросами и `pg_repository`.
* **Кеш**: in-memory `map[string]*Order` (`internal/orders/cache`) + синхронизация; прогрев кеша из БД на старте.
* **HTTP-сервер** (`internal/server`, `internal/orders/delivery/http`): JSON API + маршруты.
* **Web** (`web/`): простая страница `index.html` под Nginx.
* **Логирование** (`pkg/logger/zap_logger.go`): zap.
* **Валидация** (`pkg/validation`): проверка входящих сообщений.

## Производительность и устойчивость

* Горячий путь чтения — из кеша (микросекунды).
* Повторные запросы того же `order_uid` — существенно быстрее чтения из БД.
* Сообщения подтверждаются брокеру только после успешной фиксации транзакции в БД.
* Невалидные сообщения не сохраняются, попадают в логи.
* При рестарте кеш восстанавливается из БД — сервис сразу готов к обслуживанию чтений.

## Локальный запуск без Docker

1. Настройте `config-local.yml` под локальные хосты/порты.
2. Примените миграции (или дайте это сделать сервису при старте).
3. Запустите:

```bash
go run ./cmd/api
```

Тесты:

```bash
go test ./...
```



## Структура проекта

```
.
├── README.md
├── cmd
│   └── api
│       └── main.go
├── config
│   ├── config-local.yml
│   └── config.go
├── docker
│   └── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── internal
│   ├── kafka
│   │   ├── consumers.go
│   │   └── kafka.go
│   ├── orders
│   │   ├── cache
│   │   │   └── cache.go
│   │   ├── consumer.go
│   │   ├── delivery
│   │   │   ├── http
│   │   │   │   ├── handlers.go
│   │   │   │   └── routes.go
│   │   │   └── kafka
│   │   │       ├── handlers.go
│   │   │       └── routes.go
│   │   ├── delivery.go
│   │   ├── mocks
│   │   │   └── repository.go
│   │   ├── pg_repository.go
│   │   ├── repository
│   │   │   ├── pg_repository.go
│   │   │   └── sql_queries.go
│   │   ├── usecase
│   │   │   ├── usecase.go
│   │   │   └── usecase_test.go
│   │   └── usecase.go
│   └── server
│       ├── handlers.go
│       └── server.go
├── migrations
│   ├── 000001_create_initial_tables.down.sql
│   ├── 000001_create_initial_tables.up.sql
│   └── embed.go
├── models
│   ├── delivery.go
│   ├── item.go
│   ├── order.go
│   └── payment.go
├── pkg
│   ├── logger
│   │   └── zap_logger.go
│   └── validation
│       └── validation.go
└── web
    ├── Dockerfile
    ├── index.html
    └── nginx.conf
```

---

## Полезные команды

Пересобрать и перезапустить:

```bash
docker compose up -d --build
```

Посмотреть логи:

```bash
docker compose logs -f api
docker compose logs -f web
docker compose logs -f postgres
docker compose logs -f kafka
```

Остановить и удалить контейнеры/сети:

```bash
docker compose down -v
```

---

## Траблшутинг

* **`connection refused` к БД**: проверьте `db.dsn`/порты и что Postgres поднят.
* **Сообщения не читаются**: убедитесь, что `kafka.brokers`, `topic`, `group_id` совпадают с реальной конфигурацией кластера/compose.
* **404 по `GET /order/{id}`**: заказа нет ни в кеше, ни в БД — сначала опубликуйте валидное сообщение в Kafka.
* **Падение на старте**: смотрите логи — нередко проблема в конфиге (неверный DSN/брокер)
