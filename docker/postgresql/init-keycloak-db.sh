#!/bin/bash
set -e

# This script initializes the Keycloak database and user
# It runs automatically when PostgreSQL starts for the first time

echo "Initializing Keycloak database..."

# Create Keycloak database and user
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    -- Create Keycloak user
    DO \$\$
    BEGIN
        IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'keycloak') THEN
            CREATE USER keycloak WITH PASSWORD 'keycloak123';
        END IF;
    END
    \$\$;

    -- Create Keycloak database
    SELECT 'CREATE DATABASE keycloak OWNER keycloak'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'keycloak')\gexec

    -- Grant privileges to Keycloak user
    GRANT ALL PRIVILEGES ON DATABASE keycloak TO keycloak;
    
    -- Connect to Keycloak database and set up permissions
    \c keycloak;
    GRANT ALL ON SCHEMA public TO keycloak;
    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO keycloak;
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO keycloak;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON TABLES TO keycloak;
    ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL ON SEQUENCES TO keycloak;
EOSQL

echo "Keycloak database initialization completed successfully."