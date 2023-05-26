#!/bin/sh

# 默认监听路径，可以用 ENV 覆盖
WATCH_PATHS="${WATCH_PATHS:-/etc/nginx/nginx.conf /etc/nginx/conf.d}"
RELOAD_CMD="${RELOAD_COMMAND:-nginx -s reload}"

echo "[agent] watching: $WATCH_PATHS"
echo "[agent] reload command: $RELOAD_CMD"

inotifywait -m -e close_write,create,modify $WATCH_PATHS | while read path action file; do
  echo "[agent] change detected at $path$file ($action), reloading..."
  $RELOAD_CMD
done
docker build -t /reload-agent:latest docker/reload-agent