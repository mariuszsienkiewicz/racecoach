# Deploy RaceCoach on AWS (EC2 + RDS + S3 + HTTPS)

Runbook for the public demo. **Local development is unchanged:** use plain
`docker compose up` (Postgres + MinIO + Vite). This path is opt-in via a second
Compose file and a server-only env file.

## Topology

| Piece | Where |
| --- | --- |
| Frontend (static) + Caddy TLS | EC2 Docker |
| Symfony API + RabbitMQ + Go workers | EC2 Docker |
| PostgreSQL | RDS |
| Object storage | S3 |
| LLM | Hosted OpenAI-compatible API (optional; OpenRouter lesson) |

Browser -> `https://$DOMAIN` -> Caddy -> `/` static SPA, `/api` -> backend.

## Prerequisites

- EC2 (e.g. `t3.small`) with Docker Compose **v2.24+** (`!reset` in overlay)
- Elastic IP + Security Group: **22** (your IP), **80/443** (public)
- RDS Postgres 16; SG allows **5432** from the **EC2 security group**
- S3 bucket + IAM user with least-privilege keys
- Domain A record -> Elastic IP
- **Cloudflare:** use **DNS only** (grey cloud) while Caddy obtains Let’s Encrypt. Orange proxy fights origin auto-HTTPS.

## Files in this repo

| Path | Role |
| --- | --- |
| [docker-compose.yml](../docker-compose.yml) | Local default (unchanged) |
| [docker-compose.aws.yml](../docker-compose.aws.yml) | AWS overlay: prod frontend, Caddy, RDS/S3 env, no MinIO depends |
| [.env.aws.example](../.env.aws.example) | Placeholders - copy to `.env.aws` on the server |
| [deploy/Caddyfile](../deploy/Caddyfile) | Uses `{$DOMAIN}` from env (not hard-coded host) |
| [frontend/Dockerfile.prod](../frontend/Dockerfile.prod) | `npm run build` -> nginx |

## Server bootstrap (once)

```bash
git clone <repo> racecoach && cd racecoach
cp .env.aws.example .env.aws
# edit .env.aws: DOMAIN, DATABASE_URL, POSTGRES_*, S3_*, CORS_ALLOW_ORIGIN

mkdir -p backend/config/jwt
openssl genpkey -out backend/config/jwt/private.pem -aes256 -algorithm rsa \
  -pkeyopt rsa_keygen_bits:4096 -pass pass:CHANGE_ME
openssl pkey -in backend/config/jwt/private.pem -passin pass:CHANGE_ME \
  -pubout -out backend/config/jwt/public.pem
# set JWT_PASSPHRASE in compose/env to match, or keep Compose default for demo
```

## Bring up (explicit services - skip Postgres/MinIO containers)

```bash
docker compose --env-file .env.aws \
  -f docker-compose.yml -f docker-compose.aws.yml \
  up -d --build \
  rabbitmq backend frontend caddy \
  fit-metrics fit-structure fit-features fit-summary fit-chat
```

Verify:

```bash
dig +short "$DOMAIN"          # must be the Elastic IP
docker compose --env-file .env.aws -f docker-compose.yml -f docker-compose.aws.yml logs -f caddy
curl -sI "https://$DOMAIN"
```

Demo login (seeded by backend entrypoint): `coach@racecoach.local` / `coach123`.

## Local vs AWS

| | Local | AWS demo |
| --- | --- | --- |
| Command | `docker compose up --build` | `-f docker-compose.yml -f docker-compose.aws.yml` + `.env.aws` |
| DB | Compose Postgres | RDS |
| Files | MinIO | S3 |
| UI | Vite `:5173` | Caddy `:443` + nginx SPA |
| Secrets | Compose defaults / `.env` | `.env.aws` on server only |

Optional on the laptop: a gitignored `docker-compose.override.yml` for pointing local app at RDS/S3 while developing - never required for contributors.

## LLM on AWS (OpenRouter)

Workers speak any **OpenAI-compatible** chat API (`LLM_BASE_URL` + `LLM_MODEL` + `LLM_API_KEY`). Locally that is usually Ollama, on the demo host use a hosted provider so you do not run a GPU.

**Cheap default - [OpenRouter](https://openrouter.ai):**

1. Create an API key (do not commit it).
2. In server `.env.aws` set:

```bash
LLM_BASE_URL=https://openrouter.ai/api/v1
LLM_MODEL=openrouter/free
LLM_API_KEY=sk-or-v1-...
```

`openrouter/free` routes to free models ($0 tokens; daily request caps - higher after a small credit top-up). See OpenRouter docs for current limits.

3. Recreate only the LLM workers:

```bash
docker compose --env-file .env.aws \
  -f docker-compose.yml -f docker-compose.aws.yml \
  up -d --force-recreate --no-deps fit-summary fit-chat
```

**Notes**

- `fit-summary` requests `response_format: json_object`. If a free model rejects that, pick another OpenRouter model that supports JSON (or a cheap paid slug) and keep the same three env vars.
- Failed summaries land in RabbitMQ queue `fit-summary.dlq` (invalid JSON retries a few times first, bad model/auth 404 goes straight to DLQ). After fixing `LLM_*`, republish the body to exchange `messages` with routing key `activity.features.ready` and **clear** header `x-retry-count` (see [workers/README.md](../workers/README.md)).
- While `fit-summary` waits on the LLM it PATCHes `/api/activities/{id}/pipeline-heartbeat` so the UI can tell in-progress from stuck.
- Groq or other OpenAI-compatible hosts work the same way - only change `LLM_*`.
