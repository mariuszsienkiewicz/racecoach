# RaceCoach workers (Go)

One Go module, multiple binaries under `cmd/`. Shared code lives in `internal/`.

Workers own RabbitMQ topology for activity processing. Symfony publishes small
reference events; workers pull FIT bytes from MinIO/S3, write JSON artifacts
beside the `.fit`, PATCH activity state via the Symfony API, then publish the
next pipeline event.

## Assumptions

- **Store → reference → process** - the queue never carries the FIT file, only
  ids + storage location (`bucket`, `objectKey`, checksum, …).
- **Topic exchange `messages`** - one routing key per stage; each worker has its
  own queue bound to the key it consumes. Fan-out is by binding, not by
  competing on one shared queue.
- **At-least-once delivery** - Ack after successful side effects; Nack+requeue
  on retryable failures (network, 5xx); drop on poison messages (bad JSON, …).
- **Idempotency** - before heavy work, `GET` the activity; if the expected
  artifact key is already set and status is `analyzing`/`ready`, short-circuit
  by republishing the downstream event and Ack.
- **Artifacts in object storage** - metrics/structure (and later features) live
  as `*.metrics.json` / `*.structure.json` next to the FIT; the DB stores the
  object key pointer, not the full payload.
- **Status model** - after metrics (and later stages) the activity stays
  `analyzing`. `ready` is reserved for a later completion signal (summary).
- **Auth** - workers call Symfony with `X-Worker-Token` (`WORKER_API_TOKEN`),
  not end-user JWT.

## Pipeline

```text
activity.uploaded
       │
       ▼
  fit-metrics  →  Put *.metrics.json  →  PATCH /metrics  →  activity.metrics.ready
       │
       ▼
 fit-structure  →  Put *.structure.json  →  PATCH /structure  →  activity.structure.ready
       │
       ▼
 fit-features   (planned)  →  activity.features.ready
       │
       ▼
 fit-summary    (planned)  →  status ready
```

Same worker skeleton each time: consume → (idempotent check) → transform →
PutObject → PATCH → publish next → Ack. Only the middle step changes
(parse FIT metrics, parse laps, assemble AI features, narrative summary).

## Status

| Binary | Role | Consumes | Publishes | Status |
| --- | --- | --- | --- | --- |
| `fit-metrics` | distance, duration, avg HR, … | `activity.uploaded` | `activity.metrics.ready` | done |
| `fit-structure` | laps / workout structure | `activity.metrics.ready` | `activity.structure.ready` | done |
| `fit-features` | AI-ready feature payload | `activity.structure.ready` (TBD) | `activity.features.ready` | planned |
| `fit-summary` | narrative summary | `activity.features.ready` | completion / `ready` | planned |

## Layout

```text
workers/
  cmd/
    fit-metrics/
    fit-structure/
  internal/
    domain/         # event + artifact DTOs (no I/O)
    amqp/           # dial, declare, bind, consume, publish
    storage/        # MinIO/S3 GetObject / PutObject
    api/            # Symfony HTTP client (worker token)
    fit/            # FIT parsing (metrics, structure)
    fitmetrics/     # worker A loop + process
    fitstructure/   # worker B loop + process
    config/         # env loading
  Dockerfile        # builds both binaries; default CMD is fit-metrics
```

## Run locally

```bash
cd workers
go run ./cmd/fit-metrics
go run ./cmd/fit-structure
```

Required env (same names as Compose):

| Variable | Purpose |
| --- | --- |
| `RABBITMQ_URL` | AMQP URL |
| `S3_ENDPOINT` / `S3_REGION` / `S3_ACCESS_KEY` / `S3_SECRET_KEY` / `S3_BUCKET` | object storage |
| `S3_USE_PATH_STYLE` | `true` for MinIO |
| `API_BASE_URL` | Symfony base URL |
| `WORKER_API_TOKEN` | must match backend `WORKER_API_TOKEN` |

## Run with Docker Compose

From the monorepo root:

```bash
docker compose up -d --build fit-metrics fit-structure
```

Both services share the workers image; `fit-structure` overrides the command to
`/fit-structure`.
