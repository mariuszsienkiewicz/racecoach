# Messaging & object storage (dev)

## Mental model

```text
Upload request
   → validate .fit
   → PutObject to MinIO (bytes)
   → INSERT activity (objectKey + bucket)
   → RabbitMQ event (references only)
   → worker downloads by bucket/key
```

## RabbitMQ

- UI: http://localhost:15672 (`racecoach` / `racecoach`)
- Event: `App\Message\ActivityUploadedMessage`
- Consume: `docker compose exec backend php bin/console messenger:consume async -vv`

## MinIO / S3

- API: http://localhost:9000
- Console: http://localhost:9001 (`racecoach` / `racecoachsecret`)
- Bucket: `racecoach-fits` (created by `minio-init`)
- Object key shape: `users/{userId}/fits/{YYYY}/{mm}/{random}.fit`

Backend env:

- `S3_ENDPOINT=http://minio:9000`
- `S3_ACCESS_KEY` / `S3_SECRET_KEY`
- `S3_BUCKET=racecoach-fits`
- `S3_USE_PATH_STYLE=true` (required for MinIO in Docker DNS)

Prod swap later: same code, change endpoint/credentials to real AWS S3 and usually set path-style to false.
