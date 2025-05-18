local wrk = require "wrk"
local json = require "cjson"

local counter = 0

local titles = { "Backend Developer", "Frontend Developer", "DevOps Engineer", "QA Specialist", "Data Analyst" }
local workFormats = { "Remote", "Onsite", "Hybrid" }
local employments = { "Full-time", "Part-time", "Contract" }
local schedules = { "5/2", "Flexible", "Shift" }
local experiences = { "No experience", "1–3 years", "3–6 years", "6+ years" }

local function randomInt(min, max)
    return math.random(min, max)
end

-- Установка метода и заголовков
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Accept"] = "application/json"
wrk.headers["X-CSRF-Token"] = "O6EbReb0sR8eoOc5/YF36JC/uWwe61f+ndPXKWhtPbA=" 
wrk.headers["Cookie"] = "csrf_token=O6EbReb0sR8eoOc5/YF36JC/uWwe61f+ndPXKWhtPbA=; Path=/; Expires=Sun, 18 May 2025 07:25:32 GMT; HttpOnly; SameSite=Strict"
wrk.thread = function(thread)
    math.randomseed(os.time() + thread:get("id"))
end

request = function()
    counter = counter + 1
    local salaryFrom = randomInt(60000, 100000)
    local salaryTo = salaryFrom + randomInt(20000, 80000)

    local job = {
        EmployerID = randomInt(1, 100),
        Title = titles[randomInt(1, #titles)],
        IsActive = true,
        WorkFormat = workFormats[randomInt(1, #workFormats)],
        Employment = employments[randomInt(1, #employments)],
        Schedule = schedules[randomInt(1, #schedules)],
        WorkingHours = randomInt(20, 60),
        SalaryFrom = salaryFrom,
        SalaryTo = salaryTo,
        TaxesIncluded = true,
        Experience = experiences[randomInt(1, #experiences)],
        Description = "Some description here...",
        City = "Moskow",
        Tasks = "You will write code.",
        Requirements = "Know Go, please.",
        OptionalRequirements = "Docker is a plus."
    }

    local body = json.encode(job)
    return wrk.format("POST", "/api/v1/vacancy/vacancies", nil, body)
end
