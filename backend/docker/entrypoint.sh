#!/bin/sh
set -e

cd /app

if [ -f composer.json ]; then
  composer install --prefer-dist --no-interaction
fi

echo "Waiting for PostgreSQL..."
until pg_isready -h "${POSTGRES_HOST:-database}" -p "${POSTGRES_PORT:-5432}" -U "${POSTGRES_USER:-racecoach}" >/dev/null 2>&1; do
  sleep 1
done

echo "Waiting for RabbitMQ..."
php -r '
$dsn = getenv("MESSENGER_TRANSPORT_DSN") ?: "";
if (!str_starts_with($dsn, "amqp://")) {
    exit(0);
}
$host = parse_url($dsn, PHP_URL_HOST) ?: "rabbitmq";
$port = parse_url($dsn, PHP_URL_PORT) ?: 5672;
for ($i = 0; $i < 60; $i++) {
    $fp = @fsockopen($host, (int) $port, $errno, $errstr, 1);
    if ($fp) {
        fclose($fp);
        exit(0);
    }
    sleep(1);
}
fwrite(STDERR, "RabbitMQ is not reachable\n");
exit(1);
'

php bin/console doctrine:database:create --if-not-exists --no-interaction || true
php bin/console doctrine:migrations:migrate --no-interaction --allow-no-migration || php bin/console doctrine:schema:update --force --no-interaction
php bin/console app:create-user "${DEMO_USER_EMAIL:-coach@racecoach.local}" "${DEMO_USER_PASSWORD:-coach123}" || true

exec "$@"
