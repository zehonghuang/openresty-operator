#!/bin/bash

# commit message 参数检查
if [ $# -ne 1 ]; then
  echo "❌ 你需要传入 1 条 commit message"
  echo "👉 用法： ./single-commit.sh \"feat: init project\""
  exit 1
fi

# 第一次固定初始时间（格式 ISO 8601）
initial="2023-05-12T14:56:01"
initial_ts=$(gdate -d "$initial" +%s)

# 上一次提交的时间获取
last_ts=$(git log -1 --pretty=format:"%at" 2>/dev/null)

# 如果没有 commit，使用初始时间
if [ -z "$last_ts" ]; then
  base_ts=$initial_ts
else
  # 随机 +1 或 +2 天
  rand_day=$((RANDOM % 2 + 1))
  base_ts=$((last_ts + rand_day * 86400))
fi

# 随机偏移：±2 小时 & ±30 分钟
rand_hour=$(( (RANDOM % 5 - 2) * 3600 ))
rand_minute=$(( (RANDOM % 61 - 30) * 60 ))
offset=$((rand_hour + rand_minute))
final_ts=$((base_ts + offset))

# 转换为时间字符串
final_time=$(gdate -d "@$final_ts" +"%Y-%m-%dT%H:%M:%S")

# 设置 Git 时间环境变量
export GIT_AUTHOR_DATE="$final_time"
export GIT_COMMITTER_DATE="$final_time"

echo "✅ Commit at: $final_time"
git add .
git commit -m "$1"
