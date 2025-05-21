-- 文件：/etc/lua/normalize_utils.lua
local M = {}

function M.get(obj, path)
    for part in path:gmatch("[^%.]+") do
        if type(obj) ~= "table" then return nil end
        obj = obj[part]
    end
    return obj
end

function M.set(obj, path, value)
    local parts = {}
    for part in path:gmatch("[^%.]+") do table.insert(parts, part) end
    for i = 1, #parts - 1 do
        obj[parts[i]] = obj[parts[i]] or {}
        obj = obj[parts[i]]
    end
    obj[parts[#parts]] = value
end

function M.delete(obj, path)
    local parts = {}
    for part in path:gmatch("[^%.]+") do table.insert(parts, part) end
    for i = 1, #parts - 1 do
        if type(obj[parts[i]]) ~= "table" then return end
        obj = obj[parts[i]]
    end
    obj[parts[#parts]] = nil
end

return M