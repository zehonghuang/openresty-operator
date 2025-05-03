local cjson = require("cjson.safe")
local dict = ngx.shared.secrets_store
local secret_root = "/usr/local/openresty/lualib/secrets"

local _M = {}

local function join_path(a, b)
    return a:match("/$") and a .. b or a .. "/" .. b
end

function _M.reload()
    local cmd = "find " .. secret_root .. " -type l -name keys.json"
    ngx.log(ngx.INFO, "[secrets-loader] starting to run: ", cmd)
    local p = io.popen(cmd)
    if not p then
        ngx.log(ngx.WARN, "[secrets-loader] failed to run: ", cmd)
        return
    end

    for filepath in p:lines() do
        local f = io.open(filepath, "r")
        if f then
            local content = f:read("*a")
            f:close()

            local ok, data = pcall(cjson.decode, content)
            if ok and data then
                for dict_key, header_value in pairs(data) do
                    dict:set(dict_key, header_value)
                end
                ngx.log(ngx.INFO, "[secrets-loader] loaded: ", filepath)
            else
                ngx.log(ngx.ERR, "[secrets-loader] decode failed: ", filepath)
            end
        else
            ngx.log(ngx.WARN, "[secrets-loader] cannot open file: ", filepath)
        end
    end

    p:close()
end

return _M