#!/bin/bash

# Load environment variables from .env.production
source .env.production

# Required env vars (from .env.production)
#   POSTGRES_USER=...
#   POSTGRES_PASSWORD=...
#   POSTGRES_DB=...
#   DB_CONTAINER_NAME=your_pg_container
#   DB_NETWORK_NAME=your_docker_network

# Fail fast on errors
set -euo pipefail
CONTAINER="brainwars_pgsql"

# Dynamically get the network name of the DB container
DB_NETWORK_NAME=$(docker inspect "$CONTAINER" \
  | jq -r '.[0].NetworkSettings.Networks | keys[0]')

# Timestamp for backup file
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
OUTPUT_FILE="./dbbackup/backup_POSTGRES_DB_$TIMESTAMP.sql"

echo "Starting backup of database '$PG_DB' from container 'brainwars_pgsql'..."

docker run -rm --network "$DB_NETWORK_NAME" \
  -e PGPASSWORD="$PG_PASSWORD" \
  postgres \
  pg_dump -h "brainwars_pgsql" -U "$PG_USER" "$PG_DB" > "$OUTPUT_FILE"

echo "Backup complete: $OUTPUT_FILE"
