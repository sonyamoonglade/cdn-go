#!/bin/bash

# DEV_APP_SRC=$(echo $(cd -))
# DEV_BUCKETS_SRC=$(echo $DEV_APP_SRC/buckets)
docker-compose -f docker/docker-compose.dev.yaml --env-file ./.env up --build
