# Crypto-Pay

Multichain Crypto Payment Gateway — backend-платформа, через которую интернет-магазин может принимать оплату в криптовалюте.

```
Мерчант создаёт invoice
        ↓
Система выдаёт уникальный адрес
        ↓
Пользователь отправляет криптовалюту
        ↓
Scanner обнаруживает транзакцию
        ↓
Система ждёт нужное количество подтверждений
        ↓
Invoice становится paid
        ↓
Мерчанту отправляется webhook
```

Целевые сети (в планах):

- **EVM** (account-based) — например Sepolia / Ethereum-compatible
- **Bitcoin** Regtest/Testnet (UTXO)

## Стек

| Слой | Технология |
|------|------------|
| Язык | Go 1.26+ |
| HTTP | Echo |
| БД | PostgreSQL |
| ORM / SQL | GORM |
| Миграции | golang-migrate |
| DI | явный composition root (`internal/app`) |

## Архитектура

Модульный монолит:

- **Bounded Contexts** — независимые бизнес-модули
- **DDD** внутри модуля
- **Hexagonal / Clean Architecture** на уровне модуля
- **CQRS** — commands через aggregates, queries читают DTO из PostgreSQL
- **Transactional outbox** — integration events в одной TX с записью aggregate
- **Event Sourcing в v1 не используется**
- Синхронные handlers инжектятся напрямую (без reflection-bus)

### Bounded contexts

| Модуль | Назначение | Статус |
|--------|------------|--------|
| `merchant` | Мерчанты и API keys | реализован |
| `invoice` | Счета на оплату (create/cancel/paid, адрес) | частично |
| `payment` | Платежи (observed → confirmed → invoice paid) | частично |
| `blockchain` | Адреса + observation/confirm (scanner stub) | частично |
| `ledger` | Учёт балансов (credit + balances) | частично |
| `withdrawal` | Выводы | каркас |
| `webhook` | Исходящие уведомления мерчанту | каркас |

### Структура репозитория

```
cmd/
  crypto-pay/     # HTTP API
  migrator/       # миграции БД
  scanner/        # наблюдение за сетями (stub)
  worker/         # outbox / webhooks / jobs (stub)

internal/
  app/            # bootstrap, config
  platform/       # database, outbox, eventbus, httpserver, transaction, observability
  modules/        # bounded contexts
migrations/
```

Модуль:

```
internal/modules/<bc>/
  module.go
  contracts.go
  events.go
  internal/
    domain/
    application/{command,query,ports}
    infrastructure/{write,read}
    delivery/http/          # единственное место с Echo в модуле
```

## Быстрый старт

### Требования

- Go 1.26+
- PostgreSQL 17+ (или Docker)

### Конфигурация

Создайте `.env` в корне:

```env
APP_ENV=development
APP_PORT=8080
APP_URL=http://localhost:8080
APP_VERSION=dev

DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=crypto_pay
DB_SSLMODE=disable

# минимум 32 символа; также используется как pepper для API keys
JWT_SECRET=change-me-to-a-long-random-secret-key
```

### Локальный запуск

```bash
# поднять Postgres (пример)
docker compose up -d postgres

# API (миграции применяются при старте)
go run ./cmd/crypto-pay
```

Проверка:

```bash
curl -s http://localhost:8080/health
curl -s http://localhost:8080/health/ready
```

### Docker

```bash
docker compose up -d --build
```

Миграции вручную (опционально):

```bash
docker compose --profile tools run --rm migrator
# или локально:
go run ./cmd/migrator -command=up
```

## API

Базовый префикс: `/api/v1`

### Health

| Method | Path | Описание |
|--------|------|----------|
| `GET` | `/health` | liveness |
| `GET` | `/health/live` | liveness |
| `GET` | `/health/ready` | readiness (ping Postgres) |

### Merchant

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| `POST` | `/api/v1/merchants` | нет | создать мерчанта + первый API key |
| `GET` | `/api/v1/merchants/:id` | API key | получить мерчанта |
| `POST` | `/api/v1/merchants/:id/api-keys` | API key | выпустить ключ |
| `GET` | `/api/v1/merchants/:id/api-keys` | API key | список ключей (без секрета) |
| `DELETE` | `/api/v1/merchants/:id/api-keys/:keyId` | API key | revoke ключа |

Авторизация защищённых роутов:

- заголовок `X-API-Key: <key>`, или
- `Authorization: Bearer <key>` / `Authorization: ApiKey <key>`

Ключ хранится только как `SHA-256(pepper:plaintext)`; plaintext возвращается **один раз** при создании.

Пример:

```bash
# создать мерчанта
curl -s -X POST http://localhost:8080/api/v1/merchants \
  -H 'Content-Type: application/json' \
  -d '{"name":"Acme","webhook_url":"https://example.com/hook"}'

# дальше с ключом из ответа
curl -s http://localhost:8080/api/v1/merchants/<merchant_id> \
  -H "X-API-Key: <api_key>"
```

## Процессы

| Бинарь | Роль |
|--------|------|
| `crypto-pay` | HTTP API |
| `migrator` | применение / откат миграций |
| `scanner` | blockchain observation (заглушка) |
| `worker` | outbox relay, webhooks, фоновые job'ы (заглушка) |

## Разработка

```bash
go build ./cmd/crypto-pay
go test ./...
```

Правила архитектуры для агентов/команды: `.cursor/rules/architecture.mdc`.

## Roadmap (кратко)

1. Invoice + выделение депозитного адреса
2. Blockchain scanner (EVM + Bitcoin)
3. Payment confirmations → invoice paid
4. Webhook delivery через worker + outbox/event bus
5. Ledger / Withdrawal
