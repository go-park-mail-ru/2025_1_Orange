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
wrk.headers["X-CSRF-Token"] = "FMbTwqnyf+ySlzyXasoC2N8obZ1MtcrorfRAXZHrwMs=" 
wrk.headers["Cookie"] = "session_id=69782128-d3cc-48a3-b323-34ea4ad5c9d7; csrf_token=FMbTwqnyf+ySlzyXasoC2N8obZ1MtcrorfRAXZHrwMs="
wrk.thread = function(thread)
    math.randomseed(os.time() + thread:get("id"))
end

request = function()
    counter = counter + 1
    local salaryFrom = randomInt(60000, 100000)
    local salaryTo = salaryFrom + randomInt(20000, 80000)

    local job = {
        EmployerID = randomInt(1, 100),
        Title = titles[randomInt(1, #titles)]:gsub("Developer", "QA"),  -- модификация случайного заголовка
        IsActive = true,
        WorkFormat = math.random() > 0.5 and "hybrid" or "remote",  -- 50/50
        Employment = "full_time",
        Schedule = "5/2",
        WorkingHours = randomInt(30, 50),
        SalaryFrom = randomInt(80000, 120000),
        SalaryTo = randomInt(150000, 250000),
        TaxesIncluded = math.random() > 0.5,  -- случайно true/false
        Experience = experiences[randomInt(1, #experiences)],
        Specialization = "Информационные технологии",
        City = "Москва",
        Description = "Аудит-логирование — сервис...",  -- можно сделать массив и выбирать случайно
        Tasks = "Написание компонентных, интеграционных...",
        Requirements = "Опыт самостоятельного написания...",
        OptionalRequirements = "Опыт работы с gRPC...",
        Skills = {"REST API", "Kafka", "Python", "MySQL", "SQL", "gRPC"}  -- можно сделать массив массивов и выбирать случайно
    }

    local body = json.encode(job)
    return wrk.format("POST", "/api/v1/vacancy/vacancies", nil, body)
end
