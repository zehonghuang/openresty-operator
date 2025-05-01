local _M = {}

local prometheus
local metric_latency
local metric_total
local metric_errors

function _M.init()
    prometheus = require("prometheus").init("prometheus_metrics")

    metric_latency = prometheus:histogram(
        "upstream_latency_seconds",
        "Upstream response time in seconds",
        {"upstream"}
    )

    metric_total = prometheus:counter(
        "upstream_requests_total",
        "Total upstream requests",
        {"upstream", "status"}
    )

    metric_errors = prometheus:counter(
        "upstream_errors_total",
        "Total upstream errors by type",
        {"upstream", "error_type"}
    )
end

function _M.record()
    local addr = (ngx.var.upstream_addr or "unknown"):match("^[^,]+") or "none"
    local status = ngx.status
    local latency = tonumber(ngx.var.upstream_response_time) or 0
    local upstream_status = ngx.var.upstream_status or ""

    metric_latency:observe(latency, {addr})
    metric_total:inc(1, {addr, tostring(status)})

    if status >= 500 then
        metric_errors:inc(1, {addr, "http_" .. status})
    elseif upstream_status:find("timeout") then
        metric_errors:inc(1, {addr, "timeout"})
    end
end

return _M