# 🔄 OpenResty Reload Agent

`reload-agent` 是一个用于 OpenResty 容器内的 sidecar 工具，支持实时监听配置文件变更并触发 `nginx -s reload`。

## 📦 镜像特性

- ✅ 基于 Alpine 构建，体积小巧
- 🔍 使用 `inotifywait` 实时监听
- 🚀 自动执行 `nginx -s reload` 无需重启容器
- 🔄 支持监听多个路径，适配 Nginx 的 include 模式

## ⚙️ 默认行为

| 环境变量 | 默认值 | 说明 |
|----------|--------|------|
| `WATCH_PATHS` | `/etc/nginx/nginx.conf /etc/nginx/conf.d` | 监听的路径，可为空格分隔多个 |
| `RELOAD_COMMAND` | `nginx -s reload` | 执行的 reload 命令 |

## 🧪 使用方式（在 Kubernetes 中）

确保 Pod 启用 `shareProcessNamespace: true`，然后注入此 sidecar：

```yaml
spec:
  shareProcessNamespace: true
  containers:
    - name: openresty
      image: gintonic1glass/openresty:with-prometheus
    - name: reload-agent
      image: gintonic1glass/reload-agent:latest
      env:
        - name: WATCH_PATHS
          value: "/etc/nginx/nginx.conf /etc/nginx/conf.d"
