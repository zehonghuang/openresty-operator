#!/bin/sh

RELOAD_CMD="nginx -s reload"

WATCH_DIRS="/etc/nginx/nginx.conf"

# 🔍 递归查找 conf.d 下所有包含 .conf 文件的目录
for dir in $(find /etc/nginx/conf.d -type f -name "*.conf" -exec dirname {} \; | sort -u); do
  echo "[agent] will watch directory: $dir"
  WATCH_DIRS="$WATCH_DIRS $dir"
done

echo "[agent] watching paths: $WATCH_DIRS"
inotifywait -m -e create,modify,delete,move,close_write $WATCH_DIRS |
while read path action file; do
  echo "[agent] change detected: $path$file ($action), reloading..."
  $RELOAD_CMD
done
