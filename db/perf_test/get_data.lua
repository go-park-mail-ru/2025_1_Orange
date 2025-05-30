local counter = 0
local max_id = 100000 

wrk.method = "GET"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Cookie"] = "csrf_token=kdmykfS8zMIsIZiwBtX26KZCrZ3X7iqafQiq5fzon5A=; session_id=0ea13ee4-df3b-48e6-80d7-81e61fecd980"

request = function()
    counter = counter + 1
    local id = (counter % max_id) + 1
    return wrk.format(nil, "/api/v1/vacancy/vacancy/" ..id)
end