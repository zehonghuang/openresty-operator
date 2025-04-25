-- round_robin.lua
local rr_state = {}
local _M = {}

function _M.init(servers)
    rr_state.index = 1
    rr_state.pool = {}

    for _, s in ipairs(servers) do
        local weight = s.weight or 1
        for _ = 1, weight do
            table.insert(rr_state.pool, s.address)
        end
    end
end

function _M.pick()
    if #rr_state.pool == 0 then
        return nil
    end
    local addr = rr_state.pool[rr_state.index]
    rr_state.index = rr_state.index % #rr_state.pool + 1
    return addr
end

return _M
