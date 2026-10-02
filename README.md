# RaceCoach

Local coaching desk for endurance athletes: upload a `.fit` workout, run a Go analysis pipeline, then ask an AI coach about that session with **live streaming** replies.

## Highlights

- **Polyglot monorepo** - React 19 frontend, Symfony 7.4 API, Go workers
- **Event-driven FIT pipeline** - RabbitMQ topic exchange; MinIO holds `.fit` + JSON artifacts; DB stores keys, not blobs
- **Multi-turn Ask AI over SSE** - token stream from LLM -> fit-chat -> Symfony `StreamedResponse` -> live drawer bubble; history persisted per activity
- **Separated auth** - user JWT for the browser; `X-Worker-Token` for pipeline workers; Bearer service token for fit-chat
- **CI quality gates** - GitHub Actions: Go test/vet/build, PHPUnit, PHP-CS-Fixer, PHPStan (level 6), frontend lint + build

## Stack

| Layer | Tech |
| --- | --- |
| Frontend | React 19, Vite, Tailwind CSS 4, shadcn/ui |
| Backend | Symfony 7.4 LTS, PHP 8.5, Lexik JWT, PostgreSQL 16 |
| Messaging | RabbitMQ (topic exchange + per-stage queues, delayed retry / DLQ on summary) |
| Object storage | MinIO (S3-compatible) in Docker |
| Workers | Go services in `workers/` |
| LLM (optional) | Local: Ollama. AWS demo: OpenAI-compatible host (e.g. OpenRouter) |
| Dev / CI | Docker Compose; GitHub Actions |

## Architecture

```text
Browser (React)
   │  JWT
   ▼
Symfony API  ──upload──►  MinIO (.fit + JSON artifacts)
   │                         ▲
   │  activity.uploaded      │ Get/PutObject
   ▼                         │
RabbitMQ ──► fit-metrics ──► fit-structure ──► fit-features ──► fit-summary
                                                                      │
                                                              LLM (Ollama)

Ask AI (SSE, out of band - not via RabbitMQ):
React ──JWT──► Symfony StreamedResponse ──Bearer──► fit-chat SSE ──► LLM stream
                 │                                    │
                 ├─ persist athlete before stream     ├─ GetObject features
                 └─ persist coach after `done`        └─ token / done / error events
```

Pipeline stages write artifacts beside the FIT object (`*.metrics.json`, `*.structure.json`, …), PATCH activity state on Symfony with a shared worker token, then publish the next event. Ask AI stays synchronous HTTP between Symfony and fit-chat, but the body is an **SSE** stream (`token` / `done` / `error`), not a single JSON reply.

Deeper notes: [docs/architecture.md](docs/architecture.md), [docs/messaging-and-storage.md](docs/messaging-and-storage.md), [workers/README.md](workers/README.md).

## What works today

- JWT auth + activity list
- `.fit` upload -> MinIO -> full analysis pipeline -> activity `ready` with coach summary
- Multi-turn Ask AI on a finished activity (live tokens in the drawer; conversation saved per activity)
- Calendar + week overview from real activities

## Coming soon

- Adaptive training plans (dashboard plan panel + landing preview)
- Device sync (Coros / Garmin OAuth) - `.fit` upload works now

## Live demo

- **https://racecoach.heaps.pl** - EC2 + RDS + S3 + Caddy HTTPS; LLM via OpenRouter (or any OpenAI-compatible API)

Local vs AWS (including LLM env): [docs/aws-deploy-ec2.md](docs/aws-deploy-ec2.md).

## Quick start (local)

**Requirements:** Docker + Docker Compose. For AI summary / Ask AI, run [Ollama](https://ollama.com/) on the host and pull the model:

```bash
ollama pull llama3.1:8b
```

From the repo root:

```bash
cp backend/.env.example backend/.env

mkdir -p backend/config/jwt
openssl genpkey -out backend/config/jwt/private.pem -aes256 -algorithm rsa -pkeyopt rsa_keygen_bits:4096 -pass pass:racecoach_dev_jwt
openssl pkey -in backend/config/jwt/private.pem -passin pass:racecoach_dev_jwt -pubout -out backend/config/jwt/public.pem

docker compose up --build
```

Demo user is seeded on backend start.

| Service | URL |
| --- | --- |
| Frontend | http://localhost:5173 |
| API | http://localhost:8000 |
| Adminer | http://localhost:8080 - PostgreSQL, server `database`, user/password/db `racecoach` |
| RabbitMQ | http://localhost:15672 - `racecoach` / `racecoach` |
| MinIO console | http://localhost:9001 - `racecoach` / `racecoachsecret` |

**Demo login:** `coach@racecoach.local` / `coach123`

Without Ollama, upload and metrics/structure/features still run; summary retries/DLQs and Ask AI fail until the LLM is available.

### AWS demo (optional)

Same monorepo, second Compose file - **does not change** local `docker compose up`:

```bash
cp .env.aws.example .env.aws   # on the server only; fill RDS/S3/DOMAIN
docker compose --env-file .env.aws \
  -f docker-compose.yml -f docker-compose.aws.yml \
  up -d --build rabbitmq backend frontend caddy \
  fit-metrics fit-structure fit-features fit-summary fit-chat
```

Details: [docs/aws-deploy-ec2.md](docs/aws-deploy-ec2.md). Domain is set via `DOMAIN` env (Caddyfile uses `{$DOMAIN}`), not hard-coded in git beyond the example.

## API (high level)

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| `POST` | `/api/login` | - | `{ "email", "password" }` -> `{ "token" }` |
| `GET` | `/api/me` | JWT | Current user |
| `GET` | `/api/activities` | JWT | Activity list |
| `POST` | `/api/activities/fit` | JWT | Multipart `fitFile` |
| `GET` | `/api/activities/{id}/chat` | JWT | Persisted conversation |
| `POST` | `/api/activities/{id}/chat` | JWT | Ask AI - `text/event-stream` (`token` / `done` / `error`) |

Workers call internal PATCH endpoints with header `X-Worker-Token` (`WORKER_API_TOKEN`). Symfony calls `fit-chat` with `Authorization: Bearer` (`CHAT_SERVICE_TOKEN`). Those tokens are different on purpose.

## Quality / CI

On push and pull requests, [`.github/workflows/ci.yml`](.github/workflows/ci.yml) runs:

| Job | Checks |
| --- | --- |
| Workers | `go test ./…`, `go vet`, build all five binaries |
| Backend | PHPUnit, PHP-CS-Fixer (`composer cs-check`), PHPStan level 6 |
| Frontend | `npm run lint`, `npm run build` |

Backend tooling from Compose (preferred over host PHP):

```bash
docker compose exec backend composer cs-check
docker compose exec backend composer phpstan
docker compose exec backend composer test
```

## Configuration

- Compose injects env for all services (see `docker-compose.yml`).
- For host-side Symfony / tooling, copy `backend/.env.example` -> `backend/.env`.
- Dev secrets in Compose (`change-me`, `racecoachsecret`, …) are local only - change them before any shared deployment.
- Never commit `backend/config/jwt/*.pem` or a filled `.env` with real secrets.

## Repo layout

```text
frontend/              React app (+ Dockerfile.prod for demo)
backend/               Symfony API
workers/               Go pipeline + fit-chat
deploy/Caddyfile       TLS reverse proxy (DOMAIN from env)
docker-compose.yml     Local stack
docker-compose.aws.yml AWS / HTTPS overlay
docs/                  Architecture, messaging, AWS deploy
.github/               CI workflows
```
