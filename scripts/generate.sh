#!/usr/bin/env bash
mkdir -p internal/generated
protoc -I ./ lock.proto --go_out=.