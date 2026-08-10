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
| `invoice` | Счета на оплату (create/cancel/paid/expire, адрес) | реализован |
| `payment` | Платежи (observed → confirmed → invoice paid) | реализован |
| `blockchain` | HD-адреса (EVM xpub) + Sepolia scanner | реализован |
| `ledger` | Учёт балансов (credit + balances) | реализован |
| `withdrawal` | Выводы (request + debit + complete facade) | частично |
| `webhook` | Исходящие уведомления мерчанту (enqueue + retry delivery) | частично |

### Структура репозитория

```
cmd/
  crypto-pay/     # HTTP API
  migrator/       # миграции БД
  scanner/        # EVM Sepolia observation
  worker/         # expire invoices + webhook delivery

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

# EVM Sepolia deposit addresses (account xpub at m/44'/60'/0')
# Path used: m/44'/60'/0'/0/{index}. Private keys never enter the process.
EVM_SEPOLIA_XPUB=

# Scanner (cmd/scanner) — required for on-chain observation
EVM_SEPOLIA_RPC_URL=
# EVM_SEPOLIA_START_BLOCK=0
# SCANNER_POLL_INTERVAL=3s
# SCANNER_BLOCK_BATCH=20
# CONFIRMATION_THRESHOLDS=evm:sepolia=1
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

### Webhook deliveries

| Method | Path | Auth | Описание |
|--------|------|------|----------|
| `GET` | `/api/v1/webhook-deliveries` | API key | список доставок (`?status=&limit=`) |
| `GET` | `/api/v1/webhook-deliveries/:id` | API key | одна доставка |

События v1: `invoice.paid`, `withdrawal.completed` (если у мерчанта задан `webhook_url`).

Доставка (worker):

- `POST` на snapshot URL
- headers: `X-Webhook-Event`, `X-Webhook-Delivery-Id`, `X-Webhook-Timestamp`, `X-Webhook-Signature: sha256=<hex>`
- подпись: `HMAC-SHA256(JWT_SECRET, "{timestamp}.{body}")`
- retry: 408/429/5xx + network, backoff; non-retryable 4xx / max attempts → `failed`

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
| `scanner` | EVM Sepolia poll → observe/confirm |
| `worker` | expire due invoices + webhook delivery (`ProcessDue`); outbox relay (пока stub) |

## Разработка

```bash
go build ./cmd/crypto-pay
go test ./...
```

Правила архитектуры для агентов/команды: `.cursor/rules/architecture.mdc`.

## Roadmap (кратко)

Подробный backlog до статуса «реализован» по всем BC: [docs/backlog-complete-modules.md](docs/backlog-complete-modules.md).

Фаза 1 (money-in) закрыта: ledger credit, confirmation policy, expire job, Sepolia scanner + HD xpub.

Дальше: outbox relay / async webhooks, per-merchant secrets, withdrawal pipeline.
