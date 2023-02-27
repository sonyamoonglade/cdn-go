#!/bin/bash

# run app
docker-compose -f docker/docker-compose.dev.yaml --env-file ./.env up --build -d

APP_SRC=$(cat .env | grep "APP_SRC" | cut -d "=" -f2)
MONGO_DB_NAME=$(cat .env | grep "MONGO_DB_NAME" | cut -d "=" -f2)
MONGO_USER=$(cat .env | grep "MONGO_USER" | cut -d "=" -f2)
MONGO_PWD=$(cat .env | grep "MONGO_PWD" | cut -d "=" -f2)
MONGO_URI="mongodb://$MONGO_USER:$MONGO_PWD@localhost:27017/$MONGO_DB_NAME?authSource=admin"

# run migrations
docker run -v $APP_SRC/migrations:/migrations --network host --rm migrate/migrate -path=/migrations/ -database $MONGO_URI up

