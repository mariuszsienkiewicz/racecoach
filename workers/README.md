# RaceCoach workers (Go)

One Go module, multiple binaries under `cmd/`. Shared code lives in `internal/`.

Go owns RabbitMQ topology for the activity pipeline. Symfony publishes a small
reference event after upload; workers pull FIT bytes from MinIO/S3, write JSON
artifacts beside the `.fit`, PATCH activity state via the Symfony API, then
publish the next routing key.

`fit-chat` is **not** a queue consumer: it is a sync HTTP service Symfony calls
for Ask AI (Bearer `CHAT_SERVICE_TOKEN`).

## Assumptions

- **Store → reference → process** - the queue never carries the FIT file, only
  ids + storage location (`bucket`, `objectKey`, checksum, …).
- **Topic exchange `messages`** - one routing key per stage; each worker has its
  own queue bound to the key it consumes. Symfony’s Messenger transport declares
  the exchange with **empty `queues: []`** so PHP does not create a catch-all
  `messages` queue.
- **At-least-once delivery** - Ack after successful side effects. Retryable
  failures use either in-process backoff + Nack(requeue) (metrics → features)
  or broker TTL retry + DLQ (`fit-summary`). Poison messages are Rejected.
- **Idempotency** - before heavy work, `GET` the activity; if the expected
  artifact key is already set and status is far enough along, short-circuit by
  republishing the downstream event and Ack.
- **Artifacts in object storage** - JSON lives next to the FIT; the DB stores
  object key pointers, not full payloads:
  `*.metrics.json`, `*.structure.json`, `*.features.json`, `*.summary.json`.
- **Status model** - stages after upload keep the activity in `analyzing`.
  `fit-summary` PATCHes the coach summary and moves status to `ready`.
- **Auth**
  - Pipeline workers → Symfony: header `X-Worker-Token` (`WORKER_API_TOKEN`).
  - Symfony → `fit-chat`: `Authorization: Bearer` (`CHAT_SERVICE_TOKEN`).
  - These tokens are intentionally different. End-user JWT never reaches workers.
- **QoS** - consumers use prefetch `1` where set, so backoff / LLM work actually
  pauses that consumer instead of stacking unacked deliveries.

## Pipeline

```text
activity.uploaded
       │
       ▼
  fit-metrics   → Put *.metrics.json   → PATCH /metrics   → activity.metrics.ready
       │
       ▼
 fit-structure  → Put *.structure.json → PATCH /structure → activity.structure.ready
       │
       ▼
 fit-features   → Put *.features.json  → PATCH /features  → activity.features.ready
       │
       ▼
  fit-summary   → Put *.summary.json   → PATCH /summary   → activity.summary.ready
                  (LLM; status → ready)                      (+ retry queue / DLQ)
```

Same skeleton for queue workers: consume → idempotent check → transform →
PutObject → PATCH → publish next → Ack. Only the middle step changes.

Ask AI (out of band):

```text
React ──JWT──► Symfony ──Bearer──► fit-chat :8081
                                    │
                                    ├─ GetObject features (keys from request body)
                                    ├─ optional summary text from gateway
                                    └─ LLM CompletionRaw → { "reply" }
```

`fit-chat` must **not** call Symfony mid-request (avoids PHP worker deadlock).
Artifact keys / summary come in the JSON body from the gateway.

## Binaries

| Binary | Role | Consumes / listens | Publishes / responds | Status |
| --- | --- | --- | --- | --- |
| `fit-metrics` | distance, duration, HR, … from FIT | `activity.uploaded` | `activity.metrics.ready` | done |
| `fit-structure` | laps / workout structure | `activity.metrics.ready` | `activity.structure.ready` | done |
| `fit-features` | AI-ready feature payload | `activity.structure.ready` | `activity.features.ready` | done |
| `fit-summary` | coach narrative via LLM | `activity.features.ready` | `activity.summary.ready` | done |
| `fit-chat` | Ask AI HTTP gateway target | `POST /v1/chat`, `GET /healthz` | `{ "reply" }` | done |

### Transient failure strategy

| Stage | Retryable (API / MinIO / publish) | LLM down / bad JSON |
| --- | --- | --- |
| metrics → features | `RequeueAfterBackoff` (~30s) then Nack(requeue) | n/a |
| summary | TTL queue `fit-summary.retry.30s` (DLX back to `activity.features.ready`), header `x-retry-count`, max **5** delayed attempts → `fit-summary.dlq` | retryable LLM → same delayed path; non-retryable / invalid JSON → Reject |

## Layout

```text
workers/
  cmd/
    fit-metrics/
    fit-structure/
    fit-features/
    fit-summary/
    fit-chat/
  internal/
    domain/         # event + artifact DTOs (no I/O)
    amqp/           # dial, declare, bind, consume, publish, backoff helpers
    storage/        # MinIO/S3 GetObject / PutObject
    api/            # Symfony HTTP client (worker token)
    fit/            # FIT parsing (metrics, structure)
    fitmetrics/     # worker loop + process
    fitstructure/
    fitfeatures/    # merge metrics+structure → features artifact
    fitsummary/     # LLM summary + retry/DLQ topology
    fitchat/        # HTTP server + chat prompts
    coach/          # ForCoachPrompt - strip recovery traps before LLM
    llm/            # OpenAI-compatible client (Ollama)
    config/         # env loading
  Dockerfile        # builds all five binaries; default CMD is fit-metrics
```

## Environment

Shared by Compose / local runs (`config.Load`):

| Variable | Purpose |
| --- | --- |
| `RABBITMQ_URL` | AMQP URL (pipeline workers) |
| `S3_ENDPOINT` / `S3_REGION` / `S3_ACCESS_KEY` / `S3_SECRET_KEY` / `S3_BUCKET` | object storage |
| `S3_USE_PATH_STYLE` | `true` for MinIO |
| `API_BASE_URL` | Symfony base URL |
| `WORKER_API_TOKEN` | must match backend `WORKER_API_TOKEN` |
| `LLM_BASE_URL` | OpenAI-compatible base (Compose: `http://host.docker.internal:11434/v1`) |
| `LLM_MODEL` | default `llama3.1:8b` |
| `LLM_API_KEY` | default `ollama` |
| `CHAT_HTTP_ADDR` | `fit-chat` listen addr (default `:8081`) |
| `CHAT_SERVICE_TOKEN` | required for `fit-chat`; must match Symfony |

`LLM_*` is required for `fit-summary` and `fit-chat`. Without Ollama on the host,
earlier pipeline stages still complete; summary retries/DLQs and chat returns 503.

## Run locally

```bash
cd workers
go run ./cmd/fit-metrics
go run ./cmd/fit-structure
go run ./cmd/fit-features
go run ./cmd/fit-summary
go run ./cmd/fit-chat
```

## Run with Docker Compose

From the monorepo root:

```bash
docker compose up -d --build fit-metrics fit-structure fit-features fit-summary fit-chat
```

All five share the workers image; non-default services override `command` to
`/fit-structure`, `/fit-features`, `/fit-summary`, or `/fit-chat`. Summary and
chat reach Ollama via `host.docker.internal`.
