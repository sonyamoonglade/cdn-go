#!/bin/bash

logpath=/data/logs/$(date '+%Y-%m-%d_%H-%M-%S').log
cfgpath=/app/config.yaml

mkdir -p /data/logs /data/buckets
touch $logpath
animakuro-cdn -debug=false -buckets-path=/data/buckets -logs-path=$logpath -config-path=$cfgpath
