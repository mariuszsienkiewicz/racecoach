# Messaging & object storage (dev)

## Mental model

```text
Upload request
   -> validate .fit
   -> PutObject to MinIO (bytes)
   -> INSERT activity (objectKey + bucket)
   -> RabbitMQ event on topic exchange `messages` (references only)
   -> Go workers download by bucket/key, write JSON artifacts, PATCH API, publish next key
```

Symfony Messenger is **publish-only** for activity events (`queues: []` on the
`async` transport). Queue topology and consumers live in `workers/` - see
`workers/README.md`.

## RabbitMQ

- UI: http://localhost:15672 (`racecoach` / `racecoach`)
- Exchange: `messages` (topic)
- First event: `App\Message\ActivityUploadedMessage` -> routing key `activity.uploaded`
- Pipeline: `activity.metrics.ready` -> `activity.structure.ready` ->
  `activity.features.ready` -> `activity.summary.ready`
- `fit-summary` also declares `fit-summary.retry.30s` (TTL + DLX) and
  `fit-summary.dlq`

## MinIO / S3

- API: http://localhost:9000
- Console: http://localhost:9001 (`racecoach` / `racecoachsecret`)
- Bucket: `racecoach-fits` (created by `minio-init`)
- Object key shape: `users/{userId}/fits/{YYYY}/{mm}/{random}.fit`
- Artifacts beside the FIT: `*.metrics.json`, `*.structure.json`,
  `*.features.json`, `*.summary.json`

Backend env:

- `S3_ENDPOINT=http://minio:9000`
- `S3_ACCESS_KEY` / `S3_SECRET_KEY`
- `S3_BUCKET=racecoach-fits`
- `S3_USE_PATH_STYLE=true` (required for MinIO in Docker DNS)

Prod swap later: same code, change endpoint/credentials to real AWS S3 and usually set path-style to false.

## Related

- [architecture.md](architecture.md) - why reference events, dual tokens, SSE chat
- [../workers/README.md](../workers/README.md) - queue topology, retry/DLQ, fit-chat SSE
- [../README.md](../README.md) - product overview and quick start

