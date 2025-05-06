#!/bin/sh
# this script is used by the Docker container to wait for the database to be ready
# and then run the database migrations before starting the main application
# it is used in the Dockerfile to run the migrations before starting the application
# it is also used in the docker-compose.yml file to run the migrations before starting the application 

set -e

echo "Waiting for database..."
while ! nc -z db 5432; do
  echo "Waiting for database to become available..."
  sleep 1
done
echo "Database is ready!"

# Load environment variables from .env
if [ -f /app/.env ]; then
    echo "Loading environment from .env..."
    set -a
    . /app/.env
    set +a
else
    echo ".env file not found!"
    exit 1
fi

# Set goose environment
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="$DATABASE_URL"

# Extract connection details for psql check
PGHOST=$(echo "$DATABASE_URL" | sed -E 's|.*@([^:/]+).*|\1|')
PGDATABASE=$(echo "$DATABASE_URL" | sed -E 's|.*/([^?]+).*|\1|')
PGUSER=$(echo "$DATABASE_URL" | sed -E 's|postgres://([^:]+):.*|\1|')
PGPASSWORD=$(echo "$DATABASE_URL" | sed -E 's|postgres://[^:]+:([^@]+)@.*|\1|')
export PGPASSWORD

goose -dir /app/migrations up

echo "Migration step done."
exec "$@"
