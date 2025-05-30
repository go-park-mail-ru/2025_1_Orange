local function randomInt(min, max)
  return math.random(min, max)
end

local function randomString(length)
  local chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
  local result = ""
  for i = 1, length do
    local index = math.random(1, #chars)
    result = result .. chars:sub(index, index)
  end
  return result
end

local function pickRandom(list)
  return list[randomInt(1, #list)]
end

-- Параметры
local titles = {"Backend Developer", "Frontend Developer", "DevOps Engineer", "QA Tester", "Project Manager"}
local specializations = {"Software Engineering", "DevOps", "Quality Assurance", "Management"}
local cities = {"Moscow", "Saint Petersburg", "Kazan", "Novosibirsk"}
local employments = {"full_time", "part_time", "contract", "internship", "freelance", "watch"}
local schedules = {"5/2", "2/2", "6/1", "3/3", "on_weekend", "by_agreement"}
local workFormats = {"office", "hybrid", "remote", "traveling"}
local experiences = {"no_experience", "1_3_years", "3_6_years", "6_plus_years"}
local allSkills = {"Go", "Docker", "PostgreSQL", "Kubernetes", "Redis", "REST", "CI/CD"}

local title = pickRandom(titles)
local specialization = pickRandom(specializations)
local city = pickRandom(cities)
local employment = pickRandom(employments)
local schedule = pickRandom(schedules)
local workingHours = randomInt(1, 96)
local workFormat = pickRandom(workFormats)
local salaryFrom = randomInt(15000, 500000)
local salaryTo = randomInt(salaryFrom, 1000000)
local taxesIncluded = "true"
local experience = pickRandom(experiences)
local description = randomString(50)
local tasks = randomString(50)
local requirements = randomString(50)
local optionalRequirements = randomString(50)

-- Случайные 2-3 уникальных скилла
local selectedSkills = {}
local used = {}
for i = 1, randomInt(2, 3) do
  local skill
  repeat
    skill = pickRandom(allSkills)
  until not used[skill]
  used[skill] = true
  table.insert(selectedSkills, '"' .. skill .. '"')
end
local skillsJson = "[" .. table.concat(selectedSkills, ",") .. "]"

-- Установка метода и заголовков
wrk.method = "POST"
wrk.path = "/api/v1/vacancy/vacancies"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["X-CSRF-Token"] = "kdmykfS8zMIsIZiwBtX26KZCrZ3X7iqafQiq5fzon5A="
wrk.headers["Cookie"] = "csrf_token=kdmykfS8zMIsIZiwBtX26KZCrZ3X7iqafQiq5fzon5A=; session_id=0ea13ee4-df3b-48e6-80d7-81e61fecd980"
wrk.thread = function(thread)
    math.randomseed(os.time() + thread:get("id"))
end

wrk.body = [[
{
  "title": "]] .. title .. [[",
  "specialization": "]] .. specialization .. [[",
  "city": "]] .. city .. [[",
  "employment": "]] .. employment .. [[",
  "schedule": "]] .. schedule .. [[",
  "working_hours": ]] .. workingHours .. [[,
  "work_format": "]] .. workFormat .. [[",
  "salary_from": ]] .. salaryFrom .. [[,
  "salary_to": ]] .. salaryTo .. [[,
  "taxes_included": ]] .. tostring(taxesIncluded) .. [[,
  "experience": "]] .. experience .. [[",
  "description": "]] .. description .. [[",
  "tasks": "]] .. tasks .. [[",
  "requirements": "]] .. requirements .. [[",
  "optional_requirements": "]] .. optionalRequirements .. [[",
  "skills": ]] .. skillsJson .. [[
}
]]

print("status,content_length,body_snippet")

response = function(status, headers, body)
  local now = os.time()
  local latency_ms = (wrk.latency and wrk.latency:percentile(50.0)) or "N/A"

  print(string.format("[LOG] Status: %d | Time: %s | Latency: %s ms", status, os.date("%Y-%m-%d %H:%M:%S", now), latency_ms))

  if status ~= 200 then
    print("[ERROR BODY]: " .. tostring(body))
  end
end