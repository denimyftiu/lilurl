#!/bin/bash
migrate \
    -path ./migrations \
    -database "postgresql://$POSTGRES_DB_USER:$POSTGRES_DB_PASSWORD@$POSTGRES_DB_HOST:$POSTGRES_DB_PORT/$POSTGRES_DB_NAME?sslmode=disable" \
    up
