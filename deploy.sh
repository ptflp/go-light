#!/bin/sh
git stash
git pull
sleep 3
docker restart golightserver_"$1"