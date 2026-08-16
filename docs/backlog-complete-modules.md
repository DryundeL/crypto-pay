# Backlog: довести BC до «реализован»

Критерий готовности: один happy-path без ручных вызовов facade:

`create invoice → реальный адрес → scanner видит tx → N confirmations → invoice paid → ledger credit → webhook delivered`

и отдельно:

`request withdrawal → broadcast → confirmed → completed → webhook`

Текущие статусы модулей: см. таблицу Bounded contexts в [README.md](../README.md).

---

## Фаза 0 — фундамент (без этого всё остальное врёт)

### PR-0.1 Schema alignment ✅

**Проблема:** код писал в `invoice_invoices` / `payment_payments`, миграции создавали `invoices` / `payments`; FK payments ссылался на несуществующую таблицу.

**Сделано:**

- [x] код и `000004`/`000005` на простых именах: `invoices`, `payments`
- [x] FK `payments.invoice_id → invoices(id)`
- [x] индексы `(network, address)` и partial `expires_at WHERE status = 'pending'`

**Done when:** `go run ./cmd/crypto-pay` + create invoice/payment реально пишут/читают из Postgres.

### PR-0.2 Закрыть simulate-оплату с API key ✅

Сейчас мерчант с API key может дергать observe/confirm и «оплатить» сам себя.

**Сделано:**

- [x] убраны публичные write-роуты: `POST /payments/observed|confirmed`, весь blockchain HTTP
- [x] observe/confirm остаются in-process (facade/`cmd/scanner`); merchant API: `GET /payments/:id` (+ invoice/ledger/withdrawal/webhook reads)

**Done when:** внешний API key не может перевести invoice в paid.

---

## Фаза 1 — money-in happy path

### PR-1.1 Ledger credit on paid ✅

**Сделано:**

- [x] sync notifier: `invoice.paid` → `ledger.PostJournal` (idempotency `invoice_paid:{invoice_id}`)
- [x] wiring в bootstrap / scanner через `app.NewLedgerPaidNotifier` (webhook — outbox consumer, PR-2.2)

**Done when:** после paid у мерчанта растёт available balance; повтор paid/credit идемпотентен.

### PR-1.2 Confirmation policy ✅

**Сделано:**

- [x] конфиг порогов: `evm:sepolia=1`, `btc:regtest=1` (dev/staging); prod-значения отдельно; override `CONFIRMATION_THRESHOLDS`
- [x] `Confirm` только если `confirmations >= required`
- [x] обновлять `confirmations` на повторных observations (не только status flip)

**Done when:** invoice не становится paid на 0 confirmations.

### PR-1.3 Invoice expire job ✅

**Сделано:**

- [x] в `cmd/worker`: poll `expires_at < now AND status=pending` → `ExpireInvoice`
- [x] индекс по `(status, expires_at)` (`000010`)
- [ ] опционально webhook `invoice.expired` (v1.1)

**Done when:** просроченный invoice уходит в `expired` без ручного вызова.

### PR-1.4 Scanner v1 (хотя бы одна сеть) ✅

**Сделано:**

**EVM Sepolia:**

- [x] RPC poll новых блоков (`cmd/scanner` + `ethclient`)
- [x] match watched addresses из `blockchain_addresses`
- [x] вызов `blockchain.RecordObservation` / `ConfirmTransaction` → payment по policy
- [x] cursor/checkpoint в `blockchain_scan_cursors` (`000011`)
- [x] reorg v1: scan tip lagged на `required-1` блоков (deep reorg later)

**Адреса:**

- [x] EVM: HD из account-level xpub (`EVM_SEPOLIA_XPUB`, path `m/44'/60'/0'/0/{index}`)
- [x] BTC: hash-placeholder до Bitcoin scanner

**Done when:** перевод на выданный адрес в testnet → invoice paid без HTTP simulate.

---

## Фаза 2 — async reliability

