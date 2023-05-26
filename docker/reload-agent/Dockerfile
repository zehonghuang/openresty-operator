# syntax=docker/dockerfile:1
FROM alpine:3.19

# 安装 inotify 工具用于监听文件系统变化
RUN apk add --no-cache inotify-tools openresty

# 拷贝 reload 脚本
COPY reload-agent.sh /usr/local/bin/reload-agent
RUN chmod +x /usr/local/bin/reload-agent

# 默认执行 agent
ENTRYPOINT ["/usr/local/bin/reload-agent"]
