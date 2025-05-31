local balancer = require("ngx.balancer")
local random_weighted = require("utils.random_weighted")

local _M = {}

function _M.randomWeightedBalance(servers)
    local server = random_weighted.init(servers).pick()
    if not server or not server.host or not server.port then
        ngx.log(ngx.ERR, "no valid upstream server found")
        return ngx.exit(502)
    end

    ngx.ctx.server_host = server.host

    local ip = server.host
    if server.ips and #server.ips > 0 then
        ip = server.ips[math.random(#server.ips)]
    end

    local ok, err = balancer.set_current_peer(ip, server.port, server.host)
    if not ok then
        ngx.log(ngx.ERR, "failed to set current peer: ", err)
        return ngx.exit(502)
    end

end

return _M