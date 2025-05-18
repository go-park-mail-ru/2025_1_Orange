local counter = 0
local max_id = 100000 

wrk.method = "GET"
wrk.headers["Content-Type"] = "application/json"

request = function()
    counter = counter + 1
    local id = (counter % max_id) + 1
    return wrk.format(nil, "/api/v1/vacancy/" ..id)
end