# OpenResty Operator

A Kubernetes Operator to declaratively manage OpenResty configurations using CRDs.

## Features

- 🎯 CRD support for OpenRestyApp, ServerBlock, Location, Upstream
- 🔄 Automatic ConfigMap rendering & NGINX reload
- 📊 Prometheus metrics (via Lua and upstream probes)
- 🔐 Multi-tenant namespace isolation
- ⚙️ Rate limit & cache policy injection (via CRD)

## Installation

### From Local Chart

```bash
helm install openresty-operator ./charts/openresty-operator
```

### (Optional) Add from Helm Repo (for hosted charts)

```bash
helm repo add openresty-operator https://huangzehong.me/openresty-operator
helm install openresty-operator openresty-operator/openresty-operator
```

## Parameters

### ⚙️ Common Parameters

| Name                     | Description                                                      | Value                               |
|--------------------------|------------------------------------------------------------------|-------------------------------------|
| `installCRDs`            | Install CRDs when installing the chart                           | `true`                              |
| `replicaCount`           | Number of operator controller replicas                           | `1`                                 |
| `image.repository`       | Operator image repository                                        | `gintonic1glass/openresty-operator` |
| `image.tag`              | Operator image tag                                               | `"v0.3.2"`                          |
| `image.pullPolicy`       | Image pull policy                                                | `IfNotPresent`                      |
| `imagePullSecrets`       | List of image pull secrets                                       | `[]`                                |
| `nameOverride`           | Partially override chart name                                    | `""`                                |
| `fullnameOverride`       | Fully override release name                                      | `""`                                |

### 👤 RBAC and ServiceAccount

| Name                        | Description                                      | Value    |
|-----------------------------|--------------------------------------------------|----------|
| `serviceAccount.create`     | Create a dedicated ServiceAccount                | `true`   |
| `serviceAccount.name`       | Name of the ServiceAccount to use                | `""`     |
| `rbac.create`               | Create required RBAC roles and bindings          | `true`   |

### 📦 Pod Deployment

| Name             | Description                           | Value |
|------------------|---------------------------------------|-------|
| `resources`       | Pod resource requests and limits      | `{}`  |
| `nodeSelector`    | Node selector rules                   | `{}`  |
| `tolerations`     | Pod tolerations                       | `[]`  |
| `affinity`        | Pod affinity                          | `{}`  |

### 🚀 OpenResty Instance

| Name                               | Description                                                     | Value                                      |
|------------------------------------|-----------------------------------------------------------------|--------------------------------------------|
| `openresty.name`                   | Name of the OpenResty instance                                  | `openresty-app`                            |
| `openresty.image`                  | OpenResty image with Prometheus support                         | `gintonic1glass/openresty:with-prometheus` |
| `openresty.replicas`               | Number of OpenResty pods                                        | `1`                                        |
| `openresty.http.accessLog`         | Access log output path                                          | `/dev/stdout`                              |
| `openresty.http.errorLog`          | Error log output path                                           | `/dev/stderr`                              |
| `openresty.http.gzip`              | Enable gzip compression                                         | `true`                                     |
| `openresty.http.serverRefs`        | Referenced ServerBlock names                                    | `[...]`                                    |
| `openresty.http.upstreamRefs`      | Referenced Upstream names                                       | `[...]`                                    |
| `openresty.metrics.enable`         | Enable Prometheus metrics endpoint                              | `true`                                     |
| `openresty.metrics.listen`         | Metrics listening port                                          | `"9090"`                                   |
| `openresty.metrics.path`           | Metrics endpoint path                                           | `"/metrics"`                               |
| `openresty.serviceMonitor.enabled` | Create a ServiceMonitor resource (requires Prometheus Operator) | `true`                                     |

### 🔗 `upstreams`

| Name                  | Description                                        | Value   |
|-----------------------|----------------------------------------------------|---------|
| `upstreams`           | List of upstream definitions                       | `[...]` |
| `upstreams[].name`    | Name of the upstream (referenced in `proxyPass`)   | `""`    |
| `upstreams[].servers` | List of backend servers in `host:port` format      | `[]`    |

### 🌐 `servers` (ServerBlock)

| Name                     | Description                                      | Value     |
|--------------------------|--------------------------------------------------|-----------|
| `servers`                | List of `ServerBlock` definitions                | `[...]`   |
| `servers[].name`         | Unique server name (used as `server_name`)       | `""`      |
| `servers[].listen`       | Listen port (string)                             | `"80"`    |
| `servers[].locationRefs` | List of location names to include                | `[""]`    |

### 📍 `locations`

| Name                                          | Description                                                           | Value   |
|-----------------------------------------------|-----------------------------------------------------------------------|---------|
| `locations`                                   | List of location configurations                                       | `[...]` |
| `locations[].name`                            | Unique name for this location resource                                | `""`    |
| `locations[].entries[].path`                  | URL path prefix to match (must start with `/`)                        | `""`    |
| `locations[].entries[].proxyPass`             | Backend URL or upstream name (e.g. `http://svc`, `https://upstream/`) | `""`    |
| `locations[].entries[].enableUpstreamMetrics` | Enable Prometheus metrics for this path                               | `true`  |
| `locations[].entries[].accessLog`             | Enable access log (default: true)                                     | `false` |
| `locations[].entries[].extra`                 | List of additional raw Nginx directives                               | `[""]`  |

### 🔁 Reload Agent

| Name                             | Description                                          | Value  |
|----------------------------------|------------------------------------------------------|--------|
| `reloadAgent.policies.window`    | Time window for evaluating reload trigger (seconds ) | `90`   |
| `reloadAgent.policies.maxEvents` | Max file changes within window to trigger reload     | `12`   |

## Example Usage

After installing the operator, you can apply custom resources:

```bash
kubectl apply -f examples/upstream.yaml
kubectl apply -f examples/location.yaml
kubectl apply -f examples/serverblock.yaml
kubectl apply -f examples/openrestyapp.yaml
```

The operator will automatically:
- Render ConfigMaps for upstreams, locations, and server blocks
- Assemble a full nginx.conf
- Deploy OpenResty pods with all referenced configuration

## CRDs Installed

This chart installs the following CRDs:

- `openrestyapps.openresty.huangzehong.me`
- `serverblocks.openresty.huangzehong.me`
- `locations.openresty.huangzehong.me`
- `upstreams.openresty.huangzehong.me`
- `ratelimitpolicies.openresty.huangzehong.me`

## License

MIT © 2025 huangzehong