#!/bin/bash
set -e

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE DATABASE workflow_service;
	\c workflow_service
    CREATE ROLE workflow_service WITH LOGIN PASSWORD '$DB_PASSWORD';
    CREATE EXTENSION IF NOT EXISTS citext;
EOSQL
