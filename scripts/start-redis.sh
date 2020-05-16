#!/usr/bin/env bash

if ! [[ -x "$(command -v docker)" ]]; then
    echo 'could not find docker executable, abort.'
    exit 1
fi

docker run -d --rm --name go-lock-redis -p 6379:6379 redis