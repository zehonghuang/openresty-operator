FROM openresty/openresty:alpine

LABEL maintainer="your-name <you@example.com>"
LABEL description="OpenResty with Prometheus metrics support"

# 安装 prometheus 模块 + curl + bash
RUN opm get knyar/nginx-lua-prometheus && \
    apk add --no-cache curl bash

# 可选：暴露 Nginx 默认端口
EXPOSE 80

# 可选：设置工作目录
WORKDIR /usr/local/openresty/nginx

# 启动 nginx（可改成 supervisord 或其他入口）
CMD ["openresty", "-g", "daemon off;"]
