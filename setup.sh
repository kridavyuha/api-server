#!/bin/bash

# Stop and remove existing PostgreSQL container

make stop_db

# Start PostgreSQL database using Makefile

make start_db

# Wait for PostgreSQL to start
sleep 5

# Create a new database named 'db'
docker exec -it db psql -U postgres -c "CREATE DATABASE db;"

echo "Database 'db' created successfully."

# Stop and remove existing redis container

make stop_redis

# Start redis server using Makefile

make start_redis

echo "Redis server started successfully."
