-- random_weighted.lua
local _M = {}

local servers = {}
local total_weight = 0

function _M.init(input)
    servers = {}
    total_weight = 0
    for _, s in ipairs(input) do
        local weight = s.weight or 1
        total_weight = total_weight + weight
        table.insert(servers, { address = s.address, weight = weight })
    end
end

function _M.pick()
    if total_weight == 0 or #servers == 0 then
        return nil
    end

    local rand = math.random() * total_weight
    local cumulative = 0

    for _, s in ipairs(servers) do
        cumulative = cumulative + s.weight
        if rand <= cumulative then
            return s.address
        end
    end

    return servers[#servers].address
end

return _M