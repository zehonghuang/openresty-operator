#!/bin/bash

# 用法: ./single-commit.sh "your commit message" [--same-day]

# 参数校验
if [ $# -lt 1 ]; then
  echo "❌ 用法: ./single-commit.sh \"commit message\" [--same-day]"
  exit 1
fi

message="$1"
same_day=false

# 第二个参数可选：是否 same-day 模式
if [ "$2" == "--same-day" ]; then
  same_day=true
fi

# 自动适配 macOS 或 Linux 的 date/gdate
if command -v gdate &> /dev/null; then
  DATE="gdate"
else
  DATE="date"
fi

# 初始时间
initial="2023-05-12T14:56:01"
initial_ts=$($DATE -d "$initial" +%s)

# 获取上一次提交时间（秒）
last_ts=$(git log -1 --pretty=format:"%at" 2>/dev/null)

# 没有上一次就用初始
if [ -z "$last_ts" ]; then
  base_ts=$initial_ts
else
  if [ "$same_day" = true ]; then
    # 同一天加 30~45 分钟
    rand_min=$((70 + RANDOM % 16))  # 30-45 分钟
    base_ts=$((last_ts + rand_min * 60))
  else
    # 随机 +1 or +2 天
    rand_day=$((RANDOM % 2 + 1))
    base_ts=$((last_ts + rand_day * 86400))

    # 再加一个 ±2 小时 ±30 分钟的随机偏移
    rand_hour=$(( (RANDOM % 5 - 2) * 3600 ))
    rand_minute=$(( (RANDOM % 61 - 30) * 60 ))
    offset=$((rand_hour + rand_minute))
    base_ts=$((base_ts + offset))
  fi
fi

# 转换为 ISO 时间格式
final_time=$($DATE -d "@$base_ts" +"%Y-%m-%dT%H:%M:%S")

export GIT_AUTHOR_DATE="$final_time"
export GIT_COMMITTER_DATE="$final_time"

echo "✅ Commit at: $final_time"
git add .
git commit -m "$message"