### PR-2.1 Outbox Relayer ✅

**Сделано:**

- [x] реализация `outbox.Relayer` (claim pending → publish → mark sent)
- [x] минимальный bus: in-process в worker (Watermill можно подставить позже)
- [x] worker loop: relay + expire + webhook `ProcessDue`

**Done when:** `outbox_messages.status` доходит до `sent`; consumers видят события.

### PR-2.2 Webhook enqueue из events (убрать sync notch) ✅

**Сделано:**

- [x] убрать sync webhook enqueue после commit из invoice/withdrawal facades
- [x] consumer: `invoice.paid` / `withdrawal.completed` → `webhook.Enqueue`
- [x] at-least-once + текущая idempotency `event:source_id` остаются

**Done when:** падение API после MarkPaid не теряет webhook (delivery появится после relay).

### PR-2.3 Per-merchant webhook secret ✅

**Сделано:**

- [x] `merchants.webhook_secret` (generate on create / rotate endpoint)
- [x] подпись этим секретом, не `JWT_SECRET`
- [x] документ верификации для мерчанта

**Done when:** два мерчанта не шарят один signing key.

---

## Фаза 3 — money-out

### PR-3.1 Withdrawal TX consistency

**Сделать:**

- debit ledger + save withdrawal в **одной** DB transaction (или saga с compensating credit)
- сейчас debit до save — при фейле save деньги уходят «в никуда»

### PR-3.2 Withdrawal pipeline worker

Статусы довести до реальности:

`requested → approved? → broadcast → completed | rejected`

**Сделать:**

- команды Approve / Broadcast / Reject / Complete
- worker: взять requested → build+sign tx → broadcast → track confirmations → `Complete`
- на reject: credit back
- убрать из публичного контракта статусы, которых нет в домене, **или** реализовать их

**Done when:** request → on-chain tx → completed → webhook, без ручного `Complete()`.

---

## Фаза 4 — operational polish (можно после «реализован»)

| Item | Зачем |
|------|--------|
| Webhook replay / requeue HTTP | ops при сбое у мерчанта |
| Metrics + alerts (due lag, failed deliveries, scanner tip lag) | SRE |
| Amount tolerance (under/overpay) | реальные платежи |
| Reorg / double-spend handling | safety |
| Bitcoin UTXO scanner | второй network из README |
| Integration e2e (testcontainers + anvil/regtest) | не ломать flow регрессией |

---

## Порядок PR (рекомендуемый)

```text
0.1 schema
  → 0.2 lock down public write APIs
    → 1.1 ledger credit
      → 1.2 confirmations policy
        → 1.3 expire job
          → 1.4 scanner (одна сеть)
            → 2.1 outbox relayer
              → 2.2 async webhook enqueue
                → 2.3 per-merchant secrets
                  → 3.1 withdrawal TX fix
                    → 3.2 withdrawal worker
```

Параллелить можно: `1.3` с `1.1/1.2`; `2.3` с `2.1`; `3.x` после credit.

---

## Когда менять статусы в README

| Модуль | «реализован» после |
|--------|---------------------|
| **invoice** | 0.1 + 1.3 (+ webhook expired опционально) |
| **payment** | 0.1 + 0.2 + 1.2 |
| **blockchain** | 1.4 (реальный scanner + не-placeholder адреса) |
| **ledger** | 1.1 |
| **withdrawal** | 3.1 + 3.2 |
| **webhook** | 2.2 + 2.3 (ProcessDue уже есть) |

До `1.4` весь продукт честно остаётся «частично», даже если отдельные BC выглядят полными.

---

## Оценка объёма (грубо)

| Фаза | Effort |
|------|--------|
| 0 | 0.5–1 day |
| 1.1–1.3 | 1–2 days |
| 1.4 scanner+keys | 3–7 days (основная сложность) |
| 2 | 2–4 days |
| 3 | 3–6 days |
