#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE identity_service;
	\c identity_service
    CREATE ROLE identity_service WITH LOGIN PASSWORD '$DB_PASSWORD';
    CREATE EXTENSION IF NOT EXISTS citext;
EOSQL
