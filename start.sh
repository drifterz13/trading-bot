#!/bin/bash
echo "mode $1"
docker run --env-file=".env" -v data:/data mixz-robo "$1"