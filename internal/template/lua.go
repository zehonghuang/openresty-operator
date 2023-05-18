package template

const DefaultInitLua = `
    prometheus = require("prometheus").init("prometheus_metrics")

    metric_upstream_latency = prometheus:histogram(
        "upstream_latency_seconds",
        "Upstream response time in seconds",
        {"upstream"}
    )

    metric_upstream_total = prometheus:counter(
        "upstream_requests_total",
        "Total upstream requests",
        {"upstream", "status"}
    )
`
