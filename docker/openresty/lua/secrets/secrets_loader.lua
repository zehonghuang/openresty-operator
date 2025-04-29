local cjson = require("cjson.safe")
local dict = ngx.shared.secrets_store

local secret_root = "/usr/local/openresty/lualib/secrets"

local function scan_secrets(dir, prefix)
    local p = io.popen("ls -1 " .. dir)
    if not p then
        ngx.log(ngx.WARN, "[secrets-loader] cannot open directory: ", dir)
        return
    end

    for entry in p:lines() do
        local full_path = dir .. "/" .. entry
        -- 简单判断是目录
        local f = io.open(full_path, "r")
        if not f then
            -- 是目录
            scan_secrets(full_path, prefix .. "/" .. entry)
        else
            -- 是文件
            if entry == "keys.json" then
                local content = f:read("*a")
                f:close()

                local data = cjson.decode(content)
                if data then
                    for dict_key, header_value in pairs(data) do
                        dict:set(dict_key, header_value)
                    end
                    ngx.log(ngx.INFO, "[secrets-loader] loaded keys from ", full_path)
                else
                    ngx.log(ngx.ERR, "[secrets-loader] failed to decode json: ", full_path)
                end
            else
                f:close()
            end
        end
    end
    p:close()
end

local _M = {}

function _M.reload()
    scan_secrets(secret_root, "")
end

return _M