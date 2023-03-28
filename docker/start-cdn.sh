#!/bin/bash

logpath=/data/logs/$(date '+%Y-%m-%d_%H-%M-%S').log
cfgpath=/app/config.yaml
bucketspath=/data/buckets

mkdir -p /data/logs $bucketspath

touch $logpath

cdn -debug=false -buckets-path=$bucketspath -logs-path=$logpath -config-path=$cfgpath -strict-log=true
