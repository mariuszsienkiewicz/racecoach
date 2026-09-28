# RaceCoach

Local coaching desk for endurance athletes: upload a `.fit` workout, let a Go pipeline analyze it, then ask an AI coach about that session.

## Stack

| Layer | Tech |
| --- | --- |
| Frontend | React 19, Vite, Tailwind CSS 4, shadcn/ui |
| Backend | Symfony 7.4 LTS, PHP 8.5, Lexik JWT, PostgreSQL 16 |
| Messaging | RabbitMQ (topic exchange + per-stage queues, delayed retry / DLQ on summary) |
| Object storage | MinIO (S3-compatible) in Docker |
| Workers | Go services in `workers/` |
| LLM (optional) | Ollama on the host (`llama3.1:8b` by default) |
| Dev | Docker Compose |

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
                                                                      │
Ask AI: React ──► Symfony ──► fit-chat (sync HTTP) ──► MinIO + LLM
```

Pipeline stages write artifacts beside the FIT object (`*.metrics.json`, `*.structure.json`, …), PATCH activity state on Symfony with a shared worker token, then publish the next event. Ask AI does **not** go through RabbitMQ: Symfony calls `fit-chat` over HTTP with a separate service token.

## Quick start

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

Demo user is seeded on backend start (`coach@racecoach.local` / `coach123`).

| Service | URL |
| --- | --- |
| Frontend | http://localhost:5173 |
| API | http://localhost:8000 |
| Adminer | http://localhost:8080 - PostgreSQL, server `database`, user/password/db `racecoach` |
| RabbitMQ | http://localhost:15672 - `racecoach` / `racecoach` |
| MinIO console | http://localhost:9001 - `racecoach` / `racecoachsecret` |

**Demo login:** `coach@racecoach.local` / `coach123`

Without Ollama, upload and metrics/structure/features still run; summary and Ask AI will fail or retry until the LLM is available.

## What works today

- JWT auth + activity list
- `.fit` upload → MinIO → full analysis pipeline → activity `ready` with coach summary
- Ask AI on a finished activity (optimistic chat UI)
- Calendar + week overview from real activities

## Coming soon

- Adaptive training plans (dashboard plan panel + landing preview)
- Device sync (Coros / Garmin OAuth) - `.fit` upload works now

## API (high level)

| Method | Path | Auth | Notes |
| --- | --- | --- | --- |
| `POST` | `/api/login` | - | `{ "email", "password" }` → `{ "token" }` |
| `GET` | `/api/me` | JWT | Current user |
| `GET` | `/api/activities` | JWT | Activity list |
| `POST` | `/api/activities/fit` | JWT | Multipart `fitFile` |
| `POST` | `/api/activities/{id}/chat` | JWT | Ask AI about one activity |

Workers call internal PATCH endpoints with header `X-Worker-Token` (`WORKER_API_TOKEN`). Symfony calls `fit-chat` with `Authorization: Bearer` (`CHAT_SERVICE_TOKEN`). Those tokens are different on purpose.

## Configuration

- Compose injects env for all services (see `docker-compose.yml`).
- For host-side Symfony / tooling, copy `backend/.env.example` → `backend/.env`.
- Dev secrets in Compose (`change-me`, `racecoachsecret`, …) are local only - change them before any shared deployment.
- Never commit `backend/config/jwt/*.pem` or a filled `.env` with real secrets.

## Repo layout

```text
frontend/   React app
backend/    Symfony API
workers/    Go pipeline + fit-chat
docs/       Deeper notes (messaging, storage, …)
```

See `workers/README.md` for worker conventions and queue topology.
