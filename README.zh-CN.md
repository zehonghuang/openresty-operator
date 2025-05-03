[English](./README.md) | 中文
<p align="center">
  <img src="./docs/images/logo-tight.png" alt="OpenResty Operator Logo">
</p>
<p align="center">
  <b>一个轻量级的 Kubernetes Operator，用于将 OpenResty 用作内部 API 网关的配置与管理。</b>
</p>

# OpenResty Operator

![GitHub release (latest by tag)](https://img.shields.io/github/v/tag/zehonghuang/openresty-operator?label=release)
![Release](https://github.com/zehonghuang/openresty-operator/actions/workflows/release.yaml/badge.svg)
![License](https://img.shields.io/badge/license-MIT-blue)
[![Artifact Hub](https://img.shields.io/endpoint?url=https://artifacthub.io/badge/repository/openresty-operator)](https://artifacthub.io/packages/search?repo=openresty-operator)

## TL;DR

🚀 **OpenResty Operator** 是一个轻量级的 Kubernetes Operator，用于将 OpenResty（Nginx）作为内部 API 网关进行管理。 

- ✅ 特别适用于**代理大量第三方 API 的场景**，具备极简开销与高度可观测性。

- 🛠️ **通过 CRD 实现声明式管理**：将 Location、ServerBlock 和 Upstream 配置为原生 Kubernetes 资源。

- 🔁 **配置热更新，无需重启容器**：内置 reload agent 可即时应用配置更改。

- 📊 **原生支持 Prometheus 监控**：内建上游健康检查、DNS 解析状态与配置引用状态等指标。

- 🎯 **无需 etcd、无需 Admin API、零额外负担**——只需 OpenResty 和这个 Operator。

## 适用人群

✅ 你需要在内部系统中代理多个第三方 API

✅ 你希望通过 GitOps 和 CRD 实现配置管理，而非依赖图形界面

✅ 你倾向于使用透明、轻量的 Nginx/OpenResty 网关方案

✅ 你认为 APISIX 或 Kong 在你的场景中过于重型

## 为什么选择 OpenResty + Operator？

我希望构建一个更加 **Infrastructure-Friendly** 的方案：

- **基于 OpenResty 原始配置层级**：配置即 Nginx，完全对标原生语法，具备更强的可控性；
- **模块化资源抽象**：以 CRD 表达 `Location`、`ServerBlock`、`Upstream` 等核心组件，具有清晰的引用逻辑与版本控制；
- **Kubernetes 原生生态兼容**：天然支持 GitOps 管理，方便与 ArgoCD / Flux 等工具集成；
- **零依赖、可落地**：无需额外组件，只需 OpenResty 镜像和 Operator 即可运行。

## 功能特性

- **灵活的配置建模**  
  使用 `Location`、`ServerBlock`、`Upstream` 等 CRD 描述 Nginx 配置结构，支持任意组合，适用于数量多、分布广、维护难度高的第三方 API 管理场景。

- **自动渲染与部署**  
  基于 `OpenResty` 自动拼接多个配置块，生成统一的 `nginx.conf` 并部署为 OpenResty 实例。

- **配置变更热更新**  
  内置`reload agent`，无需重启容器，即可动态应用配置变更。

- **引用校验与版本控制**  
  配置引用支持 version、ready 校验机制，确保资源引用始终一致、可追踪。

- **原生监控集成**  
  内置 Prometheus metrics 采集能力，可视化展示 upstream 状态与资源引用等状况。

## 快速开始

### 1. 安装 Operator

推荐方式：使用 Helm 安装。

```bash
helm repo add openresty-operator https://huangzehong.me/openresty-operator
helm install openresty openresty-operator/openresty-operator
```

如果你从源码部署，也可以直接应用原始 YAML：

```bash
kubectl apply -f config/crd/bases/
kubectl apply -k config/smaples/
```

### 2. 定义配置资源

示例：一个简单的 Location / ServerBlock / Upstream 配置。

```yaml
apiVersion: openresty.huangzehong.me/v1alpha1
kind: Location
metadata:
  name: location-sample
spec:
  entries:
    - path: /sample-api/
      proxyPass: http://upstream-sample/
      enableUpstreamMetrics: true
      headers:
        - key: Host
          value: $host
        - key: X-Real-IP
          value: $remote_addr
        - key: X-Forwarded-For
          value: $proxy_add_x_forwarded_for
        - key: X-Forwarded-Proto
          value: $scheme
        - key: X-Content-Type-Options
          value: nosniff
        - key: Access-Control-Allow-Origin
          value: "*"
      accessLog: false
---
apiVersion: openresty.huangzehong.me/v1alpha1
kind: ServerBlock
metadata:
  name: serverblock-sample
spec:
  listen: "80"
  locationRefs:
    - location-sample
---
apiVersion: openresty.huangzehong.me/v1alpha1
kind: Upstream
metadata:
  name: upstream-sample
spec:
  servers:
    - example.com:80
    - www.baidu.com:443
    - invalid.domain.local:8080
```

### 3. 创建 OpenResty 实例

```yaml
apiVersion: openresty.huangzehong.me/v1alpha1
kind: OpenResty
metadata:
  name: openresty-sample
spec:
  image: gintonic1glass/openresty:with-prometheus
  http:
    include:
      - mime.types
    logFormat: |
      $remote_addr - $remote_user [$time_local] "$request" ...
    clientMaxBodySize: 16m
    gzip: true
    extra:
      - sendfile on;
      - tcp_nopush on;
    serverRefs:
      - serverblock-sample
    upstreamRefs:
      - upstream-sample
  metrics:
    enable: true
    listen: "8080"
    path: "/metrics"
```

## 📈 指标与监控

OpenResty Operator 默认导出多种 Prometheus 指标，便于观测配置状态与流量健康状况，适配常见的云原生监控栈（Prometheus + Grafana）：

- `openresty_crd_ref_status`：追踪各类 CRD（如 ServerBlock、Location、Upstream）之间的引用关系和就绪状态。
- `openresty_upstream_dns_ready`：展示 upstream DNS 解析成功率与可达性。
- `openresty_request_total` 与 `openresty_response_status`：分析各个 upstream 的请求量与状态码分布。
- 支持通过 Lua 扩展自定义业务级 metrics。

## 📊 Grafana Dashboard 示例

OpenResty Operator 导出的指标可以通过 Grafana 进行可视化。以下是一个仪表盘示例，展示了：

- CRD 数量、Ready 状态、引用结构
- Upstream 的 DNS 健康度和响应情况；
- 近期配置异常与告警事件（如路径冲突、域名无法解析等）

![OpenResty Operator Grafana Dashboard](./docs/images/grafana-dashboard-02.png)
![OpenResty Operator Grafana Dashboard](./docs/images/grafana-dashboard-03.png)

> 📊 官方 Grafana Dashboard 已上线，用于配合 Prometheus 监控 OpenResty Operator。  
> 查看或导入地址：[Dashboard #23321](https://grafana.com/grafana/dashboards/23321)

## 许可证

MIT License. 详见 [LICENSE](LICENSE)。