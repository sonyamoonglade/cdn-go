#!/bin/bash

APP_SRC=$(cat .env | grep "APP_SRC" | cut -d "=" -f2)

APP_SRC=$APP_SRC docker-compose -f ./docker/docker-compose.dev.yaml down
