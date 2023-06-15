# OpenResty Operator

[![License](https://img.shields.io/github/license/zehonghuang/openresty-operator)](LICENSE)

**OpenResty Operator** 是一个 Kubernetes 原生的 Operator，用于声明式管理 OpenResty 实例和配置。通过 CRD 资源将复杂的 Nginx/OpenResty 配置模块化，支持热重载、Prometheus metrics 暴露与可观测性增强，适用于多项目多服务部署场景。

---

## ✨ Features

- 🧩 **模块化配置管理**：支持 Http、ServerBlock、Location、Upstream 多 CRD 管理
- 🔁 **配置热更新**：支持通过 Sidecar 监听 ConfigMap 实现自动热重载
- 📊 **内置 Metrics**：内建 Prometheus Lua 支持，可选 metrics server 暴露
- 🔎 **状态可观测性**：展示 CRD 引用健康状态，结合 Grafana 轻松可视化
- 🧵 **轻量部署**：支持通过 Helm 安装，资源开销小、无侵入性

---

## 📦 CRD 概览

| Kind            | 描述                    |
|-----------------|-----------------------|
| `OpenResty`     | 声明一个完整 OpenResty 应用   |
| `ServerBlock`   | 配置 server 区块          |
| `Location`      | 配置 location 区块        |
| `Upstream`      | 配置 upstream 服务器组      |

---

## 🚀 快速开始

### 1. 安装 CRD 与 Operator

```bash
helm repo add openresty-operator https://zehonghuang.github.io/openresty-operator/charts
helm install openresty-operator openresty-operator/openresty-operator
```

> 若你正在本地开发，可使用：
> ```bash
> helm install openresty-operator ./charts/openresty-operator
> ```

### 2. 创建 OpenResty 应用

```yaml
apiVersion: openresty.huangzehong.me/v1alpha1
kind: OpenResty
metadata:
  name: openresty-sample
spec:
  replicas: 1
  image: gintonic1glass/openresty:with-prometheus
  metrics:
    enable: true
    listen: "9090"
  http:
    accessLog: /var/log/nginx/access.log
    errorLog: /var/log/nginx/error.log
    gzip: true
    serverRefs:
      - serverblock-sample
    upstreamRefs:
      - upstream-sample
```

---

## 📈 可观测性

支持以下 Prometheus 指标：

- `openresty_crd_ref_status`
- `openresty_upstream_dns_resolvable`
- `openresty_upstream_server_alive_total`

推荐结合 Grafana Dashboard 使用，见 `/deploy/grafana/`。

---

## 🧪 开发 & 贡献

```bash
make install
make run
```
---

## 📄 License

[MIT](./LICENSE)