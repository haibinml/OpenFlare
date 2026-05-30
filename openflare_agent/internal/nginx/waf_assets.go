package nginx

import "openflare-agent/internal/protocol"

const openRestyWAFCheckLua = `local cjson = require "cjson.safe"

local config_dict = ngx.shared.openflare_waf_config

local function read_file(path)
    local f = io.open(path, "r")
    if not f then
        return nil
    end
    local content = f:read("*a")
    f:close()
    return content
end

local function load_config()
    local paths = {
        "__OPENFLARE_RUNTIME_CONFIG_DIR__/waf_config.json",
        "/etc/nginx/openflare-lua/waf_config.json",
        "/usr/local/openresty/nginx/conf/waf_config.json"
    }
    for _, path in ipairs(paths) do
        local content = read_file(path)
        if content and content ~= "" then
            local hash = ngx.md5(content)
            if config_dict:get("_config_hash") == hash then
                local cached = config_dict:get("_config_json")
                if cached then
                    local decoded = cjson.decode(cached)
                    if decoded then
                        return decoded
                    end
                end
            end
            local decoded = cjson.decode(content)
            if decoded then
                config_dict:set("_config_hash", hash, 0)
                config_dict:set("_config_json", content, 0)
                return decoded
            end
        end
    end
    return nil
end

local function list_contains(items, value)
    if not items or not value or value == "" then
        return false
    end
    for _, item in ipairs(items) do
        if item == value then
            return true
        end
    end
    return false
end

local function parse_ipv4(value)
    local a, b, c, d = string.match(value or "", "^(%d+)%.(%d+)%.(%d+)%.(%d+)$")
    if not a then
        return nil
    end
    a, b, c, d = tonumber(a), tonumber(b), tonumber(c), tonumber(d)
    if a > 255 or b > 255 or c > 255 or d > 255 then
        return nil
    end
    return ((a * 256 + b) * 256 + c) * 256 + d
end

local function ipv4_in_cidr(ip, cidr)
    local base, bits = string.match(cidr or "", "^([^/]+)/(%d+)$")
    if not base then
        return false
    end
    bits = tonumber(bits)
    if not bits or bits < 0 or bits > 32 then
        return false
    end
    local ip_num = parse_ipv4(ip)
    local base_num = parse_ipv4(base)
    if not ip_num or not base_num then
        return false
    end
    if bits == 0 then
        return true
    end
    local mask = 4294967295 - (2 ^ (32 - bits) - 1)
    return (ip_num - (ip_num % (2 ^ (32 - bits)))) == (base_num - (base_num % (2 ^ (32 - bits))))
end

local function ip_matches(items, ip)
    if not items or not ip or ip == "" then
        return false
    end
    for _, item in ipairs(items) do
        if item == ip then
            return true
        end
        if string.find(item, "/", 1, true) and ipv4_in_cidr(ip, item) then
            return true
        end
    end
    return false
end

local function lookup_country(ip)
    local ok, maxminddb = pcall(require, "resty.maxminddb")
    if not ok or not maxminddb then
        return nil
    end
    local paths = {
        "__OPENFLARE_RUNTIME_CONFIG_DIR__/GeoLite2-Country.mmdb",
        "/etc/openflare/GeoLite2-Country.mmdb",
        "/usr/local/share/openflare/GeoLite2-Country.mmdb"
    }
    for _, path in ipairs(paths) do
        local opened = pcall(maxminddb.init, path)
        if opened then
            local res, err = maxminddb.lookup(ip)
            if res and res.country and res.country.iso_code then
                return string.upper(res.country.iso_code)
            end
        end
    end
    return nil
end

local function group_by_id(config)
    local result = {}
    for _, group in ipairs(config.rule_groups or {}) do
        result[tostring(group.id)] = group
    end
    return result
end

local function active_groups(config, groups)
    local site = ngx.var.openflare_waf_site or ""
    local ids = (config.site_rule_groups or {})[site]
    local result = {}
    for _, group in ipairs(config.rule_groups or {}) do
        if group.is_global then
            result[#result + 1] = group
        end
    end
    if ids then
        local by_id = group_by_id(config)
        for _, id in ipairs(ids) do
            local group = by_id[tostring(id)]
            if group and not group.is_global then
                result[#result + 1] = group
            end
        end
    end
    return result
end

local function exit_with_group(group)
    ngx.status = tonumber(group.block_status_code) or 418
    local body = group.block_response_body or ""
    if body ~= "" then
        ngx.header["Content-Type"] = "text/html; charset=utf-8"
        ngx.say(body)
    end
    return ngx.exit(ngx.status)
end

local config = load_config()
if not config then
    return
end

local ip = ngx.var.remote_addr or ""
local groups = active_groups(config)

for _, group in ipairs(groups) do
    if ip_matches(group.ip_whitelist, ip) then
        return
    end
end

local country = nil
for _, group in ipairs(groups) do
    if group.country_whitelist and #group.country_whitelist > 0 then
        country = country or lookup_country(ip)
        if list_contains(group.country_whitelist, country) then
            return
        end
    end
end

for _, group in ipairs(groups) do
    if ip_matches(group.ip_blacklist, ip) then
        return exit_with_group(group)
    end
end

for _, group in ipairs(groups) do
    if group.country_blacklist and #group.country_blacklist > 0 then
        country = country or lookup_country(ip)
        if list_contains(group.country_blacklist, country) then
            return exit_with_group(group)
        end
    end
end
`

func ManagedWAFLuaFiles() []protocol.SupportFile {
	return []protocol.SupportFile{
		{Path: "waf/check.lua", Content: openRestyWAFCheckLua},
	}
}
