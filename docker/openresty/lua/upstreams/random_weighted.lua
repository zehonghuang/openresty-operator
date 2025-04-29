-- random_weighted.lua
local _M = {}

local servers = {}
local total_weight = 0

function _M.init(input)
    servers = input
    total_weight = 0
    for _, s in ipairs(servers) do
        total_weight = total_weight + (s.weight or 1)
    end
end

function _M.pick()
    local rand = math.random() * total_weight
    local cumulative = 0

    for _, s in ipairs(servers) do
        cumulative = cumulative + (s.weight or 1)
        if rand <= cumulative then
            return s.address
        end
    end
end

return _M
