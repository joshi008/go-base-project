#!/bin/bash

# Database setup script for go-base-project
# Usage: ./scripts/setup_db.sh

set -e

DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-go_base_db}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"

echo "Setting up database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"

# Create database if it doesn't exist
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -p $DB_PORT -tc "SELECT 1 FROM pg_database WHERE datname = '$DB_NAME'" | grep -q 1 || \
  PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -p $DB_PORT -c "CREATE DATABASE $DB_NAME"

echo "Database created or already exists"

# Run migrations
echo "Running migrations..."
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -p $DB_PORT -d $DB_NAME -f migrations/init.sql

echo "✅ Database setup complete!"
echo ""
echo "To insert test data, run:"
echo "PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -p $DB_PORT -d $DB_NAME << EOF"
echo "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com');"
echo "INSERT INTO users (name, email) VALUES ('Jane Smith', 'jane@example.com');"
echo "EOF"
