#!/bin/bash

# Load environment variables from .env.production
set -a
source .env.production
set +a

# Fail fast on errors
set -euo pipefail

# Required env vars in .env.production:
# POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, DB_CONTAINER_NAME

# Validate input
if [ $# -ne 1 ]; then
  echo "Usage: $0 <path-to-sql-dump>"
  exit 1
fi

DUMP_FILE="$1"
if [ ! -f "$DUMP_FILE" ]; then
  echo " File not found: $DUMP_FILE"
  exit 1
fi

# Dynamically get the network name of the container
DB_NETWORK_NAME=$(docker inspect "$DB_CONTAINER_NAME" \
  | jq -r '.[0].NetworkSettings.Networks | keys[0]')

echo "Detected Docker network: $DB_NETWORK_NAME"
echo "Restoring database '$POSTGRES_DB' into container '$DB_CONTAINER_NAME' from '$DUMP_FILE'..."

# Optionally drop and recreate the database (optional safety step)
echo "Dropping and recreating the database '$POSTGRES_DB'..."
docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" "$DB_CONTAINER_NAME" \
  psql -U "$POSTGRES_USER" -c "DROP DATABASE IF EXISTS $POSTGRES_DB;"
docker exec -e PGPASSWORD="$POSTGRES_PASSWORD" "$DB_CONTAINER_NAME" \
  psql -U "$POSTGRES_USER" -c "CREATE DATABASE $POSTGRES_DB;"

# Restore using temporary postgres container
docker run --rm \
  --network "$DB_NETWORK_NAME" \
  -v "$(pwd):/backup" \
  -e PGPASSWORD="$POSTGRES_PASSWORD" \
  postgres \
  psql -h "$DB_CONTAINER_NAME" -U "$POSTGRES_USER" -d "$POSTGRES_DB" -f "/backup/$(basename "$DUMP_FILE")"

echo "✅ Restore complete."
