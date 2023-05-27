#!/bin/sh

# 🔍 动态收集所有 conf 路径
build_watch_dirs() {
  WATCH_DIRS=""
  for dir in $(find /etc/nginx/conf.d -type l -name "*.conf" -exec dirname {} \; | sort -u); do
    echo "[agent] will watch directory: $dir"
    WATCH_DIRS="$WATCH_DIRS $dir"
  done
}

# 🔁 重载 nginx master
reload_nginx() {
  pid=$(ps | grep 'nginx: master' | grep -v grep | awk '{print $1}')
  if [ -z "$pid" ]; then
    echo "[agent] ❌ nginx master PID not found"
  else
    echo "[agent] ✅ reloading nginx (pid=$pid)"
    kill -HUP "$pid"
  fi
}

# 🧠 主循环：即使 inotifywait 异常退出，也能自动重启监听
while true; do
  build_watch_dirs
  echo "[agent] watching paths: $WATCH_DIRS"

  inotifywait -m -e create,modify,delete,move,close_write $WATCH_DIRS 2>/dev/null |
  while read path action file; do
    echo "[agent] change detected: $path$file ($action), reloading..."
    reload_nginx
  done

  echo "[agent] ⚠️ inotifywait exited unexpectedly. Restarting in 3s..."
  sleep 3
done
