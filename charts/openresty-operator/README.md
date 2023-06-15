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
helm repo add openresty-operator https://zehonghuang.github.io/openresty-operator/charts
helm install openresty-operator openresty-operator/openresty-operator
```

## Values

| Key                                | Default                                    | Description                        |
|------------------------------------|--------------------------------------------|------------------------------------|
| `replicaCount`                     | `1`                                        | Number of Operator pods            |
| `image.repository`                 | `gintonic1glass/openresty`                 | Operator image repository          |
| `image.tag`                        | `with-prometheus`                          | Image tag                          |
| `image.pullPolicy`                 | `IfNotPresent`                             | Image pull policy                  |
| `serviceAccount.create`            | `true`                                     | Whether to create a ServiceAccount |
| `serviceAccount.name`              | `""`                                       | Name override for ServiceAccount   |
| `rbac.create`                      | `true`                                     | Whether to create RBAC resources   |
| `resources`                        | `{}`                                       | Pod resource requests/limits       |
| `nodeSelector`                     | `{}`                                       | Node selector                      |
| `tolerations`                      | `[]`                                       | Tolerations                        |
| `affinity`                         | `{}`                                       | Affinity rules                     |
| `openresty.name`                   | "openresty-sample"                         | Name of OpenResty custom resource  |
| `openresty.replicas`               | 1                                          | Number of OpenResty replicas       |
| `openresty.image`                  | "gintonic1glass/openresty:with-prometheus" | OpenResty image to deploy          |
| `openresty.http.include`           | []                                         | Additional include directives      |
| `openresty.http.logFormat`         | ""                                         | Log format string                  |
| `openresty.http.accessLog`         | "/dev/stdout"                              | Access log path                    |
| `openresty.http.errorLog`          | "/dev/stderr"                              | Error log path                     |
| `openresty.http.clientMaxBodySize` | "16m"                                      | Maximum client body size           |
| `openresty.http.gzip`              | true                                       | Enable gzip                        |
| `openresty.http.extra`             | []                                         | Additional raw nginx directives    |
| `openresty.http.serverRefs`        | []                                         | Referenced ServerBlock names       |
| `openresty.http.upstreamRefs`      | []                                         | Referenced Upstream names          |
| `openresty.metrics.enable`         | true                                       | Enable Prometheus metrics server   |
| `openresty.metrics.listen`         | "9090"                                     | Metrics server listen port         |
| `openresty.metrics.path`           | "/metrics"                                 | Path to expose metrics             |

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