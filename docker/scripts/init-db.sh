#!/bin/bash
set -e

# Database and user configurations
declare -A DATABASES=(
  ["orders_db"]="orders_user:admin"
  ["users_db"]="users_user:admin"
  ["product_db"]="product_user:admin"
)

echo "Starting database initialization..."

for db in "${!DATABASES[@]}"; do
  IFS=':' read -r user password <<< "${DATABASES[$db]}"

  echo "Creating database: $db with user: $user"

  psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create user if not exists
    DO \$\$
    BEGIN
      IF NOT EXISTS (SELECT FROM pg_user WHERE usename = '$user') THEN
        CREATE USER $user WITH PASSWORD '$password';
      END IF;
    END
    \$\$;

    -- Create database if not exists
    SELECT 'CREATE DATABASE $db OWNER $user'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = '$db')\gexec

    -- Grant privileges
    GRANT ALL PRIVILEGES ON DATABASE $db TO $user;

    -- Connect to the new database and set up schema permissions
    \c $db
    GRANT ALL ON SCHEMA public TO $user;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO $user;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO $user;
EOSQL

  echo "✓ Database $db created successfully"
done

echo "Database initialization completed!"