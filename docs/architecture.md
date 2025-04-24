# 架构总览

OpenResty Operator 是一个模块化的 Kubernetes 原生 API 网关框架，专为内网服务和第三方 API 接入优化。其架构基于声明式 CRD 控制流，具备自动配置渲染与热更新能力。

---

## 🧱 组件结构图

```
             ┌────────────────────┐
             │     OpenResty      │
             └────────────────────┘
                    ▲      ▲
                    │      │
      ┌─────────────┘      └──────────────┐
┌──────────────┐                 ┌────────────────┐
│ ServerBlock  │ <────────────── │   Upstream     │
└──────────────┘                 └────────────────┘
       ▲
       │
┌──────────────┐
│  Location    │
└──────────────┘
```

---

## 📦 CRD 功能职责说明

### `OpenResty`
- 顶层资源，定义 OpenResty 实例。
- 配置镜像、metrics、serverRefs 与 upstreamRefs。
- 负责组合多个 ConfigMap，生成最终的 nginx.conf。

### `ServerBlock`
- 映射为 Nginx 中的 `server {}` 块。
- 配置监听端口，引用多个 `Location`。
- 每个 ServerBlock 渲染为独立的 `server.conf` 并挂载到 Pod。

### `Location`
- 定义一个或多个 Nginx `location {}` 配置项。
- 支持 proxy_pass、headers、timeout、日志、metrics 注入等。
- 渲染为 `location.conf` 通过 ConfigMap 挂载。

### `Upstream`
- 配置上游服务节点（IP 或域名:端口）。
- 支持 DNS 解析追踪，输出相关 Prometheus 指标。

---

## 🔁 配置渲染与部署流程

1. 用户声明 Location / ServerBlock / Upstream 等 CR 资源。
2. 各自的 Controller 监听变更，生成对应 ConfigMap。
3. `OpenRestyReconciler` 汇总所有引用资源，渲染 nginx.conf。
4. 部署或更新 OpenResty Pod，挂载相关配置。
5. Pod 内部 reload agent 监听配置变更，调用 `nginx -s reload` 实现热更新。

---

## 🔍 可观测性设计

- 每个 Location 可启用 `enableUpstreamMetrics`，注入 Lua 代码采集指标。
- `openresty_crd_ref_status` 指标暴露 CRD 依赖关系与就绪状态。
- 完全兼容 Prometheus + Grafana 监控体系。

---

## ⚙️ GitOps 友好特性

- 所有 CRD 均为 namespace-scoped，便于隔离部署。
- 支持 ArgoCD / Flux 等 GitOps 工具自动部署。
- 提供 Helm Chart，支持默认值与参数覆盖配置。

---

该架构使 OpenResty Operator 具备：
- 易部署
- 模块化易扩展
- 高度可维护性

适用于需要精细控制第三方 API 接入与流量治理的场景。

