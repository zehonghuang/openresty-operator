FROM --platform=linux/amd64 openresty/openresty:alpine

LABEL maintainer="zehong.huang <zehong.hongframe.huang@gmail.com>"
LABEL description="OpenResty with Prometheus metrics support"

RUN apk add --no-cache curl bash perl && \
    opm get knyar/nginx-lua-prometheus

# 可选：暴露 Nginx 默认端口
EXPOSE 80

COPY lua/upstreams/ /usr/local/openresty/lualib/upstreams/
COPY lua/secrets/ /usr/local/openresty/lualib/secrets/
COPY lua/metrics/ /usr/local/openresty/lualib/

# 可选：设置工作目录
WORKDIR /usr/local/openresty/nginx

# 启动 nginx（可改成 supervisord 或其他入口）
CMD ["openresty", "-g", "daemon off;"]
