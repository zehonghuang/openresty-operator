-- utils/random_weighted.lua
local _M = {
    servers = {},
    total_weight = 0,
}


function _M.init(input)
    _M.servers = {}
    _M.total_weight = 0

    for _, s in ipairs(input) do
        local weight = s.weight or 1
        _M.total_weight = _M.total_weight + weight
        table.insert(_M.servers, {
            host = s.host,
            port = s.port,
            weight = weight,
            ips = s.ips
        })
    end

    return _M
end

function _M.pick()
    if _M.total_weight == 0 or #_M.servers == 0 then
        return nil
    end

    local rand = math.random() * _M.total_weight
    local cumulative = 0

    for _, s in ipairs(_M.servers) do
        cumulative = cumulative + s.weight
        if rand <= cumulative then
            return s
        end
    end

    return _M.servers[#_M.servers]
end

return _M
