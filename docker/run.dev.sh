#!/bin/bash

DEV_APP_SRC=$(echo $(cd -))
DEV_BUCKETS_SRC=$(echo $DEV_APP_SRC/buckets)

APP_SRC=${DEV_APP_SRC} \
BUCKETS_SRC=${DEV_BUCKETS_SRC} \
docker-compose -f docker/docker-compose.dev.yaml --env-file ./.env up --build
