# Architecture notes

Short interview / CV talking points. For queue and MinIO details see
[messaging-and-storage.md](messaging-and-storage.md); for worker conventions see
[../workers/README.md](../workers/README.md).

## Why reference events (not the FIT on the bus)

Upload stores the `.fit` in MinIO first, then publishes a small RabbitMQ message
with `activityId`, `bucket`, `objectKey`, checksum, etc. Workers download by key.

That keeps brokers light, allows reprocessing without re-uploading, and matches
a store -> reference -> process pattern you can later point at real S3.

## Why two machine tokens (plus user JWT)

| Caller | Credential | Purpose |
| --- | --- | --- |
| Browser | Lexik JWT | User-facing API |
| Pipeline workers | `X-Worker-Token` (`WORKER_API_TOKEN`) | PATCH activity state after each stage |
| Symfony -> fit-chat | `Authorization: Bearer` (`CHAT_SERVICE_TOKEN`) | Ask AI inference only |

End-user JWTs never reach workers. Worker and chat tokens stay separate so a leak
in one path does not grant the other.

## Why Ask AI is SSE (not one JSON reply)

The analysis pipeline is async (queues). Chat is interactive: athletes expect
tokens as they arrive. The path is:

```text
React (fetch + SSE parse)
  -> Symfony StreamedResponse
    -> fit-chat SSE
      -> OpenAI-compatible LLM stream (Ollama)
```

`EventSource` cannot send POST + JWT, so the frontend uses `fetch` and a manual
SSE parser. Events: `token` (delta), `done` (full reply + persisted messages),
`error`.

## Where persistence happens

Symfony owns the database for chat:

1. Validate activity ownership and readiness
2. Persist the **athlete** message
3. Stream tokens from fit-chat to the client
4. On `done`, persist the **coach** message; on mid-stream failure, athlete stays,
   coach is not written

fit-chat must not call Symfony mid-request (avoids tying up PHP while waiting on
the LLM and risking deadlock with worker PATCHes). Artifact keys, summary, and
history are passed in the chat request body.

## Pipeline vs chat

| Concern | Pipeline (metrics -> summary) | Ask AI |
| --- | --- | --- |
| Transport | RabbitMQ topic | Direct HTTP + SSE |
| LLM | fit-summary (async, retry/DLQ) | fit-chat (request-scoped stream) |
| Failure | Nack / TTL retry / DLQ | SSE `error` event (or JSON 4xx before stream) |
